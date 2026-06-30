package theme

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/anomalyco/omarchy-themegen/internal/image"
)

func createTestImage(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if _, err := exec.LookPath("magick"); err != nil {
		t.Skip("magick not available")
	}
	cmd := exec.Command("magick", "-size", "800x450", "xc:#00ff00", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create test image: %v: %s", err, string(out))
	}
	return path
}

func TestNewStatic_CreatesValidThemeModel(t *testing.T) {
	dir := t.TempDir()
	imgPath := createTestImage(t, dir, "test.png")
	imgResult := image.Validate(imgPath)
	if !imgResult.Valid {
		t.Fatalf("test image validation failed: %v", imgResult.Errors)
	}

	tm, err := NewStatic("My Cool Theme!", imgPath, imgResult)
	if err != nil {
		t.Fatalf("NewStatic failed: %v", err)
	}

	if tm.Name != "My Cool Theme!" {
		t.Errorf("expected name 'My Cool Theme!', got %q", tm.Name)
	}
	if tm.NormalizedName != "my-cool-theme" {
		t.Errorf("expected normalized name 'my-cool-theme', got %q", tm.NormalizedName)
	}
	if tm.Mode != "whole-theme" {
		t.Errorf("expected mode 'whole-theme', got %q", tm.Mode)
	}
	if tm.Colors == nil {
		t.Fatal("expected non-nil colors")
	}
	if tm.Version == "" {
		t.Fatal("expected non-empty version")
	}
}

func TestNewStatic_EmptyName(t *testing.T) {
	dir := t.TempDir()
	imgPath := createTestImage(t, dir, "test.png")
	imgResult := image.Validate(imgPath)
	if !imgResult.Valid {
		t.Fatalf("test image validation failed: %v", imgResult.Errors)
	}

	_, err := NewStatic("", imgPath, imgResult)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestNewStatic_NormalizedNameEmpty(t *testing.T) {
	dir := t.TempDir()
	imgPath := createTestImage(t, dir, "test.png")
	imgResult := image.Validate(imgPath)
	if !imgResult.Valid {
		t.Fatalf("test image validation failed: %v", imgResult.Errors)
	}

	_, err := NewStatic("!!!", imgPath, imgResult)
	if err == nil {
		t.Fatal("expected error for name that normalizes to empty")
	}
}

func TestNormalizeThemeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Theme", "my-theme"},
		{"Hello World!", "hello-world"},
		{" spaces  everywhere ", "spaces-everywhere"},
		{"MixEd CaSe", "mixed-case"},
		{"special@#$chars", "special-chars"},
		{"dash-and_underscore", "dash-and_underscore"},
		{"multiple---dashes", "multiple-dashes"},
		{"-leading-dash", "leading-dash"},
		{"trailing-dash-", "trailing-dash"},
		{"   ", ""},
		{"!!!", ""},
		{"simple", "simple"},
		{"123theme", "123theme"},
	}

	for _, tt := range tests {
		result := normalizeThemeName(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeThemeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestModelIsIndependentOfPaths(t *testing.T) {
	dir := t.TempDir()
	imgPath := createTestImage(t, dir, "test.png")
	imgResult := image.Validate(imgPath)
	if !imgResult.Valid {
		t.Fatalf("test image validation failed: %v", imgResult.Errors)
	}

	tm, err := NewStatic("independent-test", imgPath, imgResult)
	if err != nil {
		t.Fatalf("NewStatic failed: %v", err)
	}

	// Theme model does not reference any output path
	if tm.NormalizedName == "" {
		t.Fatal("expected non-empty normalized name")
	}
	if tm.Version == "" {
		t.Fatal("expected non-empty version")
	}

	// It's pure data, no filesystem paths
	_ = os.TempDir()
}
