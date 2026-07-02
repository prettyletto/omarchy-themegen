package validate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/prettyletto/omarchy-themegen/internal/omarchy"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

type PostExportResult struct {
	Passed           bool
	OmarchyInstalled bool
	Errors           []string
	Warnings         []string
}

func PreExport(tm *theme.ThemeModel) error {
	if tm == nil {
		return fmt.Errorf("no Theme Model provided")
	}

	if tm.NormalizedName == "" {
		return fmt.Errorf("normalized theme name is empty")
	}

	if tm.Name == "" {
		return fmt.Errorf("theme name was not provided")
	}

	if tm.SourceImage == "" {
		return fmt.Errorf("no source image specified")
	}

	if tm.ImageResult == nil || !tm.ImageResult.Valid {
		return fmt.Errorf("source image validation did not pass")
	}

	if tm.Colors == nil {
		return fmt.Errorf("no color palette in theme model")
	}

	if errs := theme.ValidateColors(tm.Colors); len(errs) > 0 {
		return fmt.Errorf("color validation failed: %v", errs)
	}

	return nil
}

func PostExport(exportDir, normalizedName string) *PostExportResult {
	result := &PostExportResult{Passed: true}

	// Check target directory exists
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		result.Errors = append(result.Errors, fmt.Sprintf("export directory does not exist: %s", exportDir))
		result.Passed = false
		return result
	}

	// Check required files exist
	requiredFiles := []string{
		"colors.toml",
		"alacritty.toml",
		"btop.theme",
		"chromium.theme",
		"foot.ini",
		"ghostty.conf",
		"gum.env.conf",
		"helix.toml",
		"hyprland-preview-share-picker.css",
		"hyprland.conf",
		"hyprlock.conf",
		"keyboard.rgb",
		"kitty.conf",
		"mako.ini",
		"preview.png",
		"preview-unlock.png",
		"obsidian.css",
		"swayosd.css",
		"walker.css",
		"waybar.css",
		"unlock.png",
		"neovim.lua",
		"README.md",
	}

	for _, f := range requiredFiles {
		path := filepath.Join(exportDir, f)
		fi, err := os.Stat(path)
		if os.IsNotExist(err) {
			result.Errors = append(result.Errors, fmt.Sprintf("required file missing: %s", f))
			result.Passed = false
		} else if err == nil && fi.Size() == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("file is empty: %s", f))
		}
	}

	// Check backgrounds directory has at least one file
	bgDir := filepath.Join(exportDir, "backgrounds")
	if entries, err := os.ReadDir(bgDir); err != nil {
		result.Errors = append(result.Errors, "backgrounds directory missing or unreadable")
		result.Passed = false
	} else if len(entries) == 0 {
		result.Errors = append(result.Errors, "backgrounds directory is empty")
		result.Passed = false
	}

	// Validate colors.toml contains all required keys
	if content, err := os.ReadFile(filepath.Join(exportDir, "colors.toml")); err == nil {
		validateColorKeys(string(content), result)
	} else {
		result.Errors = append(result.Errors, "cannot read colors.toml for validation")
		result.Passed = false
	}

	// Check if Omarchy is installed (optional)
	disc := omarchy.Discover()
	if disc.Installed {
		result.OmarchyInstalled = true
		// Run theme list check (non-mutating)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, disc.BinaryPath, "theme", "list")
		output, err := cmd.Output()
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("omarchy detected but theme list check failed: %v", err))
			result.OmarchyInstalled = false
		} else {
			_ = output
		}
	} else {
		result.OmarchyInstalled = false
		for _, diag := range disc.Diagnostics {
			result.Warnings = append(result.Warnings, "omarchy: "+diag)
		}
		result.Warnings = append(result.Warnings, fmt.Sprintf("omarchy not installed; reduced validation confidence (%s)", disc.Confidence()))
	}

	// Preview dimension validation
	if errs := checkPreviewDimensions(exportDir); len(errs) > 0 {
		result.Errors = append(result.Errors, errs...)
		result.Passed = false
	}

	return result
}

func validateColorKeys(content string, result *PostExportResult) {
	requiredKeys := []string{
		"accent", "cursor", "foreground", "background",
		"selection_foreground", "selection_background",
		"color0", "color1", "color2", "color3",
		"color4", "color5", "color6", "color7",
		"color8", "color9", "color10", "color11",
		"color12", "color13", "color14", "color15",
	}

	for _, key := range requiredKeys {
		if !strings.Contains(content, key+" = \"") {
			result.Errors = append(result.Errors, fmt.Sprintf("colors.toml missing key: %s", key))
			result.Passed = false
		}
	}
}

func checkPreviewDimensions(exportDir string) []string {
	var errs []string

	magick, err := exec.LookPath("magick")
	if err != nil {
		return nil
	}

	checks := []struct {
		file     string
		expected string
	}{
		{"preview.png", "1800x1012"},
		{"preview-unlock.png", "1920x1080"},
	}

	for _, c := range checks {
		path := filepath.Join(exportDir, c.file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, magick, "identify", "-format", "%wx%h", path)
		out, err := cmd.Output()
		cancel()
		if err != nil {
			continue
		}
		dim := strings.TrimSpace(string(out))
		if dim != c.expected {
			errs = append(errs, fmt.Sprintf("%s dimensions %s, expected %s", c.file, dim, c.expected))
		}
	}

	return errs
}
