# Sprint 7 Tasks

## Sprint Goal

Validate generated themes against real installed Omarchy behavior and close remaining Omarchy/Aether unknowns.

At the end of this sprint, validation uses installed Omarchy information when available, Neovim generation targets the verified local Aether/AetherNvim contract, and a generated theme has been manually applied and rolled back on a real Omarchy system.

This sprint intentionally does not add new product features, new Theme Surfaces, publishing, mouse support, manual color editing, or theme import.

## Assumed Complete From Sprint 6

- Recipe export and replay work.
- Finished-theme archive and reproducible archive export work.
- JSON output is stable.
- Whole-theme and component-mix exports are fully automated.
- Browser preview and TUI preview are complete.

## Task 1: Strengthen Installed Omarchy Discovery

Type: AFK

Blocked by: Sprint 6

### What to build

Add non-mutating discovery of installed Omarchy paths, scripts, and theme template locations.

Discovery should identify:

- whether Omarchy is installed;
- default Omarchy path when discoverable;
- installed theme template directory;
- user theme directory;
- theme set/list command availability;
- reduced-confidence reason when unavailable.

Do not apply themes or modify Omarchy state.

### Acceptance criteria

- [ ] Omarchy-installed environment is detected.
- [ ] Omarchy-missing environment reports reduced confidence.
- [ ] Theme template directory is detected when available.
- [ ] User theme directory is detected when available.
- [ ] Discovery is read-only.
- [ ] Errors identify which expected Omarchy path/command was missing.

## Task 2: Validate Against Installed Omarchy Templates

Type: AFK

Blocked by: Task 1

### What to build

Strengthen structural validation using installed Omarchy template behavior.

When Omarchy is installed, validation should check:

- exported theme is discoverable where expected;
- required direct files exist;
- `colors.toml` contains the observed required keys;
- direct files do not duplicate installed Omarchy-templated files unnecessarily;
- `light.mode` behavior is represented correctly for explicit light themes;
- `icons.theme` is either omitted intentionally or fixed/default.

Validation must remain non-mutating and must not call `omarchy theme set`.

### Acceptance criteria

- [ ] Validation uses installed Omarchy template information when available.
- [ ] Validation reports reduced confidence when Omarchy is unavailable.
- [ ] Validation does not apply the theme.
- [ ] Validation catches missing required direct files.
- [ ] Validation catches unexpected direct generation of templated files when detectable.
- [ ] Validation errors are actionable.

## Task 3: Verify Background And Preview Asset Behavior

Type: AFK

Blocked by: Task 2

### What to build

Verify local Omarchy expectations for generated background and preview assets.

Validation should confirm:

- source image is present under `backgrounds/`;
- `preview.png` exists and is valid;
- `preview-unlock.png` exists and is valid;
- `unlock.png` exists and is valid;
- generated preview dimensions match documented expectations where fixed;
- theme menu fallback behavior is documented if preview assets are missing.

Do not generate real Omarchy screenshots.

### Acceptance criteria

- [ ] Background asset validation is explicit.
- [ ] Preview asset validation is explicit.
- [ ] `preview.png` dimension check remains 1800x1012.
- [ ] `preview-unlock.png` dimension check remains 1920x1080.
- [ ] `unlock.png` validity is checked without requiring fixed dimensions.
- [ ] Docs reflect verified local behavior.

## Task 4: Harden Aether/Neovim Contract Detection

Type: AFK

Blocked by: Task 1

### What to build

Resolve the remaining Aether/AetherNvim uncertainty by inspecting the installed environment and supporting only the verified contract.

Behavior:

- detect the installed Aether/AetherNvim palette contract;
- generate `neovim.lua` only for a known detected contract;
- fail clearly when the contract is unknown;
- include diagnostics for unsupported setups.

Do not generate a universal compatibility file.

### Acceptance criteria

- [ ] Known installed Aether/AetherNvim contract generates `neovim.lua`.
- [ ] Unknown contract fails clearly before export.
- [ ] Failure message identifies what could not be detected.
- [ ] Generated `neovim.lua` uses Theme Model colors.
- [ ] Tests cover known and unknown contract detection.
- [ ] `QUESTIONS.md` no longer lists Aether contract as an unresolved critical question after implementation.

## Task 5: Add Omarchy Validation Tests With Fixtures

Type: AFK

Blocked by: Task 4

### What to build

Add tests for Omarchy discovery and validation without requiring a live Omarchy install.

Use filesystem fixtures for:

- installed templates present;
- Omarchy missing;
- missing required direct files;
- unknown Aether contract;
- known Aether contract;
- light-theme marker behavior.

### Acceptance criteria

- [ ] `go test ./...` passes.
- [ ] Tests do not mutate real `~/.config/omarchy`.
- [ ] Tests do not run `omarchy theme set`.
- [ ] Tests cover installed and missing Omarchy cases.
- [ ] Tests cover known and unknown Aether contracts.
- [ ] Tests cover light and dark theme validation.

## Task 6: Verify Apply Flow Against Real Omarchy

Type: HITL

Blocked by: Task 2, Task 4

### What to build

Run and document a real apply validation on an Omarchy system.

This task may call Omarchy apply only after explicit confirmation, using the existing separate apply path.

Verify:

- theme export succeeds;
- apply delegates to Omarchy;
- Waybar visible colors are not obviously broken;
- terminal colors are not obviously broken;
- Hyprland border/accent behavior is not obviously broken;
- Hyprlock/unlock behavior is not obviously broken;
- Mako and btop are not obviously broken;
- Neovim loads generated Aether theme file;
- a known-good theme can be re-applied after testing.

### Acceptance criteria

- [ ] Apply remains separate from export.
- [ ] Real Omarchy apply was tested on a generated theme.
- [ ] Known-good theme rollback was tested.
- [ ] Any mismatch is documented in the relevant spec file.
- [ ] No app code reimplements Omarchy apply internals.
- [ ] User-facing warnings reflect real apply behavior.

## Task 7: Update Documentation From Real Omarchy Findings

Type: AFK

Blocked by: Task 6

### What to build

Update project docs based on Sprint 7 validation findings.

Update only facts learned from real Omarchy behavior:

- direct generated files;
- delegated template files;
- apply behavior;
- Aether/Neovim contract;
- preview/background behavior;
- validation confidence rules.

Remove or downgrade resolved questions.

### Acceptance criteria

- [ ] `QUESTIONS.md` contains no resolved local Omarchy questions.
- [ ] `COMPONENTS.md` reflects verified surfaces.
- [ ] `EXPORTERS.md` reflects verified export/apply behavior.
- [ ] `THEME_SPEC.md` reflects verified Neovim generation rules.
- [ ] Docs clearly separate known facts from remaining unknowns.
- [ ] No new speculative surfaces are added.

## Task 8: Add Real-Omarchy Validation Report Template

Type: AFK

Blocked by: Task 7

### What to build

Add a reusable validation report template for manual Omarchy checks.

The template should record:

- source image used;
- selected mode;
- selected direction/group mix;
- export path;
- Omarchy version/source path if discoverable;
- apply result;
- rollback result;
- observed surface issues;
- follow-up actions.

This is documentation only, not telemetry.

### Acceptance criteria

- [ ] Template exists in project docs.
- [ ] Template does not require screenshots.
- [ ] Template includes rollback verification.
- [ ] Template includes surface-specific observations.
- [ ] Template makes clear that no telemetry is collected.

## Out Of Sprint

- new user-facing workflows;
- new Theme Surfaces;
- VS Code support;
- manual color editing;
- theme import/merge;
- mouse support;
- browser preview expansion;
- remote publishing;
- release packaging.
