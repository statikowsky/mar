---
title: Archive / Unarchive / Delete board cards
type: design
status: archived
created: "2026-06-10T12:29:14.789973Z"
updated: "2026-06-10T21:11:10.68725Z"
tasks:
    - T-CLI-TASK-ARCHIVE-UNARCHIVE
    - T-STORE-TASK-STATUS-COLUMN
    - T-WEB-TASK-ARCHIVE-UNARCHIVE
    - T-WEB-UI-BOARD-ARCHIVED
---
# Archive / Unarchive / Delete board cards

Design for managing task (card) lifecycle from both the CLI and the web UI,
mirroring the document archive feature (DOC-ARCHIVE).

## Background

Documents already support archive/unarchive/delete on both surfaces. Tasks do
not: `task rm --force` deletes from the CLI, but there is no archive concept and
no web actions. Unlike docs, tasks have no `status` column — they live in
kanban columns — so archiving requires a schema change.

## Goals

- Add a `status` column to tasks and migrate existing databases.
- Add `ArchiveTask` / `UnarchiveTask` store methods and `task archive` /
  `task unarchive` CLI commands.
- Filter archived cards out of the board and the default task listings.
- Add web actions: Archive in the task modal/detail, Unarchive + Delete in a
  collapsible Archived section below the board.
- Make deletion safe: only once a card is archived, confirmed in the browser.

## 1. Store layer — schema + migration

- Add `status TEXT NOT NULL DEFAULT 'active'` to the `tasks` table DDL.
- **Migration for existing DBs:** `migrate()` runs `CREATE TABLE IF NOT EXISTS`,
  which does not alter existing tables. After the create block, query
  `PRAGMA table_info(tasks)` and run `ALTER TABLE tasks ADD COLUMN status TEXT
  NOT NULL DEFAULT 'active'` only if the column is absent. Idempotent across
  re-open. Existing tasks default to `active`.
- New methods `ArchiveTask(code)` and `UnarchiveTask(code)` mirror the doc
  versions: `UPDATE tasks SET status=?, updated_at=? WHERE id=?`. Unknown code
  returns `ErrNotFound`. `DeleteTask` already exists, unchanged.
- `scanTask` and every task query (`GetTask`, `ListTasks`, `tasksInColumn`, the
  link lookups in `link.go`) gain the `status` column. The `Task` struct gets a
  `Status` field with a `json:"status"` tag.
- **Board filtering:** `Board()` / `tasksInColumn` return only `active` tasks. A
  new `ArchivedTasks()` returns archived ones for the section below the board.

## 2. CLI parity

- **`task archive T-CODE [--json]`** and **`task unarchive T-CODE [--json]`** —
  new commands mirroring `doc archive` / `doc unarchive`.
  - JSON: `{"archived": true, "code": ...}` / `{"unarchived": true, "code": ...}`
- `task rm --force` is unchanged.
- **`task list`** gains a `--status active|archived` filter, defaulting to
  active only (matches `doc list`). Keeps `task list` and `board show` from
  surprising callers with archived cards.
- **`board show`** shows only active cards.

## 3. Web endpoints

Three new POST routes mirroring the doc routes:

- `POST /task/{code}/archive` -> `ArchiveTask`
- `POST /task/{code}/unarchive` -> `UnarchiveTask`
- `POST /task/{code}/delete` -> `DeleteTask`

Behavior:

- `404` if the card is missing, `200` on success. No request body.
- **Safety guard:** `/task/{code}/delete` returns `409 Conflict` unless the card
  is already archived, enforcing archive-first server-side.
- `PRAGMA data_version` auto-bumps on write, so other tabs live-reload.

## 4. Web UI

### Board page (`/board`)

- The board renders active cards only (store filter). Below the columns, add a
  collapsible **Archived (N)** `<details>` section, rendered only when archived
  cards exist, closed by default.
- Each archived row: code, title, its column name, **Unarchive** + **Delete**
  buttons. These rows are not draggable.
- `handleBoard` additionally calls `store.ArchivedTasks()` and passes them to
  the template.

### Active-card actions

Active cards are draggable links that open a modal. To avoid click/drag
conflicts and board clutter, the **Archive** action lives in the task
detail/modal, not on the card itself:

- Task detail page + modal fragment (`/task/{code}`): an actions row — Archive
  (if active), or Unarchive + Delete (if archived) — mirroring the doc detail
  page. The modal loads this fragment, so actions appear there too.

### Interaction

A small JS handler (same shape as the docs one) POSTs to the right endpoint and
reloads on success. Reused by the board's archived section and the task
fragment. **Delete** triggers a `confirm()`; on the task detail page, a
successful delete redirects to `/board`.

## 5. Testing

TDD at each layer; `make check` (fmt + vet + test-race) must pass before each
commit.

### Store

- Migration adds the `status` column on a fresh DB and is idempotent across
  re-open (the `PRAGMA table_info` guard does not error or duplicate).
- `ArchiveTask` / `UnarchiveTask` flip status and bump `updated_at`; unknown
  code returns `ErrNotFound`.
- `Board()` excludes archived cards; `ArchivedTasks()` returns only archived.

### CLI

- `task archive` / `unarchive --json` emit the status objects; round-trip
  returns the card to active.
- `task list` excludes archived by default; `--status archived` shows them.

### Web

- `POST /task/{code}/archive|unarchive` change status; an archived card leaves
  the board and appears in the archived list.
- `POST /task/{code}/delete` -> 409 if active, 200 if archived; missing -> 404.
- Board renders the Archived `<details>` only when archived cards exist; the
  task fragment shows the right actions per status.
