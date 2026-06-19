---
title: Browser drag-and-drop card reordering
type: design
status: archived
created: "2026-06-09T18:12:21.557428Z"
updated: "2026-06-12T19:35:27.903004Z"
tasks:
    - T-9
---
The web board (`internal/web/templates/board.gohtml`, `handleBoard`) is read-only
today — cards render in `(column, position)` order but can only be moved via the
CLI (`mar task move`). This design adds browser-side reordering: drag a card
within a column or across columns and have the new column + position persist.
It introduces MAR's **first server-side mutation route**.

> [!NOTE]
> Tracked as **T-9** on the board. Complements [[doc-modal]] (open a card in a
> modal) — both are board UX work and can share the same fetch/refresh plumbing.

## Goals

- Drag a card within its column to reorder it.
- Drag a card to a different column.
- Persist the move (column + position) so a reload and other open tabs reflect it.
- Degrade gracefully: with JS disabled the board stays the ordered, read-only view.

## Non-goals

- Reordering or renaming columns in the browser (that is CLI-only; see T-4).
- Editing card content inline (separate concern).
- Multi-select / bulk drag.

## Server side — the first mutation route

The web layer is currently read-only by design. This adds one write endpoint,
loopback-only like the rest of `serve`:

#### POST /task/{code}/move

Request body: `{"column": "In progress", "after": "T-3"}` (both optional —
omitting `after` drops the card at the top of the target column; omitting
`column` keeps the current one). Handler resolves the codes and calls the
existing `store.MoveTask(code, column, after)`, which already does gap-based
midpoint positioning. Responds `200` with the updated task JSON, or `404` /
`400` on unknown code / bad column.

Reusing `MoveTask` means the browser path and the CLI path share one ordering
implementation — no duplicate logic, and the gap-based `positionAfter` already
handles "insert between two cards" correctly.

## Client side

### Drag interactions

- Mark each `.task-card` `draggable="true"`; make each `.board-col` a drop target.
- On `dragover`, compute the drop index from pointer position relative to the
  sibling cards; show an insertion indicator.
- On `drop`, determine the `after` card (the card now above the drop point, or
  none if dropping at the top) and the target column name.

### Persist + reconcile

- POST the move to `/task/{code}/move`.
- On success, leave the optimistic DOM position in place. The server bumps
  SQLite `data_version`, so the existing `/events` SSE stream broadcasts
  `reload` and **other** open tabs refresh. The acting tab does not need a full
  reload (optimistic update already matches).
- On error, revert the card to its original position and surface a message.

Prefer a small dependency-free pointer/drag implementation over a library to
keep the single-binary, no-build-step property of the web UI.

## Edge cases

- **Drop in the same spot:** no-op; skip the POST if column and neighbours are
  unchanged.
- **Empty target column:** `after` is null, card goes to the top (`MoveTask`
  already handles the empty-column case).
- **Concurrent move from the CLI:** SSE reload converges the view; last write
  wins on position, which is acceptable for a single-user local tool.
- **Position precision exhaustion:** out of scope here; the store's gap-based
  scheme already contemplates a future rebalance.

## Security

The route is bound to `127.0.0.1` like all of `serve`. It mutates only the local
store via the same validated `MoveTask` path the CLI uses; no new trust surface
beyond "anything that can already reach the loopback CLI can reorder cards."

> [!TIP]
> **Recommendation:** ship the `POST /task/{code}/move` route + dependency-free
> drag handler in `board.gohtml`, reusing `store.MoveTask` and the existing SSE
> reload. It is a self-contained increment and the natural first write route for
> the web UI.
