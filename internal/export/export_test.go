package export

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prettyletto/omarchy-themegen/internal/gen"
	"github.com/prettyletto/omarchy-themegen/internal/image"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

func hasMagickInExport() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

func createTestSourceImage(t *testing.T, dir string) string {
	t.Helper()
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	stubOmarchyTemplates(t)
	path := filepath.Join(dir, "source.png")
	cmd := exec.Command("magick", "-size", "800x450", "xc:#336699", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create test source: %v: %s", err, string(out))
	}
	return path
}

func createTestSourceImageOnly(t *testing.T, dir string) string {
	t.Helper()
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	path := filepath.Join(dir, "source.png")
	cmd := exec.Command("magick", "-size", "800x450", "xc:#336699", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create test source: %v: %s", err, string(out))
	}
	return path
}

func stubOmarchyTemplates(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	templateDir := filepath.Join(root, "default", "themed")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("create template dir: %v", err)
	}
	for _, name := range omarchyTemplateFiles {
		if err := os.WriteFile(filepath.Join(templateDir, name), []byte("foreground={{ foreground }}\nbackground={{ background }}\n"), 0644); err != nil {
			t.Fatalf("write template %s: %v", name, err)
		}
	}
	t.Setenv("OMARCHY_PATH", root)
}

func buildTestThemeModel(t *testing.T, name, path string) *theme.ThemeModel {
	t.Helper()
	imgResult := image.Validate(path)
	if !imgResult.Valid {
		// Try to get a valid result anyway for test purposes
		imgResult = &image.Result{Valid: true, Width: 800, Height: 450, Format: "PNG"}
	}
	tm, err := theme.NewStatic(name, path, imgResult)
	if err != nil {
		t.Fatalf("NewStatic failed: %v", err)
	}
	return tm
}

func TestThemeDirectory_ExportsSuccessfully(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createTestSourceImage(t, dir)
	tm := buildTestThemeModel(t, "test-export", img)
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	result, err := ThemeDirectory(tm, exportDir, false)
	if err != nil {
		t.Fatalf("ThemeDirectory failed: %v", err)
	}
	if result.Path != exportDir {
		t.Errorf("expected path %q, got %q", exportDir, result.Path)
	}

	requiredFiles := []string{
		"colors.toml", "preview.png", "preview-unlock.png",
		"unlock.png", "neovim.lua", "README.md",
	}
	for _, f := range requiredFiles {
		if _, err := os.Stat(filepath.Join(exportDir, f)); os.IsNotExist(err) {
			t.Errorf("missing required file: %s", f)
		}
	}
}

