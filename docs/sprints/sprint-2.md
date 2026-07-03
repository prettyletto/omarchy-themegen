# Sprint 2 Tasks

## Sprint Goal

Replace the Sprint 1 static color set with real offline image-derived Theme Direction generation.

At the end of this sprint, `omarchy-themegen` should take one valid source image, generate five deterministic Theme Directions, validate their palettes, allow non-interactive whole-theme selection, and export the selected direction through the Sprint 1 export path.

This sprint intentionally does not implement TUI screens, browser preview, component-mix, recipe files, or Omarchy apply.

## Assumed Complete From Sprint 1

- Go CLI skeleton exists.
- `go install` works.
- one source image is validated.
- ImageMagick `magick` dependency is validated when needed.
- static Theme Model exists.
- structural validation exists.
- self-contained Omarchy Theme Directory export exists.
- finished-theme archive export exists.
- overwrite-with-backup behavior exists.
- generated README exists.

## Task 1: Add Deterministic Generation Options

Type: AFK

Blocked by: Sprint 1

### What to build

Add generation options that control deterministic output:

- source image fingerprint;
- generator version;
- seed/style option;
- light-theme request flag.

The same image and options must produce the same three Theme Directions. A different seed/style option may produce a different set of directions.

Do not expose manual color editing.

### Acceptance criteria

- [ ] Generation options are represented independently from CLI parsing.
- [ ] Same image plus same options produces stable direction IDs and palette values.
- [ ] Different seed/style option can change generated directions.
- [ ] Light-theme request is represented but may fail clearly until light palette generation is implemented later in this sprint.
- [ ] No option allows direct hex/color editing.

## Task 2: Extract Candidate Colors From the Source Image

Type: AFK

Blocked by: Task 1

### What to build

Implement offline candidate color extraction from the validated Theme Source Image.

The implementation may use Go image libraries and/or ImageMagick `magick`. It must not use online services or AI APIs.

The extraction output should be deterministic and suitable for building three distinct palette candidates.

### Acceptance criteria

- [ ] Candidate extraction works for common opaque still image formats supported by the Sprint 1 validator.
- [ ] Candidate extraction is deterministic for the same image and options.
- [ ] Large images are processed without requiring the original full-resolution image for every analysis step.
- [ ] Extraction fails clearly when `magick` is required but missing.
- [ ] Tests cover at least two materially different source images.

## Task 3: Generate Three Palette Candidates

Type: AFK

Blocked by: Task 2

### What to build

Generate five visibly distinct palette candidates from the extracted colors.

Each candidate must include:

- background;
- foreground;
- accent;
- cursor;
- selection foreground;
- selection background;
- terminal colors `color0` through `color15`;
- internal semantic roles needed by previews/export.

This task should prioritize deterministic, readable output over advanced color theory.

### Acceptance criteria

- [ ] Exactly three candidates are produced.
- [ ] Each candidate includes every required `colors.toml` key.
- [ ] Each color is valid `#RRGGBB`.
- [ ] Candidates are deterministic for the same image/options.
- [ ] Candidates are distinct enough to be meaningfully selectable.
- [ ] No candidate depends on app-specific file formats.

## Task 4: Add Palette Contrast and Usability Validation

Type: AFK

Blocked by: Task 3

### What to build

Validate generated palettes before they become Theme Directions.

Minimum checks:

- foreground/background readability;
- accent/background distinguishability;
- selection foreground/background readability;
- terminal red/yellow/green must not collapse into near-identical colors;
- lock screen text readability against lock background.

The exact numeric thresholds should be pragmatic and documented in code comments or tests. This is a guardrail, not a full accessibility audit.

### Acceptance criteria

- [ ] Invalid low-contrast palettes are rejected or repaired before export.
- [ ] Validation reports actionable warnings/errors.
- [ ] Tests cover failing foreground/background contrast.
- [ ] Tests cover failing selection contrast.
- [ ] Tests cover terminal role collapse.
- [ ] Valid palettes from representative images pass.

## Task 5: Build Theme Directions From Valid Palettes

Type: AFK

Blocked by: Task 4

### What to build

Convert each valid palette candidate into a Theme Direction.

Each Theme Direction must include:

- direction id: `1`, `2`, or `3`;
- source image fingerprint;
- generation options;
- final palette;
- semantic roles;
- surface values needed by the current whole-theme export path;
- direction label.

Labels should be short descriptive names when accurate, otherwise `Direction 1`, `Direction 2`, or `Direction 3`.

### Acceptance criteria

