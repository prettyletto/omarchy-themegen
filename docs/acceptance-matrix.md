# Acceptance Test Matrix

## Image Validation

| Test | Expected | Status |
|------|----------|--------|
| Valid opaque image >= 800x450 | Passes validation | AFK |
| Transparent image | Rejected | AFK |
| Animated image | Rejected | AFK |
| Image < 800x450 | Rejected | AFK |
| Non-image file | Rejected | AFK |
| UI-heavy screenshot | Warning, non-blocking | AFK |
| Multiple images | Rejected | AFK |

## Generation

| Test | Expected | Status |
|------|----------|--------|
| Exactly 3 Theme Directions | Produced | AFK |
| Deterministic output | Same image+seed = same result | AFK |
| Different seed | Potentially different output | AFK |
| Light mode | 3 light directions, light.mode written | AFK |
| Dark mode default | 3 dark directions, no light.mode | AFK |

## Whole-Theme

| Test | Expected | Status |
|------|----------|--------|
| TUI whole-theme flow | Validating → Mode Select → Comparison → Naming → Confirm → Export → Result → Apply | HITL |
| CLI whole-theme export | `--direction 1-3` works | AFK |
| CLI missing direction | Clear error | AFK |

## Component-Mix

| Test | Expected | Status |
|------|----------|--------|
| TUI component-mix flow | Mode Select → Group Assign → Overrides → Naming → Export | HITL |
| CLI component-mix export | `--group_*` flags work | AFK |
| CLI missing groups | Clear error | AFK |
| Group reset | `r` key resets all to direction 1 | AFK |
| Per-surface overrides | Optional, visible before export | AFK |

## Previews

| Test | Expected | Status |
|------|----------|--------|
| Terminal image capability | Detected or fallback | AFK |
| TUI text/ANSI preview | Visible without terminal image support | AFK |
| Browser preview | `b` key, local-only, token-required | HITL |
| Browser selection sync | Browser changes update TUI | HITL |

## Export

| Test | Expected | Status |
|------|----------|--------|
| Theme directory | `~/.config/omarchy/themes/<name>` created | AFK |
| Required files | colors.toml, backgrounds/, preview.png, neovim.lua, README.md | AFK |
| Overwrite refusal | Fails by default | AFK |
| Overwrite with backup | Timestamped backup created | AFK |
| Finished-theme archive | tar.gz with correct root | AFK |
| Reproducible archive | tar.gz with theme + recipe + source image | AFK |
| light.mode | Written only for explicit light themes | AFK |

## Recipe

| Test | Expected | Status |
|------|----------|--------|
| Recipe export | JSON with provenance, no source bytes | AFK |
| Recipe replay | Same image reproduces same theme | AFK |
| Fingerprint mismatch | Clear error with both fingerprints | AFK |
| Force fingerprint | `--force-fingerprint` bypasses check | AFK |

## Validation

| Test | Expected | Status |
|------|----------|--------|
| Pre-export validation | Blocks on missing keys/bad format | AFK |
| Post-export validation | Checks files, dimensions, keys | AFK |
| Omarchy present | High/medium confidence | AFK |
| Omarchy absent | Reduced confidence, clear message | AFK |
| Aether contract known | neovim.lua generated | AFK |
| Aether contract unknown | Clear error before export | AFK |

## Apply

| Test | Expected | Status |
|------|----------|--------|
| Apply separate from export | Never triggered by export | HITL |
| Apply confirmation | Explicit confirmation required | HITL |
| Omarchy missing | Apply unavailable, clear message | AFK |
| Apply success | Theme applied, components restart | HITL |
| Apply failure | Error shown, theme not applied | HITL |
| Rollback | Known-good theme re-applied | HITL |

## Non-Goals (Verified Absent)

| Feature | Status |
|---------|--------|
| VS Code generation | Excluded |
| Manual color editing | Excluded |
| Theme import/merge | Excluded |
| Multiple source images | Excluded |
| Mouse support in TUI | Excluded |
| Remote browser preview | Excluded |
| GitHub publishing | Excluded |
| License/gitignore generation | Excluded |
