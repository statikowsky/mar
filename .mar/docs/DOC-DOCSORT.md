---
title: Sort and filter the docs list
type: design
status: active
created: "2026-06-30T23:37:23.614315616Z"
updated: "2026-06-30T23:37:23.614315616Z"
tasks:
    - T-SEARCH-BOX-ON-THE
    - T-SORT-AND-FILTER-THE
---
## Goal

Let the web **Contents** table (`internal/web/templates/index.gohtml`) be
reordered by **Updated**, **Code**, or **Document** (title), and filtered by
**type**. Optional CLI parity: a `--sort` flag on `mar doc list`.

## Current state

- `index.gohtml` renders a static table — columns Code / Document / Type /
  Updated — in whatever order `ListDocs` returns: **updated desc, code asc**
  tie-break (`internal/store/doc.go` `ListDocs` / `sort.Slice`).
- CLI `mar doc list` already supports `--type` and `--status` filtering, but
  has **no ordering flag** (locked to updated-desc).

## Approach (lazy: client-side)

Do it in the browser — the docs table is small, the rows are already fully
rendered, no server round-trip needed.

1. **Sort** — make the `<th>` headers for Updated / Code / Document
   clickable; vanilla JS sorts the `<tr>` rows in place. Repeat click toggles
   asc/desc. Sort key from existing cell text (dates already render shortDate;
   sort on the row's machine value via a `data-sort` attr if shortDate isn't
   lexically orderable). Default stays Updated desc.
2. **Filter** — a native `<select>` of the distinct types present; hide
   non-matching rows. Applies to both Active and Archived tables.

No new endpoints, no query params, no dependency — a `<script>` block in
`index.gohtml` plus `data-sort` attrs on the date cells.

`skipped:` server-side sort/filter + URL-param persistence — add when doc
counts make client-side rendering slow (YAGNI for a repo-sized list).

## CLI parity (optional follow-up)

`mar doc list --sort updated|code|title` (default `updated`). Thread a sort
key into `ListDocs`; reuse the existing `sort.Slice`. One store + one CLI
test. Type filtering already exists, so nothing to add there.

## Check

- Sort: clicking a header reorders rows and toggling reverses them.
- Filter: selecting a type hides the rest in both tables; "all" restores.
