---
title: Remove legacy docs/superpowers specs and plans
status: active
created: "2026-06-12T19:19:04.193549Z"
updated: "2026-06-12T22:43:42.584556Z"
---
From project review. AGENTS.md says specs/plans must live in MAR docs, not committed files, but git still tracked docs/superpowers/specs/2026-06-09-mar-design.md and docs/superpowers/plans/2026-06-09-mar.md (predate the rule).

Decision: removed outright rather than migrated — they were superseded by the file-store work (DOC-STORAGE, DOC-FILESTORE) and held no content worth preserving as MAR docs. The docs/ tree is now empty and deleted.
