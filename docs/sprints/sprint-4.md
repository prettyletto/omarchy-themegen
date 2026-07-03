# Sprint 4 Tasks

## Sprint Goal

Add component-level mix and match without creating a second export path.

At the end of this sprint, users can choose between whole-theme mode and component-mix mode in both TUI and non-interactive CLI. Component-mix mode presents Surface Groups first, allows per-surface overrides, resolves all choices into one Theme Model, previews the composed result with the existing TUI preview, and exports through the same validated exporter used by whole-theme mode.

This sprint intentionally does not implement browser preview, advanced terminal image protocols, recipes, reproducible archives, mouse support, or new Theme Surface plugins.

## Assumed Complete From Sprint 3

- `omarchy-themegen <image>` opens the keyboard-only TUI by default.
- Exactly three Theme Directions are generated.
- Whole-theme TUI selection works.
- Whole-theme non-interactive CLI export works.
- TUI export naming, confirmation, result view, and separate apply confirmation work.
- Export, archive export, overwrite backup, structural validation, generated previews, and README generation exist.

## Task 1: Add Composition Mode Selection

Type: AFK

Blocked by: Sprint 3

### What to build

Add an explicit choice between:

- whole-theme mode;
- component-mix mode.

The default TUI path should remain simple: whole-theme selection is the first recommended path. Component-mix is available when the user chooses it.

Non-interactive CLI must also make the mode explicit when exporting without the TUI.

### Acceptance criteria

- [ ] TUI offers whole-theme and component-mix modes before final selection.
- [ ] Whole-theme mode behavior from Sprint 3 remains unchanged.
- [ ] Non-interactive CLI can request whole-theme mode explicitly.
- [ ] Non-interactive CLI can request component-mix mode explicitly once later tasks add selections.
- [ ] Missing or contradictory mode flags fail clearly.
- [ ] Export still consumes one Theme Model, not raw TUI/CLI state.

## Task 2: Define Fixed Surface Groups In Composition

Type: AFK

Blocked by: Task 1

### What to build

Implement fixed Surface Groups as composition choices. Do not build a plugin system.

Supported groups:

- Desktop Shell;
- Terminals And TUI;
- Editor;
- Assets And System.

Each group must map only to the supported Theme Surfaces documented in `COMPONENTS.md`. VS Code remains excluded.

### Acceptance criteria

- [ ] Surface Groups are fixed and documented in user-facing selection labels.
- [ ] Each supported Theme Surface belongs to exactly one group unless deliberately marked shared.
- [ ] Unsupported/community surfaces cannot be selected.
- [ ] VS Code is not exposed.
- [ ] Group choices store source direction provenance.
- [ ] Group choices alone can resolve to one valid Theme Model.

## Task 3: Add TUI Group Selection Flow

Type: AFK

Blocked by: Task 2

### What to build

Add a keyboard-only TUI flow for component-mix group selection.

The user should be able to:

- see the three generated Theme Directions;
- choose one direction per Surface Group;
- see which direction currently owns each group;
- reset all groups to one direction;
- go back to whole-theme mode.

This flow must avoid overwhelming the user. Show groups first, not every individual surface.

### Acceptance criteria

- [ ] User can assign each Surface Group to one of the five directions.
- [ ] Current group selections are visible.
- [ ] User can reset all groups to a single direction.
- [ ] User can return to whole-theme mode without losing generated directions.
- [ ] Incomplete group selections cannot proceed to export.
- [ ] TUI remains keyboard-only.

## Task 4: Add Per-Surface Override Flow

Type: AFK

Blocked by: Task 3

### What to build

Add a second-level TUI flow for users who want to modify just one thing after group selection.

Per-surface override rules:

- overrides are optional;
- overrides can choose direction `1`, `2`, or `3`;
- overrides never allow manual color editing;
- overrides must be visible before export;
- clearing an override returns that surface to its group direction.

### Acceptance criteria

- [ ] User can open a group and override individual supported surfaces.
- [ ] User can clear an individual override.
- [ ] Overrides show both surface name and source direction.
- [ ] Overrides cannot target unsupported surfaces.
- [ ] Overrides cannot edit hex colors directly.
- [ ] The final Theme Model records override provenance.

## Task 5: Resolve Component Mix Into One Theme Model

Type: AFK

Blocked by: Task 4

### What to build

Implement composition resolution for component-mix mode.

Resolution must produce one Theme Model containing:

- final palette;
- final semantic roles;
- selected wallpaper/background;
- selected surface values;
- selected group provenance;
- per-surface override provenance;
- composition warnings.

