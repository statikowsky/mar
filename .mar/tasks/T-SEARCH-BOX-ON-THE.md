---
title: Search box on the docs list (title + code)
status: active
created: "2026-06-30T23:47:23.388863039Z"
updated: "2026-06-30T23:56:01.976947551Z"
---
Add a text search input above the web Contents table that filters rows by
document **name (title)** and **code**, case-insensitive substring. Pure
client-side, same row-hide approach as the type filter — combine the two
(AND) so search + type narrow together across active and archived tables.
Distinct from the `mar search` CLI ([[DOC-SEARCH]]), which searches title +
body server-side; this is just live filtering of the already-rendered list.
See [[DOC-DOCSORT]].
