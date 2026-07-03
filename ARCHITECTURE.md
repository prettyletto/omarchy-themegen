# ARCHITECTURE

## Architectural Position

`omarchy-themegen` is a local Go CLI/TUI tool. Omarchy remains the external system that applies themes.

The application has one durable output shape: an Omarchy theme directory. Archives and recipes are secondary artifacts derived from the same validated composition.

## Core Flow

1. Validate one opaque still source image.
2. Generate five Theme Directions.
3. Let the user select a whole direction or compose a component mix.
4. Resolve the selection into one Theme Model.
5. Render previews from that Theme Model.
6. Structurally validate the Theme Model and export target.
7. Write the Omarchy Theme Directory and optional archive/recipe.
8. Optionally apply by delegating to Omarchy after separate confirmation.

## Main Areas Of Responsibility

### Input Validation
Accepts one image path and validates via ImageMagick `magick`: readable file, still image, opaque image, minimum 800x450 dimensions, source fingerprint, UI-heavy screenshot warning.

### Theme Generation (`internal/gen/`)
Turns the image into five Theme Directions via offline palette extraction, HSL-based color math, and deterministic candidate generation. Produces foreground/background/accent/terminal colors with seed-based reproducibility.

### Composition (`internal/theme/compose.go`)
Implements Contract A: component-mix merges color roles from group-selected directions. Assets group provides master background/foreground, Terminals group provides color0-15+cursor, Desktop Shell provides accent. Per-surface overrides win over groups.

### Workflow (`internal/workflow/`)
Single entry point: `Run(Options) → Result`. Encapsulates generation → composition → validation → export → archive → recipe → reproducible archive. CLI and recipe replay both use this module.

### Preview (`internal/preview/`)
- Terminal capability detection (Kitty, iTerm2, Sixel, ANSI fallback)
- Direction and composed preview PNG generation via magick
- Preview cache keyed by fingerprint+seed+mode
- Local browser preview server with session token auth, validation, and idle timeout

### Export (`internal/export/`)
Writes Omarchy theme directory, archives, recipes, README, neovim.lua. Enforces overwrite backup policy. Supports archive-only mode (no local theme dir write).

### TUI (`internal/tui/`)
12-step keyboard-only TUI: validating → generation → mode select → comparison/group select → override select → naming → confirm → export → result → apply → done. Browser preview launch via `b` key.

### Apply
Delegates to `omarchy theme set` after separate confirmation. Uses Omarchy discovery for availability checks.

## Boundaries That Must Stay Clear

- Palette generation must not write files.
- UI state must not be exported directly.
- Browser preview must not become a second source of selection truth.
- Preview rendering must not define exported data.
- Export must not require the TUI or browser.
- Apply must not be folded into export.
- Installed Omarchy templates and local behavior are preferred when available.
- Fallback generation without Omarchy must report reduced validation confidence.

## Anti-Abstraction Rules

- Do not create a plugin system for Theme Surfaces.
- Do not create one exporter per Theme Surface.
- Do not model arbitrary app settings.
- Do not duplicate Omarchy template expansion when installed Omarchy can do it.
- Do not introduce abstractions before there are at least two real implementations.
- Do not make browser preview a web app; it is a local companion view.

## Data Ownership

- Theme Generation owns candidate directions.
- Composition owns the final Theme Model.
- Preview reads the Theme Model.
- Export reads the Theme Model.
- Validation reads generated files and the Theme Model.
- Apply reads the exported theme name/path.

No other area owns the exported truth.