The exporter must not know whether the Theme Model came from whole-theme mode or component-mix mode except through selection provenance.

### Acceptance criteria

- [ ] Component-mix selections resolve to one Theme Model.
- [ ] Whole-theme selections continue to resolve to one Theme Model.
- [ ] Exporter reads the Theme Model only.
- [ ] Composition records source direction per group.
- [ ] Composition records source direction per override.
- [ ] Theme Model contains no raw TUI state.

## Task 6: Validate Mixed Compositions

Type: AFK

Blocked by: Task 5

### What to build

Add validation specific to component-mix composition.

Validation must check:

- every group has a valid source direction;
- every override targets a supported surface;
- every override targets a valid source direction;
- the resolved Theme Model still has all required color keys;
- contrast checks still pass after mixing;
- generated preview assets can be rendered from the resolved mix.

When a mix is visually risky but structurally valid, report a warning instead of blocking.

### Acceptance criteria

- [ ] Missing group selections block export.
- [ ] Invalid direction references block export.
- [ ] Unsupported surface overrides block export.
- [ ] Required `colors.toml` keys are still validated.
- [ ] Contrast validation runs after composition.
- [ ] Warnings are visible in TUI and CLI output.

## Task 7: Add Non-Interactive CLI Component Mix

Type: AFK

Blocked by: Task 6

### What to build

Add non-interactive CLI support for component-mix mode using explicit flags.

The CLI must support:

- selecting the source direction for each Surface Group;
- selecting per-surface overrides;
- validating that all required choices are present;
- exporting without opening the TUI;
- plain-text output by default;
- JSON output when explicitly requested.

Do not add recipe files in this sprint.

### Acceptance criteria

- [ ] CLI can export a component-mix theme without opening the TUI.
- [ ] CLI requires explicit group direction choices in component-mix mode.
- [ ] CLI accepts valid per-surface overrides.
- [ ] CLI rejects unknown groups and surfaces.
- [ ] CLI rejects missing required choices.
- [ ] JSON output includes composition provenance and warnings.

## Task 8: Show Composed Preview In TUI

Type: AFK

Blocked by: Task 6

### What to build

Extend the existing terminal/TUI preview to show the current composed Theme Model.

The preview should show:

- active mode;
- selected direction per Surface Group;
- per-surface overrides;
- final background/foreground/accent swatches;
- terminal palette swatches;
- warnings from composition validation.

This is still a representative preview, not a real app screenshot.

### Acceptance criteria

- [ ] Whole-theme preview remains available.
- [ ] Component-mix preview shows group selections.
- [ ] Component-mix preview shows per-surface overrides.
- [ ] Final palette swatches reflect the resolved Theme Model.
- [ ] Warnings are visible before export confirmation.
- [ ] No browser preview is required.

## Task 9: Update Exported README For Component Mix

Type: AFK

Blocked by: Task 7

### What to build

Update generated README content to describe component-mix provenance when the exported theme was produced by component-mix mode.

Include:

- selected mode;
- group-to-direction selections;
- per-surface overrides when present;
- generator version;
- light/dark mode.

Do not include private source paths, source image bytes, recipe content, or manual-color claims.

### Acceptance criteria

- [ ] Whole-theme README content remains correct.
- [ ] Component-mix README identifies selected groups.
- [ ] Component-mix README identifies overrides when present.
- [ ] README does not expose private absolute paths.
- [ ] README does not claim browser preview or recipes are available if not implemented.

## Task 10: Add Composition Tests

Type: AFK

Blocked by: Task 8

### What to build

Add tests for component-mix behavior without requiring a real terminal.

Tests should cover:

- group selection resolution;
- per-surface override resolution;
- invalid group selection;
- invalid surface override;
- missing group selection;
- whole-theme behavior remaining stable;
- CLI component-mix export.

### Acceptance criteria

- [ ] `go test ./...` passes.
- [ ] Composition tests prove exporter receives one Theme Model.
- [ ] TUI state tests cover group selection.
- [ ] TUI state tests cover per-surface override.
- [ ] CLI tests cover non-interactive component-mix export.
- [ ] Existing whole-theme tests still pass.

## Out Of Sprint

- browser preview;
- terminal image protocol rendering beyond existing TUI preview;
- recipe export/import;
- reproducible archives;
- mouse support;
- manual color editing;
- importing existing themes;
- community extra surfaces;
- VS Code generation;
- publishing to GitHub or theme marketplaces.
