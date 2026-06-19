---
title: Make store mutations transactional (lost-update edits, RemoveColumn double-DELETE)
status: archived
created: "2026-06-12T19:19:04.13131Z"
updated: "2026-06-12T22:24:20.353824Z"
---
From project review. Store mutations are non-transactional read-modify-write sequences, which matters because the same store is shared by concurrent web handlers and CLI processes:

- `EditTask`/`EditDoc` (internal/store/task.go:291, doc.go:116) read the row, patch in memory, write back ALL columns — two concurrent edits silently clobber each other (web UI calls EditTask from pages.go:381).
- `RemoveColumn` (internal/store/column.go:200) runs `DELETE FROM tasks` and `DELETE FROM columns` as two separate statements; a failure between them loses tasks but keeps the column. Destructive, should be one transaction.
- `CreateTaskWithCode` (task.go:86) does column lookup, maxPosition, code-uniqueness probe and INSERT non-atomically; concurrent creates can tie positions or fail with raw UNIQUE errors.
- `nextTaskCode` (task.go:46) uses a deferred tx for read-then-update; under contention the lock upgrade fails with SQLITE_BUSY. Use BEGIN IMMEDIATE or UPDATE ... RETURNING.
