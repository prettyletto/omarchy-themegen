# QUESTIONS

This file contains only unresolved questions that affect implementation architecture or product behavior. Resolved decisions belong in the relevant spec files.

## Critical

1. ~~What exact Aether/AetherNvim palette contract is installed and how should `neovim.lua` be generated for it?~~
   **RESOLVED (Sprint 7)**: Standard Aether contract detected via filesystem inspection of
   `~/.local/share/nvim/lazy/aether.nvim` and similar paths. `neovim.lua` generated with the
   standard 22-key palette (accent, cursor, foreground, background, color0-color15). Unknown
   contract fails clearly before export with diagnostic paths.

2. ~~What terminal image protocols are available in the supported Omarchy terminal environment?~~
   **RESOLVED (Sprint 5)**: Runtime detection supports Kitty protocol (KITTY_WINDOW_ID, WezTerm,
   Ghostty), iTerm2 protocol (TERM_PROGRAM, imgcat), and Sixel (foot/xterm). Falls back to ANSI
   color swatches when unsupported.

3. ~~Should browser preview be implemented as static files, a local server, or a hybrid?~~
   **RESOLVED (Sprint 5)**: Local HTTP server on `127.0.0.1:0` with 16-byte hex session token,
   5-minute idle timeout. Browser selections sync back to TUI via `tea.Program.Send()`.

## Nice To Know

1. Does omitting `vscode.json` produce any noisy Omarchy behavior?
   VS Code remains excluded either way. This only affects warnings/documentation.

2. Is `.omarchy-theme.yml` used by any external theme marketplace/tooling the project wants to care about?
   Local Omarchy did not show a consumer. Exclude unless a real target appears.
