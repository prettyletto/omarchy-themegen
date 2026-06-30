# Context Glossary

This file defines project language only. It does not define implementation.

## Theme Source Image

The user-provided image or wallpaper used as the visual input for theme generation.

## Theme Direction

A complete candidate theme derived from the same source image. A direction contains a wallpaper choice, color palette, semantic color assignments, generated previews, and exportable Omarchy files.

## Theme Surface

An Omarchy-consumed surface that receives theme colors or assets, such as Waybar, Ghostty, Neovim, Hyprlock, Mako, btop, Alacritty, Foot, Kitty, Walker, SwayOSD, VS Code, Chromium, GNOME, keyboard lighting, or other files recognized by Omarchy.

## Surface Variant

A selectable visual treatment for one Theme Surface inside a Theme Direction. A Surface Variant is valid only when composition can resolve it into one exportable Theme Model.

## Selection Mode

The user's chosen workflow for building the final theme. Whole-theme mode selects one complete Theme Direction. Component-mix mode selects Surface Groups and Surface Variants, then assembles them into one exportable Theme Model.

## Surface Group

A user-facing group of related Theme Surfaces that can be selected together during component-mix mode before per-surface overrides.

## Theme Model

The internal representation of one complete generated theme. It is independent of output file formats.

## Omarchy Theme Directory

A directory that Omarchy can apply as a theme. Locally observed Omarchy installs user themes under `~/.config/omarchy/themes/<theme-name>` and applies themes through `omarchy-theme-set`.

## Export

The act of writing an Omarchy theme directory and optional archive from a selected generated theme.
