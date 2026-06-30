package gen

import (
	"math"
	"sort"

	"github.com/anomalyco/omarchy-themegen/internal/theme"
)

type PaletteCandidate struct {
	ID          int
	Label       string
	Colors      *theme.Colors
	Warnings    []string
}

const (
	minContrastForeground = 4.5
	minContrastAccent     = 3.0
	minContrastSelection  = 3.0
)

func GeneratePalettes(dominantColors []DominantColor, opts *GenerationOptions) ([]PaletteCandidate, error) {
	colors := make([]DominantColor, len(dominantColors))
	copy(colors, dominantColors)

	// Sort by lightness for dark theme background selection
	byLightness := make([]DominantColor, len(colors))
	copy(byLightness, colors)
	sort.Slice(byLightness, func(i, j int) bool {
		return byLightness[i].Lightness < byLightness[j].Lightness
	})

	// Sort by saturation for accent selection
	bySaturation := make([]DominantColor, len(colors))
	copy(bySaturation, colors)
	sort.Slice(bySaturation, func(i, j int) bool {
		return bySaturation[i].Saturation > bySaturation[j].Saturation
	})

	// Use seed to shift accent selection indices
	seedAbs := opts.Seed
	if seedAbs < 0 {
		seedAbs = -seedAbs
	}
	offset1 := seedAbs % max(1, len(bySaturation))
	offset2 := (seedAbs * 2) % max(1, len(bySaturation))
	offset3 := (seedAbs * 3) % max(1, len(bySaturation))

	mute1 := 0.0
	mute2 := 0.12 + float64(seedAbs%5)*0.02
	mute3 := 0.22 + float64(seedAbs%7)*0.02

	candidates := make([]PaletteCandidate, 3)

	candidates[0] = buildDirection(1, "Vibrant", byLightness, bySaturation, offset1, mute1)

	candidates[1] = buildDirection(2, "Balanced", byLightness, bySaturation, offset2, mute2)

	candidates[2] = buildDirection(3, "Muted", byLightness, bySaturation, offset3, mute3)

	if opts.LightMode {
		candidates = invertToLight(candidates)
	}

	return candidates, nil
}

func buildDirection(id int, label string, byLightness, bySaturation []DominantColor, accentIdx int, muteAmount float64) PaletteCandidate {
	c := PaletteCandidate{
		ID:    id,
		Label: label,
	}

	// Pick background: darkest color, but ensure it's dark enough (< 0.25 lightness)
	bg := pickBackground(byLightness)

	// Pick accent: from saturation-sorted list
	accent := pickAccent(bySaturation, accentIdx, bg)

	// Pick foreground: lightest color with enough contrast against background
	fg := pickForeground(byLightness, bg, muteAmount)

	// Generate full palette
	c.Colors = buildTerminalPalette(bg, fg, accent, bySaturation, muteAmount)

	// Validate
	c.Warnings = validatePaletteContrast(c.Colors)

	return c
}

func pickBackground(byLightness []DominantColor) RGB {
	// Find darkest colors, ensure dark enough
	for _, c := range byLightness {
		if c.Lightness < 0.35 {
			rgb := c.Color
			// Darken slightly if needed
			if c.Lightness > 0.25 {
				rgb = rgb.WithLightness(0.18)
			}
			return rgb
		}
	}
	// Fallback: use a dark neutral
	return RGB{0.08, 0.10, 0.14}
}

func pickAccent(bySaturation []DominantColor, idx int, bg RGB) RGB {
	if idx < len(bySaturation) {
		accent := bySaturation[idx].Color
		// Boost saturation if needed
		hsl := accent.ToHSL()
		if hsl.S < 0.4 {
			hsl.S = 0.55
			accent = hsl.ToRGB()
		}
		// Ensure enough contrast with background
		if ContrastRatio(accent, bg) < 3.0 {
			hsl = accent.ToHSL()
			hsl.L = math.Max(hsl.L, 0.45)
			accent = hsl.ToRGB()
		}
		return accent
	}
	return RGB{0.51, 0.67, 1.0}
}

func pickForeground(byLightness []DominantColor, bg RGB, muteAmount float64) RGB {
	// Find lightest colors with contrast
	for i := len(byLightness) - 1; i >= 0; i-- {
		c := byLightness[i].Color
		if ContrastRatio(c, bg) >= minContrastForeground {
			rgb := c
			if muteAmount > 0 {
				hsl := rgb.ToHSL()
				hsl.S = clamp(hsl.S-muteAmount*0.3, 0, 1)
				hsl.L = clamp(hsl.L+muteAmount*0.1, 0, 1)
				rgb = hsl.ToRGB()
			}
			return rgb
		}
	}
	// Fallback: light neutral
	return RGB{0.75, 0.78, 0.82}
}

func buildTerminalPalette(bg, fg, accent RGB, srcColors []DominantColor, muteAmount float64) *theme.Colors {
	c := &theme.Colors{
		Background:          bg.Hex(),
		Foreground:          fg.Hex(),
		Accent:              accent.Hex(),
		Cursor:              accent.WithLightness(clamp(accent.ToHSL().L+0.08, 0, 1)).Hex(),
		SelectionForeground: bg.Hex(),
		SelectionBackground: accent.Hex(),
	}

	// Generate terminal colors from source colors
	termColors := generateTerminalColors(bg, fg, accent, srcColors, muteAmount)

	c.Color0 = termColors[0]
	c.Color1 = termColors[1]
	c.Color2 = termColors[2]
	c.Color3 = termColors[3]
	c.Color4 = termColors[4]
	c.Color5 = termColors[5]
	c.Color6 = termColors[6]
	c.Color7 = termColors[7]
	c.Color8 = termColors[8]
	c.Color9 = termColors[9]
	c.Color10 = termColors[10]
	c.Color11 = termColors[11]
	c.Color12 = termColors[12]
	c.Color13 = termColors[13]
	c.Color14 = termColors[14]
	c.Color15 = termColors[15]

	return c
}

