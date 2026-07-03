# PROJECT

## Vision

`omarchy-themegen` turns one opaque still image into a complete, installable Omarchy theme.

The product is complete when a user can:

1. Run `omarchy-themegen <image>` and enter a keyboard-only TUI.
2. Generate five valid theme directions from that one image.
3. Choose either a whole direction or a controlled component-mix composition.
4. Preview the final composition in the terminal/TUI, with optional local browser preview.
5. Provide a theme name explicitly.
6. Export a self-contained Omarchy theme directory, and optionally a shareable archive.
7. Apply the exported theme only after a separate confirmation.

## Known Omarchy Facts

- User themes live under `~/.config/omarchy/themes/<theme-name>`.
- Official themes live under `~/.local/share/omarchy/themes/<theme-name>`.
- `omarchy-theme-set` normalizes names, copies theme files, renders templates from `colors.toml`, swaps `~/.config/omarchy/current/theme`, changes background, restarts themed components, runs app-specific theme hooks, and calls the `theme-set` hook.
- Installed Omarchy templates cover Alacritty, btop, Chromium, Foot, Ghostty, gum, Helix, Hyprland preview share picker, Hyprland, Hyprlock, keyboard RGB, Kitty, Mako, Obsidian, SwayOSD, Walker, and Waybar.
- Local Omarchy consumers exist for `preview.png`, `preview-unlock.png`, `unlock.png`, `colors.toml`, `icons.theme`, `light.mode`, and `keyboard.rgb`.
- Local Neovim loads `~/.config/nvim/lua/plugins/theme.lua`, symlinked to `~/.config/omarchy/current/theme/neovim.lua`.
- Observed Neovim themes use `bjarneo/aether.nvim` with injected palette values.

## Users

- Omarchy users who want a personalized theme from a wallpaper without learning each config format.
- Theme authors who want a complete generated starting point they can edit after export.

This is not for professional designers who need arbitrary color editing, theme import/merge, or per-widget app customization.

## Product Scope

### Input

- exactly one image per run;
- exactly one source image per generated theme;
- opaque still images only;
- reject images smaller than 800x450;
- accept screenshots/UI captures but warn when the image appears UI-heavy;
- no transparent images, animated images, video wallpapers, batch generation, or multi-wallpaper themes.

### Generation

- five theme directions;
- deterministic output for the same image and seed/style option;
- full regeneration only, never single-surface regeneration;
- explicit light-theme request; if requested, all five directions are light-theme attempts;
- no manual color editing.

### Selection

- whole-theme mode selects one complete direction;
- component-mix mode selects Surface Groups first, then allows per-surface overrides;
- component-mix may combine variants from any generated direction;
- every composition must resolve to one final Theme Model before preview or export;
- VS Code is excluded from generation, preview, and selection.

### CLI And TUI

- distributed as a single Go binary named `omarchy-themegen`;
- `go install` is a required installation path;
- `omarchy-themegen <image>` opens the TUI by default;
- the TUI must be fully keyboard-operable;
- non-interactive CLI supports the same final workflows as the TUI through explicit options or a recipe file;
- explicit non-interactive CLI fails on missing required choices;
- plain text output is default; JSON output requires an explicit flag;
- `--yes` may confirm safe file/export prompts only. It must not imply apply or browser opening.

### Preview

- terminal/TUI preview is required;
- terminal image support is detected rather than hard-coded to one protocol;
- PNG previews are generated for inspection and exported assets;
- local browser preview is optional and user-requested;
- browser preview may be interactive, but it shares the active selection session and never owns export data;
- browser preview must bind locally, use a one-time session token, ask before opening, and show the URL.

### Export

- primary output is an Omarchy theme directory;
- exported theme is self-contained and includes the source image as a background asset;
- archive export is explicit and uses the same validated theme content;
- reproducible archive additionally includes recipe and source image bytes after confirmation and privacy warning;
- recipe export is explicit and not a history system;
- README is always generated;
- README generation details are opt-in;
- no generated license, `.gitignore`, git metadata, publishing metadata, or built-in GitHub publishing.

### Install And Apply

- export target is `~/.config/omarchy/themes/<theme-name>`;
- the user must provide the theme name explicitly;
- overwrite is refused by default;
- confirmed replacement creates a timestamped backup;
- export and apply are separate actions;
- apply delegates to Omarchy after explicit confirmation;
- validation before apply is structural and non-mutating.

### Runtime Requirements

- generation is fully offline;
- no online services or AI APIs;
- ImageMagick `magick` is a runtime dependency for image processing and preview assets;
- generation and archive export can run without Omarchy installed, but validation confidence is reduced and apply is unavailable.

## Non-Goals

- generic Linux theming;
- generic exporter/plugin framework;
- importing or editing existing Omarchy themes;
- online sharing/publishing workflows;
- AI-generated wallpapers;
- manual palette editor or individual color editing;
- generation history/session database;
- mouse-dependent TUI interactions;
- arbitrary community theme extras;
- VS Code theme generation;
- direct mutation of live app configs;
- replacement of Omarchy's theme application process;
- integration into the `omarchy` command namespace by default.

## Completion Standard

The product is complete. All 17 sprints and 8 roadmap milestones have been implemented:

- One opaque still image → three Theme Directions → whole-theme or component-mix selection → validated export → optional apply.
- TUI (keyboard-only) and CLI both support complete workflows.
- Browser preview is optional, local-only, session-token-authenticated.
- Recipes and archives (finished + reproducible) work.
- Structural validation uses installed Omarchy discovery when available.
- 4,900+ lines of production Go, 150+ tests, gofmt/vet/staticcheck clean.
- v1.0.0.
