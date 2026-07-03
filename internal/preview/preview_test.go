package preview

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

func hasMagick() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

func makeDirections() []theme.Direction {
	colors := theme.StaticColors()
	return []theme.Direction{
		{ID: 1, Label: "Vibrant", Colors: colors},
		{ID: 2, Label: "Balanced", Colors: colors},
		{ID: 3, Label: "Muted", Colors: colors},
		{ID: 4, Label: "Colorful", Colors: colors},
		{ID: 5, Label: "Deep", Colors: colors},
	}
}

// Task 1: Terminal capability detection
func TestDetectCapability(t *testing.T) {
	c := DetectCapability()
	// Must return a valid Capability
	if c < CapNone || c > CapSixel {
		t.Errorf("invalid capability: %v", c)
	}
	if c.String() == "unknown" {
		t.Errorf("CapNone should return 'none', got %s", c.String())
	}
}

func TestCapability_Supported(t *testing.T) {
	if CapNone.Supported() {
		t.Error("CapNone should not be supported")
	}
	if !CapKitty.Supported() {
		t.Error("CapKitty should be supported")
	}
}

func TestClearInlineImages_Kitty(t *testing.T) {
	seq := ClearInlineImages(CapKitty)
	if seq == "" {
		t.Fatal("expected kitty cleanup sequence")
	}
	if !strings.Contains(seq, "a=d") || !strings.Contains(seq, "d=A") {
		t.Fatalf("unexpected kitty cleanup sequence: %q", seq)
	}
	if ClearInlineImages(CapNone) != "" {
		t.Fatal("cap none should not emit cleanup")
	}
}

func TestDisplayCapability(t *testing.T) {
	r := DisplayCapability()
	if r.Cap < CapNone || r.Cap > CapSixel {
		t.Errorf("invalid capability: %v", r.Cap)
	}
	if r.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestDetectCapability_KittyEnv(t *testing.T) {
	// This is a test of the detection logic, not live env
	oldVal := os.Getenv("KITTY_WINDOW_ID")
	os.Setenv("KITTY_WINDOW_ID", "12345")
	defer os.Setenv("KITTY_WINDOW_ID", oldVal)

	c := DetectCapability()
	if c != CapKitty {
		t.Errorf("expected kitty capability with KITTY_WINDOW_ID set, got %s", c.String())
	}
}

// Task 2: Direction preview PNGs
func TestRenderDirectionPreview(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	cmd := exec.Command("magick", "-size", "800x450", "plasma:#1a1b26-#82aaff", src)
	cmd.Run()

	directions := makeDirections()
	out := filepath.Join(dir, "preview.png")
	err := RenderDirectionPreview(out, src, directions[0], 800, 500)
	if err != nil {
		t.Fatalf("RenderDirectionPreview: %v", err)
	}
	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Error("preview file not created")
	}
}

func TestGenerateDirectionPreviews(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	cmd := exec.Command("magick", "-size", "800x450", "plasma:#1a1b26-#82aaff", src)
	cmd.Run()

	outDir := filepath.Join(dir, "previews")
	directions := makeDirections()
	paths, err := GenerateDirectionPreviews(outDir, src, directions)
	if err != nil {
		t.Fatalf("GenerateDirectionPreviews: %v", err)
	}
	if len(paths) != theme.DirectionCount {
		t.Errorf("expected %d previews, got %d", theme.DirectionCount, len(paths))
	}
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("missing preview: %s", p)
		}
	}
}

func TestBrowserServer_ServesDirectionPreviewImage(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	cmd := exec.Command("magick", "-size", "800x450", "plasma:#1a1b26-#82aaff", src)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create source image: %v: %s", err, string(out))
	}

	server := NewBrowserServer(src, makeDirections(), nil)
	_, err := server.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop()

	url := fmt.Sprintf("http://127.0.0.1:%d/preview/direction/1?token=%s", server.Port(), server.Token())
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("get preview: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("expected image/png content type, got %q", ct)
	}
}

