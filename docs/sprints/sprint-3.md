# Sprint 3 Tasks

## Sprint Goal

Add the complete keyboard-only TUI for the whole-theme workflow.

At the end of this sprint, `omarchy-themegen <image>` should open a TUI that lets the user generate three Theme Directions, preview them in the terminal, select one whole direction, provide a theme name, export it, and optionally confirm apply.

This sprint intentionally does not implement browser preview, component-mix, Surface Group selection, per-surface overrides, recipe files, or reproducible archives.

## Assumed Complete From Sprint 2

- Source image validation works.
- Offline image-derived palette generation works.
- Exactly three Theme Directions are generated.
- Whole-theme non-interactive CLI export works.
- Light-theme generation works when explicitly requested.
- Exported preview assets use generated direction colors.
- Structural validation, archive export, overwrite backup, and generated README exist.

## Task 1: Add TUI Application Shell

Type: AFK

Blocked by: Sprint 2

### What to build

Implement the initial keyboard-only TUI shell for `omarchy-themegen <image>`.

The shell should own the interactive run loop and route through the same generation/export paths used by the non-interactive CLI.

Required screens/states:

- source image validation status;
- generation progress/result;
- direction comparison;
- export naming;
- export confirmation;
- export result;
- optional apply confirmation after successful export.

Do not implement mouse interactions.

### Acceptance criteria

- [ ] `omarchy-themegen <image>` opens the TUI by default.
- [ ] TUI exits cleanly with keyboard input.
- [ ] TUI can display validation errors and warnings.
- [ ] TUI does not duplicate generation or export logic from CLI.
- [ ] TUI is usable without a mouse.
- [ ] `go test ./...` passes.

## Task 2: Show Source Image Validation and Generation Status

Type: AFK

Blocked by: Task 1

### What to build

Add TUI states for input validation and Theme Direction generation.

The TUI should show:

- source image path or filename;
- validation result;
- UI-heavy warning when present;
- `magick` dependency failure when relevant;
- generation progress;
- generated direction count.

### Acceptance criteria

- [ ] Valid image proceeds to generation.
- [ ] Invalid image shows a clear blocking error.
- [ ] UI-heavy warning is visible but non-blocking.
- [ ] Generation failures are displayed without crashing.
- [ ] Exactly three generated directions are shown as the success condition.

## Task 3: Implement Terminal Direction Comparison

Type: AFK

Blocked by: Task 2

### What to build

Add a terminal/TUI comparison view for the three generated Theme Directions.

The view must show enough information to choose between directions without browser preview:

- direction id;
- direction label;
- background/foreground/accent swatches;
- terminal palette swatches;
- light/dark mode indication;
- validation warnings if present.

This is a terminal representation, not pixel-perfect app preview.

### Acceptance criteria

- [ ] All three directions are visible or reachable by keyboard.
- [ ] Direction labels and ids are clear.
- [ ] Color swatches render in terminals that support ANSI color.
- [ ] Non-color fallback text remains understandable.
- [ ] Warnings are visible without blocking valid selections.
- [ ] No browser preview is required.

## Task 4: Add Whole-Theme Selection Flow

Type: AFK

Blocked by: Task 3

### What to build

Allow the user to select one complete Theme Direction in the TUI.

Selection must produce the same Theme Model shape as non-interactive whole-theme CLI selection.

Keyboard behavior should be simple and discoverable:

- move between directions;
- select direction;
- go back;
- quit.

### Acceptance criteria

- [ ] User can select one of the three directions using keyboard only.
- [ ] Selection creates one Theme Model.
- [ ] Selected Theme Model matches the corresponding generated direction.
- [ ] User can go back to comparison before export.
- [ ] Invalid/no selection cannot proceed to export.

## Task 5: Add Export Name Entry and Normalization Preview

Type: AFK

Blocked by: Task 4

### What to build

Add a TUI step where the user explicitly provides the final theme name.

The UI must show:

- entered name;
- normalized export name/path preview;
- validation errors for empty or invalid names;
- warning when the target theme already exists.

Do not silently derive the theme name from the source filename.

### Acceptance criteria

