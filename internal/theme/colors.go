package theme

import (
	"fmt"
	"regexp"
	"strings"
)

var hexColorRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type Colors struct {
	Accent             string `toml:"accent"`
	Cursor             string `toml:"cursor"`
	Foreground         string `toml:"foreground"`
	Background         string `toml:"background"`
	SelectionForeground string `toml:"selection_foreground"`
	SelectionBackground string `toml:"selection_background"`
	Color0  string `toml:"color0"`
	Color1  string `toml:"color1"`
	Color2  string `toml:"color2"`
	Color3  string `toml:"color3"`
	Color4  string `toml:"color4"`
	Color5  string `toml:"color5"`
	Color6  string `toml:"color6"`
	Color7  string `toml:"color7"`
	Color8  string `toml:"color8"`
	Color9  string `toml:"color9"`
	Color10 string `toml:"color10"`
	Color11 string `toml:"color11"`
	Color12 string `toml:"color12"`
	Color13 string `toml:"color13"`
	Color14 string `toml:"color14"`
	Color15 string `toml:"color15"`
}

func StaticColors() *Colors {
	return &Colors{
		Accent:             "#82aaff",
		Cursor:             "#c792ea",
		Foreground:         "#bbc2cf",
		Background:         "#1a1b26",
		SelectionForeground: "#1a1b26",
		SelectionBackground: "#82aaff",
		Color0:  "#1a1b26",
		Color1:  "#db4b4b",
		Color2:  "#9ece6a",
		Color3:  "#e0af68",
		Color4:  "#7aa2f7",
		Color5:  "#bb9af7",
		Color6:  "#7dcfff",
		Color7:  "#a9b1d6",
		Color8:  "#3b4261",
		Color9:  "#db4b4b",
		Color10: "#9ece6a",
		Color11: "#e0af68",
		Color12: "#7aa2f7",
		Color13: "#bb9af7",
		Color14: "#7dcfff",
		Color15: "#c0caf5",
	}
}

func ValidateColors(c *Colors) []string {
	var errs []string

	check := func(name, val string) {
		if val == "" {
			errs = append(errs, fmt.Sprintf("missing required color key: %s", name))
			return
		}
		if !hexColorRE.MatchString(val) {
			errs = append(errs, fmt.Sprintf("invalid color for %s: %q (must be #RRGGBB)", name, val))
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

	return errs
}

func (c *Colors) ToTOML() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("accent = \"%s\"\n", strings.ToLower(c.Accent)))
	b.WriteString(fmt.Sprintf("cursor = \"%s\"\n", strings.ToLower(c.Cursor)))
	b.WriteString(fmt.Sprintf("foreground = \"%s\"\n", strings.ToLower(c.Foreground)))
	b.WriteString(fmt.Sprintf("background = \"%s\"\n", strings.ToLower(c.Background)))
	b.WriteString(fmt.Sprintf("selection_foreground = \"%s\"\n", strings.ToLower(c.SelectionForeground)))
	b.WriteString(fmt.Sprintf("selection_background = \"%s\"\n", strings.ToLower(c.SelectionBackground)))
	b.WriteString(fmt.Sprintf("color0 = \"%s\"\n", strings.ToLower(c.Color0)))
	b.WriteString(fmt.Sprintf("color1 = \"%s\"\n", strings.ToLower(c.Color1)))
	b.WriteString(fmt.Sprintf("color2 = \"%s\"\n", strings.ToLower(c.Color2)))
	b.WriteString(fmt.Sprintf("color3 = \"%s\"\n", strings.ToLower(c.Color3)))
	b.WriteString(fmt.Sprintf("color4 = \"%s\"\n", strings.ToLower(c.Color4)))
	b.WriteString(fmt.Sprintf("color5 = \"%s\"\n", strings.ToLower(c.Color5)))
	b.WriteString(fmt.Sprintf("color6 = \"%s\"\n", strings.ToLower(c.Color6)))
	b.WriteString(fmt.Sprintf("color7 = \"%s\"\n", strings.ToLower(c.Color7)))
	b.WriteString(fmt.Sprintf("color8 = \"%s\"\n", strings.ToLower(c.Color8)))
	b.WriteString(fmt.Sprintf("color9 = \"%s\"\n", strings.ToLower(c.Color9)))
	b.WriteString(fmt.Sprintf("color10 = \"%s\"\n", strings.ToLower(c.Color10)))
	b.WriteString(fmt.Sprintf("color11 = \"%s\"\n", strings.ToLower(c.Color11)))
	b.WriteString(fmt.Sprintf("color12 = \"%s\"\n", strings.ToLower(c.Color12)))
	b.WriteString(fmt.Sprintf("color13 = \"%s\"\n", strings.ToLower(c.Color13)))
	b.WriteString(fmt.Sprintf("color14 = \"%s\"\n", strings.ToLower(c.Color14)))
	b.WriteString(fmt.Sprintf("color15 = \"%s\"\n", strings.ToLower(c.Color15)))

	return b.String()
}
