---
title: Task positional placement plan
type: plan
status: active
created: "2026-06-27T11:56:16.280183968Z"
updated: "2026-06-27T11:56:45.238205589Z"
tasks:
    - T-ADD-POSITIONAL-TASK-PLACEMENT
---
# Plan

## Goal

Add explicit positional placement for task create and move so users can place cards before the first item, at the top or bottom, after a known task, before a known task, or by a one-based index without hand-editing .mar/board.yml.

## Proposed CLI

- `mar task move T-CODE --column NAME [--after T-OTHER | --before T-OTHER | --first | --last | --index N]`.
- `mar task create --title T [--column NAME] [--code X] [--body -|FILE] [--after T-OTHER | --before T-OTHER | --first | --last | --index N]`.
- Keep existing move behavior: no placement flag means top, preserving drag-and-drop/web behavior and documented CLI behavior.
- Keep existing create behavior: no placement flag appends to the selected column.
- Reject multiple placement flags in one command with a clear error.
- Treat `--index` as one-based, clamping is avoided: index must be between 1 and len(column)+1 for create or move insertion after removal.

## Store shape

- Introduce a small placement type in `internal/store` instead of adding more positional pointer arguments.
- Rework creation and movement through shared placement resolution helpers so `after`, `before`, `first`, `last`, and `index` share validation.
- For move, resolve placement after removing the moved card from its previous location so moving within the same column is intuitive.
- Ensure `after` and `before` target tasks must be active and in the target column.

## Tests

- Store tests for create before, create first, create after, create index, move before, move last, move index, invalid index, and target-in-other-column failures.
- CLI tests for new flags, conflict validation, JSON output shape, and backwards compatibility for existing create/move commands.
- Update guide and README examples so agent and human docs match the new CLI.

## Compatibility

- Existing public methods remain available where practical; new option-style methods can sit alongside them if that keeps callers simple.
- Web drag-and-drop can continue to send `after` only; no web UI change is required for this feature.