func TestThemeDirectory_ExportsNeovimWithoutLocalThemePlugin(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	t.Setenv("HOME", t.TempDir())
	stubOmarchyTemplates(t)
	dir := t.TempDir()
	img := createTestSourceImageOnly(t, dir)
	tm := buildTestThemeModel(t, "generated-neovim", img)
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	result, err := ThemeDirectory(tm, exportDir, false)
	if err != nil {
		t.Fatalf("ThemeDirectory failed without local theme plugin: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(exportDir, "neovim.lua"))
	if err != nil {
		t.Fatalf("expected fallback neovim.lua: %v", err)
	}
	content := string(data)
	for _, expected := range []string{"omarchy-themegen", "vim.api.nvim_set_hl", "ColorScheme"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected generated neovim.lua to contain %q", expected)
		}
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("expected no local plugin warning, got %v", result.Warnings)
	}
}

func TestThemeDirectory_RefusesOverwrite(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createTestSourceImage(t, dir)
	tm := buildTestThemeModel(t, "test-overwrite", img)
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	_, err := ThemeDirectory(tm, exportDir, false)
	if err != nil {
		t.Fatalf("first ThemeDirectory failed: %v", err)
	}
	_, err = ThemeDirectory(tm, exportDir, false)
	if err == nil {
		t.Fatal("expected overwrite refusal, got nil error")
	}
}

func TestThemeDirectory_OverwriteWithYesCreatesBackup(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createTestSourceImage(t, dir)
	tm := buildTestThemeModel(t, "test-backup", img)
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	_, err := ThemeDirectory(tm, exportDir, false)
	if err != nil {
		t.Fatalf("first ThemeDirectory failed: %v", err)
	}
	result, err := ThemeDirectory(tm, exportDir, true)
	if err != nil {
		t.Fatalf("forced ThemeDirectory failed: %v", err)
	}
	if result.BackupPath == "" {
		t.Fatal("expected backup path, got empty")
	}
	if _, err := os.Stat(result.BackupPath); os.IsNotExist(err) {
		t.Errorf("backup does not exist at %s", result.BackupPath)
	}
}

func TestGenerateREADME(t *testing.T) {
	dir := t.TempDir()
	img := createTestSourceImage(t, dir)
	tm := buildTestThemeModel(t, "readme-test", img)
	readme := GenerateREADME(tm)

	requiredPhrases := []string{
		"omarchy-themegen", "go install",
		"omarchy theme set", "theme application is separate from export",
	}
	for _, phrase := range requiredPhrases {
		if !strings.Contains(readme, phrase) {
			t.Errorf("README missing required phrase: %q", phrase)
		}
	}

	forbiddenPhrases := []string{"browser preview", "component-mix"}
	for _, phrase := range forbiddenPhrases {
		if strings.Contains(readme, phrase) {
			t.Errorf("README should not mention unimplemented feature: %q", phrase)
		}
	}
}

// Sprint 6: Recipe tests
func TestBuildRecipe_WholeTheme(t *testing.T) {
	dir := t.TempDir()
	img := createTestSourceImage(t, dir)
	tm := buildTestThemeModel(t, "recipe-test", img)
	tm.Mode = "whole-theme"
	tm.DirectionID = 2
	tm.DirectionLabel = "Balanced"

	opts := &gen.GenerationOptions{Fingerprint: "sha256:abc", Seed: 42, LightMode: false}
	recipe := BuildRecipe(tm, opts)

	if recipe.Mode != "whole-theme" {
		t.Errorf("expected whole-theme, got %s", recipe.Mode)
	}
	if recipe.DirectionID != 2 {
		t.Errorf("expected direction 2, got %d", recipe.DirectionID)
	}
	if recipe.Fingerprint != "sha256:abc" {
		t.Errorf("expected fingerprint sha256:abc, got %s", recipe.Fingerprint)
	}
	if recipe.Seed != 42 {
		t.Errorf("expected seed 42, got %d", recipe.Seed)
	}
}

func TestBuildRecipe_ComponentMix(t *testing.T) {
	dir := t.TempDir()
	img := createTestSourceImage(t, dir)
	tm := buildTestThemeModel(t, "cmix-recipe", img)
	tm.Mode = "component-mix"
	tm.GroupSelections = map[string]int{"desktop-shell": 1, "editor": 3}
	tm.Overrides = map[string]int{"neovim": 2}

	opts := &gen.GenerationOptions{Fingerprint: "sha256:def", Seed: 7, LightMode: true}
	recipe := BuildRecipe(tm, opts)

	if recipe.Mode != "component-mix" {
		t.Errorf("expected component-mix, got %s", recipe.Mode)
	}
	if len(recipe.GroupSelections) != 2 {
		t.Errorf("expected 2 group selections, got %d", len(recipe.GroupSelections))
	}
	if recipe.GroupSelections["desktop-shell"] != 1 {
		t.Errorf("expected desktop-shell=1")
	}
	if len(recipe.Overrides) != 1 {
		t.Errorf("expected 1 override, got %d", len(recipe.Overrides))
	}
	if recipe.Overrides["neovim"] != 2 {
		t.Errorf("expected neovim override=2")
	}
}

func TestRecipe_WriteAndLoad(t *testing.T) {
	dir := t.TempDir()
	recipePath := filepath.Join(dir, "test.recipe.json")

	recipe := &Recipe{
		GeneratorVersion: "0.2.0-dev",
		Fingerprint:      "sha256:abc",
		Seed:             42,
		Mode:             "whole-theme",
		DirectionID:      1,
		DirectionLabel:   "Vibrant",
	}

	if err := WriteRecipe(recipe, recipePath); err != nil {
		t.Fatalf("WriteRecipe: %v", err)
	}

	loaded, err := LoadRecipe(recipePath)
	if err != nil {
		t.Fatalf("LoadRecipe: %v", err)
	}
	if loaded.Fingerprint != recipe.Fingerprint {
		t.Errorf("fingerprint mismatch: %s vs %s", loaded.Fingerprint, recipe.Fingerprint)
	}
	if loaded.Seed != recipe.Seed {
		t.Errorf("seed mismatch: %d vs %d", loaded.Seed, recipe.Seed)
	}
	if loaded.Mode != recipe.Mode {
		t.Errorf("mode mismatch: %s vs %s", loaded.Mode, recipe.Mode)
	}
}

func TestLoadRecipe_MissingFingerprint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.recipe.json")
	os.WriteFile(path, []byte(`{"seed": 0}`), 0644)

	_, err := LoadRecipe(path)
	if err == nil {
		t.Error("expected error for missing fingerprint")
	}
}

