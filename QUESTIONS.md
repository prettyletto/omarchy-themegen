# QUESTIONS

This file contains only unresolved questions that affect implementation architecture or product behavior. Resolved decisions belong in the relevant spec files.

## Critical

1. What exact Aether/AetherNvim palette contract is installed and how should `neovim.lua` be generated for it?

This must be answered by inspecting the installed Neovim/Aether setup during implementation. Do not generate a universal compatibility file.

2. What terminal image protocols are available in the supported Omarchy terminal environment?

This affects TUI preview implementation only. PNG preview generation remains required regardless.

3. Should browser preview be implemented as static files, a local server, or a hybrid?

Constraints already decided:

- optional;
- local-only;
- one-time token;
- user-confirmed open;
- interactive selection allowed;
- TUI remains complete without it.

## Nice To Know

1. Does omitting `vscode.json` produce any noisy Omarchy behavior?

VS Code remains excluded either way. This only affects warnings/documentation.

2. Is `.omarchy-theme.yml` used by any external theme marketplace/tooling the project wants to care about?

Local Omarchy did not show a consumer. Exclude unless a real target appears.
