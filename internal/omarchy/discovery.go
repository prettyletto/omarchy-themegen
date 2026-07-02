package omarchy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Discovery struct {
	Installed    bool
	BinaryPath   string
	UserThemeDir string
	OfficialDir  string
	TemplateDir  string
	ThemeSetCmd  string
	ThemeListCmd string
	Diagnostics  []string
}

func Discover() *Discovery {
	d := &Discovery{}

	home, err := os.UserHomeDir()
	if err != nil {
		d.Diagnostics = append(d.Diagnostics, fmt.Sprintf("cannot determine home directory: %v", err))
		return d
	}

	// Check binary
	binaryPath, err := exec.LookPath("omarchy")
	if err == nil {
		d.BinaryPath = binaryPath
		d.Installed = true
	} else {
		d.Diagnostics = append(d.Diagnostics, "omarchy binary not found on PATH")
	}

	// Check user theme directory
	userThemeDir := filepath.Join(home, ".config", "omarchy", "themes")
	if info, err := os.Stat(userThemeDir); err == nil && info.IsDir() {
		d.UserThemeDir = userThemeDir
	} else {
		d.Diagnostics = append(d.Diagnostics, fmt.Sprintf("user theme directory not found: %s", userThemeDir))
	}

	// Check official theme directory
	officialDir := filepath.Join(home, ".local", "share", "omarchy", "themes")
	if info, err := os.Stat(officialDir); err == nil && info.IsDir() {
		d.OfficialDir = officialDir
	}

	// Check template directory
	templateDir := filepath.Join(home, ".local", "share", "omarchy", "templates")
	if info, err := os.Stat(templateDir); err == nil && info.IsDir() {
		d.TemplateDir = templateDir
	} else if d.Installed {
		d.Diagnostics = append(d.Diagnostics, fmt.Sprintf("template directory not found: %s", templateDir))
	}

	// Check commands
	if d.Installed {
		d.ThemeSetCmd = "omarchy theme set"
		d.ThemeListCmd = "omarchy theme list"
	}

	return d
}

func (d *Discovery) Confidence() string {
	if d.Installed {
		if d.TemplateDir != "" {
			return "high"
		}
		return "medium"
	}
	return "reduced"
}

func (d *Discovery) ValidateThemeDir(path string) []string {
	var errs []string

	if _, err := os.Stat(path); os.IsNotExist(err) {
		errs = append(errs, fmt.Sprintf("theme directory not found: %s", path))
		return errs
	}

	requiredFiles := []string{
		"colors.toml",
		"preview.png",
		"preview-unlock.png",
		"unlock.png",
		"neovim.lua",
		"README.md",
	}
	for _, f := range requiredFiles {
		if _, err := os.Stat(filepath.Join(path, f)); os.IsNotExist(err) {
			errs = append(errs, fmt.Sprintf("required file missing: %s", f))
		}
	}

	// Check background assets
	bgDir := filepath.Join(path, "backgrounds")
	if entries, err := os.ReadDir(bgDir); err != nil || len(entries) == 0 {
		errs = append(errs, "backgrounds directory missing or empty")
	}

	// Check preview dimensions
	if errs2 := validatePreviewDimensions(path); len(errs2) > 0 {
		errs = append(errs, errs2...)
	}

	// Check for light.mode behavior
	lightModePath := filepath.Join(path, "light.mode")
	if _, err := os.Stat(lightModePath); err == nil {
		// light.mode exists — validate it
		data, err := os.ReadFile(lightModePath)
		if err != nil || len(data) == 0 {
			errs = append(errs, "light.mode file exists but is empty or unreadable")
		}
	}

	// Check colors.toml has required keys
	if content, err := os.ReadFile(filepath.Join(path, "colors.toml")); err == nil {
		validateColorKeys(string(content), &errs)
	}

	return errs
}

func validatePreviewDimensions(themeDir string) []string {
	var errs []string

	magick, err := exec.LookPath("magick")
	if err != nil {
		return errs
	}

	previewPath := filepath.Join(themeDir, "preview.png")
	if _, err := os.Stat(previewPath); err == nil {
		cmd := exec.Command(magick, "identify", "-format", "%wx%h", previewPath)
		if out, err := cmd.Output(); err == nil {
			dim := string(out)
			if dim != "1800x1012" {
				errs = append(errs, fmt.Sprintf("preview.png dimensions %s, expected 1800x1012", dim))
			}
		}
	}

	previewUnlockPath := filepath.Join(themeDir, "preview-unlock.png")
	if _, err := os.Stat(previewUnlockPath); err == nil {
		cmd := exec.Command(magick, "identify", "-format", "%wx%h", previewUnlockPath)
		if out, err := cmd.Output(); err == nil {
			dim := string(out)
			if dim != "1920x1080" {
				errs = append(errs, fmt.Sprintf("preview-unlock.png dimensions %s, expected 1920x1080", dim))
			}
		}
	}

	return errs
}

func validateColorKeys(content string, errs *[]string) {
	requiredKeys := []string{
		"accent", "cursor", "foreground", "background",
		"selection_foreground", "selection_background",
		"color0", "color1", "color2", "color3",
		"color4", "color5", "color6", "color7",
		"color8", "color9", "color10", "color11",
		"color12", "color13", "color14", "color15",
	}
	for _, key := range requiredKeys {
		search := key + " = \""
		if len(content) < len(search) || !containsStr(content, search) {
			*errs = append(*errs, fmt.Sprintf("colors.toml missing key: %s", key))
		}
	}
}

func containsStr(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
