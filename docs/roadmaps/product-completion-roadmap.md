# Product Completion Roadmap

This roadmap starts from the current repository state after the staff Go audit.

It is not a replacement for the original product plan. It is the implementation recovery plan needed to make the current codebase match the imagined finished product.

## Current Verdict

The codebase is not product-complete.

Working or mostly working:

- source image validation;
- offline palette generation;
- five Theme Directions;
- basic whole-theme CLI export;
- basic whole-theme TUI flow;
- basic Omarchy Theme Directory export;
- overwrite backup creation;
- generated README;
- finished archive creation after local export;
- recipe data structures;
- local browser preview server skeleton;
- tests and `go vet`.

Not complete enough for release:

- component-mix records provenance but does not actually change exported surface outputs;
- TUI per-surface overrides cannot be set or cleared;
- browser preview is an API shell, not a usable preview UI;
- browser selection API accepts invalid groups, surfaces, and directions;
- archive-only export is missing;
- reproducible archive silently does nothing without `--yes`;
- contrast validation warns instead of blocking or repairing unusable palettes;
- recipe replay can panic on invalid direction IDs;
- tests mutate real `$HOME`;
- Aether detection checks directory existence, not the actual palette contract;
- `main.go` owns too much product workflow logic.

## Completion Rule

The product is complete only when:

- one opaque still image produces five valid Theme Directions;
- whole-theme mode works in TUI and CLI;
- component-mix mode actually changes the exported Theme Model or is honestly narrowed;
- per-surface overrides are usable in the TUI and CLI;
- browser preview is optional, local-only, interactive, and validates requests;
- export writes an installable Omarchy Theme Directory;
- archive-only export works without writing into local Omarchy themes;
- recipe replay works without panics;
- apply remains separate from export;
- tests do not mutate real user config;
- docs match implemented behavior.

## Phase 1: Stabilize Unsafe And Misleading Behavior

Goal: remove panic paths, silent no-ops, real-user filesystem mutation, and misleading CLI behavior before adding more product behavior.

### Sprint 10: Safety And Contract Fixes

#### Task 1: Stop Tests From Mutating Real `$HOME`

Type: AFK

Blocked by: None

What to build:

Move Aether/Neovim test setup into a temp HOME or injectable discovery root. Tests must never create `~/.local/share/nvim/lazy/aether.nvim` under the developer's real home directory.

Acceptance criteria:

- [ ] `go test ./...` does not write under the real `$HOME`.
- [ ] Aether contract tests use temp directories or injected roots.
- [ ] Export tests still pass.
- [ ] Release docs remain truthful: tests do not mutate user config.

#### Task 2: Validate CLI Mode Values

Type: AFK

Blocked by: None

What to build:

Reject unknown `--mode` values before generation starts.

Acceptance criteria:

- [ ] `--mode whole-theme` works.
- [ ] `--mode component-mix` works.
- [ ] `--mode typo` fails with a clear mode error.
- [ ] Unknown mode does not fall through to whole-theme direction errors.

#### Task 3: Make Recipe Replay Non-Panicking

Type: AFK

Blocked by: None

What to build:

Validate recipe fields before indexing generated directions or resolving composition.

Acceptance criteria:

- [ ] Whole-theme recipe with missing direction fails clearly.
- [ ] Whole-theme recipe with direction outside `1..3` fails clearly.
- [ ] Component-mix recipe with invalid group direction fails clearly.
- [ ] Component-mix recipe with invalid override fails clearly.
- [ ] No malformed recipe can panic the process.

#### Task 4: Make Reproducible Archive Confirmation Explicit

Type: AFK

Blocked by: None

What to build:

Replace the current silent `--reproducible` + missing `--yes` no-op with a blocking error in non-interactive mode and an explicit confirmation in TUI mode.

Acceptance criteria:

- [ ] CLI `--reproducible` without required confirmation fails clearly.
- [ ] CLI `--reproducible --yes` creates the archive.
- [ ] Output explains source image bytes are included.
- [ ] JSON output reports the same error in machine-readable form.

#### Task 5: Normalize Error Output For Partial Writes

Type: AFK

Blocked by: Task 3, Task 4

What to build:

