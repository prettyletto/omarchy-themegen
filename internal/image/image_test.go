package image

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func hasMagick() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

func createTestPNG(t *testing.T, dir, name string, width, height int) string {
	t.Helper()
	if !hasMagick() {
		t.Skip("magick not available")
	}
	path := filepath.Join(dir, name)
	cmd := exec.Command("magick",
		"-size", "192x108",
		"xc:#ff0000",
		path,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create test PNG: %v: %s", err, string(out))
	}
	return path
}

func createTransparentPNG(t *testing.T, dir, name string) string {
	t.Helper()
	if !hasMagick() {
		t.Skip("magick not available")
	}
	path := filepath.Join(dir, name)
	cmd := exec.Command("magick",
		"-size", "192x108",
		"xc:transparent",
		path,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create transparent PNG: %v: %s", err, string(out))
	}
	return path
}

func createLargePNG(t *testing.T, dir, name string) string {
	t.Helper()
	if !hasMagick() {
		t.Skip("magick not available")
	}
	path := filepath.Join(dir, name)
	cmd := exec.Command("magick",
		"-size", "800x450",
		"xc:#0000ff",
		path,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create large PNG: %v: %s", err, string(out))
	}
	return path
}

func TestValidate_ValidOpaqueImage(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createLargePNG(t, dir, "valid.png")

	result := Validate(img)
	if !result.Valid {
		t.Fatalf("expected valid image, got errors: %v", result.Errors)
	}
	if result.Width < 800 || result.Height < 450 {
		t.Fatalf("expected dimensions >= 800x450, got %dx%d", result.Width, result.Height)
	}
	if !strings.HasPrefix(result.Format, "PNG") {
		t.Fatalf("expected PNG format, got %s", result.Format)
	}
}

func TestValidate_TransparentImageFails(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createTransparentPNG(t, dir, "transparent.png")

	result := Validate(img)
	if result.Valid {
		t.Fatal("expected transparent image to fail validation")
	}
}

func TestValidate_TinyImageFails(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createTestPNG(t, dir, "tiny.png", 100, 100)

	result := Validate(img)
	if result.Valid {
		t.Fatal("expected tiny image to fail validation")
	}
}

func TestValidate_NonExistentFile(t *testing.T) {
	result := Validate("/nonexistent/path/image.png")
	if result.Valid {
		t.Fatal("expected non-existent file to fail")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected errors for non-existent file")
	}
}

func TestValidate_DirectoryNotImage(t *testing.T) {
	dir := t.TempDir()
	result := Validate(dir)
	if result.Valid {
		t.Fatal("expected directory to fail as non-image")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected errors for directory")
	}
}

func TestValidate_NonImageFile(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available for non-image detection")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "notimage.txt")
	if err := os.WriteFile(path, []byte("not an image"), 0644); err != nil {
		t.Fatal(err)
	}

	result := Validate(path)
	if result.Valid {
		t.Fatal("expected non-image file to fail validation")
	}
}

func TestResult_ErrorsFormat(t *testing.T) {
	result := Validate("/nonexistent/image.png")
	for _, e := range result.Errors {
		if e == "" {
			t.Error("empty error message in result")
		}
	}
}
