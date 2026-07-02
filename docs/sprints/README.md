# Sprint Plan

This directory breaks the complete `omarchy-themegen` product into implementation sprints.

These are not MVP phases. Each sprint should leave production-quality behavior behind, and Sprint 9 is the explicit product-complete gate.

## Current Implementation Roadmap

The current codebase has moved ahead of these original sprint docs but is not product-complete. Use [product-completion-roadmap.md](../roadmaps/product-completion-roadmap.md) for agent handoff from the audited repository state.

## Sequence

1. `sprint-1.md`: static Theme Model, image validation, structural export.
2. `sprint-2.md`: offline image-derived Theme Direction generation.
3. `sprint-3.md`: keyboard-only TUI for whole-theme workflow.
4. `sprint-4.md`: component-mix workflow through Surface Groups and per-surface overrides.
5. `sprint-5.md`: terminal PNG preview, optional local browser preview, preview cache.
6. `sprint-6.md`: recipes, replay, reproducible archives, JSON automation output.
7. `sprint-7.md`: installed Omarchy validation, Aether/Neovim contract closure, real apply/rollback check.
8. `sprint-8.md`: UX, error, palette, preview, and file-safety hardening.
9. `sprint-9.md`: release readiness, acceptance matrix, scope guard, final product-complete review.

## Completion Rule

The product is not complete until Sprint 9 Task 8 passes.

No sprint should add:

- manual color editing;
- theme import or merge;
- multiple source images per run;
- generation history;
- VS Code generation;
- mouse-dependent TUI behavior;
- remote browser preview;
- built-in marketplace/GitHub publishing.
