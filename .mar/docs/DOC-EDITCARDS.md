---
title: Inline editing — cards and documents
type: design
status: archived
created: "2026-06-12T11:45:02.940243Z"
updated: "2026-06-12T20:44:32.435222Z"
tasks:
    - T-EDIT-CARDS-INLINE-IN
    - T-EDIT-DOCS-IN-THE
---
# Inline editing — split-pane editor for cards and documents

Allow editing a card's **title/body** and a document's **title/type/body**
directly in the web UI, edited in place (not a separate form page), saved via an
endpoint, with a live preview. Cards and documents share the same split-pane
editor pattern and the same server-side preview mechanism; they differ only in
the surface they live on and the fields they expose.

This design covers two tasks: `T-EDIT-CARDS-INLINE-IN` (cards) and
`T-EDIT-DOCS-IN-THE` (documents).

## Shared editor pattern

A **two-column editor**:
- left: a `<textarea>` holding the raw Markdown body
- right: a live-rendered HTML preview
- editable scalar fields (title, and for docs the type) sit above as inputs
- buttons: **Save** and **Cancel**

### Live preview is server-side

Markdown is rendered **server-side** via goldmark plus custom `:::` directives
(`internal/render`). The browser cannot replicate the directive rendering, so
the preview must round-trip to the server. The textarea debounces (~300ms) and
posts the current body to a preview endpoint; the response replaces the preview
pane's inner HTML. Preview is **stateless** — it never writes to the DB, so the
data-version and live-reload polling are untouched while the user types.

## Cards

- Editing lives **only in the card modal** on `/board`. The standalone
  `/task/{code}` page stays read-only.
- The task fragment gains an **"Edit" button** (next to Archive). Clicking it
  swaps the modal body into the split-pane editor. Title is a single-line
  `<input>`; body is the Markdown textarea.
- Cancel restores the read view by re-fetching the fragment (no save). Escape /
  backdrop click while editing prompts a confirm-if-dirty, then closes.
- Only active (non-archived) cards are editable; the Edit button is absent for
  archived cards.

### Card endpoints

- **`POST /task/{code}/preview`** — JSON `{body}` → rendered HTML
  (`render.RenderMarkdown`). Stateless.
- **`POST /task/{code}/edit`** — JSON `{title, body}`, mirroring
  `handleMoveTask`: `http.MaxBytesReader` (1 MiB) and `DisallowUnknownFields`.
  Calls existing `store.EditTask(code, &title, &body)`. Validation: title must be
  non-empty after trimming (else 400); 404 for unknown code.
  On success **returns the re-rendered task fragment HTML** (same output as
  `GET /task/{code}?fragment=1`); JS swaps it into the modal body, flipping the
  modal back to the updated read view. `EditTask` bumps `updated_at` and the
  data-version, so other open tabs reload via `/events/version` polling.

## Documents

- Documents render as a **standalone full page** (`/doc/{code}`) — there is no
  modal or fragment for docs. The editor therefore lives **inline on the doc
  page**: an **"Edit" button** (next to Archive) swaps the rendered body region
  into the split-pane editor.
- Editable fields: **title** (`<input>`), **type** (a `<select>` over the known
  doc types, matching the CLI's `doc edit --type`), and the Markdown **body**.
- Only active (non-archived) docs are editable; the Edit button is absent for
  archived docs.
- Cancel restores the read view (re-render of the page region or a plain reload).

### Document endpoints

- **`POST /doc/{code}/preview`** — JSON `{body}` → rendered HTML
  (`render.RenderMarkdown`). Stateless. (Shares the same render path as the card
  preview; the two preview handlers can delegate to one helper.)
- **`POST /doc/{code}/edit`** — JSON `{title, type, body}`, same guards as the
  card edit handler. Calls existing `store.EditDoc(code, &title, &type, &body)`.
  Validation: title non-empty after trimming; type must be one of the known
  types; else 400. 404 for unknown code.
  Because the doc page has **no fragment renderer**, on success the handler
  returns 200 and the **page reloads** to show the rendered result — consistent
  with how archive/delete already behave on the doc page. `EditDoc` bumps
  `updated_at`/data-version for other tabs.

## Files touched

- `internal/web/server.go` — register `POST /task/{code}/edit`,
  `POST /task/{code}/preview`, `POST /doc/{code}/edit`, `POST /doc/{code}/preview`.
- `internal/web/pages.go` — `handleEditTask`, `handlePreviewTask`,
  `handleEditDoc`, `handlePreviewDoc`; factor the task fragment-render so
  `handleTask` and `handleEditTask` share it, and a single render-preview helper
  shared by both preview handlers.
- `internal/web/templates/task.gohtml` — Edit button + editor markup (hidden
  until Edit is clicked).
- `internal/web/templates/doc.gohtml` — Edit button + editor markup + inline
  editor JS (toggle, debounced preview, save/reload, cancel).
- `internal/web/templates/board.gohtml` — card editor JS (toggle, debounced
  preview, save/cancel) in the board script, since the modal lives there.
- `internal/web/static/style.css` — two-column editor layout (shared class).
- `internal/web/server_test.go` — tests for all four endpoints.

## Testing

- `POST /task/{code}/edit` updates title and body; returns the rendered fragment;
  rejects empty title (400); 404 for unknown code.
- `POST /doc/{code}/edit` updates title, type, and body; rejects empty title and
  invalid type (400); 404 for unknown code.
- `POST /task/{code}/preview` and `POST /doc/{code}/preview` render Markdown + a
  `:::` directive and never mutate the store (data-version unchanged).
