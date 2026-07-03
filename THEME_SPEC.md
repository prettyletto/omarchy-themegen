# THEME SPEC

## Purpose

The Theme Model is the single exportable result of generation and selection. It is independent of file formats and UI state.

## Theme Direction

A Theme Direction is one generated candidate before final selection. The app always produces five.

Required direction data:

- direction id: `1`, `2`, or `3`;
- optional descriptive label;
- generation seed/style option;
- source image fingerprint;
- palette;
- semantic roles;
- preview assets or preview render inputs;
- surface variant values for supported Surface Groups.

Direction labels are selection aids only. They must not affect export.

## Theme Model

The Theme Model is created after whole-theme selection or component-mix composition.

Required data:

- user-provided theme name;
- normalized export name;
- generator version;
- source image path;
- source image fingerprint;
- source image validation results;
- selected mode: whole-theme or component-mix;
- selection provenance;
- final palette;
- final semantic roles;
- selected wallpaper asset;
- generated preview assets;
- generated README content;
- validation results.

For component-mix mode, selection provenance must include:

- selected Surface Groups;
- selected source direction per group;
- per-surface overrides;
- source direction per override;
- composition warnings.

## Color Contract

The Theme Model must emit Omarchy's observed `colors.toml` contract:

- `accent`
- `cursor`
- `foreground`
- `background`
- `selection_foreground`
- `selection_background`
- `color0` through `color15`

Every color must be `#RRGGBB`.

## Semantic Roles

Semantic roles are internal names used to produce the color contract and previews:

- background;
- foreground;
- accent;
- muted;
- border;
- selection background;
- selection foreground;
- error;
- warning;
- success;
- link/focus.

Semantic roles are internal only. Users cannot edit them manually.

## Surface Values

Surface values are not plugins. They are fixed slots in the Theme Model used by composition and preview.

Supported slots:

- wallpaper/background;
- preview assets;
- terminal palette;
- desktop shell colors;
- lock/unlock visuals;
- notification colors;
- btop colors;
- Neovim/Aether palette;
- fixed/default icons;
- light mode marker;
- keyboard RGB from accent.

VS Code is excluded.

## Component-Mix Composition Contract (Contract A)

Component-mix mode produces one final `colors.toml` by merging color roles from group-selected directions:

- **Assets And System** group (master): provides `background`, `foreground`, `selection_foreground`, `selection_background`
- **Terminals And TUI** group: provides `cursor`, `color0`–`color15` (terminal palette)
- **Desktop Shell** group: provides `accent`
- **Editor** group: no color override (Neovim consumes full palette)

If all groups select the same direction, the result is identical to whole-theme mode. Per-surface overrides win over group selections for their specific color roles.

## Generated Files From Theme Model

Directly generated:

- `colors.toml`;
- `backgrounds/<image>`;
- `preview.png`;
- `preview-unlock.png`;
- `unlock.png`;
- `neovim.lua`;
- `icons.theme` when using a fixed/default value;
- `light.mode` for explicit light themes;
- `README.md`.

Delegated to installed Omarchy templates when available:

- terminal configs;
- Waybar;
- Hyprland;
- Hyprlock;
- Mako;
- Walker;
- SwayOSD;
- btop;
- Chromium;
- Helix;
- gum;
- Obsidian;
- keyboard RGB.

## Recipes

A recipe is an explicit reproducibility artifact, not history.

Recipe data must include:

- source image fingerprint;
- generator version;
- generation seed/style option;
- selected mode;
- selected directions/groups/surfaces;
- theme name if the user chooses to include it.

Reproducible/shareable recipes may bundle source image bytes after confirmation and privacy warning.

## Light Mode

Light mode exists only when the user explicitly requested light theme generation and the final result passes validation.

For light themes, export writes `light.mode` because local Omarchy uses that file to select GNOME `prefer-light` and `Adwaita`.

## Validation Requirements

A Theme Model is exportable only when:

- source image validation passed;
- all `colors.toml` keys are present;
- all colors are valid hex RGB;
- contrast checks pass;
- required assets can be generated;
- component-mix composition resolved conflicts;
- selected Aether/Neovim contract is known or Neovim generation fails clearly.

## Explicitly Not Modeled

- arbitrary app settings;
- CSS selectors;
- user-editable color roles;
- community theme extras;
- VS Code;
- multiple source images;
- batch jobs;
- generation history.
