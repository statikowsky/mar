---
title: Checkpoint WAL into mar.db on store Close
status: archived
created: "2026-06-10T12:30:25.713199Z"
updated: "2026-06-10T21:19:04.817741Z"
---
The DB uses journal_mode=WAL. Store.Close() just calls db.Close(), which does
not always checkpoint the WAL into mar.db. Result: after CLI writes (e.g. `doc
create`), the new data lives only in .mar/mar.db-wal (which is gitignored), so
`git add .mar/mar.db` captures nothing and committing MAR-tracked docs/tasks
silently misses the content until a manual `PRAGMA wal_checkpoint(TRUNCATE)`.

Hit while committing the DOC-CARDARCHIVE spec.

Fix: run `PRAGMA wal_checkpoint(TRUNCATE)` in Store.Close() (or open with a
checkpoint-on-close setting) so the tracked mar.db is always current when the
CLI exits.
