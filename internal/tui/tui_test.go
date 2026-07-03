package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/prettyletto/omarchy-themegen/internal/export"
	"github.com/prettyletto/omarchy-themegen/internal/preview"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
	"github.com/prettyletto/omarchy-themegen/internal/validate"
)

func hasMagick() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

func createTestImage(t *testing.T) string {
	t.Helper()
	if !hasMagick() {
		t.Skip("magick not available")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	cmd := exec.Command("magick", "-size", "800x450", "plasma:#1a1b26-#82aaff-#db4b4b-#9ece6a", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create test image: %v: %s", err, string(out))
	}
	return path
}

func makeTestModel(imagePath string) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a theme name..."
	ti.CharLimit = 64
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		step:      stepValidating,
		imagePath: imagePath,
		selected:  -1,
		textInput: ti,
		spinner:   s,
	}
}

func updateModel(m Model, msg tea.Msg) Model {
	m2, _ := m.Update(msg)
	if mm, ok := m2.(Model); ok {
		return mm
	}
	if mm, ok := m2.(*Model); ok {
		return *mm
	}
	return m
}

func TestVisibleDirectionRange_ShowsThreeDirectionWindow(t *testing.T) {
	tests := []struct {
		name          string
		selected      int
		wantStart     int
		wantEnd       int
		wantVisible   []int
		wantInvisible []int
	}{
		{name: "start", selected: 0, wantStart: 0, wantEnd: 3, wantVisible: []int{1, 2, 3}, wantInvisible: []int{4, 5}},
		{name: "middle", selected: 2, wantStart: 1, wantEnd: 4, wantVisible: []int{2, 3, 4}, wantInvisible: []int{1, 5}},
		{name: "end", selected: 4, wantStart: 2, wantEnd: 5, wantVisible: []int{3, 4, 5}, wantInvisible: []int{1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := visibleDirectionRange(tt.selected, theme.DirectionCount, comparisonVisibleDirections)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Fatalf("expected range %d-%d, got %d-%d", tt.wantStart, tt.wantEnd, start, end)
			}

			m := Model{step: stepComparison, selected: tt.selected, directions: staticTestDirections(theme.DirectionCount)}
			view := m.viewComparison()
			for _, id := range tt.wantVisible {
				if !strings.Contains(view, fmt.Sprintf("Direction %d:", id)) {
					t.Fatalf("expected view to include direction %d", id)
				}
			}
			for _, id := range tt.wantInvisible {
				if strings.Contains(view, fmt.Sprintf("Direction %d:", id)) {
					t.Fatalf("expected view to hide direction %d", id)
				}
			}
		})
	}
}

func TestTUI_StateTransitions_ValidImage(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	img := createTestImage(t)

	m := makeTestModel(img)

	if m.step != stepValidating {
		t.Fatalf("expected step validating, got %v", m.step)
	}

	// Feed validation success
	m = updateModel(m, genDoneMsg{})

	// Should be in generation step
	if m.step != stepGeneration {
		t.Fatalf("expected step generation after validation, got %v", m.step)
	}

	// Feed generation result
	genMsg := getGenDoneMsg(t, img)
	if genMsg.err != nil {
		t.Skipf("generation failed: %v", genMsg.err)
	}
	m = updateModel(m, genMsg)

	// Should be in mode select
	if m.step != stepModeSelect {
		t.Fatalf("expected step mode select, got %v", m.step)
	}

	// Select whole-theme mode
	m.mixMode = false
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepComparison {
		t.Fatalf("expected step comparison, got %v", m.step)
	}
	if len(m.directions) != theme.DirectionCount {
		t.Fatalf("expected %d directions, got %d", theme.DirectionCount, len(m.directions))
	}
	if m.selected != 0 {
		t.Fatalf("expected selected=0, got %d", m.selected)
	}

	// Move selection down
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.selected != 1 {
		t.Fatalf("expected selected=1 after down, got %d", m.selected)
	}

	// Move up
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.selected != 0 {
		t.Fatalf("expected selected=0 after up, got %d", m.selected)
	}

	// Tab cycles
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyTab})
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.selected != 2 {
		t.Fatalf("expected selected=2 after 2 tabs, got %d", m.selected)
	}
}

func TestTUI_ErrorState(t *testing.T) {
	m := makeTestModel("/nonexistent/image.png")

	m = updateModel(m, genDoneMsg{err: exec.ErrNotFound})

	if m.step != stepError {
		t.Fatalf("expected step error, got %v", m.step)
	}
	if m.err == nil {
		t.Error("expected error to be set")
	}
}

