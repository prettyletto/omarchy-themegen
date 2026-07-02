package gen

import (
	"fmt"
	"math"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

const (
	MinForegroundContrast = 4.5
	MinAccentContrast     = 3.0
	MinSelectionContrast  = 3.0
	MinTermColorDistance  = 0.12
)

func validatePaletteContrast(c *theme.Colors) []string {
	var warnings []string

	bg, errBg := ParseHex(c.Background)
	fg, errFg := ParseHex(c.Foreground)
	accent, errAccent := ParseHex(c.Accent)
	selBg, errSelBg := ParseHex(c.SelectionBackground)
	selFg, errSelFg := ParseHex(c.SelectionForeground)

	if errBg != nil || errFg != nil {
		return append(warnings, "cannot validate contrast: invalid foreground/background color format")
	}

	// Foreground/background readability
	if cr := ContrastRatio(fg, bg); cr < MinForegroundContrast {
		warnings = append(warnings, fmt.Sprintf(
			"low foreground/background contrast (%.1f:1 < %.1f:1)", cr, MinForegroundContrast))
	}

	// Accent/background distinguishability
	if errAccent == nil {
		if cr := ContrastRatio(accent, bg); cr < MinAccentContrast {
			warnings = append(warnings, fmt.Sprintf(
				"low accent/background contrast (%.1f:1 < %.1f:1)", cr, MinAccentContrast))
		}
	}

	// Selection readability
	if errSelBg == nil && errSelFg == nil {
		if cr := ContrastRatio(selFg, selBg); cr < MinSelectionContrast {
			warnings = append(warnings, fmt.Sprintf(
				"low selection contrast (%.1f:1 < %.1f:1)", cr, MinSelectionContrast))
		}
	}

	// Terminal color distinctness
	termChecks := []struct {
		a, b  int
		label string
	}{
		{1, 2, "red/green"},
		{1, 3, "red/yellow"},
		{2, 3, "green/yellow"},
		{4, 5, "blue/magenta"},
		{5, 6, "magenta/cyan"},
	}

	termColors := []string{
		c.Color0, c.Color1, c.Color2, c.Color3,
		c.Color4, c.Color5, c.Color6, c.Color7,
	}

	for _, tc := range termChecks {
		a, ea := ParseHex(termColors[tc.a])
		b, eb := ParseHex(termColors[tc.b])
		if ea != nil || eb != nil {
			continue
		}
		dist := colorDistance(a, b)
		if dist < MinTermColorDistance {
			warnings = append(warnings, fmt.Sprintf(
				"terminal %s colors are nearly identical (dist %.3f < %.3f)", tc.label, dist, MinTermColorDistance))
		}
	}

	return warnings
}

func colorDistance(a, b RGB) float64 {
	dr := a.R - b.R
	dg := a.G - b.G
	db := a.B - b.B
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func ValidateThemeContrast(c *theme.Colors) []string {
	return validatePaletteContrast(c)
}

func HasCriticalContrastFailures(c *theme.Colors) bool {
	bg, errBg := ParseHex(c.Background)
	fg, errFg := ParseHex(c.Foreground)
	if errBg != nil || errFg != nil {
		return true
	}
	// Critical: foreground must be at least marginally readable (2.0 absolute minimum)
	if ContrastRatio(fg, bg) < 2.0 {
		return true
	}

	selBg, errSelBg := ParseHex(c.SelectionBackground)
	selFg, errSelFg := ParseHex(c.SelectionForeground)
	if errSelBg == nil && errSelFg == nil {
		if ContrastRatio(selFg, selBg) < 2.0 {
			return true
		}
	}
	return false
}
