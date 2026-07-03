package gen

import (
	"fmt"
	"math"
	"sort"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

type PaletteCandidate struct {
	ID       int
	Label    string
	Colors   *theme.Colors
	Warnings []string
}

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
	colorCount := max(1, len(bySaturation))
	offset1 := seedAbs % colorCount
	offset2 := (seedAbs + max(1, colorCount/3)) % colorCount
	offset3 := (seedAbs + max(2, (colorCount*2)/3)) % colorCount
	offset4 := (seedAbs + max(3, colorCount/2)) % colorCount
	offset5 := (seedAbs + max(4, (colorCount*3)/4)) % colorCount

	mute1 := 0.0
	mute2 := 0.12 + float64(seedAbs%5)*0.02
	mute3 := 0.22 + float64(seedAbs%7)*0.02

	candidates := make([]PaletteCandidate, 5)

	candidates[0] = buildDirection(1, "Vibrant", byLightness, bySaturation, offset1, mute1, directionProfile{
		bgLightness:     0.09,
		fgLightness:     0.88,
		accentLightness: 0.62,
		accentSatFloor:  0.72,
	})

	candidates[1] = buildDirection(2, "Balanced", byLightness, bySaturation, offset2, mute2, directionProfile{
		bgLightness:     0.14,
		fgLightness:     0.80,
		accentLightness: 0.54,
		accentSatFloor:  0.52,
	})

	candidates[2] = buildDirection(3, "Muted", byLightness, bySaturation, offset3, mute3, directionProfile{
		bgLightness:     0.20,
		fgLightness:     0.72,
		accentLightness: 0.46,
		accentSatFloor:  0.28,
		accentSatCeil:   0.45,
	})

	candidates[3] = buildDirection(4, "Colorful", byLightness, bySaturation, offset4, 0.04, directionProfile{
		bgLightness:     0.11,
		fgLightness:     0.86,
		accentLightness: 0.58,
		accentSatFloor:  0.86,
	})

	candidates[4] = buildDirection(5, "Deep", byLightness, bySaturation, offset5, 0.08, directionProfile{
		bgLightness:     0.06,
		fgLightness:     0.82,
		accentLightness: 0.50,
		accentSatFloor:  0.64,
	})

	if opts.LightMode {
		candidates = invertToLight(candidates)
	}

	// Filter candidates with critical contrast failures
	var valid []PaletteCandidate
	for _, c := range candidates {
		if HasCriticalContrastFailures(c.Colors) {
			c.Warnings = append(c.Warnings, "critical contrast failure — direction blocked")
			continue
		}
		valid = append(valid, c)
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("all palette candidates have critical contrast failures; cannot generate usable directions from this image")
	}

	return valid, nil
}

type directionProfile struct {
	bgLightness     float64
	fgLightness     float64
	accentLightness float64
	accentSatFloor  float64
	accentSatCeil   float64
}

func buildDirection(id int, label string, byLightness, bySaturation []DominantColor, accentIdx int, muteAmount float64, profile directionProfile) PaletteCandidate {
	c := PaletteCandidate{
		ID:    id,
		Label: label,
	}

	// Pick background: darkest color, but ensure it's dark enough (< 0.25 lightness)
	bg := pickBackground(byLightness, profile.bgLightness)

	// Pick accent: from saturation-sorted list
	accent := pickAccent(bySaturation, accentIdx, bg, profile)

	// Pick foreground: lightest color with enough contrast against background
	fg := pickForeground(byLightness, bg, muteAmount, profile.fgLightness)

	// Generate full palette
	c.Colors = buildTerminalPalette(bg, fg, accent, bySaturation, muteAmount)
	ensureReadableTextColors(c.Colors)

	// Validate
	c.Warnings = validatePaletteContrast(c.Colors)

	return c
}

func pickBackground(byLightness []DominantColor, targetLightness float64) RGB {
	// Find darkest colors, ensure dark enough
	for _, c := range byLightness {
		if c.Lightness < 0.35 {
			rgb := c.Color
			return rgb.WithLightness(targetLightness)
		}
	}
	// Fallback: use a dark neutral
	return RGB{0.08, 0.10, 0.14}.WithLightness(targetLightness)
}