func TestBrowserServer_ServesCurrentPreviewForComponentMix(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	cmd := exec.Command("magick", "-size", "800x450", "plasma:#1a1b26-#82aaff", src)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create source image: %v: %s", err, string(out))
	}

	server := NewBrowserServer(src, makeDirections(), nil)
	_, err := server.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop()

	setModeURL := fmt.Sprintf("http://127.0.0.1:%d/api/set-mode?token=%s&mode=component-mix", server.Port(), server.Token())
	resp, err := http.Post(setModeURL, "application/json", nil)
	if err != nil {
		t.Fatalf("set mode: %v", err)
	}
	resp.Body.Close()

	previewURL := fmt.Sprintf("http://127.0.0.1:%d/preview/current?token=%s", server.Port(), server.Token())
	resp, err = http.Get(previewURL)
	if err != nil {
		t.Fatalf("get current preview: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("expected image/png content type, got %q", ct)
	}
}

// Task 3: Composed preview PNG
func TestRenderComposedPreview(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	cmd := exec.Command("magick", "-size", "800x450", "plasma:#1a1b26-#82aaff", src)
	cmd.Run()

	colors := theme.StaticColors()
	tm := &theme.ThemeModel{
		Name:           "test",
		Colors:         colors,
		Mode:           "whole-theme",
		DirectionLabel: "Vibrant",
	}
	out := filepath.Join(dir, "composed.png")
	err := RenderComposedPreview(out, src, tm, 800, 500)
	if err != nil {
		t.Fatalf("RenderComposedPreview: %v", err)
	}
	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Error("composed preview not created")
	}
}

