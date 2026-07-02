package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

var omarchyTemplateFiles = []string{
	"alacritty.toml.tpl",
	"btop.theme.tpl",
	"chromium.theme.tpl",
	"foot.ini.tpl",
	"ghostty.conf.tpl",
	"gum.env.conf.tpl",
	"helix.toml.tpl",
	"hyprland-preview-share-picker.css.tpl",
	"hyprland.conf.tpl",
	"hyprlock.conf.tpl",
	"keyboard.rgb.tpl",
	"kitty.conf.tpl",
	"mako.ini.tpl",
	"obsidian.css.tpl",
	"swayosd.css.tpl",
	"walker.css.tpl",
	"waybar.css.tpl",
}

func GenerateOmarchyThemedConfigs(exportDir string, colors *theme.Colors) ([]string, error) {
	templateDir, err := omarchyTemplateDir()
	if err != nil {
		return nil, err
	}
	replacements := colorTemplateReplacements(colors)

	var written []string
	for _, templateName := range omarchyTemplateFiles {
		src := filepath.Join(templateDir, templateName)
		content, err := os.ReadFile(src)
		if err != nil {
			return written, fmt.Errorf("read Omarchy template %s: %w", templateName, err)
		}
		rendered := string(content)
		keys := make([]string, 0, len(replacements))
		for key := range replacements {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
		for _, key := range keys {
			rendered = strings.ReplaceAll(rendered, key, replacements[key])
		}

		outName := strings.TrimSuffix(templateName, ".tpl")
		if err := os.WriteFile(filepath.Join(exportDir, outName), []byte(rendered), 0644); err != nil {
			return written, fmt.Errorf("write %s: %w", outName, err)
		}
		written = append(written, outName)
	}
	return written, nil
}

func omarchyTemplateDir() (string, error) {
	if omarchyPath := os.Getenv("OMARCHY_PATH"); omarchyPath != "" {
		path := filepath.Join(omarchyPath, "default", "themed")
		if isDir(path) {
			return path, nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, ".local", "share", "omarchy", "default", "themed")
	if isDir(path) {
		return path, nil
	}
	return "", fmt.Errorf("Omarchy themed templates not found")
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func colorTemplateReplacements(colors *theme.Colors) map[string]string {
	values := map[string]string{
		"accent":               colors.Accent,
		"cursor":               colors.Cursor,
		"foreground":           colors.Foreground,
		"background":           colors.Background,
		"selection_foreground": colors.SelectionForeground,
		"selection_background": colors.SelectionBackground,
		"color0":               colors.Color0,
		"color1":               colors.Color1,
		"color2":               colors.Color2,
		"color3":               colors.Color3,
		"color4":               colors.Color4,
		"color5":               colors.Color5,
		"color6":               colors.Color6,
		"color7":               colors.Color7,
		"color8":               colors.Color8,
		"color9":               colors.Color9,
		"color10":              colors.Color10,
		"color11":              colors.Color11,
		"color12":              colors.Color12,
		"color13":              colors.Color13,
		"color14":              colors.Color14,
		"color15":              colors.Color15,
	}

	replacements := make(map[string]string)
	for key, value := range values {
		replacements["{{ "+key+" }}"] = value
		stripped := strings.TrimPrefix(value, "#")
		replacements["{{ "+key+"_strip }}"] = stripped
		replacements["{{ "+key+"_rgb }}"] = hexToRGB(value)
	}
	return replacements
}

func hexToRGB(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "0,0,0"
	}
	return fmt.Sprintf("%d,%d,%d", parseHexByte(hex[0:2]), parseHexByte(hex[2:4]), parseHexByte(hex[4:6]))
}

func parseHexByte(s string) int {
	var v int
	fmt.Sscanf(s, "%02x", &v)
	return v
}