- [ ] User must type or otherwise explicitly provide a theme name.
- [ ] Empty names are rejected.
- [ ] Normalized export name is shown before confirmation.
- [ ] Existing target theme warning is shown.
- [ ] Source filename is not silently used as the export name.

## Task 6: Add Export Confirmation and Result View

Type: AFK

Blocked by: Task 5

### What to build

Add a TUI export confirmation step and result view.

The confirmation must show:

- selected direction id/label;
- target export path;
- whether overwrite/backup is required;
- whether archive export is requested, if already supported by Sprint 1;
- that apply is separate.

The result view must show:

- exported path;
- backup path when created;
- validation status;
- warnings;
- next available actions.

### Acceptance criteria

- [ ] Export does not run until user confirms.
- [ ] Export uses existing exporter and structural validation.
- [ ] Overwrite requires the existing backup policy.
- [ ] Result view shows exported path.
- [ ] Result view shows backup path when applicable.
- [ ] Export does not apply the theme.

## Task 7: Add Separate Apply Confirmation

Type: AFK

Blocked by: Task 6

### What to build

After successful export, offer a separate confirmation to apply the theme through Omarchy.

This is the first sprint task that may call Omarchy apply, and only after export succeeds and the user explicitly confirms.

The confirmation must explain that applying restarts/reloads visible Omarchy components and may trigger hooks.

If Omarchy is missing, apply must be unavailable with a clear message.

### Acceptance criteria

- [ ] Apply is never triggered by export confirmation alone.
- [ ] Apply requires separate explicit confirmation.
- [ ] Missing Omarchy disables apply and explains why.
- [ ] Apply invokes Omarchy rather than recreating Omarchy behavior.
- [ ] Apply errors are shown clearly.
- [ ] `--yes` does not imply apply.

## Task 8: Keep Non-Interactive CLI Behavior Stable

Type: AFK

Blocked by: Task 1, Task 6

### What to build

Ensure adding the TUI does not break non-interactive CLI workflows from Sprint 2.

Explicit non-interactive commands must:

- not enter the TUI;
- fail clearly on missing required options;
- keep JSON output behavior;
- keep archive export behavior;
- keep overwrite backup behavior.

### Acceptance criteria

- [ ] Existing non-interactive whole-theme export tests still pass.
- [ ] Explicit non-interactive mode never opens the TUI.
- [ ] Missing options in non-interactive mode produce clear errors.
- [ ] JSON output remains machine-readable.
- [ ] Archive export still works.

## Task 9: Add TUI Integration Tests or Golden State Tests

Type: AFK

Blocked by: Task 6

### What to build

Add tests for the TUI state machine without requiring a real terminal.

Focus on behavior, not terminal rendering details:

- valid image advances to generation;
- invalid image blocks;
- direction selection creates the expected Theme Model;
- export name validation blocks empty names;
- export confirmation calls exporter;
- apply confirmation is separate.

### Acceptance criteria

- [ ] TUI state transitions are testable without launching a real terminal.
- [ ] Tests cover successful whole-theme export flow.
- [ ] Tests cover invalid image flow.
- [ ] Tests cover export name validation.
- [ ] Tests cover apply separation.
- [ ] `go test ./...` passes.

## Task 10: Update User-Facing Help for Interactive Workflow

Type: AFK

Blocked by: Task 6

### What to build

Update CLI help and generated README language to reflect the now-available TUI whole-theme workflow.

The docs/help should explain:

- `omarchy-themegen <image>` opens the TUI;
- keyboard-only operation;
- whole-theme selection;
- explicit theme naming;
- export/apply separation;
- non-interactive CLI remains available.

Do not claim component-mix, browser preview, recipes, or reproducible archives are available if they remain unimplemented.

### Acceptance criteria

- [ ] CLI help describes TUI default behavior.
- [ ] Generated README mentions TUI whole-theme workflow only if available.
- [ ] Generated README does not mention unimplemented features.
- [ ] Help text distinguishes export from apply.

## Out Of Sprint

- browser preview;
- terminal image protocol rendering beyond basic TUI swatches;
- component-mix selection;
- Surface Group selection;
- per-surface overrides;
- recipe export/import;
- reproducible archive mode;
- mouse support;
- theme import/merge;
- VS Code generation.

