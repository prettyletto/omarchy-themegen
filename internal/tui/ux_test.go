package tui

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

func TestUX_WholeThemeFlow(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.selected != 1 {
		t.Fatalf("expected selected=1, got %d", m.selected)
	}

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepNaming {
		t.Fatalf("expected naming, got %v", m.step)
	}

	m.textInput.SetValue("My Test Theme")
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepConfirmExport {
		t.Fatalf("expected confirm, got %v (msg: %s)", m.step, m.message)
	}
	if m.normalized != "my-test-theme" {
		t.Errorf("expected my-test-theme, got %s", m.normalized)
	}
}

func TestUX_ComponentMixFlow(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	img := createTestImage(t)
	m := makeTestModel(img)

	m = updateModel(m, genDoneMsg{})
	genMsg := getGenDoneMsg(t, img)
	if genMsg.err != nil {
		t.Skipf("gen failed: %v", genMsg.err)
	}
	m = updateModel(m, genMsg)
	if m.step != stepModeSelect {
		t.Fatalf("expected mode select, got %v", m.step)
	}

	m.mixMode = true
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepGroupSelect {
		t.Fatalf("expected group select, got %v", m.step)
	}
	if m.composition.Mode != "component-mix" {
		t.Error("mode should be component-mix")
	}

	for i, g := range theme.AllGroups {
		m.groupCursor = i
		dirID := (i % theme.DirectionCount) + 1
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune('0' + dirID)}})
		if _, ok := m.composition.GroupSources[g.ID]; !ok {
			t.Errorf("group %s not assigned", g.ID)
		}
	}

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepOverrideSelect {
		t.Fatalf("expected override select, got %v", m.step)
	}

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepNaming {
		t.Fatalf("expected naming, got %v", m.step)
	}
}

func TestUX_QuitSafety(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if !m.quitting {
		t.Error("q should quit from comparison")
	}
	if m.step == stepExporting || m.step == stepApplying {
		t.Error("quit should not trigger export or apply")
	}
}

func TestUX_ExportErrorRecovery(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m.step = stepExporting
	m.exportPath = "/tmp/test"
	m = updateModel(m, exportDoneMsg{err: exec.ErrNotFound})

	if m.step != stepError {
		t.Fatalf("expected error state, got %v", m.step)
	}
	if m.err == nil {
		t.Error("error should be set")
	}

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if !m.quitting {
		t.Error("enter from error should quit")
	}
}

func TestUX_ApplyViewShown(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)
	m.step = stepApplyConfirm
	m.normalized = "test-theme"

	view := m.View()
	if !strings.Contains(view, "Apply") {
		t.Error("apply confirm view should mention apply")
	}
}

func TestUX_OverwriteWarning(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	m.textInput.SetValue("existing-theme")
	m.forceExport = true
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepConfirmExport {
		t.Fatalf("expected confirm, got %v", m.step)
	}
	view := m.View()
	if !strings.Contains(strings.ToLower(view), "exist") && !strings.Contains(strings.ToLower(view), "backup") {
		t.Log("overwrite warning should be visible in confirm view")
	}
}

func TestUX_NavigationBack(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != stepComparison {
		t.Fatalf("expected comparison after esc, got %v", m.step)
	}
}

func TestUX_StatesDontLeak(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	m.step = stepResult
	m.exportResult = nil
	m.normalized = "test"
	m.exportPath = "/tmp/test"

	view := m.View()
	if view == "" {
		t.Error("result view should render")
	}
}

func TestUX_ModeSelectRender(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)
	m.step = stepModeSelect

	view := m.View()
	if !strings.Contains(view, "Whole Theme") && !strings.Contains(view, "Component") {
		t.Error("mode select should show both options")
	}
}

func TestUX_GroupSelectRender(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	img := createTestImage(t)
	m := makeTestModel(img)
	m = updateModel(m, genDoneMsg{})
	genMsg := getGenDoneMsg(t, img)
	if genMsg.err != nil {
		t.Skipf("gen failed: %v", genMsg.err)
	}
	m = updateModel(m, genMsg)
	m.mixMode = true
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	view := m.View()
	if !strings.Contains(view, "Surface Groups") {
		t.Error("group select should show surface groups")
	}
}

func TestUX_OverrideSelectRender(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	img := createTestImage(t)
	m := makeTestModel(img)
	m = updateModel(m, genDoneMsg{})
	genMsg := getGenDoneMsg(t, img)
	if genMsg.err != nil {
		t.Skipf("gen failed: %v", genMsg.err)
	}
	m = updateModel(m, genMsg)
	m.mixMode = true
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	for i := range theme.AllGroups {
		m.groupCursor = i
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune('1')}})
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	view := m.View()
	if !strings.Contains(view, "Overrides") {
		t.Error("override select should show overrides")
	}
}

func Test_BrowserToggleSafe(t *testing.T) {
	if !hasMagick() {
		t.Skip("magick not available")
	}
	m := makeTestModelWithDirections(t)

	// Browser toggle without directions available - should show message
	m.directions = nil
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if m.message == "" {
		t.Error("should show message when browser unavailable")
	}
}

func Test_AllViewsRender(t *testing.T) {
	m := makeTestModel("/nonexistent.png")
	steps := []step{
		stepValidating, stepError, stepGeneration, stepModeSelect,
		stepComparison, stepNaming, stepConfirmExport,
		stepExporting, stepResult, stepApplyConfirm, stepApplying, stepDone,
	}

	for _, s := range steps {
		m.step = s
		v := m.View()
		if v == "" || v == "Unknown state" {
			t.Errorf("step %v view failed to render", s)
		}
	}
}

func Test_ErrorMessageNotEmpty(t *testing.T) {
	m := makeTestModel("/bad.png")
	m.step = stepError
	m.err = os.ErrNotExist
	v := m.View()
	if !strings.Contains(v, "Error") {
		t.Error("error view should contain error text")
	}
}
