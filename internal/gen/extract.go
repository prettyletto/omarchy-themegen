package gen

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type DominantColor struct {
	Color      RGB
	Hex        string
	Frequency  float64
	Saturation float64
	Lightness  float64
}

func ExtractDominantColors(path string, count int) ([]DominantColor, error) {
	magick, err := exec.LookPath("magick")
	if err != nil {
		return nil, fmt.Errorf("ImageMagick 'magick' is not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, magick,
		path,
		"-resize", "480x270^",
		"-gravity", "center",
		"-extent", "480x270",
		"-colors", fmt.Sprintf("%d", count),
		"-unique-colors",
		"-depth", "8",
		"txt:",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("magick color extraction failed: %w", err)
	}

	colors := parseMagickTxt(string(output))
	if len(colors) == 0 {
		return nil, fmt.Errorf("no dominant colors extracted from image")
	}

	// Sort by saturation descending for accent selection
	sort.Slice(colors, func(i, j int) bool {
		return colors[i].Saturation > colors[j].Saturation
	})

	if len(colors) > count {
		colors = colors[:count]
	}

	return colors, nil
}

func parseMagickTxt(output string) []DominantColor {
	var colors []DominantColor
	totalPixels := 0.0

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: "0,0: (26,27,38)  #1A1B26  srgb(26,27,38)" or with pixel count
		// We just need the hex color
		parts := strings.Fields(line)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(p, "#") && len(p) == 7 {
				rgb, err := ParseHex(p)
				if err != nil {
					continue
				}
				hsl := rgb.ToHSL()
				colors = append(colors, DominantColor{
					Color:      rgb,
					Hex:        p,
					Frequency:  1.0,
					Saturation: hsl.S,
					Lightness:  hsl.L,
				})
				totalPixels++
				break
			}
		}
	}

	// Normalize frequencies
	for i := range colors {
		colors[i].Frequency = colors[i].Frequency / math.Max(totalPixels, 1)
	}

	return colors
}
