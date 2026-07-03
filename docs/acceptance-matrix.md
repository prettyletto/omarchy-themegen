# Acceptance Test Matrix

## Image Validation

| Test | Expected | Check Type |
|------|----------|--------|
| Valid opaque image >= 800x450 | Passes validation | Automated |
| Transparent image | Rejected | Automated |
| Animated image | Rejected | Automated |
| Image < 800x450 | Rejected | Automated |
| Non-image file | Rejected | Automated |
| UI-heavy screenshot | Warning, non-blocking | Automated |
| Multiple images | Rejected | Automated |

## Generation

| Test | Expected | Check Type |
|------|----------|--------|
| Exactly 5 Theme Directions | Produced | Automated |
| Deterministic output | Same image+seed = same result | Automated |
| Different seed | Potentially different output | Automated |
| Light mode | 5 light directions, light.mode written | Automated |
| Dark mode default | 5 dark directions, no light.mode | Automated |

## Whole-Theme

| Test | Expected | Check Type |
|------|----------|--------|
| TUI whole-theme flow | Validating → Mode Select → Comparison → Naming → Confirm → Export → Result → Apply | Manual |
| CLI whole-theme export | `--direction 1-5` works | Automated |
| CLI missing direction | Clear error | Automated |

## Component-Mix

| Test | Expected | Check Type |
|------|----------|--------|
| TUI component-mix flow | Mode Select → Group Assign → Overrides → Naming → Export | Manual |
| CLI component-mix export | `--group_*` flags work | Automated |
| CLI missing groups | Clear error | Automated |
| Group reset | `r` key resets all to direction 1 | Automated |
| Per-surface overrides | Optional, visible before export | Automated |

## Previews

| Test | Expected | Check Type |
|------|----------|--------|
| Terminal image capability | Detected or fallback | Automated |
| TUI text/ANSI preview | Visible without terminal image support | Automated |
| Browser preview | `b` key, local-only, token-required | Manual |
| Browser selection sync | Browser changes update TUI | Manual |

## Export

| Test | Expected | Check Type |
|------|----------|--------|
| Theme directory | `~/.config/omarchy/themes/<name>` created | Automated |
| Required files | colors.toml, backgrounds/, preview.png, neovim.lua, README.md | Automated |
| Overwrite refusal | Fails by default | Automated |
| Overwrite with backup | Timestamped backup created | Automated |
| Finished-theme archive | tar.gz with correct root | Automated |
| Reproducible archive | tar.gz with theme + recipe + source image | Automated |
| light.mode | Written only for explicit light themes | Automated |

## Recipe

| Test | Expected | Check Type |
|------|----------|--------|
| Recipe export | JSON with provenance, no source bytes | Automated |
| Recipe replay | Same image reproduces same theme | Automated |
| Fingerprint mismatch | Clear error with both fingerprints | Automated |
| Force fingerprint | `--force-fingerprint` bypasses check | Automated |

## Validation

| Test | Expected | Check Type |
|------|----------|--------|
| Pre-export validation | Blocks on missing keys/bad format | Automated |
| Post-export validation | Checks files, dimensions, keys | Automated |
| Omarchy present | High/medium confidence | Automated |
| Omarchy absent | Reduced confidence, clear message | Automated |
| Aether contract known | neovim.lua generated | Automated |
| Aether contract unknown | Clear error before export | Automated |

## Apply

| Test | Expected | Check Type |
|------|----------|--------|
| Apply separate from export | Never triggered by export | Manual |
| Apply confirmation | Explicit confirmation required | Manual |
| Omarchy missing | Apply unavailable, clear message | Automated |
| Apply success | Theme applied, components restart | Manual |
| Apply failure | Error shown, theme not applied | Manual |
| Rollback | Known-good theme re-applied | Manual |

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
| Generated theme license/gitignore | Excluded |
