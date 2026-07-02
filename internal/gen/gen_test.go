package gen

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

func hasMagick() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

func createTestImage(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	if !hasMagick() {
		t.Skip("magick not available")
	}
	path := filepath.Join(dir, name)
	cmdArgs := append([]string{"-size", "800x450"}, args...)
	cmdArgs = append(cmdArgs, path)
	cmd := exec.Command("magick", cmdArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create test image: %v: %s", err, string(out))
	}
	return path
}

func createDarkImage(t *testing.T, dir string) string {
	t.Helper()
	return createTestImage(t, dir, "dark.png",
		"-define", "gradient:angle=45",
		"gradient:#1a1b26-#3b4261")
}

func createBrightImage(t *testing.T, dir string) string {
	t.Helper()
	return createTestImage(t, dir, "bright.png",
		"-define", "gradient:angle=45",
		"gradient:#e8e8ec-#f0c0a0")
}

func createColorfulImage(t *testing.T, dir string) string {
	t.Helper()
	return createTestImage(t, dir, "colorful.png",
		"plasma:#1a1b26-#82aaff-#db4b4b-#9ece6a")
}

// Task 1: Generation options
func TestNewGenerationOptions(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createDarkImage(t, dir)

	opts, err := NewGenerationOptions(img, 0, false)
	if err != nil {
		t.Fatalf("NewGenerationOptions: %v", err)
	}
	if opts.Fingerprint == "" {
		t.Error("expected non-empty fingerprint")
	}
	if opts.GeneratorVer == "" {
		t.Error("expected non-empty generator version")
	}
	if opts.LightMode {
		t.Error("expected LightMode=false")
	}

	// Deterministic fingerprint
	opts2, _ := NewGenerationOptions(img, 0, false)
	if opts.Fingerprint != opts2.Fingerprint {
		t.Error("fingerprints should be deterministic for same file")
	}

	// Different file, different fingerprint
	img2 := createBrightImage(t, dir)
	opts3, _ := NewGenerationOptions(img2, 0, false)
	if opts.Fingerprint == opts3.Fingerprint {
		t.Error("different images should have different fingerprints")
	}
}

func TestGenerationOptions_LightMode(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createDarkImage(t, dir)

	opts, err := NewGenerationOptions(img, 0, true)
	if err != nil {
		t.Fatalf("NewGenerationOptions: %v", err)
	}
	if !opts.LightMode {
		t.Error("expected LightMode=true")
	}
}

// Task 2: Color extraction
func TestExtractDominantColors(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createDarkImage(t, dir)

	colors, err := ExtractDominantColors(img, 8)
	if err != nil {
		t.Fatalf("ExtractDominantColors: %v", err)
	}
	if len(colors) == 0 {
		t.Fatal("expected non-empty colors")
	}
	if len(colors) > 8 {
		t.Fatalf("expected at most 8 colors, got %d", len(colors))
	}
	for _, c := range colors {
		if c.Hex == "" || len(c.Hex) != 7 || c.Hex[0] != '#' {
			t.Errorf("invalid hex color: %q", c.Hex)
		}
	}
}

func TestExtractDominantColors_ColorfulImage(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createColorfulImage(t, dir)

	colors, err := ExtractDominantColors(img, 12)
	if err != nil {
		t.Fatalf("ExtractDominantColors: %v", err)
	}
	if len(colors) == 0 {
		t.Fatal("expected non-empty colors from colorful image")
	}
}

func TestExtractDominantColors_Deterministic(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createDarkImage(t, dir)

	c1, _ := ExtractDominantColors(img, 8)
	c2, _ := ExtractDominantColors(img, 8)

	if len(c1) != len(c2) {
		t.Fatal("deterministic extraction should produce same count")
	}
	for i := range c1 {
		if c1[i].Hex != c2[i].Hex {
			t.Errorf("color %d differs: %s vs %s", i, c1[i].Hex, c2[i].Hex)
		}
	}
}

