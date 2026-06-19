---
title: Create cards and documents in the web UI
type: design
status: archived
created: "2026-06-12T12:33:30.737938Z"
updated: "2026-06-12T20:44:03.032277Z"
tasks:
    - T-CREATE-CARDS-IN-THE
    - T-CREATE-DOCS-IN-THE
---
# Create cards and documents in the web UI

Add web-UI affordances to create a new kanban **card** and a new **document**,
which are currently CLI-only. Both reuse the existing split-pane editor
(Markdown source + server-rendered live preview) introduced for inline editing.

This design covers two tasks: `T-CREATE-CARDS-IN-THE` (cards) and
`T-CREATE-DOCS-IN-THE` (documents).

## Background

- `store.CreateTask(title, body, columnName)` auto-generates the task code from
  the title slug and appends the card to the end of the named column (empty name
  = first column). No caller-supplied code.
- `store.CreateDoc(code, title, docType, body)` requires a caller-supplied code
  (normalized to `DOC-<CODE>`, must be unique), a title, a valid type, and body.
- Markdown (including custom `:::` directives) renders **server-side** only, so a
  live preview must round-trip to the server.

## Card creation (board modal)

- A **"New card"** button sits near the board heading on `/board`.
- Clicking it opens the **existing card modal** with a create-mode split-pane
  editor (title `<input>` + Markdown source/preview, Save/Cancel), fetched via
  **`GET /task/new?fragment=1`** — mirroring how an existing card opens.
- **`POST /task`** with JSON `{title, body}`:
  - validates non-empty title after trimming (else 400);
  - creates in the **first column** (To Do) via `store.CreateTask` (code
    auto-generated from the title);
  - returns 201.
- On success the board **reloads** so the new card appears in its column,
  consistent with how move/archive/delete already trigger a reload.
- Cancel / Escape / backdrop close the modal, reusing the editor's dirty-confirm
  guard.

## Document creation (dedicated page)

- A **"New document"** button sits on the docs index page (`/`).
- It navigates to **`GET /doc/new`**, a full page with a create form reusing the
  split-pane editor: **code** `<input>`, **title** `<input>`, **type**
  `<select>` (over `store.DocTypes`), and Markdown **body**/preview.
- **Code auto-suggest**: as the user types the title, JS fills the code field
  from a slug of the title (e.g. "Auth design" -> `AUTH-DESIGN`: uppercase,
  non-alphanumerics to hyphens, collapse repeats, trim). Auto-fill stops once the
  user edits the code field themselves, after which the code stays manual.
- **`POST /doc`** with JSON `{code, title, type, body}`:
  - validates non-empty title and a valid type (else 400);
  - a malformed code is 400; a **duplicate** code is **409**;
  - on success returns the created doc's code (JSON) and the page **redirects to
    `/doc/{code}`**.

## Shared preview endpoint

Preview is **stateless** — it only renders the posted body Markdown to HTML and
never touches the store. The create forms have no code yet, so a per-code
preview route does not fit. Consolidate to a single **`POST /preview`** with
JSON `{body}` -> rendered HTML, and point **both create and edit** forms at it.
This **replaces** the `POST /task/{code}/preview` and `POST /doc/{code}/preview`
routes added for inline editing (net: one preview endpoint instead of three).

## Files touched

- `internal/web/server.go` — add `GET /task/new`, `POST /task`, `GET /doc/new`,
  `POST /doc`, `POST /preview`; remove `POST /task/{code}/preview` and
  `POST /doc/{code}/preview`.
- `internal/web/pages.go` — add `handleNewTaskForm`, `handleCreateTask`,
  `handleNewDocForm`, `handleCreateDoc`, `handlePreview`; remove
  `handlePreviewTask` and `handlePreviewDoc`. Reuse `decodeJSON` and
  `writePreview`.
- `internal/web/templates/tasknew.gohtml` — create-card editor fragment
  (rendered into the modal).
- `internal/web/templates/docnew.gohtml` — create-document full page.
- `internal/web/templates/board.gohtml` — "New card" button + open/save JS;
  repoint the editor preview fetch to `/preview`.
- `internal/web/templates/index.gohtml` — "New document" button.
- `internal/web/templates/doc.gohtml` — repoint the editor preview fetch to
  `/preview`.
- `internal/web/static/style.css` — minor button styling; reuse editor classes.
- `internal/web/server_test.go` — tests below.

## Testing

- `POST /task` creates a card in the first column with the given title/body;
  rejects empty title (400); returns 201.
- `GET /task/new?fragment=1` returns the create editor fragment (no layout
  chrome).
- `POST /doc` creates a doc; rejects empty title and invalid type (400); returns
  409 for a duplicate code; succeeds for a fresh code.
- `GET /doc/new` returns the create page.
- `POST /preview` renders Markdown + a `:::` directive and never mutates the
  store (data-version unchanged).
