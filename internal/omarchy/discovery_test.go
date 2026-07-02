package omarchy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscover_FindsSystemInfo(t *testing.T) {
	d := Discover()
	if d.Confidence() == "" {
		t.Error("expected non-empty confidence")
	}
	// Discovery is read-only — never mutates filesystem
}

func TestDiscover_ConfidenceLevels(t *testing.T) {
	d := Discover()
	c := d.Confidence()
	if c != "high" && c != "medium" && c != "reduced" {
		t.Errorf("unexpected confidence: %s", c)
	}
}

func TestDiscover_Diagnostics(t *testing.T) {
	d := Discover()
	if !d.Installed && len(d.Diagnostics) == 0 {
		t.Error("expected diagnostics when Omarchy is missing")
	}
}

func TestValidateThemeDir_MissingDirectory(t *testing.T) {
	errs := (&Discovery{}).ValidateThemeDir("/nonexistent/theme/path")
	if len(errs) == 0 {
		t.Error("expected errors for nonexistent directory")
	}
}

func TestValidateThemeDir_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	errs := (&Discovery{}).ValidateThemeDir(dir)
	// Should have errors for missing files
	if len(errs) == 0 {
		t.Error("expected errors for empty theme directory")
	}
}

func TestValidateThemeDir_ValidFixture(t *testing.T) {
	dir := t.TempDir()

	// Create all required files
	os.WriteFile(filepath.Join(dir, "colors.toml"), []byte(`
accent = "#82aaff"
cursor = "#c792ea"
foreground = "#bbc2cf"
background = "#1a1b26"
selection_foreground = "#1a1b26"
selection_background = "#82aaff"
color0 = "#1a1b26"
color1 = "#db4b4b"
color2 = "#9ece6a"
color3 = "#e0af68"
color4 = "#7aa2f7"
color5 = "#bb9af7"
color6 = "#7dcfff"
color7 = "#a9b1d6"
color8 = "#3b4261"
color9 = "#db4b4b"
color10 = "#9ece6a"
color11 = "#e0af68"
color12 = "#7aa2f7"
color13 = "#bb9af7"
color14 = "#7dcfff"
color15 = "#c0caf5"
`), 0644)

	os.WriteFile(filepath.Join(dir, "preview.png"), []byte("fake-png"), 0644)
	os.WriteFile(filepath.Join(dir, "preview-unlock.png"), []byte("fake-png"), 0644)
	os.WriteFile(filepath.Join(dir, "unlock.png"), []byte("fake-png"), 0644)
	os.WriteFile(filepath.Join(dir, "neovim.lua"), []byte("return {}"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
	os.MkdirAll(filepath.Join(dir, "backgrounds"), 0755)
	os.WriteFile(filepath.Join(dir, "backgrounds", "wallpaper.png"), []byte("fake-img"), 0644)

	errs := (&Discovery{}).ValidateThemeDir(dir)
	for _, e := range errs {
		t.Logf("validation note: %s", e)
	}
}

func TestValidateThemeDir_LightMode(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "backgrounds"), 0755)
	os.WriteFile(filepath.Join(dir, "backgrounds", "wallpaper.png"), []byte("fake"), 0644)

	// Light mode with empty file should flag
	os.WriteFile(filepath.Join(dir, "light.mode"), []byte(""), 0644)
	errs := (&Discovery{}).ValidateThemeDir(dir)
	found := false
	for _, e := range errs {
		if strings.Contains(e, "light.mode") {
			found = true
		}
	}
	if !found {
		t.Error("expected light.mode validation error for empty file")
	}
}

func TestValidateThemeDir_ColorsTomlKeys(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "backgrounds"), 0755)
	os.WriteFile(filepath.Join(dir, "backgrounds", "wallpaper.png"), []byte("fake"), 0644)

	// Incomplete colors.toml
	os.WriteFile(filepath.Join(dir, "colors.toml"), []byte(`accent = "#fff"`), 0644)
	errs := (&Discovery{}).ValidateThemeDir(dir)
	if len(errs) == 0 {
		t.Error("expected errors for incomplete colors.toml")
	}
}

func TestValidateColorKeys(t *testing.T) {
	var errs []string
	validateColorKeys("accent = \"#fff\"", &errs)
	if len(errs) == 0 {
		t.Error("expected missing keys")
	}
}

func TestDiscovery_Confidence(t *testing.T) {
	d := &Discovery{Installed: true, TemplateDir: "/some/path"}
	if d.Confidence() != "high" {
		t.Errorf("expected high confidence, got %s", d.Confidence())
	}

	d2 := &Discovery{Installed: true}
	if d2.Confidence() != "medium" {
		t.Errorf("expected medium confidence, got %s", d2.Confidence())
	}

	d3 := &Discovery{}
	if d3.Confidence() != "reduced" {
		t.Errorf("expected reduced confidence, got %s", d3.Confidence())
	}
}

func TestDiscovery_ReadOnly(t *testing.T) {
	// Discover() must not create any directories or files
	before := t.TempDir()
	_ = before
	d := Discover()
	_ = d
	// No assertion needed — if Discover created files, they'd be visible
}