Ensure export errors say whether files were written and where. This is especially important after backup/overwrite, preview generation, Neovim generation, archive creation, recipe writing, and post-export validation.

Acceptance criteria:

- [ ] Export failure before writing reports no files written.
- [ ] Export failure after writing reports target path and partial state.
- [ ] Backup failure prevents replacement.
- [ ] Archive failure is reported without hiding theme export result.
- [ ] JSON and plain text both include warnings/errors.

## Phase 2: Restore A Clean Workflow Architecture

Goal: reduce `main.go` coupling so CLI, TUI, browser, and recipe replay use the same product path.

### Sprint 11: Shared Workflow Core

#### Task 1: Extract A Product Workflow Module

Type: AFK

Blocked by: Sprint 10

What to build:

Move generation, selection resolution, validation, export, archive, recipe, reproducible archive, and result assembly behind one workflow module.

Do not create a plugin system. Do not invent generic interfaces. The module should express this product's actual flow:

source image + generation options + selection + export options -> Theme Model + export result.

Acceptance criteria:

- [ ] CLI whole-theme export uses the workflow module.
- [ ] CLI component-mix export uses the workflow module.
- [ ] Recipe replay uses the workflow module.
- [ ] TUI export uses the workflow module or the same lower-level functions.
- [ ] `main.go` no longer assembles Theme Models, archives, recipes, and JSON directly.

#### Task 2: Define One Result Shape For CLI And TUI

Type: AFK

Blocked by: Task 1

What to build:

Create one internal result shape for successful and failed runs. It should include paths, warnings, validation confidence, recipe/archive outputs, backup path, mode, selected directions, and apply status when relevant.

Acceptance criteria:

- [ ] Plain-text output renders from the shared result.
- [ ] JSON output renders from the shared result.
- [ ] TUI result screen renders from the shared result.
- [ ] Error paths carry the same information as success paths where available.

#### Task 3: Centralize Selection Validation

Type: AFK

Blocked by: Task 1

What to build:

Put whole-theme, component-mix, recipe replay, browser state, and TUI selection validation in one place.

Acceptance criteria:

- [ ] Direction IDs are validated once.
- [ ] Surface Group IDs are validated once.
- [ ] Surface override names are validated once.
- [ ] CLI, TUI, browser, and recipe replay produce equivalent validation errors.
- [ ] Tests cover each selection input path.

#### Task 4: Remove Stale Static Theme Paths From Product Code

Type: AFK

Blocked by: Task 1

What to build:

Keep static fixtures for tests if useful, but remove static Theme Model construction from runtime product paths.

Acceptance criteria:

- [ ] Runtime export uses generated Theme Directions.
- [ ] Static colors are test fixtures only.
- [ ] Generated README never says "Static" for normal product runs.
- [ ] Tests are explicit when they use static fixtures.

## Phase 3: Make Component-Mix Real

Goal: component-mix must either affect exported files or be narrowed honestly. The current state is unacceptable because it records selections but exports one master direction.

### Sprint 12: Component-Mix Semantics

#### Task 1: Decide The Real Mixing Contract

Type: HITL

Blocked by: Sprint 11

What to decide:

Choose one concrete contract:

- Contract A: component-mix affects only the single final color contract by selecting a base direction plus named color-role sources.
- Contract B: component-mix directly generates per-surface files for selected surfaces, accepting that this duplicates some Omarchy template behavior.
- Contract C: remove per-surface mix and keep only whole-theme plus group-level preview provenance.

Recommendation:

Choose Contract A unless real Omarchy template behavior proves per-surface direct files are necessary. It keeps export simple and avoids a per-surface exporter system.

Acceptance criteria:

- [ ] Decision is recorded in `THEME_SPEC.md`.
- [ ] `COMPONENTS.md` matches the decision.
- [ ] `ARCHITECTURE.md` names the selected composition rule.
- [ ] Out-of-scope behavior is removed from docs/help.

#### Task 2: Implement Real Composition Resolution

Type: AFK

Blocked by: Task 1

What to build:

Implement the selected composition contract so group and override choices affect the final Theme Model, not only metadata.

Acceptance criteria:

