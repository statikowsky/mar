---
title: Scratchpad MVP implementation plan
type: plan
status: active
created: "2026-07-11T18:03:33.353656Z"
updated: "2026-07-11T18:03:33.353656Z"
tasks:
    - T-IMPLEMENT-SCRATCHPAD-MVP
---
## Objective

Implement the vertical Scratchpad MVP described in [[DOC-SCRATCHPAD]] without introducing third-party browser dependencies.

## Slices

1. Add a versioned `.mar/scratchpad.yml` store model with stable `S-N` IDs, validation, atomic writes, optimistic version checks, and table-driven tests.
2. Add `mar scratch show|add|edit|rm` with JSON output and CLI tests.
3. Add `/scratchpad` plus JSON mutation endpoints, navigation links, CSRF protection through the existing middleware, and handler tests.
4. Build DOM-based spatial notes with create/edit, drag, resize, colors, selection, duplicate/delete, autosave state, undo/redo, and pan/zoom/fit controls.
5. Add an accessible list view and keyboard operations, plus touch-compatible pointer events.
6. Add note promotion into task/document creation flows and persisted links.
7. Update README and `mar guide`, run doc lint and `make check`, visually verify desktop/mobile behavior, and commit.

## Scope decisions

- One scratchpad per repository.
- Plain multiline text with mar wiki-link rendering; no rich text.
- One atomic YAML snapshot; viewport stays in browser local storage.
- Stale mutation versions return HTTP 409 and never overwrite store changes.
- Freehand drawing, shapes, connectors, images, multiple boards, and collaboration remain outside this implementation.

## Verification

- Store and CLI tests cover empty state, CRUD, validation, persistence, stable IDs, and stale versions.
- Web tests cover page rendering, mutation routes, conflicts, and promotion.
- Browser checks cover desktop spatial interaction and compact list view.
- `make check` passes.