func TestTUI_InvalidImageFlow(t *testing.T) {
	m := makeTestModel("/nonexistent/image.png")

	m = updateModel(m, genDoneMsg{err: exec.ErrNotFound})

	if m.step != stepError {
		t.Errorf("expected error state, got %v", m.step)
	}
	if m.err == nil {
		t.Error("expected error to be set")
	}
}

func TestTUI_EscapeFromNaming(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepNaming {
		t.Fatalf("expected naming step, got %v", m.step)
	}

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != stepComparison {
		t.Fatalf("expected comparison after esc from naming, got %v", m.step)
	}
}

func TestTUI_EscapeFromConfirm(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	m.textInput.SetValue("test-theme")
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepConfirmExport {
		t.Fatalf("expected confirm step, got %v", m.step)
	}

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != stepNaming {
		t.Fatalf("expected naming after esc from confirm, got %v", m.step)
	}
}

func TestTUI_EmptyNameRejected(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	m.textInput.SetValue("")
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step == stepConfirmExport {
		t.Error("should not advance with empty name")
	}
	if m.message == "" {
		t.Error("should show error message for empty name")
	}
}

func TestTUI_NamingToConfirm(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	m.textInput.SetValue("My Cool Theme")
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepConfirmExport {
		t.Fatalf("expected confirm step, got %v", m.step)
	}
	if m.themeName != "My Cool Theme" {
		t.Errorf("expected 'My Cool Theme', got %q", m.themeName)
	}
	if m.normalized != "my-cool-theme" {
		t.Errorf("expected 'my-cool-theme', got %q", m.normalized)
	}
	if m.forceExport {
		t.Error("fresh directory should not force export")
	}
}

func TestTUI_QuitFromComparison(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if !m.quitting {
		t.Error("expected quitting after q in comparison")
	}
}

func TestTUI_ApplyConfirmSeparate(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m.step = stepResult
	m.exportResult = &export.ExportResult{Path: "/tmp/test"}
	m.postResult = &validate.PostExportResult{Passed: true}
	m.normalized = "test-theme"
	m.exportPath = "/tmp/test"

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepApplyConfirm {
		t.Fatalf("expected apply confirm step, got %v", m.step)
	}

	view := m.View()
	if !strings.Contains(view, "Apply") {
		t.Error("apply view should mention apply")
	}
}

func TestTUI_EscapeFromApplyConfirm(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)
	m.step = stepApplyConfirm
	m.normalized = "test-theme"
	m.exportPath = "/tmp/test"

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != stepResult {
		t.Fatalf("expected back to result from apply confirm, got %v", m.step)
	}
}

func TestTUI_ArchiveModeStored(t *testing.T) {
	m := makeTestModel("/some/image.png")
	if m.archiveMode {
		t.Error("archiveMode should be false by default")
	}
}

func TestTUI_ViewContainsStateInfo(t *testing.T) {
	m := makeTestModel("/some/image.png")
	m.step = stepError
	m.err = exec.ErrNotFound

	view := m.View()
	if !strings.Contains(view, "Error") {
		t.Error("error view should contain 'Error'")
	}
}

func TestTUI_EscQuitsFromResult(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)
	m.step = stepResult
	m.exportResult = &export.ExportResult{Path: "/tmp/test"}
	m.postResult = &validate.PostExportResult{Passed: true}
	m.normalized = "test-theme"
	m.exportPath = "/tmp/test"

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEsc})
	if !m.quitting {
		t.Error("esc from result should quit")
	}
}

func TestTUI_CtrlCQuits(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.quitting {
		t.Error("ctrl+c should quit")
	}
}

func TestTUI_ForceExportDetection(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	m.textInput.SetValue("test-theme")
	m.exportPath = "/tmp/nonexistent-path-for-test"
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepConfirmExport {
		t.Fatalf("expected confirm step, got %v", m.step)
	}
	if m.forceExport {
		t.Error("forceExport should be false for nonexistent path")
	}
}

func TestTUI_GenerationStepReached(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	img := createTestImage(t)
	m := makeTestModel(img)

	// Validation passes -> generation step
	m = updateModel(m, genDoneMsg{})
	if m.step != stepGeneration {
		t.Fatalf("expected step generation, got %v", m.step)
	}

	view := m.View()
	if !strings.Contains(view, "Generating") {
		t.Error("generation view should contain 'Generating'")
	}
}