- [ ] Changing Desktop Shell group changes final model values relevant to Desktop Shell.
- [ ] Changing Terminals And TUI group changes final terminal palette values.
- [ ] Changing Editor group changes final Neovim/Aether values.
- [ ] Changing Assets And System group changes wallpaper/preview/light/system values where applicable.
- [ ] Per-surface override wins over group selection.
- [ ] Exported files reflect the resolved Theme Model.

#### Task 3: Revalidate Mixed Themes After Resolution

Type: AFK

Blocked by: Task 2

What to build:

Run color, composition, preview, and export validation against the resolved mixed Theme Model.

Acceptance criteria:

- [ ] Invalid mixed contrast blocks or repairs before export.
- [ ] Unsupported surface override blocks export.
- [ ] Missing group selection blocks export.
- [ ] Warnings are shown before export.
- [ ] Tests cover valid and invalid mixed compositions.

#### Task 4: Update Recipe Semantics For Real Mixes

Type: AFK

Blocked by: Task 2

What to build:

Ensure recipes reproduce the actual mixed result, not just provenance labels.

Acceptance criteria:

- [ ] Component-mix recipe replay reproduces the same Theme Model.
- [ ] Fingerprint mismatch still blocks by default.
- [ ] Recipe includes enough selection data for the chosen mixing contract.
- [ ] Tests compare exported `colors.toml` or equivalent final model values.

## Phase 4: Finish TUI Product Workflow

Goal: TUI must actually support the workflows promised in the product scope.

### Sprint 13: TUI Completion

#### Task 1: Implement Per-Surface Override Controls

Type: AFK

Blocked by: Sprint 12

What to build:

Make the TUI override screen usable with keyboard-only controls.

Acceptance criteria:

- [ ] User can move through surfaces.
- [ ] User can set selected surface to direction `1`, `2`, or `3`.
- [ ] User can clear an override.
- [ ] User can return to group selection.
- [ ] Override choices are visible before export.
- [ ] Tests prove override set/clear behavior.

#### Task 2: Add TUI Archive, Recipe, And Reproducible Options

Type: AFK

Blocked by: Sprint 11

What to build:

Expose export artifact options in the TUI export flow.

Acceptance criteria:

- [ ] User can export only a local theme directory.
- [ ] User can request finished archive.
- [ ] User can request recipe export.
- [ ] User can request reproducible archive.
- [ ] Source image byte bundling requires explicit confirmation.
- [ ] Result screen shows every generated path.

#### Task 3: Preserve State Across Back Navigation

Type: AFK

Blocked by: Task 1

What to build:

Make TUI back navigation predictable. Generated directions, group choices, overrides, theme name, and artifact choices should be preserved unless the user deliberately resets them.

Acceptance criteria:

- [ ] Back from naming preserves selections.
- [ ] Back from confirmation preserves theme name.
- [ ] Switching whole-theme/component-mix has explicit reset behavior.
- [ ] Browser updates do not unexpectedly erase TUI choices.
- [ ] Tests cover back navigation.

#### Task 4: Make Apply Availability Accurate

Type: AFK

Blocked by: Sprint 11

What to build:

Use Omarchy discovery, not only `exec.LookPath("omarchy")`, to determine apply availability and display validation confidence.

Acceptance criteria:

- [ ] Apply unavailable state matches Omarchy discovery.
- [ ] Apply confirmation explains what will happen.
- [ ] Apply remains separate from export.
- [ ] `--yes` never implies apply.
- [ ] Tests cover Omarchy missing state.

## Phase 5: Finish Browser Preview

Goal: browser preview must be optional, local-only, interactive, and synchronized with the active selection session.

### Sprint 14: Browser Preview Completion

#### Task 1: Validate Browser Selection Requests

Type: AFK

Blocked by: Sprint 11

What to build:

Reject invalid browser requests before they reach the active selection session.

Acceptance criteria:

- [ ] Invalid direction ID returns a 400 response.
- [ ] Unknown Surface Group returns a 400 response.
- [ ] Unknown surface returns a 400 response.
- [ ] Invalid mode returns a 400 response.
- [ ] TUI state is unchanged after invalid browser requests.

#### Task 2: Build Real Browser Selection UI

Type: AFK

Blocked by: Task 1, Sprint 12

What to build:

