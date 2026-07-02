# Sprint 5 Tasks

## Sprint Goal

Complete preview support for the final product without making preview the source of truth.

At the end of this sprint, users can inspect whole-theme and component-mix results in the TUI, view generated PNG previews, optionally open a local browser companion, and keep browser/TUI selections synchronized through the same active Theme Model session.

This sprint intentionally does not implement recipes, reproducible archives, theme publishing, mouse support, manual color editing, or theme import.

## Assumed Complete From Sprint 4

- Whole-theme and component-mix modes both work in TUI and CLI.
- Surface Group selection and per-surface overrides resolve to one Theme Model.
- Composition validation runs before export.
- TUI shows a representative composed preview.
- Exported README includes whole-theme or component-mix provenance.

## Task 1: Add Terminal Image Capability Detection

Type: AFK

Blocked by: Sprint 4

### What to build

Detect whether the current terminal supports inline image display suitable for showing generated PNG previews.

Detection must:

- be runtime-based;
- avoid hardcoding one terminal emulator as the only supported option;
- fall back to text/ANSI preview when unsupported;
- expose a clear capability result to the TUI.

Do not make terminal image support required for using the app.

### Acceptance criteria

- [ ] TUI can determine whether inline image preview is available.
- [ ] Unsupported terminals fall back cleanly to text/ANSI preview.
- [ ] Missing support does not block generation, selection, export, or apply.
- [ ] Capability result is visible in diagnostics/help output.
- [ ] Tests cover supported, unsupported, and unknown capability outcomes.

## Task 2: Render Direction Preview PNGs For TUI Use

Type: AFK

Blocked by: Task 1

### What to build

Generate preview PNGs for each of the three Theme Directions before final selection.

These are selection previews, not exported theme assets. They should use the same deterministic preview rendering rules as exported assets.

Each direction preview should show:

- source wallpaper treatment;
- background/foreground/accent;
- terminal palette;
- bar-like surface;
- terminal-like surface;
- notification-like surface;
- lock/unlock surface.

### Acceptance criteria

- [ ] Three direction preview PNGs are generated.
- [ ] Preview PNGs are deterministic for the same image/options.
- [ ] Preview PNGs reflect each direction's colors.
- [ ] Preview generation failures are surfaced clearly.
- [ ] TUI can reference generated preview PNG paths.
- [ ] Exported assets remain generated from the final Theme Model, not stale direction previews.

## Task 3: Render Composed Theme Preview PNG

Type: AFK

Blocked by: Task 2, Sprint 4

### What to build

Generate a preview PNG for the current composed Theme Model in both whole-theme and component-mix modes.

The composed preview must update when:

- selected whole-theme direction changes;
- Surface Group selections change;
- per-surface overrides change;
- light/dark mode changes.

### Acceptance criteria

- [ ] Whole-theme selection produces a composed preview PNG.
- [ ] Component-mix selection produces a composed preview PNG.
- [ ] Preview updates when group selections change.
- [ ] Preview updates when per-surface overrides change.
- [ ] Preview reflects final Theme Model colors.
- [ ] Preview rendering does not mutate the Theme Model.

## Task 4: Display PNG Previews In TUI When Supported

Type: AFK

Blocked by: Task 3

### What to build

Integrate terminal image display into the TUI when capability detection says it is supported.

Behavior:

- show inline direction previews when comparing directions;
- show inline composed preview before export;
- retain keyboard-only navigation;
- fall back to text/ANSI preview when image display fails.

### Acceptance criteria

- [ ] Direction comparison can show inline PNG previews when supported.
- [ ] Composed preview can show inline PNG preview when supported.
- [ ] Image display failure falls back without crashing.
- [ ] Text/ANSI preview remains available.
- [ ] TUI remains keyboard-only.
- [ ] Tests cover fallback behavior without requiring a real terminal image protocol.

## Task 5: Add Preview Cache

Type: AFK

Blocked by: Task 3

### What to build

Cache expensive preview artifacts by stable preview identity.

Cache identity must include:

- source image fingerprint;
- generator version;
- seed/style option;
- light-theme flag;
- selected mode;
- selected group/override provenance for composed previews.

The cache is internal only. It must not become generation history.

### Acceptance criteria

