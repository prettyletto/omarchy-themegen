package validate

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/anomalyco/omarchy-themegen/internal/export"
	"github.com/anomalyco/omarchy-themegen/internal/image"
	"github.com/anomalyco/omarchy-themegen/internal/theme"
)

func hasMagickInValidate() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

func buildTestModel(t *testing.T, name string) *theme.ThemeModel {
	t.Helper()
	if !hasMagickInValidate() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "source.png")
	cmd := exec.Command("magick", "-size", "800x450", "xc:#112233", imgPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create test image: %v: %s", err, string(out))
	}

	imgResult := image.Validate(imgPath)
	if !imgResult.Valid {
		t.Fatalf("image validation: %v", imgResult.Errors)
	}

	tm, err := theme.NewStatic(name, imgPath, imgResult)
	if err != nil {
		t.Fatalf("NewStatic: %v", err)
	}
	return tm
}

func TestPreExport_ValidModelPasses(t *testing.T) {
	tm := buildTestModel(t, "validate-test")
	if err := PreExport(tm); err != nil {
		t.Errorf("PreExport failed: %v", err)
	}
}

func TestPreExport_NilModelFails(t *testing.T) {
	if err := PreExport(nil); err == nil {
		t.Fatal("expected error for nil model")
	}
}

func TestPreExport_EmptyNameFails(t *testing.T) {
	tm := buildTestModel(t, "validate-test")
	tm.Name = ""
	if err := PreExport(tm); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestPreExport_EmptyNormalizedNameFails(t *testing.T) {
	tm := buildTestModel(t, "validate-test")
	tm.NormalizedName = ""
	if err := PreExport(tm); err == nil {
		t.Fatal("expected error for empty normalized name")
	}
}

func TestPreExport_NoColorsFails(t *testing.T) {
	tm := buildTestModel(t, "validate-test")
	tm.Colors = nil
	if err := PreExport(tm); err == nil {
		t.Fatal("expected error for nil colors")
	}
}

func TestPreExport_InvalidImageResultFails(t *testing.T) {
	tm := buildTestModel(t, "validate-test")
	tm.ImageResult = nil
	if err := PreExport(tm); err == nil {
		t.Fatal("expected error for nil image result")
	}
}

func TestPreExport_BadColorsFail(t *testing.T) {
	tm := buildTestModel(t, "validate-test")
	tm.Colors.Accent = ""
	if err := PreExport(tm); err == nil {
		t.Fatal("expected error for missing accent color")
	}
}

func TestPostExport_MissingFiles(t *testing.T) {
	if !hasMagickInValidate() {
		t.Skip("magick not available")
	}
	tm := buildTestModel(t, "post-validate")
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	// Export without full contents
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := PostExport(exportDir, tm.NormalizedName)
	if result.Passed {
		t.Fatal("expected validation to fail for incomplete export")
	}
	if !result.OmarchyInstalled {
		// Omarchy may or may not be installed; fine either way
		t.Log("omarchy not installed (expected reduced confidence)")
	}
}

func TestPostExport_ValidExportPasses(t *testing.T) {
	if !hasMagickInValidate() {
		t.Skip("magick not available")
	}
	tm := buildTestModel(t, "post-valid")
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	_, err := export.ThemeDirectory(tm, exportDir, false)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	result := PostExport(exportDir, tm.NormalizedName)
	if !result.Passed {
		t.Errorf("expected validation to pass, got errors: %v", result.Errors)
	}
}

func TestPostExport_ChecksColorsKeys(t *testing.T) {
	if !hasMagickInValidate() {
		t.Skip("magick not available")
	}
	tm := buildTestModel(t, "post-colors")
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	_, err := export.ThemeDirectory(tm, exportDir, false)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Verify colors.toml has all required keys
	content, err := os.ReadFile(filepath.Join(exportDir, "colors.toml"))
	if err != nil {
		t.Fatalf("cannot read colors.toml: %v", err)
	}

	result := &PostExportResult{Passed: true}
	validateColorKeys(string(content), result)

	if !result.Passed {
		t.Errorf("color key validation failed: %v", result.Errors)
	}
}