func pickAccent(bySaturation []DominantColor, idx int, bg RGB, profile directionProfile) RGB {
	if idx < len(bySaturation) {
		accent := bySaturation[idx].Color
		hsl := accent.ToHSL()
		hsl.S = math.Max(hsl.S, profile.accentSatFloor)
		if profile.accentSatCeil > 0 {
			hsl.S = math.Min(hsl.S, profile.accentSatCeil)
		}
		hsl.L = profile.accentLightness
		accent = hsl.ToRGB()
		// Ensure enough contrast with background
		if ContrastRatio(accent, bg) < 3.0 {
			hsl = accent.ToHSL()
			hsl.L = math.Max(hsl.L, 0.58)
			accent = hsl.ToRGB()
		}
		return accent
	}
	return RGB{0.51, 0.67, 1.0}
}

func pickForeground(byLightness []DominantColor, bg RGB, muteAmount float64, targetLightness float64) RGB {
	// Find lightest colors with contrast
	for i := len(byLightness) - 1; i >= 0; i-- {
		c := byLightness[i].Color
		if ContrastRatio(c, bg) >= MinForegroundContrast {
			rgb := c
			if muteAmount > 0 {
				hsl := rgb.ToHSL()
				hsl.S = clamp(hsl.S-muteAmount*0.3, 0, 1)
				hsl.L = targetLightness
				rgb = hsl.ToRGB()
			} else {
				rgb = rgb.WithLightness(targetLightness)
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
		slot  int
		hue   float64
		label string
	}{
		{1, 0.0, "red"},
		{2, 0.33, "green"},
		{3, 0.17, "yellow"},
		{4, 0.6, "blue"},
		{5, 0.83, "magenta"},
		{6, 0.5, "cyan"},
	}

	accentFamily := hueDiversity(srcColors) < 0.18
	used := map[int]bool{}
	for _, t := range hueTargets {
		idx, best, dist := findBestHueColor(srcColors, t.hue, used)
		if best != nil && (!accentFamily || dist <= 0.14) {
			rgb := best.Color
			hsl := rgb.ToHSL()
			hsl.L = clamp(hsl.L+0.05, 0.25, 0.55)
			hsl.S = clamp(hsl.S-muteAmount, 0.3, 1.0)
			colors[t.slot] = hsl.ToRGB().Hex()
			used[idx] = true
		} else if accentFamily {
			colors[t.slot] = accentFamilyTermColor(accent, t.slot, muteAmount).Hex()
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

func ensureReadableTextColors(c *theme.Colors) {
	bg, err := ParseHex(c.Background)
	if err != nil {
		return
	}
	if fg, err := ParseHex(c.Foreground); err == nil {
		c.Foreground = ensureReadableAgainst(bg, fg, MinForegroundContrast).Hex()
	}
	if color7, err := ParseHex(c.Color7); err == nil {
		c.Color7 = ensureReadableAgainst(bg, color7, MinForegroundContrast).Hex()
	}
	if color15, err := ParseHex(c.Color15); err == nil {
		c.Color15 = ensureReadableAgainst(bg, color15, MinForegroundContrast).Hex()
	}
	if selBg, err := ParseHex(c.SelectionBackground); err == nil {
		if selFg, err := ParseHex(c.SelectionForeground); err == nil {
			c.SelectionForeground = ensureReadableAgainst(selBg, selFg, MinSelectionContrast).Hex()
		}
	}
}

func ensureReadableAgainst(bg, fg RGB, minContrast float64) RGB {
	if ContrastRatio(fg, bg) >= minContrast {
		return fg
	}

	hsl := fg.ToHSL()
	for _, l := range textLightnessLadder(bg, hsl.L) {
		candidate := HSL{H: hsl.H, S: hsl.S, L: l}.ToRGB()
		if ContrastRatio(candidate, bg) >= minContrast {
			return candidate
		}
	}

	black := RGB{0, 0, 0}
	white := RGB{1, 1, 1}
	if ContrastRatio(black, bg) > ContrastRatio(white, bg) {
		return black
	}
	return white
}

func textLightnessLadder(bg RGB, current float64) []float64 {
	if bg.Luminance() < 0.5 {
		return []float64{math.Max(current, 0.78), 0.84, 0.90, 0.96, 1.0}
	}
	return []float64{math.Min(current, 0.24), 0.18, 0.12, 0.06, 0.0}
}

func findBestHueColor(src []DominantColor, targetHue float64, used map[int]bool) (int, *DominantColor, float64) {
	bestIdx := -1
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
			bestIdx = i
		}
	}
	return bestIdx, best, bestDist
}

func hueDiversity(src []DominantColor) float64 {
	var hues []float64
	for _, c := range src {
		hsl := c.Color.ToHSL()
		if hsl.S >= 0.12 {
			hues = append(hues, hsl.H)
		}
	}
	if len(hues) < 2 {
		return 0
	}
	maxGap := 0.0
	sort.Float64s(hues)
	for i := 1; i < len(hues); i++ {
		if gap := hues[i] - hues[i-1]; gap > maxGap {
			maxGap = gap
		}
	}
	if wrapGap := hues[0] + 1 - hues[len(hues)-1]; wrapGap > maxGap {
		maxGap = wrapGap
	}
	return 1 - maxGap
}

func accentFamilyTermColor(accent RGB, slot int, muteAmount float64) RGB {
	hsl := accent.ToHSL()
	type transform struct {
		hueShift float64
		sat      float64
		light    float64
	}
	transforms := map[int]transform{
		1: {0.00, 0.90, 0.42},
		2: {0.06, 0.55, 0.36},
		3: {0.10, 0.68, 0.44},
		4: {-0.08, 0.50, 0.48},
		5: {-0.04, 0.70, 0.50},
		6: {0.04, 0.60, 0.52},
	}
	t := transforms[slot]
	h := hsl.H + t.hueShift
	for h < 0 {
		h += 1
	}
	for h >= 1 {
		h -= 1
	}
	return HSL{
		H: h,
		S: clamp(math.Max(hsl.S, t.sat)-muteAmount*0.35, 0.35, 1),
		L: t.light,
	}.ToRGB()
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
		bg, errBg := ParseHex(c.Background)
		fg, errFg := ParseHex(c.Foreground)

		if errBg != nil || errFg != nil {
			candidates[i].Warnings = append(candidates[i].Warnings,
				"light inversion failed: invalid foreground/background colors")
			continue
		}

		// Light theme: background is light, foreground is dark
		newBg := fg.WithLightness(clamp(fg.ToHSL().L+0.08, 0.85, 0.97))
		newFg := bg.WithLightness(clamp(bg.ToHSL().L+0.05, 0.08, 0.22))

		c.Background = newBg.Hex()
		c.Foreground = newFg.Hex()

		// Adjust accent for light bg
		accent, errAcc := ParseHex(c.Accent)
		if errAcc == nil {
			accentHsl := accent.ToHSL()
			accentHsl.L = clamp(accentHsl.L+0.1, 0.25, 0.6)
			c.Accent = accentHsl.ToRGB().Hex()
			// Selection bg must contrast with light background
			selHsl := accent.ToHSL()
			selHsl.L = clamp(selHsl.L-0.15, 0.2, 0.4)
			c.SelectionBackground = selHsl.ToRGB().Hex()
		}

		c.SelectionForeground = newBg.Hex()

		// Make terminal colors darker for light bg
		for _, idx := range []*string{
			&c.Color1, &c.Color2, &c.Color3, &c.Color4, &c.Color5, &c.Color6,
			&c.Color7, &c.Color8, &c.Color9, &c.Color10, &c.Color11, &c.Color12,
			&c.Color13, &c.Color14, &c.Color15,
		} {
			rgb, err := ParseHex(*idx)
			if err != nil {
				continue
			}
			hsl := rgb.ToHSL()
			hsl.L = clamp(hsl.L-0.25, 0.15, 0.7)
			*idx = hsl.ToRGB().Hex()
		}

		c.Color0 = newBg.Hex()
		ensureReadableTextColors(c)
		candidates[i].Warnings = validatePaletteContrast(c)

		candidates[i].Colors = c
		candidates[i].Label = candidates[i].Label + " Light"
	}
	return candidates
}
