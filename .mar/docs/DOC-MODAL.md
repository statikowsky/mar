---
title: Open a kanban card in a modal
type: design
status: archived
created: "2026-06-09T18:23:11.951957Z"
updated: "2026-06-12T19:36:34.87627Z"
tasks:
    - T-6
---
Clicking a board card today navigates to the standalone task page
(`/task/{code}`, `task.gohtml`), leaving the board. This design opens the task
inline as a modal overlay on the board instead, as progressive enhancement over
the existing link.

> [!NOTE]
> Tracked as **T-6**. Shares the board's fetch/refresh plumbing with
> [[doc-reorder]] (drag-and-drop), which added the first board interactivity.

## Goals

- Click a card -> the task (title, code, column, rendered notes, linked docs)
  appears in a modal over the board; the board stays in place behind it.
- Close via Escape, a close button, or clicking the backdrop.
- Deep-linkable: a card's task is reachable by URL so a link can open it.
- Degrade gracefully: with JS off, the card link still navigates to the full
  `/task/{code}` page.

## Non-goals

- Editing the task in the modal (read-only, like the current task page).
- A modal for docs (separate, if wanted later).

## Server

Add a fragment view of the task page so the modal can fetch just the content,
not the whole layout chrome:

#### GET /task/{code}?fragment=1

Returns the rendered task content only (the inner `content` template: title,
meta, body, linked docs) without the `<html>`/layout wrapper or the SSE script.
Plain `GET /task/{code}` is unchanged — full page, the no-JS fallback. One new
branch in `handleTask` keyed off the query param; reuses the same view model.

## Client

### Open

- Delegate clicks on `.task-card`. If the drag handler just fired (the
  suppress-click flag from [[doc-reorder]]) or a modifier key is held, do
  nothing special. Otherwise `preventDefault`, fetch
  `/task/{code}?fragment=1`, inject the HTML into a modal container, show it,
  and set `location.hash = "#task/{code}"` for deep-linking.

### Close + deep-link

- Escape, close button, or backdrop click hides the modal and clears the hash.
- On load, if the hash matches `#task/{code}`, open that card's modal — so a
  shared/bookmarked URL lands on the open task.

## Interaction with drag-and-drop

Both behaviours live on the same `.task-card`. The drag handler already sets a
one-shot `suppressClick` flag on drop; the modal click handler must honour it so
a drag never also opens the modal. A plain click (no drag) opens the modal.

## Edge cases

- **Unknown code in hash:** fragment fetch 404s -> ignore, leave board as-is.
- **Card removed while modal open (other tab's SSE reload):** the board reloads
  on the SSE `reload` event; simplest is to let the reload close the modal.
- **No-JS:** anchors still navigate to `/task/{code}` (unchanged).

## Security

The fragment is the same rendered, sanitized content as the full page (Markdown
already escaped by goldmark; only the pre-rendered body is `template.HTML`). No
new trust surface; loopback-only like the rest of `serve`.

> [!TIP]
> **Recommendation:** add the `?fragment=1` branch to `handleTask` and a small
> dependency-free modal in `board.gohtml`, reusing the suppress-click flag so drag
> and click stay distinct. Self-contained, no new dependencies, no-JS safe.
