---
title: Add --json to all mutating commands
status: archived
created: "2026-06-10T10:20:16.440324Z"
updated: "2026-06-10T21:19:05.029942Z"
---
`--json` is only wired on read + core-create commands. It is **missing** on
many mutating commands an agent calls constantly:

- task: move, rm, link
- doc: move, archive, rm, link
- column: add, move, rename, rm

Effect: an agent gets structured output creating a task but free text
("Moved T-X") when moving it, forcing it to switch parsing strategies
mid-workflow. Add `--json` everywhere a command produces a result.