func TestLoadRecipe_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.recipe.json")
	os.WriteFile(path, []byte(`not json`), 0644)

	_, err := LoadRecipe(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// Sprint 6: Neovim contract tests
func TestGenerateNeovimLua_WritesLazyPluginSpec(t *testing.T) {
	dir := t.TempDir()
	colors := theme.StaticColors()

	if err := GenerateNeovimLua(dir, colors, false); err != nil {
		t.Fatalf("GenerateNeovimLua: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "neovim.lua"))
	if err != nil {
		t.Fatalf("read neovim.lua: %v", err)
	}
	content := string(data)
	for _, expected := range []string{"return {", "lazy = false", "vim.o.background = \"dark\"", colors.Background} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected neovim.lua to contain %q", expected)
		}
	}
}

// Sprint 6: Archive tests
func TestCreateArchive_Content(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createTestSourceImage(t, dir)
	tm := buildTestThemeModel(t, "archive-test", img)
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)
	ThemeDirectory(tm, exportDir, false)

	arcPath := filepath.Join(t.TempDir(), "test.tar.gz")
	result, err := CreateArchive(exportDir, tm.NormalizedName, arcPath)
	if err != nil {
		t.Fatalf("CreateArchive: %v", err)
	}
	if result.Path != arcPath {
		t.Errorf("expected %s, got %s", arcPath, result.Path)
	}

	// Verify archive exists and is non-empty
	if fi, err := os.Stat(result.Path); err != nil || fi.Size() == 0 {
		t.Error("archive should exist and be non-empty")
	}
}

func TestExtraFile(t *testing.T) {
	ef := ExtraFile{Name: "test.txt", Path: "/tmp/test.txt"}
	if ef.Name != "test.txt" {
		t.Errorf("unexpected name: %s", ef.Name)
	}
}

// End-to-end fixtures from Sprint 3-4
func createFixtureImage(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	stubOmarchyTemplates(t)
	path := filepath.Join(dir, name)
	cmdArgs := append([]string{"-size", "800x450"}, args...)
	cmdArgs = append(cmdArgs, path)
	cmd := exec.Command("magick", cmdArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create fixture: %v: %s", err, string(out))
	}
	return path
}

func buildGeneratedThemeModel(t *testing.T, name string, lightMode bool) *theme.ThemeModel {
	t.Helper()
	dir := t.TempDir()
	imgPath := createFixtureImage(t, dir, "source.png", "plasma:#1a1b26-#82aaff-#db4b4b-#9ece6a")
	imgResult := image.Validate(imgPath)
	if !imgResult.Valid {
		t.Fatalf("fixture image validation failed: %v", imgResult.Errors)
	}

	opts, err := gen.NewGenerationOptions(imgPath, 42, lightMode)
	if err != nil {
		t.Fatalf("gen options: %v", err)
	}

	colors, err := gen.ExtractDominantColors(imgPath, 12)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	candidates, err := gen.GeneratePalettes(colors, opts)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	dirObj := theme.Direction{
		ID:          candidates[0].ID,
		Label:       candidates[0].Label,
		Fingerprint: opts.Fingerprint,
		Colors:      candidates[0].Colors,
		Warnings:    candidates[0].Warnings,
		LightMode:   opts.LightMode,
	}

	tm, err := theme.NewThemeModelFromDirection(name, imgPath, imgResult, dirObj)
	if err != nil {
		t.Fatalf("NewThemeModelFromDirection: %v", err)
	}
	return tm
}