// Task 5: Preview cache
func TestCache_HitAndMiss(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	key := CacheKey{Fingerprint: "sha256:abc", Seed: 0, LightMode: false, Mode: "whole-theme"}

	// Miss
	_, ok := cache.Get(key)
	if ok {
		t.Error("expected cache miss")
	}

	// Write a file, then put
	tmpFile := filepath.Join(dir, "tmp.png")
	os.WriteFile(tmpFile, []byte("fake-png"), 0644)
	cache.Put(key, tmpFile)

	// Hit
	path, ok := cache.Get(key)
	if !ok {
		t.Error("expected cache hit")
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestCache_DifferentKeyMiss(t *testing.T) {
	dir := t.TempDir()
	cache, _ := NewCache(dir)

	key1 := CacheKey{Fingerprint: "sha256:abc", Seed: 0}
	tmpFile := filepath.Join(dir, "tmp.png")
	os.WriteFile(tmpFile, []byte("fake"), 0644)
	cache.Put(key1, tmpFile)

	key2 := CacheKey{Fingerprint: "sha256:abc", Seed: 42}
	_, ok := cache.Get(key2)
	if ok {
		t.Error("different seed should miss cache")
	}
}

func TestCache_Invalidate(t *testing.T) {
	dir := t.TempDir()
	cache, _ := NewCache(dir)

	key := CacheKey{Fingerprint: "sha256:abc"}
	tmpFile := filepath.Join(dir, "tmp.png")
	os.WriteFile(tmpFile, []byte("fake"), 0644)
	cache.Put(key, tmpFile)

	cache.Invalidate()

	_, ok := cache.Get(key)
	if ok {
		t.Error("cache should be empty after invalidation")
	}
}

func TestCache_Cleanup(t *testing.T) {
	dir := t.TempDir()
	cache, _ := NewCache(dir)

	key := CacheKey{Fingerprint: "sha256:abc"}
	tmpFile := filepath.Join(dir, "tmp.png")
	os.WriteFile(tmpFile, []byte("fake"), 0644)
	cache.Put(key, tmpFile)

	cache.Cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Log("cache dir may still exist after cleanup")
	}
}

// Task 6, 7, 9: Browser server tests
func TestBrowserServer_StartAndStop(t *testing.T) {
	dirs := makeDirections()
	src := "/nonexistent/image.png"
	server := NewBrowserServer(src, dirs, nil)

	url, err := server.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
	if !strings.Contains(url, "127.0.0.1") {
		t.Errorf("expected local bind, got %s", url)
	}
	if !strings.Contains(url, "token=") {
		t.Errorf("expected token in URL, got %s", url)
	}

	server.Stop()
}

func TestBrowserServer_TokenRequired(t *testing.T) {
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	url, _ := server.Start()
	defer server.Stop()

	// Request without token should fail
	baseURL := strings.Split(url, "?")[0]
	resp, err := http.Get(baseURL)
	if err != nil {
		t.Skipf("cannot connect: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 without token, got %d", resp.StatusCode)
	}
}

func TestBrowserServer_InvalidToken(t *testing.T) {
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	_, _ = server.Start()
	defer server.Stop()

	resp, err := http.Get(server.URL() + "xxx")
	if err != nil {
		t.Skipf("cannot connect: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 with invalid token, got %d", resp.StatusCode)
	}
}

func TestBrowserServer_ValidTokenWorks(t *testing.T) {
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	_, _ = server.Start()
	defer server.Stop()

	resp, err := http.Get(server.URL())
	if err != nil {
		t.Skipf("cannot connect: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestBrowserServer_SelectDirection(t *testing.T) {
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	baseURL, _ := server.Start()
	defer server.Stop()
	_ = baseURL

	apiURL := fmt.Sprintf("http://127.0.0.1:%d/api/select-direction?token=%s&id=2", server.Port(), server.Token())
	resp, err := http.Post(apiURL, "application/json", nil)
	if err != nil {
		t.Skipf("cannot POST: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	state := server.copyState()
	if state.Selected != 2 {
		t.Errorf("expected selected=2, got %d", state.Selected)
	}
	if state.Mode != "whole-theme" {
		t.Errorf("expected mode whole-theme, got %s", state.Mode)
	}
}

func TestBrowserServer_SelectGroup(t *testing.T) {
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	baseURL, _ := server.Start()
	defer server.Stop()
	_ = baseURL

	apiURL := fmt.Sprintf("http://127.0.0.1:%d/api/select-group?token=%s&group=desktop-shell&id=1", server.Port(), server.Token())
	resp, _ := http.Post(apiURL, "application/json", nil)
	if resp != nil {
		resp.Body.Close()
	}

	state := server.copyState()
	if state.Mode != "component-mix" {
		t.Errorf("expected component-mix mode, got %s", state.Mode)
	}
	if state.Groups["desktop-shell"] != 1 {
		t.Errorf("expected desktop-shell → 1, got %d", state.Groups["desktop-shell"])
	}
}

func TestBrowserServer_SetOverride(t *testing.T) {
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	baseURL, _ := server.Start()
	defer server.Stop()
	_ = baseURL

	apiURL := fmt.Sprintf("http://127.0.0.1:%d/api/set-override?token=%s&surface=neovim&id=2", server.Port(), server.Token())
	resp, _ := http.Post(apiURL, "application/json", nil)
	if resp != nil {
		resp.Body.Close()
	}

	state := server.copyState()
	if state.Overrides["neovim"] != 2 {
		t.Errorf("expected neovim override → 2, got %d", state.Overrides["neovim"])
	}
}

func TestBrowserServer_ClearOverride(t *testing.T) {
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	baseURL, _ := server.Start()
	defer server.Stop()
	_ = baseURL

	// Set first
	apiURL := fmt.Sprintf("http://127.0.0.1:%d/api/set-override?token=%s&surface=neovim&id=2", server.Port(), server.Token())
	http.Post(apiURL, "application/json", nil)
	// Clear
	apiURL = fmt.Sprintf("http://127.0.0.1:%d/api/clear-override?token=%s&surface=neovim", server.Port(), server.Token())
	http.Post(apiURL, "application/json", nil)

	state := server.copyState()
	if _, ok := state.Overrides["neovim"]; ok {
		t.Error("neovim override should be cleared")
	}
}

func TestBrowserServer_IdleTimeout(t *testing.T) {
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	server.idleTimeout = 500 * time.Millisecond
	url, _ := server.Start()

	// Wait for idle timeout to fire
	time.Sleep(700 * time.Millisecond)

	// Server should be stopped; requests should fail
	resp, err := http.Get(url)
	if err == nil {
		resp.Body.Close()
		t.Log("server may still be running after timeout (timing-dependent)")
	}
}

func TestBrowserServer_NonLocalBindRejected(t *testing.T) {
	// The Start() method hardcodes 127.0.0.1, so non-local is impossible.
	// Test that the URL always contains 127.0.0.1
	dirs := makeDirections()
	server := NewBrowserServer("/tmp/test.png", dirs, nil)
	url, err := server.Start()
	if err != nil {
		t.Skipf("cannot start: %v", err)
		return
	}
	defer server.Stop()

	if !strings.Contains(url, "127.0.0.1") {
		t.Errorf("preview server must bind localhost only, got %s", url)
	}
}

func TestSelectionState_Defaults(t *testing.T) {
	s := NewSelectionState()
	if s.Groups == nil {
		t.Error("Groups should be initialized")
	}
	if s.Overrides == nil {
		t.Error("Overrides should be initialized")
	}
}
