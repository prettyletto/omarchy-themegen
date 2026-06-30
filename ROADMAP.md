# ROADMAP

This is an implementation sequence toward the complete product. It is not an MVP/future split.

Each milestone should leave behind production-quality behavior, not throwaway scaffolding.

## Milestone 1: Omarchy Contract Lockdown

Outcome: implementation has a verified local contract to target.

Work:

- inspect installed Omarchy templates and theme scripts;
- record required generated files;
- record files delegated to Omarchy templates;
- verify local behavior for preview assets, `light.mode`, `icons.theme`, keyboard RGB, and background cycling;
- decide the Neovim/Aether contract detection strategy.

Exit criteria:

- docs contain no unanswered Omarchy behavior question that can be answered locally;
- a static hand-authored theme directory can be structurally validated.

## Milestone 2: Theme Model And Validation

Outcome: one complete Theme Model can be created, validated, and written without image analysis.

Work:

- represent the `colors.toml` contract;
- represent required assets;
- represent selection provenance;
- validate names, colors, source image metadata, and required files;
- generate a static test theme directory.

Exit criteria:

- invalid colors, missing assets, bad names, and missing dependencies produce clear errors;
- validation is structural and non-mutating.

## Milestone 3: Image-To-Directions

Outcome: one valid source image produces exactly three Theme Directions.

Work:

- validate one opaque still image;
- reject transparent, tiny, animated, and batch inputs;
- extract candidate palettes;
- produce semantic roles and terminal colors;
- generate direction labels;
- support deterministic regeneration with a seed/style option;
- support explicit light-theme generation.

Exit criteria:

- at least five representative wallpapers produce three valid directions each;
- directions are deterministic for the same input/options;
- unusable palettes are rejected before selection.

## Milestone 4: Composition

Outcome: whole-theme and component-mix workflows both produce one Theme Model.

Work:

- implement whole-theme selection;
- implement Surface Group selection;
- implement per-surface overrides;
- support cross-direction mixing;
- parse CLI flags and recipe files into the same composition path;
- reject invalid compositions.

Exit criteria:

- TUI, CLI flags, browser selections, and recipe files all resolve through the same composition rules;
- export never consumes raw UI state.

## Milestone 5: Preview

Outcome: users can inspect generated directions and the composed Theme Model.

Work:

- render terminal/TUI previews;
- detect terminal image capability;
- generate required PNG assets;
- implement optional local browser preview with one-time token;
- keep browser and TUI synchronized through one active session.

Exit criteria:

- TUI workflow is complete without browser preview;
- PNG assets match the final Theme Model;
- browser preview is local-only and optional.

## Milestone 6: Export, Archive, And Recipe

Outcome: selected themes can be written safely.

Work:

- write self-contained theme directory;
- generate README;
- enforce overwrite backup policy;
- write finished-theme archive;
- write reproducible archive when requested;
- write recipe files when requested;
- support JSON output mode.

Exit criteria:

- export never silently overwrites;
- export never silently applies;
- archives are theme-repo-compatible;
- recipes reproduce the selected composition through the CLI.

## Milestone 7: Apply And Real Omarchy Validation

Outcome: exported themes work on real Omarchy.

Work:

- apply only after separate confirmation;
- verify generated themes through Omarchy;
- verify Waybar, terminals, Hyprland, Hyprlock, Mako, btop, Neovim, background, and menus;
- verify rollback by applying a known-good theme.

Exit criteria:

- a generated theme can be exported, applied, used, and replaced;
- documentation reflects any mismatch found during real apply.

## Milestone 8: Usability Hardening

Outcome: the product feels finished for a non-designer.

Work:

- refine errors;
- refine direction labels;
- tune palette scoring from real outputs;
- test keyboard-only TUI navigation;
- test CLI missing-option behavior;
- test no-Omarchy/reduced-confidence behavior;
- test dependency failures for `magick`.

Exit criteria:

- a user can create, preview, export, and optionally apply a theme without reading Omarchy internals.