Replace placeholder browser HTML with usable local controls for:

- whole-theme direction selection;
- component-mix mode;
- Surface Group assignment;
- per-surface override set/clear;
- current warnings;
- current resolved preview.

Acceptance criteria:

- [ ] Browser displays all five directions.
- [ ] Browser displays Surface Groups.
- [ ] Browser can set and clear overrides through visible controls.
- [ ] Browser shows current selection state.
- [ ] Browser shows local-only URL/token behavior.
- [ ] Browser does not own export data.

#### Task 3: Serve Only Session Assets

Type: AFK

Blocked by: Task 2

What to build:

Serve generated preview/session assets without exposing arbitrary filesystem paths.

Acceptance criteria:

- [ ] Browser can view generated preview PNGs.
- [ ] Browser cannot request arbitrary local paths.
- [ ] Assets are scoped to active session.
- [ ] Session shutdown makes assets unavailable.

#### Task 4: Ask Before Opening Browser

Type: AFK

Blocked by: Task 2

What to build:

Honor the product rule that opening a browser is explicit. The app may show a URL without opening the browser.

Acceptance criteria:

- [ ] TUI starts local server only when user asks.
- [ ] TUI does not run `xdg-open` without confirmation.
- [ ] CLI browser mode does not hang unless explicitly interactive.
- [ ] URL is shown when opening is declined or fails.

## Phase 6: Export And Omarchy Correctness

Goal: exported artifacts must be correct, safe, and aligned with real Omarchy behavior.

### Sprint 15: Exporter Correctness

#### Task 1: Add Archive-Only Export

Type: AFK

Blocked by: Sprint 11

What to build:

Allow finished-theme archive creation without writing to `~/.config/omarchy/themes`.

Acceptance criteria:

- [ ] CLI can create finished archive only.
- [ ] CLI can create reproducible archive only.
- [ ] Archive root is the normalized theme directory.
- [ ] Archive-only mode does not mutate local Omarchy themes.
- [ ] Tests assert no local theme directory write.

#### Task 2: Make Overwrite Replacement Transactional Enough

Type: AFK

Blocked by: Sprint 10

What to build:

Avoid leaving users with a moved old theme and partial new theme when replacement export fails.

Acceptance criteria:

- [ ] Existing theme is backed up before replacement.
- [ ] Failed replacement reports backup path and partial path.
- [ ] Failed replacement does not silently destroy the previous theme.
- [ ] If automatic rollback is implemented, it is tested.
- [ ] Errors clearly say what was mutated.

#### Task 3: Verify Direct Files Vs Omarchy Templates

Type: AFK

Blocked by: Sprint 11

What to build:

Use installed Omarchy discovery to ensure export writes only direct files and source facts Omarchy expects.

Acceptance criteria:

- [ ] Export writes `colors.toml`.
- [ ] Export writes background and preview assets.
- [ ] Export writes `neovim.lua` only when contract is known.
- [ ] Export writes `light.mode` only for explicit light themes.
- [ ] Export does not generate templated surface files unnecessarily.
- [ ] Reduced confidence is reported when Omarchy is unavailable.

#### Task 4: Harden Aether Contract Detection

Type: AFK

Blocked by: Task 3

What to build:

Replace directory-exists detection with actual contract detection for the installed Aether/AetherNvim setup.

Acceptance criteria:

- [ ] Known contract is detected by inspecting relevant installed files.
- [ ] Unknown contract fails clearly before export.
- [ ] Generated `neovim.lua` matches detected contract.
- [ ] Tests use temp HOME/fixtures only.
- [ ] `QUESTIONS.md` is updated when the contract is resolved.

## Phase 7: Palette And Preview Quality

Goal: generated outputs must be consistently usable, not merely syntactically valid.

### Sprint 16: Usability Validation

#### Task 1: Enforce Palette Usability

Type: AFK

Blocked by: Sprint 12

What to build:

Turn critical contrast failures into blocking errors or deterministic repairs before Theme Directions are offered.

Acceptance criteria:

- [ ] Foreground/background failure blocks or repairs.
- [ ] Selection contrast failure blocks or repairs.
- [ ] Terminal red/yellow/green collapse blocks or repairs.
- [ ] Lock screen readability is checked.
- [ ] Valid representative images still produce five directions.

