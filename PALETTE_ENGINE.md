# PALETTE ENGINE

## Purpose

The Palette Engine converts one valid opaque still image into five Theme Directions.

It works fully offline. It may use ImageMagick `magick`, but it must validate the dependency before operations that require it.

## Inputs

- one source image;
- source image fingerprint;
- generation seed/style option;
- optional explicit light-theme request.

Invalid inputs:

- transparent images;
- animated images;
- video;
- images smaller than 800x450;
- multiple images;
- batch input.

Screenshots/UI captures are accepted, but the engine should warn when the image appears UI-heavy.

## Outputs

Exactly three Theme Directions.

Each direction must include:

- palette;
- semantic roles;
- full terminal color contract;
- surface values needed by composition;
- validation warnings;
- short descriptive label when accurate, otherwise `Direction 1` / `Direction 2` / `Direction 3`.

## Required Behavior

- deterministic output for the same image, seed/style option, and light-mode request;
- visibly distinct directions;
- readable foreground/background;
- distinguishable accent/background;
- readable selection colors;
- coherent terminal colors;
- no missing Omarchy color keys.

If the user requests light generation, all five directions must be light-theme attempts. The engine must not infer light mode from image brightness alone.

## Validation

The engine rejects:

- invalid source image type;
- non-opaque image;
- image smaller than 800x450;
- palettes missing required color roles;
- palettes failing contrast checks.

The engine warns, but does not reject, when:

- source image appears UI-heavy;
- palette confidence is low;
- a descriptive direction label would be misleading.

## Algorithm Policy

Do not choose or document a specific palette extraction algorithm yet.

Algorithm choice must be driven by real wallpaper outputs and exported Omarchy themes. Implementation agents should optimize for deterministic, readable, coherent output before advanced color theory.

## Anti-Complexity Rules

- no manual color editing;
- no arbitrary direction count;
- no single-surface regeneration;
- no per-app palette scoring unless a concrete contrast failure requires it;
- no online services or AI APIs;
- no palette knobs exposed as user-facing design controls.

## Regeneration

Regeneration is full only: all five directions are regenerated with a different deterministic seed/style option.

The engine must never regenerate one Theme Surface in isolation.
