---
title: Archive / Unarchive / Delete documents
type: design
status: archived
created: "2026-06-10T12:02:44.987545Z"
updated: "2026-06-10T21:11:10.66718Z"
tasks:
    - T-CLI-ADD-DOC-UNARCHIVE
    - T-STORE-ADD-UNARCHIVEDOC
    - T-WEB-ARCHIVE-UNARCHIVE-DELETE
    - T-WEB-UI-DOC-ACTIONS
---
# Archive / Unarchive / Delete documents

Design for managing document lifecycle from both the CLI and the web UI.

## Background

The store and CLI already support archiving (`doc archive`) and hard-deleting
(`doc rm --force`) documents. The web UI lists archived docs inline (greyed via
a CSS class) with no actions, and there is no way to restore an archived doc on
any surface. This work adds a full lifecycle: archive, restore (unarchive), and
delete, with parity between CLI and web.

## Goals

- Add a missing `UnarchiveDoc` store primitive and a `doc unarchive` CLI command.
- Add web actions (Archive / Unarchive / Delete) on both the index rows and the
  doc detail page.
- Separate archived docs into a collapsible section below the active documents.
- Make permanent deletion safe: only allowed once a doc is archived, and
  confirmed in the browser.

## 1. Store layer

- **`UnarchiveDoc(code)`** — new method mirroring `ArchiveDoc`:
  `UPDATE docs SET status='active', updated_at=? WHERE id=?`. Unknown code
  returns `ErrNotFound`.
- `ArchiveDoc` and `DeleteDoc` already exist and are reused unchanged.

## 2. CLI parity

- **`doc unarchive DOC-CODE [--json]`** — new command mirroring `doc archive`.
  Calls `store.UnarchiveDoc`.
  - Plain output: `Unarchived DOC-X`
  - JSON: `{"unarchived": true, "code": "DOC-X"}` (matches the status-object
    convention used by the other mutating commands).
- `doc archive` and `doc rm --force` are unchanged.

## 3. Web endpoints

Three new POST routes, mirroring the existing `POST /task/{code}/move` pattern
(fetch from JS, check `resp.ok`, rely on the live-reload poll to refresh):

- `POST /doc/{code}/archive` -> `ArchiveDoc`
- `POST /doc/{code}/unarchive` -> `UnarchiveDoc`
- `POST /doc/{code}/delete` -> `DeleteDoc` (hard delete)

Behavior:

- `404` if the doc is missing, `200` on success. No request body.
- **Safety guard:** `/doc/{code}/delete` returns `409 Conflict` if the doc is
  still `active` (not archived). This enforces the "delete only when archived"
  rule server-side so it cannot be bypassed by a stray request.
- SQLite's `PRAGMA data_version` auto-bumps on any write, so other open tabs
  live-reload automatically; no manual version bump is needed.

## 4. Web UI

### Index page (`/`)

Split the single Contents table into two:

- **Contents** — active docs only.
- **Archived** — a collapsible `<details>` section below, rendered only when
  archived docs exist, closed by default.
- Each row gains an action cell:
  - Active row: **Archive** button.
  - Archived row: **Unarchive** + **Delete** buttons.

`handleIndex` calls `ListDocs("", "active")` and `ListDocs("", "archived")` and
passes both `Docs` and `Archived` to the template.

Sketch:

```
Contents
  Code   Document          Type     Updated
  AUTH   Auth design       design   2026-06-03   [Archive]

> Archived (2)        <- collapsible, closed by default
    LOGIN  Old login flow  design   2026-05-01   [Unarchive] [Delete]
```

### Doc detail page (`/doc/{code}`)

An actions row near the meta line:

- If active: **Archive** button.
- If archived: an "Archived" indicator plus **Unarchive** and **Delete**
  buttons.

The template already receives `Doc` (with `.Status`), so no new handler data is
required.

### Interaction

A small inline script POSTs to the right endpoint and reloads on success, same
shape as the board's move fetch. **Delete** triggers a `confirm()` first:
"Permanently delete DOC-X? This cannot be undone."

## 5. Testing

TDD at each layer; `make check` (fmt + vet + test-race) must pass before each
commit.

### Store

- `UnarchiveDoc` flips `status` back to `active` and bumps `updated_at`.
- Unknown code returns `ErrNotFound`.

### CLI

- `doc unarchive --json` emits `{"unarchived": true, "code": ...}`.
- Round-trip archive -> unarchive returns the doc to `active` (assert via
  `doc show --json`).

### Web

Using the existing httptest harness (add a `post` helper if needed):

- `POST /doc/{code}/archive` -> doc becomes archived; index moves it into the
  Archived list.
- `POST /doc/{code}/unarchive` -> back to active.
- `POST /doc/{code}/delete` on an archived doc -> 200, row gone.
- `POST /doc/{code}/delete` on an active doc -> 409, doc still present.
- Missing code on any route -> 404.
- Index renders the Archived `<details>` section only when archived docs exist.