- [ ] Exactly three Theme Directions are produced from a valid image.
- [ ] Direction IDs are stable.
- [ ] Direction labels are present.
- [ ] Labels fall back to numbered names when descriptive naming is not confident.
- [ ] Each direction can become a Theme Model through whole-theme selection.
- [ ] Direction data does not contain filesystem export paths.

## Task 6: Wire Whole-Theme CLI Selection to Generated Directions

Type: AFK

Blocked by: Task 5

### What to build

Replace the Sprint 1 static Theme Model path for non-interactive whole-theme export.

The CLI must support choosing one generated direction explicitly, then export that direction through the existing export path.

Expected behavior:

- `omarchy-themegen <image>` still opens TUI intent or reports TUI not implemented.
- non-interactive whole-theme export requires explicit theme name and direction choice.
- missing direction/name in explicit non-interactive mode fails clearly.

Do not implement component-mix.

### Acceptance criteria

- [ ] Non-interactive whole-theme export uses generated direction colors, not the Sprint 1 static palette.
- [ ] Direction choice must be explicit in non-interactive mode.
- [ ] Invalid direction IDs fail clearly.
- [ ] Missing required options fail clearly.
- [ ] Exported `colors.toml` matches the selected generated direction.
- [ ] Existing structural validation still runs before and after export.

## Task 7: Implement Explicit Light-Theme Generation

Type: AFK

Blocked by: Task 4

### What to build

Support explicit light-theme generation for whole-theme directions.

When the user requests light generation:

- all five directions must be light-theme attempts;
- generated palettes must pass contrast validation;
- export writes `light.mode`;
- image brightness alone must not trigger light generation.

If the image cannot produce valid light palettes, fail clearly instead of silently falling back to dark.

### Acceptance criteria

- [ ] Light mode is generated only when explicitly requested.
- [ ] Explicit light generation produces three light-theme directions or fails clearly.
- [ ] Exported light themes include `light.mode`.
- [ ] Non-light exports do not include `light.mode`.
- [ ] Tests cover explicit light generation.
- [ ] Tests confirm bright images do not automatically trigger light generation.

## Task 8: Improve Generated Preview Assets With Real Direction Colors

Type: AFK

Blocked by: Task 5

### What to build

Replace placeholder preview/unlock assets with simple generated assets that reflect the selected Theme Direction colors.

This is not the full Preview Engine/TUI. It only improves exported assets:

- `preview.png` at 1800x1012;
- `preview-unlock.png` at 1920x1080;
- `unlock.png` as a valid logo/mark image.

The assets should use the selected direction's background, foreground, accent, terminal palette, and source image.

### Acceptance criteria

- [ ] Generated assets use colors from the selected direction.
- [ ] `preview.png` remains 1800x1012.
- [ ] `preview-unlock.png` remains 1920x1080.
- [ ] `unlock.png` is valid and visually tied to the selected direction.
- [ ] Assets are deterministic for the same Theme Model.
- [ ] Asset generation failures block export with clear errors.

## Task 9: Update Exported README With Generation Summary

Type: AFK

Blocked by: Task 6

### What to build

Update the generated README to mention that the theme was generated from an image-derived direction.

The README should include:

- selected direction label;
- selected direction id;
- whether light mode was requested;
- generator version.

Do not include full generation details, private absolute paths, source image bytes, or recipe content unless a later opt-in recipe task implements that behavior.

### Acceptance criteria

- [ ] README identifies selected direction id and label.
- [ ] README identifies light/dark generation mode.
- [ ] README does not include private absolute source paths.
- [ ] README does not claim component-mix, TUI, browser preview, or recipe support is available if not implemented yet.

## Task 10: Add Representative End-to-End Fixtures

Type: AFK

Blocked by: Task 6, Task 8

### What to build

Add representative test fixtures and end-to-end tests for generated whole-theme export.

Tests should cover:

- dark wallpaper-like image;
- bright wallpaper-like image;
- UI-heavy screenshot-like image;
- explicit light request;
- invalid low-contrast handling if fixture can trigger it.

The tests should verify generated `colors.toml`, required assets, structural validation, and deterministic output.

### Acceptance criteria

- [ ] `go test ./...` passes.
- [ ] End-to-end export test uses generated palettes.
- [ ] Output is deterministic for same image/options.
- [ ] UI-heavy fixture produces a warning but can still export.
- [ ] Explicit light fixture writes `light.mode`.
- [ ] Tests do not require Omarchy apply.

## Out Of Sprint

- TUI screens;
- browser preview;
- terminal image preview;
- component-mix selection;
- Surface Group composition UI;
- recipe export/import;
- reproducible archive mode;
- Omarchy apply integration;
- online services or AI APIs;
- manual color editing;
- theme import/merge;
- VS Code generation.
