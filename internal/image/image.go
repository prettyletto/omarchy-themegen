package image

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Valid    bool
	Path     string
	Width    int
	Height   int
	Format   string
	Errors   []string
	Warnings []string
}

func Validate(path string) *Result {
	r := &Result{Path: path}

	// Check file exists and is readable
	fi, err := os.Stat(path)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("cannot read %q: %v", path, err))
		return r
	}
	if fi.IsDir() {
		r.Errors = append(r.Errors, fmt.Sprintf("%q is a directory, not an image", path))
		return r
	}

	// Check magick is available for deeper inspection
	magickPath, err := exec.LookPath("magick")
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("ImageMagick 'magick' is not installed or not on PATH; required for image validation"))
		return r
	}

	// Use magick identify to get image details
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, magickPath, "identify", "-verbose", path)
	output, err := cmd.Output()
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("failed to identify %q: %v (is it a valid image?)", path, err))
		return r
	}

	rawFormat := extractIdentifyField(string(output), "Format")
	if rawFormat == "" {
		r.Errors = append(r.Errors, fmt.Sprintf("%q does not appear to be a valid image", path))
		return r
	}
	// Extract base format (e.g., "PNG" from "PNG (Portable Network Graphics)")
	r.Format = strings.SplitN(rawFormat, " ", 2)[0]

	// Check for animation (multiple frames/scenes)
	scenes := extractIdentifyField(string(output), "Scene")
	if scenes != "" {
		s, _ := strconv.Atoi(scenes)
		if s > 0 {
			r.Errors = append(r.Errors, fmt.Sprintf("%q appears to be animated (scene %d)", path, s))
			return r
		}
	}

	// Check dimensions
	geom := extractIdentifyField(string(output), "Geometry")
	if geom != "" {
		base := strings.Split(geom, "+")[0]
		dimParts := strings.Split(base, "x")
		if len(dimParts) == 2 {
			// trim possible letter suffixes
			wStr := strings.TrimRightFunc(dimParts[0], func(r rune) bool { return r < '0' || r > '9' })
			hStr := strings.TrimRightFunc(dimParts[1], func(r rune) bool { return r < '0' || r > '9' })
			r.Width, _ = strconv.Atoi(wStr)
			r.Height, _ = strconv.Atoi(hStr)
		}
	}

	if r.Width < 800 || r.Height < 450 {
		r.Errors = append(r.Errors, fmt.Sprintf(
			"image dimensions %dx%d are below minimum 800x450", r.Width, r.Height))
		return r
	}

	// Check for transparency (alpha channel)
	alpha := extractIdentifyField(string(output), "Alpha")
	channels := extractIdentifyField(string(output), "Channel(s)")
	if alpha != "" || strings.Contains(strings.ToLower(channels), "alpha") {
		transparent, err := hasTransparentPixels(magickPath, path)
		if err != nil {
			r.Errors = append(r.Errors, fmt.Sprintf("transparency check failed for %q: %v", path, err))
			return r
		}
		if transparent {
			r.Errors = append(r.Errors, fmt.Sprintf("%q has transparent pixels; opaque images only", path))
			return r
		}
	}

	// UI-heavy detection (warn only)
	if isUIHeavy(r) {
		r.Warnings = append(r.Warnings, "image appears UI-heavy; screenshot/UI capture may produce poor themes")
	}

	r.Valid = true
	return r
}

func extractIdentifyField(output, field string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, field+":") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, field+":"))
			return val
		}
	}
	return ""
}

func hasTransparentPixels(magickPath, path string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, magickPath, path, "-alpha", "extract", "-negate", "-format", "%[fx:mean]", "info:")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("cannot check transparency: %w", err)
	}
	mean, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return false, fmt.Errorf("cannot parse transparency result %q: %w", string(output), err)
	}
	// negated alpha: opaque pixels are 0, transparent pixels are 1
	return mean > 0.001, nil
}

var uiCaptureFormats = map[string]bool{
	"PNG": true,
}

func isUIHeavy(r *Result) bool {
	if r.Height < 720 && r.Width < 1280 {
		return false
	}

	aspectRatio := float64(r.Width) / float64(r.Height)

	if uiCaptureFormats[strings.ToUpper(r.Format)] {
		if (aspectRatio > 1.7 && aspectRatio < 1.8) ||
			(aspectRatio > 1.58 && aspectRatio < 1.62) {
			return true
		}
	}

	return false
}