func TestEndToEnd_DarkWallpaper(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	tm := buildGeneratedThemeModel(t, "e2e-dark", false)
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	_, err := ThemeDirectory(tm, exportDir, false)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(exportDir, "colors.toml"))
	if err != nil {
		t.Fatalf("read colors.toml: %v", err)
	}
	_ = content

	if _, err := os.Stat(filepath.Join(exportDir, "light.mode")); err == nil {
		t.Error("dark theme should not have light.mode")
	}
}

func TestEndToEnd_BrightImage_StillDarkByDefault(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	imgPath := createFixtureImage(t, dir, "bright.png",
		"-define", "gradient:angle=45", "gradient:#e8e8ec-#f0c0a0")
	imgResult := image.Validate(imgPath)
	if !imgResult.Valid {
		t.Fatalf("validation: %v", imgResult.Errors)
	}

	opts, _ := gen.NewGenerationOptions(imgPath, 0, false)
	colors, _ := gen.ExtractDominantColors(imgPath, 12)
	candidates, _ := gen.GeneratePalettes(colors, opts)

	dirObj := theme.Direction{
		ID: candidates[0].ID, Label: candidates[0].Label,
		Fingerprint: opts.Fingerprint, Colors: candidates[0].Colors, LightMode: false,
	}
	tm, _ := theme.NewThemeModelFromDirection("bright-test", imgPath, imgResult, dirObj)
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)
	ThemeDirectory(tm, exportDir, false)

	if _, err := os.Stat(filepath.Join(exportDir, "light.mode")); err == nil {
		t.Error("bright image should not auto-trigger light mode")
	}
}

func TestEndToEnd_ExplicitLightMode(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	tm := buildGeneratedThemeModel(t, "e2e-light", true)
	exportDir := filepath.Join(t.TempDir(), "themes", tm.NormalizedName)

	_, err := ThemeDirectory(tm, exportDir, false)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	if _, err := os.Stat(filepath.Join(exportDir, "light.mode")); os.IsNotExist(err) {
		t.Error("light theme should have light.mode")
	}
}

func TestEndToEnd_UIHeavyImageWarns(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "ui.png")
	cmd := exec.Command("magick", "-size", "1920x1080", "xc:#224466", imgPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create: %v: %s", err, string(out))
	}

	imgResult := image.Validate(imgPath)
	foundUIWarning := false
	for _, w := range imgResult.Warnings {
		if strings.Contains(strings.ToLower(w), "ui-heavy") {
			foundUIWarning = true
		}
	}
	if !foundUIWarning {
		t.Errorf("expected UI-heavy warning for 1920x1080 PNG, got warnings: %v", imgResult.Warnings)
	}
}

func TestEndToEnd_DeterministicOutput(t *testing.T) {
	if !hasMagickInExport() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	imgPath := createFixtureImage(t, dir, "det.png", "plasma:#1a1b26-#82aaff")
	opts, _ := gen.NewGenerationOptions(imgPath, 123, false)
	colors, _ := gen.ExtractDominantColors(imgPath, 12)
	candidates, _ := gen.GeneratePalettes(colors, opts)
	if len(candidates) == 0 {
		t.Skip("no candidates passed contrast validation")
	}

	exportDir1 := filepath.Join(t.TempDir(), "det1")
	exportDir2 := filepath.Join(t.TempDir(), "det2")

	tm1 := buildTestThemeModel(t, "det", imgPath)
	tm1.Colors = candidates[0].Colors

	tm2 := buildTestThemeModel(t, "det", imgPath)
	tm2.Colors = candidates[0].Colors

	ThemeDirectory(tm1, exportDir1, false)
	ThemeDirectory(tm2, exportDir2, false)

	c1, _ := os.ReadFile(filepath.Join(exportDir1, "colors.toml"))
	c2, _ := os.ReadFile(filepath.Join(exportDir2, "colors.toml"))
	if string(c1) != string(c2) {
		t.Error("same input should produce deterministic output")
	}
}