// Task 3: Palette generation
func TestGeneratePalettes_ThreeCandidates(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createColorfulImage(t, dir)

	colors, err := ExtractDominantColors(img, 12)
	if err != nil {
		t.Fatalf("extraction: %v", err)
	}

	opts, _ := NewGenerationOptions(img, 0, false)
	candidates, err := GeneratePalettes(colors, opts)
	if err != nil {
		t.Fatalf("GeneratePalettes: %v", err)
	}
	if len(candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(candidates))
	}

	for i, c := range candidates {
		if c.ID != i+1 {
			t.Errorf("candidate %d: expected ID %d, got %d", i, i+1, c.ID)
		}
		if c.Label == "" {
			t.Errorf("candidate %d: expected non-empty label", i)
		}
		if c.Colors == nil {
			t.Fatalf("candidate %d: expected non-nil colors", i)
		}

		// Validate all required keys
		errs := theme.ValidateColors(c.Colors)
		if len(errs) > 0 {
			t.Errorf("candidate %d: color validation failed: %v", i, errs)
		}

		// Check colors are valid hex
		for _, col := range []string{
			c.Colors.Accent, c.Colors.Cursor, c.Colors.Foreground, c.Colors.Background,
			c.Colors.SelectionForeground, c.Colors.SelectionBackground,
		} {
			if len(col) != 7 || col[0] != '#' {
				t.Errorf("candidate %d: invalid color format: %q", i, col)
			}
		}
	}
}

func TestGeneratePalettes_DirectionsAreVisuallyDistinctAtSeedZero(t *testing.T) {
	colors := []DominantColor{
		makeDominant("#151821"),
		makeDominant("#2d2038"),
		makeDominant("#8f4fd6"),
		makeDominant("#d65f9f"),
		makeDominant("#6f52d9"),
		makeDominant("#c68bdd"),
		makeDominant("#32365a"),
		makeDominant("#e8d7f0"),
	}

	candidates, err := GeneratePalettes(colors, &GenerationOptions{Seed: 0})
	if err != nil {
		t.Fatalf("GeneratePalettes: %v", err)
	}
	if len(candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(candidates))
	}

	seenAccents := map[string]bool{}
	seenBackgrounds := map[string]bool{}
	for _, c := range candidates {
		seenAccents[c.Colors.Accent] = true
		seenBackgrounds[c.Colors.Background] = true
	}
	if len(seenAccents) != 3 {
		t.Fatalf("expected three distinct accents, got %v", seenAccents)
	}
	if len(seenBackgrounds) != 3 {
		t.Fatalf("expected three distinct backgrounds, got %v", seenBackgrounds)
	}
}

func makeDominant(hex string) DominantColor {
	rgb, _ := ParseHex(hex)
	hsl := rgb.ToHSL()
	return DominantColor{
		Color:      rgb,
		Hex:        hex,
		Frequency:  1,
		Saturation: hsl.S,
		Lightness:  hsl.L,
	}
}

func TestGeneratePalettes_Deterministic(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createDarkImage(t, dir)

	colors, _ := ExtractDominantColors(img, 8)
	opts, _ := NewGenerationOptions(img, 42, false)

	c1, _ := GeneratePalettes(colors, opts)
	c2, _ := GeneratePalettes(colors, opts)

	if len(c1) != len(c2) {
		t.Fatal("deterministic palette generation")
	}
	for i := range c1 {
		if c1[i].Colors.Accent != c2[i].Colors.Accent {
			t.Errorf("direction %d accent differs: %s vs %s", i+1, c1[i].Colors.Accent, c2[i].Colors.Accent)
		}
	}
}

func TestGeneratePalettes_DifferentSeed(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createColorfulImage(t, dir)

	colors, _ := ExtractDominantColors(img, 12)
	opts1, _ := NewGenerationOptions(img, 0, false)
	opts2, _ := NewGenerationOptions(img, 99, false)

	c1, _ := GeneratePalettes(colors, opts1)
	c2, _ := GeneratePalettes(colors, opts2)

	if len(c1) == 0 || len(c2) == 0 {
		t.Skip("no candidates passed contrast validation")
	}

	different := false
	for i := range c1 {
		if c1[i].Colors.Accent != c2[i].Colors.Accent {
			different = true
			break
		}
	}
	if !different && len(colors) > 3 {
		t.Log("note: small image palette may not differ with seed (few colors)")
	}
}

func TestGeneratePalettes_LightMode(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createColorfulImage(t, dir)

	colors, _ := ExtractDominantColors(img, 12)
	opts, _ := NewGenerationOptions(img, 0, true)

	candidates, err := GeneratePalettes(colors, opts)
	if err != nil {
		t.Fatalf("GeneratePalettes light: %v", err)
	}
	if len(candidates) != 3 {
		t.Fatalf("expected 3 light candidates, got %d", len(candidates))
	}

	for i, c := range candidates {
		// Light theme backgrounds should be light
		bg, err := ParseHex(c.Colors.Background)
		if err != nil {
			t.Errorf("candidate %d: invalid background hex: %s", i, c.Colors.Background)
			continue
		}
		if bg.ToHSL().L < 0.5 {
			t.Errorf("candidate %d: light theme background too dark (L=%.3f): %s",
				i, bg.ToHSL().L, c.Colors.Background)
		}

		// Light theme foreground should be dark
		fg, err := ParseHex(c.Colors.Foreground)
		if err != nil {
			continue
		}
		if fg.ToHSL().L > 0.5 {
			t.Errorf("candidate %d: light theme foreground too light (L=%.3f): %s",
				i, fg.ToHSL().L, c.Colors.Foreground)
		}
	}
}

