---
title: Resolve column_id to name in task JSON
status: archived
created: "2026-06-10T11:39:05.149858Z"
updated: "2026-06-10T21:19:04.931435Z"
---
`task list --json` and `task show --json` emit `column_id` (an opaque DB id)
with no column name. `board show --json` maps id->name, but an agent using
list/show must cross-reference against the board to learn which column a task
is in.

Add a `column` (name) field alongside `column_id` in task show/list JSON so
each task is self-describing. Store already has the mapping
(`columnNames()` equivalent in web layer).

Found during agent-usability re-audit.
