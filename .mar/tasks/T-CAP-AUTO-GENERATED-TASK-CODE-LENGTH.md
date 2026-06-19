---
title: Cap auto-generated task code length
status: archived
created: "2026-06-10T11:50:22.494233Z"
updated: "2026-06-10T21:19:04.913912Z"
---
Auto-generated codes slug the *entire* title, producing unwieldy codes like
T-RESOLVE-COLUMN-ID-TO-NAME-IN-TASK-JSON (8 words) that are painful to type
and reference for both humans and agents.

Cap the slug at the first 4 words (T-RESOLVE-COLUMN-ID-TO). Collision suffix
and the empty-title -> T-<seq> fallback are unchanged. Callers can still pass
an explicit --code to override entirely.

Found during agent-usability work.