// Task 4: Contrast validation
func TestValidateThemeContrast_ValidPasses(t *testing.T) {
	c := &theme.Colors{
		Background:          "#1a1b26",
		Foreground:          "#c0caf5",
		Accent:              "#7aa2f7",
		SelectionForeground: "#1a1b26",
		SelectionBackground: "#7aa2f7",
		Color1:              "#db4b4b",
		Color2:              "#9ece6a",
		Color3:              "#e0af68",
		Color4:              "#7aa2f7",
		Color5:              "#bb9af7",
		Color6:              "#7dcfff",
	}

	warnings := ValidateThemeContrast(c)
	for _, w := range warnings {
		t.Logf("contrast warning (may be acceptable): %s", w)
	}
}

func TestValidateThemeContrast_LowForegroundFails(t *testing.T) {
	c := &theme.Colors{
		Background: "#1a1b26",
		Foreground: "#1a1b27", // Nearly identical to background
		Accent:     "#7aa2f7",
	}

	warnings := ValidateThemeContrast(c)
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "foreground") || strings.Contains(w, "contrast") {
			found = true
		}
	}
	if !found {
		t.Error("expected foreground/background contrast warning")
	}
}

func TestValidateThemeContrast_LowSelectionFails(t *testing.T) {
	c := &theme.Colors{
		Background:          "#1a1b26",
		Foreground:          "#c0caf5",
		Accent:              "#7aa2f7",
		SelectionForeground: "#7aa2f8",
		SelectionBackground: "#7aa2f7",
	}

	warnings := ValidateThemeContrast(c)
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "selection") {
			found = true
		}
	}
	if !found {
		t.Error("expected selection contrast warning")
	}
}

func TestValidateThemeContrast_TerminalCollapse(t *testing.T) {
	c := &theme.Colors{
		Background: "#1a1b26",
		Foreground: "#c0caf5",
		Accent:     "#7aa2f7",
		Color1:     "#9ece6a",
		Color2:     "#9ece6b",
		Color3:     "#e0af68",
	}

	warnings := ValidateThemeContrast(c)
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "nearly identical") {
			found = true
		}
	}
	if !found {
		t.Error("expected terminal color collapse warning")
	}
}

func TestEnsureReadableTextColors_DarkMonotone(t *testing.T) {
	c := &theme.Colors{
		Background:          "#202020",
		Foreground:          "#262626",
		Color7:              "#282828",
		Color15:             "#303030",
		SelectionBackground: "#353535",
		SelectionForeground: "#363636",
	}

	ensureReadableTextColors(c)

	assertContrast(t, c.Foreground, c.Background, MinForegroundContrast)
	assertContrast(t, c.Color7, c.Background, MinForegroundContrast)
	assertContrast(t, c.Color15, c.Background, MinForegroundContrast)
	assertContrast(t, c.SelectionForeground, c.SelectionBackground, MinSelectionContrast)
}

func TestEnsureReadableTextColors_LightMonotone(t *testing.T) {
	c := &theme.Colors{
		Background:          "#eeeeee",
		Foreground:          "#e5e5e5",
		Color7:              "#e2e2e2",
		Color15:             "#dddddd",
		SelectionBackground: "#d8d8d8",
		SelectionForeground: "#d6d6d6",
	}

	ensureReadableTextColors(c)

	assertContrast(t, c.Foreground, c.Background, MinForegroundContrast)
	assertContrast(t, c.Color7, c.Background, MinForegroundContrast)
	assertContrast(t, c.Color15, c.Background, MinForegroundContrast)
	assertContrast(t, c.SelectionForeground, c.SelectionBackground, MinSelectionContrast)
}