func generateTerminalColors(bg, fg, accent RGB, srcColors []DominantColor, muteAmount float64) [16]string {
	var colors [16]string

	// color0 = background
	colors[0] = bg.Hex()
	// color7 = foreground (normal white)
	colors[7] = fg.Hex()

	// Use source colors for red/green/yellow/blue/magenta/cyan
	// Strategy: find colors in source that are close to these hues
	hueTargets := []struct {
		slot    int
		hue     float64
		label   string
	}{
		{1, 0.0, "red"},
		{2, 0.33, "green"},
		{3, 0.17, "yellow"},
		{4, 0.6, "blue"},
		{5, 0.83, "magenta"},
		{6, 0.5, "cyan"},
	}

	used := map[int]bool{}
	for _, t := range hueTargets {
		best := findBestHueColor(srcColors, t.hue, used)
		if best != nil {
			rgb := best.Color
			// Adjust to ensure visibility on dark bg
			hsl := rgb.ToHSL()
			hsl.L = clamp(hsl.L+0.05, 0.25, 0.55)
			hsl.S = clamp(hsl.S-muteAmount, 0.3, 1.0)
			colors[t.slot] = hsl.ToRGB().Hex()
			used[bestIdx(srcColors, best)] = true
		} else {
			colors[t.slot] = defaultTermColor(t.slot).Hex()
		}
	}

	// Bright colors (color8-color15) = lighter versions of color0-color7
	for i := 8; i < 16; i++ {
		base, err := ParseHex(colors[i-8])
		if err != nil {
			base = RGB{0.5, 0.5, 0.5}
		}
		hsl := base.ToHSL()
		if i-8 == 0 {
			// bright black = dark gray
			hsl.L = clamp(hsl.L+0.15, 0.15, 0.35)
		} else {
			hsl.L = clamp(hsl.L+0.2, 0.3, 0.8)
		}
		colors[i] = hsl.ToRGB().Hex()
	}

	return colors
}

func findBestHueColor(src []DominantColor, targetHue float64, used map[int]bool) *DominantColor {
	var best *DominantColor
	bestDist := math.MaxFloat64

	for i, c := range src {
		if used[i] {
			continue
		}
		hsl := c.Color.ToHSL()
		if hsl.S < 0.1 {
			continue
		}
		dist := math.Abs(hsl.H - targetHue)
		if dist > 0.5 {
			dist = 1.0 - dist
		}
		if dist < bestDist {
			bestDist = dist
			best = &src[i]
		}
	}
	return best
}

func bestIdx(src []DominantColor, target *DominantColor) int {
	for i := range src {
		if &src[i] == target {
			return i
		}
	}
	return -1
}

func defaultTermColor(slot int) RGB {
	defaults := map[int]RGB{
		1: {0.86, 0.29, 0.29},
		2: {0.62, 0.81, 0.42},
		3: {0.88, 0.69, 0.41},
		4: {0.48, 0.64, 0.97},
		5: {0.73, 0.60, 0.97},
		6: {0.49, 0.81, 1.0},
	}
	return defaults[slot]
}

func invertToLight(candidates []PaletteCandidate) []PaletteCandidate {
	for i := range candidates {
		c := candidates[i].Colors
		// Swap foreground/background with adjustments
		bg, _ := ParseHex(c.Background)
		fg, _ := ParseHex(c.Foreground)

		// Light theme: background is light, foreground is dark
		newBg := fg.WithLightness(clamp(fg.ToHSL().L+0.08, 0.85, 0.97))
		newFg := bg.WithLightness(clamp(bg.ToHSL().L+0.05, 0.08, 0.22))

		c.Background = newBg.Hex()
		c.Foreground = newFg.Hex()

		// Adjust accent for light bg
		accent, _ := ParseHex(c.Accent)
		accentHsl := accent.ToHSL()
		accentHsl.L = clamp(accentHsl.L+0.1, 0.25, 0.6)
		c.Accent = accentHsl.ToRGB().Hex()

		// Selection on light bg
		c.SelectionForeground = newBg.Hex()
		c.SelectionBackground = accentHsl.ToRGB().Hex()

		// Make terminal colors darker for light bg
		for _, idx := range []*string{
			&c.Color1, &c.Color2, &c.Color3, &c.Color4, &c.Color5, &c.Color6,
			&c.Color7, &c.Color8, &c.Color9, &c.Color10, &c.Color11, &c.Color12,
			&c.Color13, &c.Color14, &c.Color15,
		} {
			rgb, _ := ParseHex(*idx)
			hsl := rgb.ToHSL()
			hsl.L = clamp(hsl.L-0.25, 0.15, 0.7)
			*idx = hsl.ToRGB().Hex()
		}

		c.Color0 = newBg.Hex()

		candidates[i].Colors = c
		candidates[i].Label = candidates[i].Label + " Light"
	}
	return candidates
}