- [ ] Re-rendering the same direction can reuse cached preview output.
- [ ] Changing seed/style option invalidates the preview cache.
- [ ] Changing component-mix selections invalidates the composed preview cache.
- [ ] Cache entries are safe to delete.
- [ ] Cache paths are not stored in exported Theme Model provenance.
- [ ] Tests cover cache hit and invalidation behavior.

## Task 6: Add Local Browser Preview Server

Type: AFK

Blocked by: Task 3

### What to build

Implement an optional local browser companion preview.

Constraints:

- user-requested only;
- bind only to a local interface;
- use an ephemeral port unless explicitly configured;
- require a one-time session token in the URL;
- show the URL in the TUI/CLI;
- ask before using `xdg-open` or equivalent;
- stop when the preview session ends.

Browser preview must render the same directions and current composition as the TUI.

### Acceptance criteria

- [ ] Browser preview never starts automatically.
- [ ] Server binds only locally.
- [ ] Browser URL contains a one-time token.
- [ ] URL is shown even when browser open is declined.
- [ ] `xdg-open` runs only after explicit confirmation.
- [ ] Server stops when the session ends.

## Task 7: Add Browser Direction And Mix Selection

Type: AFK

Blocked by: Task 6

### What to build

Allow browser preview to change the active selection session.

Browser interactions may include:

- choosing whole-theme direction;
- switching to component-mix mode;
- assigning Surface Groups;
- setting per-surface overrides;
- clearing overrides.

Browser state must not become export truth. It must update the same active selection session that the TUI uses, and export must still consume the validated Theme Model.

### Acceptance criteria

- [ ] Browser can select whole-theme direction.
- [ ] Browser can select component-mix groups.
- [ ] Browser can set and clear per-surface overrides.
- [ ] TUI reflects browser selection changes.
- [ ] Export consumes the resolved Theme Model, not browser-local state.
- [ ] Invalid browser requests are rejected and shown as errors.

## Task 8: Add CLI Browser Preview Entry Point

Type: AFK

Blocked by: Task 6

### What to build

Add explicit CLI support for opening or starting browser preview for a generated session.

Expected behavior:

- CLI can generate previews and print a local URL;
- CLI can optionally ask before opening the browser;
- non-interactive mode must not hang waiting for browser interaction unless explicitly requested;
- JSON output includes preview URL and server lifecycle status when requested.

### Acceptance criteria

- [ ] Browser preview can be requested from CLI.
- [ ] CLI prints local URL.
- [ ] CLI does not open the browser without confirmation or explicit open flag.
- [ ] Non-interactive export still works without browser preview.
- [ ] JSON output includes preview metadata when requested.
- [ ] Browser preview errors do not corrupt generation/export state.

## Task 9: Harden Preview Security And Lifecycle

Type: AFK

Blocked by: Task 7, Task 8

### What to build

Add lifecycle and security guardrails for local preview.

Required checks:

- reject non-local bind addresses by default;
- reject missing/invalid tokens;
- expire token after first valid session use or session end;
- avoid serving arbitrary filesystem paths;
- serve only generated preview/session assets;
- close idle sessions.

### Acceptance criteria

- [ ] Non-local bind is rejected.
- [ ] Missing token cannot access preview session.
- [ ] Invalid token cannot access preview session.
- [ ] Server does not expose arbitrary files.
- [ ] Generated assets are served only for the active session.
- [ ] Idle sessions close automatically.

## Task 10: Add Preview End-To-End Tests

Type: AFK

Blocked by: Task 9

### What to build

Add tests for preview behavior across TUI, CLI, PNG rendering, cache, and browser companion.

Tests should cover:

- preview PNG generation for directions;
- composed preview update after component-mix changes;
- terminal image fallback;
- cache invalidation;
- local browser server startup;
- token rejection;
- browser selection updating the active Theme Model session.

### Acceptance criteria

- [ ] `go test ./...` passes.
- [ ] Tests do not require opening a real browser.
- [ ] Tests do not require a real terminal image protocol.
- [ ] Tests prove preview does not bypass Theme Model validation.
- [ ] Tests prove export remains possible without browser preview.
- [ ] Tests prove browser preview is local-only.

## Out Of Sprint

- recipe export/import;
- reproducible archives;
- generation history;
- mouse support in TUI;
- manual color editing;
- real Omarchy component screenshots;
- launching Waybar/Ghostty/Hyprlock for preview;
- theme import/merge;
- publishing to remote services.
