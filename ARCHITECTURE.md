# ARCHITECTURE

## Architectural Position

`omarchy-themegen` is a local Go CLI/TUI tool. Omarchy remains the external system that applies themes.

The application has one durable output shape: an Omarchy theme directory. Archives and recipes are secondary artifacts derived from the same validated composition.

## Core Flow

1. Validate one opaque still source image.
2. Generate exactly three Theme Directions.
3. Let the user select a whole direction or compose a component mix.
4. Resolve the selection into one Theme Model.
5. Render previews from that Theme Model.
6. Structurally validate the Theme Model and export target.
7. Write the Omarchy Theme Directory and optional archive/recipe.
8. Optionally apply by delegating to Omarchy after separate confirmation.

## Main Areas Of Responsibility

### Input Validation

Accepts one image path and validates:

- readable file;
- still image;
- opaque image;
- minimum dimensions;
- source fingerprint;
- UI-heavy screenshot warning.

It does not extract palettes or write theme files.

### Theme Generation

Turns the image into exactly three Theme Directions.

It owns:

- image-derived palette candidates;
- semantic color assignment;
- terminal color contract;
- direction labels;
- deterministic seed/style option handling;
- light-theme request handling.

It does not know output paths or Omarchy file layout.

### Composition

Turns user choices into one Theme Model.

It owns:

- whole-theme selection;
- Surface Group selection;
- per-surface overrides;
- cross-direction mixing;
- selection provenance;
- conflict warnings.

Composition is the key architectural seam. TUI, browser preview, CLI flags, and recipe files all feed the same composition rules. Export never consumes UI state directly.

### Preview

Renders the current Theme Model for decision support.

It owns:

- TUI/terminal preview;
- PNG preview assets;
- optional local browser companion preview;
- browser/TUI selection synchronization for the active session.

Previews are representative, not authoritative. Export correctness comes from the Theme Model and validation.

### Export

Writes artifacts from one validated Theme Model.

It owns:

- Omarchy theme directory export;
- archive export;
- recipe export;
- README generation;
- overwrite backup policy;
- structural post-write validation.

It does not run `omarchy theme set`.

### Apply

Delegates to Omarchy after export and explicit confirmation.

It owns only:

- presenting the apply consequence;
- invoking Omarchy;
- surfacing errors.

It does not implement Omarchy's theme switching behavior.

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
