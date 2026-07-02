package theme

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prettyletto/omarchy-themegen/internal/image"
)

var (
	errNameRequired = errors.New("theme name is required")
	errNameEmpty    = errors.New("theme name normalizes to empty string")
	errBadColors    = errors.New("color validation failed")
)

type ThemeModel struct {
	Name           string
	NormalizedName string
	Version        string
	SourceImage    string
	ImageResult    *image.Result
	Colors         *Colors
	Mode           string // "whole-theme" or "component-mix"
	DirectionID    int
	DirectionLabel string
	LightMode      bool

	GroupSelections     map[string]int
	Overrides           map[string]int
	CompositionWarnings []string
}

func NewStatic(name, sourceImage string, imgResult *image.Result) (*ThemeModel, error) {
	if name == "" {
		return nil, errNameRequired
	}

	normalized := normalizeThemeName(name)
	if normalized == "" {
		return nil, errNameEmpty
	}

	colors := StaticColors()
	if errs := ValidateColors(colors); len(errs) > 0 {
		return nil, fmt.Errorf("static colors validation failed: %s", strings.Join(errs, "; "))
	}

	return &ThemeModel{
		Name:           name,
		NormalizedName: normalized,
		Version:        "0.1.0-dev",
		SourceImage:    sourceImage,
		ImageResult:    imgResult,
		Colors:         colors,
		Mode:           "whole-theme",
		DirectionID:    0,
		DirectionLabel: "Static",
	}, nil
}

func NormalizeForExport(name string) string {
	return normalizeThemeName(name)
}

func normalizeThemeName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))

	var b strings.Builder
	lastHyphen := false
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastHyphen = false
		} else if r == '_' {
			b.WriteRune('_')
			lastHyphen = false
		} else if r == '-' {
			if !lastHyphen && b.Len() > 0 {
				b.WriteRune('-')
			}
			lastHyphen = true
		} else {
			if !lastHyphen && b.Len() > 0 {
				b.WriteRune('-')
			}
			lastHyphen = true
		}
	}

	result := strings.Trim(b.String(), "-_")
	return result
}