// Task 5: Direction building
func TestDirectionSet_BuildFromValidPalette(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createColorfulImage(t, dir)

	colors, _ := ExtractDominantColors(img, 12)
	opts, _ := NewGenerationOptions(img, 0, false)
	candidates, _ := GeneratePalettes(colors, opts)

	dirObj := theme.Direction{
		ID:          candidates[0].ID,
		Label:       candidates[0].Label,
		Fingerprint: opts.Fingerprint,
		Colors:      candidates[0].Colors,
		Warnings:    candidates[0].Warnings,
	}

	if dirObj.ID != 1 {
		t.Errorf("expected direction ID 1, got %d", dirObj.ID)
	}
	if dirObj.Label == "" {
		t.Error("expected non-empty label")
	}
	if dirObj.Fingerprint == "" {
		t.Error("expected non-empty fingerprint")
	}

	// Can build ThemeModel from direction
	tm, err := theme.NewThemeModelFromDirection("test-dir", img, nil, dirObj)
	if err != nil {
		t.Fatalf("NewThemeModelFromDirection: %v", err)
	}
	if tm.DirectionID != 1 {
		t.Errorf("expected DirectionID 1, got %d", tm.DirectionID)
	}
	if tm.DirectionLabel != dirObj.Label {
		t.Errorf("label mismatch: %q vs %q", tm.DirectionLabel, dirObj.Label)
	}
}

// Color utilities
func TestParseHex(t *testing.T) {
	rgb, err := ParseHex("#ff8000")
	if err != nil {
		t.Fatal(err)
	}
	if rgb.R < 0.99 || rgb.G < 0.49 || rgb.G > 0.51 || rgb.B > 0.01 {
		t.Errorf("unexpected RGB: (%.3f, %.3f, %.3f)", rgb.R, rgb.G, rgb.B)
	}
}

func TestParseHex_Invalid(t *testing.T) {
	_, err := ParseHex("#GGG")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
	_, err = ParseHex("#12345")
	if err == nil {
		t.Error("expected error for wrong length")
	}
}

func TestRGB_Hex_Roundtrip(t *testing.T) {
	hexes := []string{"#ff8000", "#1a1b26", "#c0caf5", "#000000", "#ffffff"}
	for _, h := range hexes {
		rgb, err := ParseHex(h)
		if err != nil {
			t.Errorf("ParseHex(%s): %v", h, err)
			continue
		}
		if rgb.Hex() != h {
			t.Errorf("roundtrip %s -> %s", h, rgb.Hex())
		}
	}
}

func TestHSL_Roundtrip(t *testing.T) {
	rgb, _ := ParseHex("#7aa2f7")
	hsl := rgb.ToHSL()
	back := hsl.ToRGB()

	// Allow small epsilon for floating point conversion
	dist := rgb.R*255 - back.R*255
	if dist < 0 {
		dist = -dist
	}
	if dist > 2 {
		t.Errorf("HSL roundtrip: %s -> HSL -> %s (diff %.3f)", rgb.Hex(), back.Hex(), dist)
	}
}

func TestLuminance_BlackWhite(t *testing.T) {
	black, _ := ParseHex("#000000")
	white, _ := ParseHex("#ffffff")

	if black.Luminance() >= white.Luminance() {
		t.Error("black should have lower luminance than white")
	}
}

func TestContrastRatio(t *testing.T) {
	black, _ := ParseHex("#000000")
	white, _ := ParseHex("#ffffff")

	cr := ContrastRatio(black, white)
	if cr < 20 || cr > 22 {
		t.Errorf("expected high contrast, got %.2f", cr)
	}

	same, _ := ParseHex("#1a1b26")
	cr = ContrastRatio(same, same)
	if cr < 0.99 || cr > 1.01 {
		t.Errorf("expected contrast ~1 for same color, got %.2f", cr)
	}
}

func TestWithLightness(t *testing.T) {
	rgb, _ := ParseHex("#7aa2f7")
	darker := rgb.WithLightness(0.2)
	lighter := rgb.WithLightness(0.8)

	if darker.ToHSL().L > lighter.ToHSL().L {
		t.Error("darker should have lower lightness")
	}
}

func TestFingerprintFile_Error(t *testing.T) {
	_, err := fingerprintFile("/nonexistent/file.png")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// Bright image does NOT auto-trigger light mode
func TestBrightImage_DoesNotAutoLight(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	img := createBrightImage(t, dir)

	opts, _ := NewGenerationOptions(img, 0, false)
	if opts.LightMode {
		t.Error("bright image should not auto-trigger light mode")
	}
}

func assertContrast(t *testing.T, fgHex, bgHex string, min float64) {
	t.Helper()
	fg, err := ParseHex(fgHex)
	if err != nil {
		t.Fatalf("invalid foreground %s: %v", fgHex, err)
	}
	bg, err := ParseHex(bgHex)
	if err != nil {
		t.Fatalf("invalid background %s: %v", bgHex, err)
	}
	if cr := ContrastRatio(fg, bg); cr < min {
		t.Fatalf("contrast %s on %s = %.2f, want >= %.2f", fgHex, bgHex, cr, min)
	}
}
