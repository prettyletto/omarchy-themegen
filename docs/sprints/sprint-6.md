# Sprint 6 Tasks

## Sprint Goal

Finish reproducibility and automation artifacts.

At the end of this sprint, users can export recipes, replay recipes through the CLI, export reproducible archives with explicit source-image bundling, and consume stable JSON output for automation.

This sprint intentionally does not perform real Omarchy apply validation, final UX hardening, release packaging, new Theme Surfaces, manual color editing, theme import/merge, publishing, or mouse support.

## Assumed Complete From Sprint 5

- Whole-theme and component-mix modes work in TUI and CLI.
- TUI previews work without browser preview.
- Terminal image preview is used when supported and falls back cleanly.
- Optional browser preview is local-only, tokenized, user-requested, and synchronized with the TUI session.
- Export consumes one validated Theme Model.

## Task 1: Add Explicit Recipe Export

Type: AFK

Blocked by: Sprint 5

### What to build

Add explicit recipe export for the final Theme Model.

Recipe export must be user-requested and must not create generation history.

Recipe data must include:

- generator version;
- source image fingerprint;
- generation seed/style option;
- light-theme flag;
- selected mode;
- whole-theme direction when relevant;
- Surface Group selections when relevant;
- per-surface overrides when relevant;
- theme name only when the user chooses to include it.

Do not include source image bytes in a plain recipe.

### Acceptance criteria

- [ ] Recipe export is never automatic.
- [ ] Recipe includes enough selection data to reproduce the composition from the same source image and generator version.
- [ ] Recipe includes component-mix provenance when relevant.
- [ ] Recipe does not include source image bytes.
- [ ] Recipe does not include private absolute source paths.
- [ ] Generated README mentions recipe export only when the user requested it.

## Task 2: Add CLI Recipe Replay

Type: AFK

Blocked by: Task 1

### What to build

Allow the non-interactive CLI to read a recipe and regenerate/export the selected Theme Model using an explicitly provided source image.

Replay must:

- validate the provided source image;
- compare its fingerprint to the recipe fingerprint;
- fail clearly on mismatch unless a deliberate override flag is provided;
- use the same composition path as TUI and CLI flags;
- export through the existing exporter.

### Acceptance criteria

- [ ] CLI can replay a recipe with the original source image.
- [ ] Fingerprint mismatch fails clearly by default.
- [ ] Replay uses normal Theme Model validation.
- [ ] Replay supports whole-theme recipes.
- [ ] Replay supports component-mix recipes.
- [ ] Replay does not open the TUI unless explicitly requested.

## Task 3: Add Reproducible Archive Export

Type: AFK

Blocked by: Task 2

### What to build

Add explicit reproducible archive export.

The archive must include:

- finished Omarchy Theme Directory;
- recipe file;
- source image bytes.

Because source images may be private, the user must explicitly confirm source image bundling. Generic `--yes` may confirm only when the command explicitly requested reproducible archive mode and the warning is shown in non-interactive output.

### Acceptance criteria

- [ ] Reproducible archive export is explicit.
- [ ] Archive contains the finished theme directory.
- [ ] Archive contains the recipe.
- [ ] Archive contains source image bytes only after confirmation or explicit reproducible archive request.
- [ ] Finished-theme archive behavior remains unchanged.
- [ ] Archive root remains the normalized theme directory.

## Task 4: Add Recipe And Archive TUI Options

Type: AFK

Blocked by: Task 3

### What to build

Expose recipe and reproducible archive options in the TUI export flow.

The user must be able to:

- export the theme normally;
- optionally export a recipe;
- optionally export a finished-theme archive;
- optionally export a reproducible archive;
- see a privacy warning before source image bytes are bundled.

Do not make any of these options default except normal theme export.

### Acceptance criteria

- [ ] TUI offers recipe export as an explicit option.
- [ ] TUI offers finished-theme archive export as an explicit option.
- [ ] TUI offers reproducible archive export as an explicit option.
- [ ] Source image byte bundling requires explicit confirmation.
- [ ] Export result shows recipe/archive paths when created.
- [ ] Apply remains separate from every export option.

## Task 5: Finalize JSON Output For Automation

Type: AFK

Blocked by: Task 3

### What to build

Finalize explicit JSON output for automation.

JSON output should describe command results without replacing human-readable defaults. Include:

- success/failure status;
- warnings;
- exported path;
- archive path when created;
- recipe path when created;
- backup path when created;
- selected mode;
- selected direction/group/override provenance;
- validation confidence when Omarchy is missing;
- apply status when apply was explicitly requested.

### Acceptance criteria

- [ ] Plain text remains default.
- [ ] JSON output is emitted only when explicitly requested.
- [ ] JSON output is valid machine-readable JSON.
- [ ] JSON output includes warnings and validation confidence.
- [ ] JSON output includes recipe/reproducible archive paths when relevant.
- [ ] JSON output never includes source image bytes.

## Task 6: Add Automation Contract Tests

Type: AFK

Blocked by: Task 5

### What to build

Add tests that lock down recipe, archive, replay, and JSON behavior.

Tests should cover:

- whole-theme recipe export and replay;
- component-mix recipe export and replay;
- fingerprint mismatch rejection;
- finished-theme archive content;
- reproducible archive content;
- JSON success output;
- JSON failure output.

### Acceptance criteria

- [ ] `go test ./...` passes.
- [ ] Recipe replay tests do not require the TUI.
- [ ] Archive tests verify archive root and required files.
- [ ] Reproducible archive tests verify source image inclusion only in reproducible mode.
- [ ] JSON tests verify valid JSON on success and failure.
- [ ] Tests do not invoke `omarchy theme set`.

## Task 7: Update Generated README For Recipes And Archives

Type: AFK

Blocked by: Task 4

### What to build

Update generated theme README content to reflect requested reproducibility artifacts.

README behavior:

- normal theme export explains installation and apply only;
- recipe export mentions the recipe path and replay command shape;
- reproducible archive export explains that source image bytes are included;
- finished-theme archive export explains how to extract into Omarchy themes.

Do not include source image bytes, private absolute source paths, or recipe content directly in the README.

### Acceptance criteria

- [ ] README remains generated for every exported theme.
- [ ] README mentions recipe export only when requested.
- [ ] README mentions reproducible archive privacy only when relevant.
- [ ] README includes no private absolute source paths.
- [ ] README distinguishes finished-theme archive from reproducible archive.

## Task 8: Document Sprint 6 CLI/TUI Behavior

Type: AFK

Blocked by: Task 7

### What to build

Update user-facing command help and project docs for completed recipe/archive automation behavior.

Documentation must cover:

- explicit recipe export;
- recipe replay with source image validation;
- finished-theme archive;
- reproducible archive with source image bytes;
- explicit JSON output;
- no generation history.

### Acceptance criteria

- [ ] CLI help describes recipe export and replay.
- [ ] CLI help describes archive modes.
- [ ] CLI help explains JSON is opt-in.
- [ ] Docs explain reproducible archive privacy.
- [ ] Docs state recipes are not history.
- [ ] Docs do not claim real Omarchy apply validation is complete in this sprint.

## Out Of Sprint

- real Omarchy apply validation;
- Aether/Neovim contract hardening beyond existing behavior;
- final UX hardening;
- release packaging;
- new Theme Surfaces;
- VS Code support;
- manual color editing;
- importing or editing existing themes;
- multiple source images per run;
- generation history/session database;
- mouse support;
- remote browser preview;
- GitHub/theme marketplace publishing;
- license selection;
- community extra app configs.
