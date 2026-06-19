---
title: 'Misleading ''unknown flag: --json'' error on unsupported commands'
status: archived
created: "2026-06-10T10:20:16.459082Z"
updated: "2026-06-10T21:19:05.011473Z"
---
`mar task move T-1 --json` returns `{"error":"unknown flag: --json"}`.

The README/CLAUDE.md claim "All commands accept --json", so an agent assumes
it is universal and may misdiagnose this as "task not found". Resolved
naturally once --json is universal (see sibling task); until then the docs
overstate coverage. Fix docs and/or flag.
