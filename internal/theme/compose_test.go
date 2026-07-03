package theme

import (
	"testing"
)

func makeTestDirections() []Direction {
	colors := StaticColors()
	return []Direction{
		{ID: 1, Label: "Vibrant", Fingerprint: "sha256:aaa", Colors: colors, LightMode: false},
		{ID: 2, Label: "Balanced", Fingerprint: "sha256:aaa", Colors: colors, LightMode: false},
		{ID: 3, Label: "Muted", Fingerprint: "sha256:aaa", Colors: colors, LightMode: false},
	}
}

func TestComposition_WholeTheme(t *testing.T) {
	dirs := makeTestDirections()
	c := NewComposition("whole-theme")
	c.Directions = dirs

	tm, err := c.Resolve("test", "/img.png", nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if tm.Mode != "whole-theme" {
		t.Errorf("expected whole-theme, got %s", tm.Mode)
	}
	if tm.DirectionID != 1 {
		t.Errorf("expected direction 1, got %d", tm.DirectionID)
	}
	if len(tm.GroupSelections) != 4 {
		t.Errorf("expected 4 group selections, got %d", len(tm.GroupSelections))
	}
}

func TestComposition_ComponentMix(t *testing.T) {
	dirs := makeTestDirections()
	c := NewComposition("component-mix")
	c.Directions = dirs

	// Set groups
	c.SetGroupSource(GroupDesktopShell.ID, 1)
	c.SetGroupSource(GroupTerminalsAndTUI.ID, 2)
	c.SetGroupSource(GroupEditor.ID, 3)
	c.SetGroupSource(GroupAssetsAndSystem.ID, 1)

	tm, err := c.Resolve("mixed-test", "/img.png", nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if tm.Mode != "component-mix" {
		t.Errorf("expected component-mix, got %s", tm.Mode)
	}
	if len(tm.GroupSelections) != 4 {
		t.Errorf("expected 4 group selections, got %d", len(tm.GroupSelections))
	}
	if tm.GroupSelections[GroupDesktopShell.ID] != 1 {
		t.Errorf("desktop shell expected dir 1, got %d", tm.GroupSelections[GroupDesktopShell.ID])
	}
	if tm.GroupSelections["editor"] != 3 {
		t.Errorf("editor expected dir 3, got %d", tm.GroupSelections["editor"])
	}
}

func TestComposition_ComponentMix_Overrides(t *testing.T) {
	dirs := makeTestDirections()
	c := NewComposition("component-mix")
	c.Directions = dirs

	c.SetGroupSource(GroupDesktopShell.ID, 1)
	c.SetGroupSource(GroupTerminalsAndTUI.ID, 2)
	c.SetGroupSource(GroupEditor.ID, 3)
	c.SetGroupSource(GroupAssetsAndSystem.ID, 1)

	// Override neovim
	if err := c.SetOverride("neovim", 1); err != nil {
		t.Fatalf("SetOverride: %v", err)
	}

	tm, err := c.Resolve("mixed-override", "/img.png", nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if len(tm.Overrides) == 0 {
		t.Error("expected overrides in resolved model")
	}
	if tm.Overrides["neovim"] != 1 {
		t.Errorf("expected neovim override to 1, got %d", tm.Overrides["neovim"])
	}
}

func TestComposition_Validation_MissingGroups(t *testing.T) {
	c := NewComposition("component-mix")
	// No groups set - should fail validation
	errs := c.Validate()
	if len(errs) != 4 {
		t.Errorf("expected 4 missing group errors, got %d", len(errs))
	}
}

func TestComposition_Validation_InvalidDirection(t *testing.T) {
	c := NewComposition("component-mix")
	invalidDirID := DirectionCount + 1
	err := c.SetGroupSource(GroupDesktopShell.ID, invalidDirID)
	if err == nil {
		t.Errorf("expected error for invalid direction %d", invalidDirID)
	}
}

func TestComposition_Validation_UnknownGroup(t *testing.T) {
	c := NewComposition("component-mix")
	err := c.SetGroupSource("unknown-group", 1)
	if err == nil {
		t.Error("expected error for unknown group")
	}
}

func TestComposition_Validation_UnknownSurface(t *testing.T) {
	c := NewComposition("component-mix")
	err := c.SetOverride("spotify", 1)
	if err == nil {
		t.Error("expected error for unknown surface (spotify)")
	}
}

func TestComposition_Validation_VSCodeBlocked(t *testing.T) {
	if ValidSurface("vscode") {
		t.Error("VS Code should not be a valid surface")
	}
}

func TestComposition_ClearOverride(t *testing.T) {
	c := NewComposition("component-mix")
	c.SetOverride("neovim", 2)
	if len(c.Overrides) != 1 {
		t.Error("expected 1 override")
	}
	c.ClearOverride("neovim")
	if len(c.Overrides) != 0 {
		t.Error("expected 0 overrides after clear")
	}
	c.ClearOverride("nonexistent") // should not panic
}

func TestComposition_GroupByID(t *testing.T) {
	g, ok := GroupByID("desktop-shell")
	if !ok || g.ID != GroupDesktopShell.ID {
		t.Errorf("expected desktop-shell group, got %v", ok)
	}

	g, ok = GroupByID("Desktop Shell")
	if !ok || g.ID != GroupDesktopShell.ID {
		t.Errorf("expected desktop-shell group by label")
	}

	_, ok = GroupByID("nonexistent")
	if ok {
		t.Error("expected not found for nonexistent group")
	}
}

func TestComposition_ValidSurface(t *testing.T) {
	if !ValidSurface("waybar") {
		t.Error("waybar should be valid")
	}
	if !ValidSurface("neovim") {
		t.Error("neovim should be valid")
	}
	if ValidSurface("vscode") {
		t.Error("vscode should not be valid")
	}
	if ValidSurface("firefox") {
		t.Error("firefox should not be valid")
	}
}

func TestComposition_SurfacesBelongToOneGroup(t *testing.T) {
	seen := map[string]bool{}
	for _, g := range AllGroups {
		for _, s := range g.Surfaces {
			if seen[s] {
				t.Errorf("surface %s appears in multiple groups", s)
			}
			seen[s] = true
		}
	}
}

func TestNewComposition_NilMaps(t *testing.T) {
	c := NewComposition("whole-theme")
	if c.GroupSources == nil {
		t.Error("GroupSources should be initialized")
	}
	if c.Overrides == nil {
		t.Error("Overrides should be initialized")
	}
}

func TestComposition_ResolveWithoutDirections(t *testing.T) {
	c := NewComposition("whole-theme")
	_, err := c.Resolve("test", "/img.png", nil)
	if err == nil {
		t.Error("expected error when no directions present")
	}
}

// Sprint 12: Real composition mixing
func TestComposition_RealMixing_TerminalsGroup(t *testing.T) {
	dirs := makeRealDirections()
	c := NewComposition("component-mix")
	c.Directions = dirs

	// All groups → same direction should produce identical colors
	c.SetGroupSource(GroupAssetsAndSystem.ID, 1)
	c.SetGroupSource(GroupTerminalsAndTUI.ID, 1)
	c.SetGroupSource(GroupDesktopShell.ID, 1)
	c.SetGroupSource(GroupEditor.ID, 1)

	tm, _ := c.Resolve("real", "/img.png", nil)
	if tm.Colors.Color0 != dirs[0].Colors.Color0 {
		t.Error("all-group mix should match master terminal palette")
	}

	// Terminals → different direction should mix terminal palette
	c2 := NewComposition("component-mix")
	c2.Directions = dirs
	c2.SetGroupSource(GroupAssetsAndSystem.ID, 1)
	c2.SetGroupSource(GroupTerminalsAndTUI.ID, 2) // Different terminals
	c2.SetGroupSource(GroupDesktopShell.ID, 1)
	c2.SetGroupSource(GroupEditor.ID, 1)

	tm2, _ := c2.Resolve("real2", "/img.png", nil)
	if tm2.Colors.Color0 != dirs[1].Colors.Color0 {
		t.Error("different terminals group should use its own terminal colors")
	}
	if tm2.Colors.Background != dirs[0].Colors.Background {
		t.Error("master (Assets) background should come from master direction")
	}
}

func TestComposition_RealMixing_DesktopGroup(t *testing.T) {
	dirs := makeRealDirections()
	c := NewComposition("component-mix")
	c.Directions = dirs
	c.SetGroupSource(GroupAssetsAndSystem.ID, 1)
	c.SetGroupSource(GroupTerminalsAndTUI.ID, 1)
	c.SetGroupSource(GroupDesktopShell.ID, 2) // Different desktop
	c.SetGroupSource(GroupEditor.ID, 1)

	tm, _ := c.Resolve("real3", "/img.png", nil)
	if tm.Colors.Accent != dirs[1].Colors.Accent {
		t.Errorf("desktop group should override accent: got %s, expected %s",
			tm.Colors.Accent, dirs[1].Colors.Accent)
	}
}

func TestComposition_OverrideWinsOverGroup(t *testing.T) {
	dirs := makeRealDirections()
	c := NewComposition("component-mix")
	c.Directions = dirs
	c.SetGroupSource(GroupAssetsAndSystem.ID, 1)
	c.SetGroupSource(GroupTerminalsAndTUI.ID, 1)
	c.SetGroupSource(GroupDesktopShell.ID, 1)
	c.SetGroupSource(GroupEditor.ID, 1)
	c.SetOverride("neovim", 2) // Override editor

	tm, _ := c.Resolve("real4", "/img.png", nil)
	if tm.Colors.Accent != dirs[1].Colors.Accent {
		t.Errorf("neovim override should change accent: got %s, expected %s",
			tm.Colors.Accent, dirs[1].Colors.Accent)
	}
}

func makeRealDirections() []Direction {
	c1 := StaticColors()
	c2 := StaticColors()
	c3 := StaticColors()

	// Make direction 2 meaningfully different
	c2.Accent = "#ff8040"
	c2.Color0 = "#0a0b16"
	c2.Color1 = "#ff4444"
	c2.Cursor = "#ff8040"

	// Make direction 3 meaningfully different
	c3.Accent = "#40ff80"
	c3.Color0 = "#001a00"

	return []Direction{
		{ID: 1, Label: "Vibrant", Colors: c1, LightMode: false},
		{ID: 2, Label: "Balanced", Colors: c2, LightMode: false},
		{ID: 3, Label: "Muted", Colors: c3, LightMode: false},
	}
}
