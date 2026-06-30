package theme

import (
	"strings"
	"testing"
)

func TestStaticColors_AllKeysPresent(t *testing.T) {
	c := StaticColors()

	check := func(name, val string) {
		if val == "" {
			t.Errorf("static color %s is empty", name)
		}
	}

	check("accent", c.Accent)
	check("cursor", c.Cursor)
	check("foreground", c.Foreground)
	check("background", c.Background)
	check("selection_foreground", c.SelectionForeground)
	check("selection_background", c.SelectionBackground)
	check("color0", c.Color0)
	check("color1", c.Color1)
	check("color2", c.Color2)
	check("color3", c.Color3)
	check("color4", c.Color4)
	check("color5", c.Color5)
	check("color6", c.Color6)
	check("color7", c.Color7)
	check("color8", c.Color8)
	check("color9", c.Color9)
	check("color10", c.Color10)
	check("color11", c.Color11)
	check("color12", c.Color12)
	check("color13", c.Color13)
	check("color14", c.Color14)
	check("color15", c.Color15)
}

func TestValidateColors_ValidPasses(t *testing.T) {
	c := StaticColors()
	errs := ValidateColors(c)
	if len(errs) > 0 {
		t.Errorf("expected valid static colors, got errors: %v", errs)
	}
}

func TestValidateColors_MissingKey(t *testing.T) {
	c := StaticColors()
	c.Accent = ""
	errs := ValidateColors(c)
	if len(errs) == 0 {
		t.Fatal("expected error for missing accent")
	}
}

func TestValidateColors_InvalidFormat(t *testing.T) {
	c := StaticColors()
	c.Foreground = "not-a-color"
	errs := ValidateColors(c)
	if len(errs) == 0 {
		t.Fatal("expected error for invalid color format")
	}
}

func TestValidateColors_NonHex(t *testing.T) {
	c := StaticColors()
	c.Background = "#GGGGGG"
	errs := ValidateColors(c)
	if len(errs) == 0 {
		t.Fatal("expected error for non-hex color")
	}
}

func TestValidateColors_ShortHex(t *testing.T) {
	c := StaticColors()
	c.Color0 = "#FFF"
	errs := ValidateColors(c)
	if len(errs) == 0 {
		t.Fatal("expected error for short hex color")
	}
}

func TestColors_ToTOML_ContainsAllKeys(t *testing.T) {
	c := StaticColors()
	toml := c.ToTOML()

	requiredKeys := []string{
		"accent", "cursor", "foreground", "background",
		"selection_foreground", "selection_background",
		"color0", "color1", "color2", "color3",
		"color4", "color5", "color6", "color7",
		"color8", "color9", "color10", "color11",
		"color12", "color13", "color14", "color15",
	}

	for _, key := range requiredKeys {
		if !strings.Contains(toml, key+" = \"") {
			t.Errorf("TOML output missing key: %s", key)
		}
	}
}

func TestColors_ToTOML_Lowercase(t *testing.T) {
	c := StaticColors()
	// Set uppercase to test lowercasing
	c.Accent = "#AABBCC"
	toml := c.ToTOML()

	if !strings.Contains(toml, "\"#aabbcc\"") {
		t.Errorf("expected lowercase hex in TOML, got: %s", toml)
	}
}
