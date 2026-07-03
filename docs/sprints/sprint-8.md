# Sprint 8 Tasks

## Sprint Goal

Harden the complete user experience until it is suitable for non-designers using the final product.

At the end of this sprint, the TUI, CLI, previews, validation, errors, and generated documentation should behave consistently across expected success and failure paths.

This sprint intentionally does not add new features, new surfaces, publishing, manual color editing, theme import, or mouse support.

## Assumed Complete From Sprint 7

- Installed Omarchy discovery and structural validation work.
- Aether/Neovim contract detection is verified.
- Real apply/rollback validation has been performed.
- Docs reflect real Omarchy behavior.
- Recipe/archive/JSON automation is complete.

## Task 1: Harden Command And Help Text

Type: AFK

Blocked by: Sprint 7

### What to build

Review every command, flag, prompt, and help message for consistency.

Help text must clearly explain:

- default TUI workflow;
- non-interactive CLI workflow;
- whole-theme mode;
- component-mix mode;
- light generation;
- browser preview;
- recipe export/replay;
- archive modes;
- export/apply separation.

### Acceptance criteria

- [ ] Running without required arguments prints useful usage text.
- [ ] TUI default behavior is documented.
- [ ] CLI non-interactive behavior is documented.
- [ ] Missing required CLI options produce actionable errors.
- [ ] Help text does not mention unsupported features.
- [ ] Help text consistently names Theme Directions, Surface Groups, and Theme Model outputs.

## Task 2: Harden TUI Navigation And Recovery

Type: AFK

Blocked by: Sprint 7

### What to build

Harden keyboard-only TUI navigation across the complete flow.

The user must be able to:

- move forward and backward without losing generated directions unexpectedly;
- quit safely;
- recover from validation errors;
- recover from export errors;
- return from browser preview;
- understand when apply is available or unavailable.

### Acceptance criteria

- [ ] Every TUI screen has a keyboard path forward or back.
- [ ] Quit never applies or overwrites.
- [ ] Validation errors do not crash the TUI.
- [ ] Export errors report whether files were written.
- [ ] Browser preview can be closed without losing selection.
- [ ] Apply unavailable state is clear when Omarchy is missing.

## Task 3: Harden Error Messages And Warnings

Type: AFK

Blocked by: Sprint 7

### What to build

Normalize errors and warnings across CLI, TUI, JSON, and browser preview.

Focus areas:

- missing `magick`;
- missing Omarchy;
- invalid image;
- transparent image;
- animated image;
- tiny image;
- UI-heavy screenshot warning;
- invalid theme name;
- overwrite refusal;
- backup failure;
- preview rendering failure;
- browser preview unavailable;
- recipe fingerprint mismatch;
- unknown Neovim/Aether contract;
- apply failure.

Errors should say what failed, whether anything was written, and what to do next.

### Acceptance criteria

- [ ] Blocking errors are clearly marked as blocking.
- [ ] Warnings are clearly non-blocking.
- [ ] Errors distinguish pre-write failure from post-write failure.
- [ ] Export failure reports whether files were written.
- [ ] Backup failure prevents replacement.
- [ ] JSON output includes the same warnings/errors in machine-readable form.

## Task 4: Tune Palette Scoring From Representative Images

Type: AFK

Blocked by: Sprint 7

### What to build

Tune deterministic palette scoring using representative real images.

Use a fixed fixture set:

- dark wallpaper-like image;
- bright wallpaper-like image;
- colorful image;
- low-saturation image;
- UI-heavy screenshot-like image;
- image with one dominant accent color.

Do not add online services, AI APIs, or manual color editing.

### Acceptance criteria

- [ ] Each fixture produces five valid Theme Directions.
- [ ] Directions are visibly distinct enough to choose from.
- [ ] Foreground/background contrast passes.
- [ ] Selection contrast passes.
- [ ] Terminal role collapse checks pass.
- [ ] Results remain deterministic.

## Task 5: Tune Preview Readability

Type: AFK

Blocked by: Task 4

### What to build

Improve preview readability without making previews authoritative.

Focus on:

- terminal palette swatches;
- selected Surface Group provenance;
- per-surface override visibility;
- light/dark indication;
- warning visibility;
- source wallpaper treatment.

Do not launch real Omarchy components or create real app screenshots.

### Acceptance criteria

- [ ] TUI preview remains useful without terminal image support.
- [ ] PNG preview shows enough information to compare directions.
- [ ] Component-mix provenance is visible before export.
- [ ] Warnings are visible before export.
- [ ] Preview rendering remains deterministic.
- [ ] Export validation does not depend on preview appearance.

## Task 6: Harden File Safety And Rollback Behavior

Type: AFK

Blocked by: Sprint 7

### What to build

Audit all filesystem-writing paths.

Ensure:

- overwrite refuses by default;
- replacement always backs up first;
- backup failure blocks replacement;
- partial export failure reports written files;
- archive export does not mutate local themes unless requested;
- recipe export does not write unexpected files;
- apply remains separate.

### Acceptance criteria

- [ ] Every write path has a clear target.
- [ ] Existing theme directories are never overwritten silently.
- [ ] Backup path is reported.
- [ ] Partial failures are reported accurately.
- [ ] Archive-only export does not write to `~/.config/omarchy/themes`.
- [ ] Apply is never triggered by export, recipe, archive, or `--yes` alone.

## Task 7: Add Full UX Regression Tests

Type: AFK

Blocked by: Tasks 1, 2, 3, 6

### What to build

Add regression tests for complete UX flows without relying on a real terminal or browser.

Test:

- whole-theme TUI-equivalent flow;
- component-mix TUI-equivalent flow;
- CLI whole-theme export;
- CLI component-mix export;
- browser selection sync;
- export error handling;
- overwrite refusal;
- backup replacement;
- recipe replay;
- reduced-confidence validation when Omarchy is missing.

### Acceptance criteria

- [ ] `go test ./...` passes.
- [ ] Tests cover successful whole-theme and component-mix flows.
- [ ] Tests cover representative failure paths.
- [ ] Tests do not invoke real `xdg-open`.
- [ ] Tests do not invoke `omarchy theme set`.
- [ ] Tests do not mutate real user Omarchy directories.

## Task 8: Update User Documentation For Hardened Behavior

Type: AFK

Blocked by: Task 7

### What to build

Update project README/help docs and generated-theme README language to match hardened behavior.

Docs must explain:

- install requirements;
- image requirements;
- warnings vs errors;
- whole-theme vs component-mix;
- terminal and browser preview behavior;
- export/apply separation;
- backup behavior;
- recipe and archive behavior;
- known exclusions.

### Acceptance criteria

- [ ] Docs match implemented CLI/TUI behavior.
- [ ] Docs explain `magick` requirement.
- [ ] Docs explain opaque still image requirement.
- [ ] Docs explain source screenshot warning.
- [ ] Docs explain privacy implications of reproducible archives.
- [ ] Docs list exclusions without implying they are planned soon.

## Out Of Sprint

- new Theme Surfaces;
- VS Code support;
- manual color editing;
- importing or editing existing themes;
- multiple source images;
- generation history;
- mouse support;
- remote browser preview;
- marketplace publishing;
- license selection.
