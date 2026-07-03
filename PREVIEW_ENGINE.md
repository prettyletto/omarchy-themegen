# PREVIEW ENGINE

## Purpose

Preview gives the user enough visual evidence to choose and validate a final theme composition. It is not the source of export truth.

## Required Outputs

- TUI/terminal comparison of five directions;
- TUI/terminal preview of the current composed Theme Model;
- PNG `preview.png` at 1800x1012;
- PNG `preview-unlock.png` at 1920x1080;
- PNG `unlock.png` suitable for Plymouth/SDDM use;
- optional local browser preview when the user requests it.

`unlock.png` dimensions vary across stock Omarchy themes, so validity matters more than a fixed size.

## Preview Content

Previews must show enough to compare:

- wallpaper/background;
- background color;
- foreground color;
- accent color;
- terminal palette;
- bar-like surface;
- terminal-like surface;
- notification-like surface;
- lock/unlock surface.

Previews are representative mockups. They must not claim pixel accuracy for Waybar, Ghostty, Hyprlock, or other apps.

## TUI Preview

The TUI must be complete without browser preview.

Requirements:

- keyboard-only operation;
- terminal image capability detection;
- PNG fallback when terminal image display is unavailable;
- no mouse dependency.

## Browser Preview

Browser preview is optional, local-only, and user-requested.

It may:

- render the same directions and current composition as HTML;
- allow interactive selection of directions, Surface Groups, and supported surfaces;
- synchronize changes with the active TUI session.

It must:

- bind only to a local interface;
- use an ephemeral port unless explicitly configured;
- include a one-time session token in the URL;
- ask before opening a browser;
- use `xdg-open` or equivalent only after confirmation;
- show the local URL even when opening fails or is declined;
- stop when the preview session ends;
- never own export data.

Browser and TUI share one active selection session. Export consumes the validated Theme Model, not browser state.

## Rendering Strategy

Use the simplest deterministic renderer that produces the required PNGs and browser/TUI preview surfaces.

Allowed:

- ImageMagick `magick`;
- HTML/CSS if it is the simplest reliable preview renderer;
- generated PNG files;
- terminal image protocols detected at runtime.

Avoid:

- launching real Omarchy components;
- live screenshots;
- per-surface preview tabs;
- browser-only UI;
- unsupported community extra previews.

## Caching

Cache only expensive generated artifacts:

- source image fingerprint;
- generator version;
- seed/style option;
- rendered preview PNGs.

The cache is internal. It is not a generation history or session browser.

## Validation Role

Preview may reveal readability problems, but validation must be structural and color-contract based. A good-looking preview does not prove the theme is valid.
