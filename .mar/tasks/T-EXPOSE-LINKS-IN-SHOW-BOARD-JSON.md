---
title: Expose links in show/board JSON
status: archived
created: "2026-06-10T11:39:05.130002Z"
updated: "2026-06-10T21:19:04.949722Z"
---
Links are effectively **write-only** from the CLI: `task link` / `doc link`
create them, but no command surfaces existing links in JSON. An agent that
creates a link can't later discover it without dropping to SQL.

The store and web UI already support this:
- `store.DocsForTask`, `store.TasksForDoc`, `store.DocCodesForTask`
- web board shows `DocCodes` per task; doc page shows related tasks

Add linked codes to JSON output:
- `task show --json`  -> include linked doc codes
- `doc show --json`   -> include linked task codes
- (optional) `board show --json` -> include per-task doc codes, matching the web board

Found during agent-usability re-audit.
