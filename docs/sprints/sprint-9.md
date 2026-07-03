# Sprint 9 Tasks

## Sprint Goal

Prepare the finished product for local release and long-term maintenance.

At the end of this sprint, `omarchy-themegen` should have stable install instructions, version output, release artifacts, final acceptance tests, and documentation that matches the complete product scope.

This sprint intentionally does not add new features. Any feature request discovered here should be rejected or documented as out of scope unless the product scope is deliberately reopened.

## Assumed Complete From Sprint 8

- Complete TUI and CLI workflows are hardened.
- Whole-theme and component-mix modes are stable.
- Preview, browser companion, recipes, archives, validation, and apply separation are complete.
- Real Omarchy apply validation has been performed.
- User-facing errors and docs are consistent.

## Task 1: Add Version And Build Metadata Output

Type: AFK

Blocked by: Sprint 8

### What to build

Add stable version/build metadata output for support and reproducibility.

Output should include:

- app version;
- generator version;
- build commit/date when available;
- Go version when useful;
- whether Omarchy was detected;
- whether `magick` was detected.

Do not make version output require Omarchy.

### Acceptance criteria

- [ ] CLI has a version command or flag.
- [ ] Version output works without Omarchy installed.
- [ ] Version output works without a source image.
- [ ] JSON mode can emit version metadata.
- [ ] Recipe/export metadata uses the same generator version.
- [ ] Tests cover version output.

## Task 2: Lock Down Installation Instructions

Type: AFK

Blocked by: Task 1

### What to build

Finalize installation instructions for users and generated theme README content.

Instructions must cover:

- `go install`;
- binary name: `omarchy-themegen`;
- required runtime dependency: `magick`;
- Omarchy requirement for apply and high-confidence validation;
- no network requirement for generation;
- no online/AI services.

### Acceptance criteria

- [ ] Project README has clear install instructions.
- [ ] Generated theme README has clear app install instructions.
- [ ] Docs explain `magick` requirement.
- [ ] Docs explain Omarchy is not required for archive generation but is required for apply.
- [ ] Docs state generation is fully offline.
- [ ] Docs do not claim package-manager support unless implemented.

## Task 3: Add Release Build Checks

Type: AFK

Blocked by: Task 1

### What to build

Add release-quality build checks that implementation agents can run before publishing binaries or tags.

Checks should include:

- formatting;
- tests;
- vet/static checks already used by the project;
- build from clean checkout assumptions;
- no dependency on local Omarchy for normal tests;
- no mutation of user Omarchy directories.

Do not require network access for normal checks once dependencies are present.

### Acceptance criteria

- [ ] Documented release check command exists.
- [ ] `go test ./...` is part of the release check.
- [ ] Formatting check is part of the release check.
- [ ] Build check is part of the release check.
- [ ] Release checks do not call `omarchy theme set`.
- [ ] Release checks do not mutate real user config directories.

## Task 4: Add Final Acceptance Test Matrix

Type: AFK

Blocked by: Sprint 8

### What to build

Create a final acceptance test matrix for the complete product.

The matrix must cover:

- valid dark image;
- valid bright image;
- UI-heavy screenshot warning;
- invalid transparent image;
- invalid animated image;
- invalid tiny image;
- whole-theme TUI path;
- component-mix TUI path;
- whole-theme CLI path;
- component-mix CLI path;
- browser preview path;
- recipe replay path;
- finished archive path;
- reproducible archive path;
- no-Omarchy reduced confidence path;
- real Omarchy apply/rollback path.

### Acceptance criteria

- [ ] Acceptance matrix exists in project docs.
- [ ] Each row has expected observable result.
- [ ] Each row identifies AFK automated or HITL manual status.
- [ ] Matrix includes apply/rollback as HITL only.
- [ ] Matrix includes out-of-scope exclusions.
- [ ] Matrix is referenced by release documentation.

## Task 5: Final Documentation Consistency Review

Type: AFK

Blocked by: Task 4

### What to build

Review all planning and user-facing docs for contradictions.

Docs to check:

- `PROJECT.md`;
- `ARCHITECTURE.md`;
- `THEME_SPEC.md`;
- `COMPONENTS.md`;
- `PALETTE_ENGINE.md`;
- `PREVIEW_ENGINE.md`;
- `EXPORTERS.md`;
- `ROADMAP.md`;
- `RISKS.md`;
- `QUESTIONS.md`;
- sprint docs;
- project README if present.

### Acceptance criteria

- [ ] Docs agree on five Theme Directions.
- [ ] Docs agree on one source image per run.
- [ ] Docs agree on no manual color editing.
- [ ] Docs agree on VS Code exclusion.
- [ ] Docs agree that apply is separate from export.
- [ ] `QUESTIONS.md` contains only genuinely unresolved questions.

## Task 6: Final Scope Guard Review

Type: AFK

Blocked by: Task 5

### What to build

Audit implementation and docs for accidental scope creep.

Reject or remove:

- theme import/merge;
- manual color editing;
- multiple image batch mode;
- generation history/session database;
- remote browser preview;
- built-in GitHub publishing;
- license selection;
- community extra app configs;
- mouse-dependent TUI behavior;
- VS Code generation.

### Acceptance criteria

- [ ] Unsupported features are absent from CLI help.
- [ ] Unsupported features are absent from TUI flows.
- [ ] Unsupported features are absent from generated README claims.
- [ ] Unsupported files are not generated.
- [ ] Scope exclusions are documented.
- [ ] Any accidental implementation is removed or disabled before release.

## Task 7: Final Release Candidate Run

Type: HITL

Blocked by: Task 6

### What to build

Run a release-candidate validation on a real Omarchy machine.

Use one representative source image and validate:

- TUI whole-theme export;
- TUI component-mix export;
- CLI whole-theme export;
- CLI component-mix export;
- browser preview optional path;
- recipe replay;
- finished archive extraction;
- reproducible archive extraction/replay;
- apply and rollback with explicit confirmation.

### Acceptance criteria

- [ ] Release candidate run was completed.
- [ ] Apply was explicitly confirmed.
- [ ] Known-good rollback was completed.
- [ ] Any defects were filed or fixed.
- [ ] Docs were updated for observed behavior.
- [ ] No out-of-scope features were added during release validation.

## Task 8: Mark Product Scope Complete

Type: HITL

Blocked by: Task 7

### What to build

Perform the final product-scope review.

The product is complete only when:

- one opaque still image produces five Theme Directions;
- whole-theme selection works;
- component-mix selection works;
- TUI and CLI both support complete workflows;
- TUI works without browser preview;
- optional browser preview is local-only;
- export writes an installable Omarchy Theme Directory;
- apply is separate and delegated to Omarchy;
- recipes and archives work;
- docs match behavior;
- known exclusions are explicit.

### Acceptance criteria

- [ ] Product owner confirms final scope is coherent.
- [ ] No critical blockers remain.
- [ ] `QUESTIONS.md` has no critical unresolved implementation blockers.
- [ ] Final docs point to current behavior, not plans.
- [ ] Release checks pass.
- [ ] The finished product can be used without reading Omarchy internals.

## Out Of Sprint

- adding new features;
- new Theme Surfaces;
- VS Code support;
- manual color editing;
- importing or editing existing themes;
- multiple source images;
- generation history;
- mouse support;
- remote browser preview;
- GitHub/theme marketplace publishing;
- license selection;
- package-manager publishing unless separately requested.