#### Task 2: Improve Direction Distinctness

Type: AFK

Blocked by: Task 1

What to build:

Ensure the three Theme Directions are meaningfully different for representative images.

Acceptance criteria:

- [ ] Direction labels are truthful or fallback to numbered names.
- [ ] Directions differ in at least accent and palette treatment.
- [ ] Low-saturation images still produce three valid choices or fail clearly.
- [ ] Tests cover dark, bright, colorful, muted, and UI-heavy fixtures.

#### Task 3: Make Preview Reflect Resolved Theme Model

Type: AFK

Blocked by: Sprint 12

What to build:

Ensure TUI, PNG, browser, and exported preview assets are rendered from the resolved Theme Model.

Acceptance criteria:

- [ ] Whole-theme preview reflects selected direction.
- [ ] Component-mix preview reflects resolved mix.
- [ ] Per-surface override provenance is visible.
- [ ] Exported `preview.png` and `preview-unlock.png` reflect final model.
- [ ] Preview rendering does not define export truth.

## Phase 8: Final Release Readiness

Goal: release only after the product behavior, tests, docs, and manual Omarchy validation agree.

### Sprint 17: Release Candidate

#### Task 1: Update Docs To Match Reality

Type: AFK

Blocked by: Sprints 10-16

What to build:

Update docs after implementation, removing claims for behavior that was narrowed or not implemented.

Acceptance criteria:

- [ ] `PROJECT.md` matches behavior.
- [ ] `ARCHITECTURE.md` matches module responsibilities.
- [ ] `THEME_SPEC.md` matches composition semantics.
- [ ] `COMPONENTS.md` matches supported surfaces.
- [ ] `PREVIEW_ENGINE.md` matches TUI/browser behavior.
- [ ] `EXPORTERS.md` matches archive/apply behavior.
- [ ] `QUESTIONS.md` contains no resolved blockers.

#### Task 2: Final Acceptance Matrix Run

Type: AFK

Blocked by: Task 1

What to build:

Run the automated acceptance matrix and update statuses.

Acceptance criteria:

- [ ] Valid image path passes.
- [ ] Invalid image paths fail correctly.
- [ ] Whole-theme CLI path passes.
- [ ] Component-mix CLI path passes.
- [ ] Recipe replay passes.
- [ ] Finished archive passes.
- [ ] Reproducible archive passes.
- [ ] Tests do not invoke `omarchy theme set`.

#### Task 3: Real Omarchy Apply And Rollback

Type: HITL

Blocked by: Task 2

What to build:

On a real Omarchy machine, validate export, apply, visible surfaces, and rollback.

Acceptance criteria:

- [ ] Apply is explicitly confirmed.
- [ ] Waybar/terminal/Hyprland/Hyprlock/Mako/btop/Neovim are checked.
- [ ] Known-good theme rollback is completed.
- [ ] Observed mismatches are fixed or documented.
- [ ] No code reimplements Omarchy apply internals.

#### Task 4: Product Complete Review

Type: HITL

Blocked by: Task 3

What to build:

Final review against the completion rule in this roadmap.

Acceptance criteria:

- [ ] Product owner confirms scope is coherent.
- [ ] No critical blockers remain.
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] Release checks pass.
- [ ] The product can be used without reading Omarchy internals.

## Agent Handoff Order

Pass work to agents in this order:

1. Sprint 10: safety fixes.
2. Sprint 11: shared workflow core.
3. Sprint 12: real component-mix semantics.
4. Sprint 13: TUI completion.
5. Sprint 14: browser preview completion.
6. Sprint 15: exporter correctness.
7. Sprint 16: palette and preview quality.
8. Sprint 17: release candidate.

Do not allow agents to start Sprint 13, 14, 15, or 16 before Sprint 12 decides and implements real component-mix semantics. Otherwise they will build UI and preview on top of fake composition.

## Scope Guard

Do not add:

- manual color editing;
- theme import or merge;
- multiple source images;
- generation history;
- VS Code generation;
- remote browser preview;
- GitHub or marketplace publishing;
- package-manager publishing;
- arbitrary community app configs;
- generic plugin systems.
