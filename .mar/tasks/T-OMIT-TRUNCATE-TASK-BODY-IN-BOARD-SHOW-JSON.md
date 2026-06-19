---
title: Omit/truncate task Body in board show --json
status: archived
created: "2026-06-10T10:20:16.495856Z"
updated: "2026-06-10T21:19:04.970145Z"
---
`board show --json` inlines the full `Body` of every task. On a real board
this is significant token bloat for every board read (a hot path for agents).

Consider omitting Body from the board listing (fetch via `task show`/`doc show`
when needed), or adding a flag to include it.