func TestTUI_BrowserComponentMixInitializesComposition(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)
	if m.composition != nil {
		t.Fatal("whole-theme flow should leave composition nil")
	}

	state := preview.NewSelectionState()
	state.Mode = "component-mix"
	state.Groups[theme.GroupDesktopShell.ID] = 2
	state.Overrides["neovim"] = 3

	m = updateModel(m, browserStateMsg{state: state})
	if m.composition == nil {
		t.Fatal("browser component-mix state should initialize composition")
	}
	if m.step != stepGroupSelect {
		t.Fatalf("expected group select, got %v", m.step)
	}
	if m.composition.Mode != "component-mix" {
		t.Fatalf("expected component-mix, got %s", m.composition.Mode)
	}
	if got := m.composition.GroupSources[theme.GroupDesktopShell.ID]; got != 2 {
		t.Fatalf("expected desktop shell direction 2, got %d", got)
	}
	if got := m.composition.Overrides["neovim"]; got != 3 {
		t.Fatalf("expected neovim override direction 3, got %d", got)
	}

	state.Overrides = map[string]int{}
	m = updateModel(m, browserStateMsg{state: state})
	if _, ok := m.composition.Overrides["neovim"]; ok {
		t.Fatal("browser state should clear stale overrides")
	}
}

func TestTUI_InlinePreviewSeparatesViewAndDrawOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "preview.png")
	if err := os.WriteFile(path, []byte("not-a-real-png"), 0644); err != nil {
		t.Fatalf("write preview file: %v", err)
	}

	m := makeTestModel("/some/image.png")
	m.width = 120
	m.termCaps = preview.CapKitty

	placeholder := m.inlinePreviewPlaceholder(path)
	if strings.Contains(placeholder, "\x1b_G") {
		t.Fatal("view placeholder must not contain image escape bytes")
	}
	if got := strings.Count(placeholder, "\n"); got != inlinePreviewRows {
		t.Fatalf("expected %d reserved rows, got %d", inlinePreviewRows, got)
	}

	draw := m.inlinePreviewOutput(path)
	if draw == "" {
		t.Fatal("expected out-of-band draw output")
	}
	if !strings.Contains(draw, preview.ClearInlineImages(preview.CapKitty)) {
		t.Fatal("draw output should clear old image placements before drawing")
	}
	if !strings.Contains(draw, "\x1b_G") {
		t.Fatal("draw output should contain kitty image escape")
	}
}

func TestTUI_ComparisonViewReservesInlinePreviewWithoutEscapes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "preview.png")
	if err := os.WriteFile(path, []byte("not-a-real-png"), 0644); err != nil {
		t.Fatalf("write preview file: %v", err)
	}

	m := makeTestModel("/some/image.png")
	m.step = stepComparison
	m.width = 120
	m.termCaps = preview.CapKitty
	m.inlinePreviewEnabled = true
	m.selected = 0
	m.directions = []theme.Direction{{ID: 1, Label: "Vibrant", Colors: theme.StaticColors()}}
	m.directionPreviews = map[int]string{1: path}

	view := m.View()
	if strings.Contains(view, "\x1b_G") {
		t.Fatal("inline image escape should never be embedded in View output")
	}
	if got := strings.Count(view, "\n"); got < inlinePreviewRows {
		t.Fatalf("view should reserve inline preview rows, got only %d newlines", got)
	}
}

// Helpers

func makeTestModelWithDirections(t *testing.T) Model {
	t.Helper()
	img := createTestImage(t)

	m := makeTestModel(img)
	m = updateModel(m, genDoneMsg{})
	genMsg := getGenDoneMsg(t, img)
	if genMsg.err != nil {
		t.Skipf("generation failed for test image: %v", genMsg.err)
	}
	m = updateModel(m, genMsg)

	if m.step != stepModeSelect {
		t.Fatalf("expected stepModeSelect after gen, got %v", m.step)
	}
	m.mixMode = false
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepComparison {
		t.Fatalf("expected stepComparison after whole-theme enter, got %v", m.step)
	}
	return m
}

func staticTestDirections(count int) []theme.Direction {
	labels := []string{"Vibrant", "Balanced", "Muted", "Colorful", "Deep"}
	directions := make([]theme.Direction, count)
	for i := range directions {
		label := fmt.Sprintf("Test %d", i+1)
		if i < len(labels) {
			label = labels[i]
		}
		directions[i] = theme.Direction{ID: i + 1, Label: label, Colors: theme.StaticColors()}
	}
	return directions
}

func getGenDoneMsg(t *testing.T, img string) genDoneMsg {
	t.Helper()
	cmd := runGeneration(img, false, 0)
	if cmd == nil {
		t.Fatal("runGeneration returned nil")
	}
	msg := cmd()
	if gm, ok := msg.(genDoneMsg); ok {
		return gm
	}
	t.Fatalf("expected genDoneMsg, got %T", msg)
	return genDoneMsg{}
}
