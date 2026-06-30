# RISKS

## Technical Risks

### Omarchy Drift

Risk: installed Omarchy templates or scripts change and generated themes stop matching real behavior.

Reduction: prefer installed Omarchy templates and script behavior when available. Keep fallback generation limited to the documented contract and report reduced confidence without Omarchy.

### Theme Model Becomes Too Broad

Risk: the Theme Model turns into a generic app-settings object.

Reduction: keep it limited to Omarchy's color contract, required assets, fixed Surface slots, and selection provenance.

### Component-Mix Incoherence

Risk: cross-direction surface mixing creates unreadable or visually incoherent themes.

Reduction: all mixes resolve into one Theme Model, regenerate previews, and rerun validation before export.

### Neovim/Aether Drift

Risk: generated `neovim.lua` targets the wrong Aether palette contract.

Reduction: detect the installed contract and target only that contract. Fail clearly when detection fails.

### Missing Dependencies

Risk: `magick` or Omarchy is missing.

Reduction: validate `magick` before image/preview operations. Allow generation/archive export without Omarchy but report reduced confidence and disable apply.

### Preview Mismatch

Risk: preview looks good while real Omarchy output is poor.

Reduction: previews are representative only. Real apply validation remains required before claims are treated as final.

## UX Risks

### Too Much Choice

Risk: all-surface component-mix overwhelms non-designers.

Reduction: present Surface Groups first, keep whole-theme mode simple, and make per-surface overrides a second step.

### Browser Preview Scope Creep

Risk: browser preview becomes a second application with independent state.

Reduction: browser preview is optional, local-only, tokenized, and tied to one active selection session. Export consumes the Theme Model, not browser-local state.

### Unsafe Actions

Risk: users overwrite or apply themes accidentally.

Reduction: overwrite is refused by default, replacement creates backups, apply is separate, and `--yes` cannot imply apply or browser open.

### TUI Complexity

Risk: selection, preview, export, recipe, and apply flows become hard to navigate.

Reduction: keyboard-only operation is mandatory. Missing non-interactive CLI options fail clearly instead of falling into ambiguous prompts.

## Maintenance Risks

### Premature Plugin Architecture

Risk: Theme Surfaces become plugins/exporters before there are real extension requirements.

Reduction: fixed Surface slots only. No plugin system. No per-surface exporters.

### Duplicating Omarchy

Risk: the app reimplements Omarchy template logic and diverges.

Reduction: write `colors.toml` and direct files only where needed. Let installed Omarchy templates generate supported Theme Surface configs.

### Recipe And Archive Confusion

Risk: recipes become hidden history or archives become a second output format.

Reduction: recipes are explicit artifacts. Archives contain the same theme directory content, with reproducible extras only after opt-in.

## Scope Risks

### Product Becomes A Theme IDE

Risk: importing themes, editing colors, editing app settings, and supporting community extras expand the product beyond image-to-theme.

Reduction: no theme import, no manual color editing, no arbitrary app settings, no community extras unless explicitly brought into scope later.

### Publishing Creep

Risk: share/export becomes GitHub publishing and license management.

Reduction: stop at a theme-repo-compatible folder/archive. Users publish themselves.

## Performance Risks

### Large Image Processing

Risk: large wallpapers make generation slow.

Reduction: downsample for analysis while preserving original image for export.

### Preview Rendering Cost

Risk: PNG/browser previews slow down selection.

Reduction: cache rendered previews by image fingerprint, generator version, seed/style option, and Theme Model identity. Avoid per-app screenshots.
