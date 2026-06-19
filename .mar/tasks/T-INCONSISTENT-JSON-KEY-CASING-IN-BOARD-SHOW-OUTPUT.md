---
title: Inconsistent JSON key casing in board show output
status: archived
created: "2026-06-10T10:20:16.47772Z"
updated: "2026-06-10T21:19:04.992157Z"
---
`board show --json` mixes conventions in one payload: columns use lowercase
(`id`, `name`, `tasks`) but embedded task objects use Go PascalCase (`ID`,
`Code`, `Title`, `CreatedAt`, ...).

Agents parsing the payload must handle both casings at once. Standardize on
lowercase/snake_case JSON tags across all serialized structs (task, doc,
column).
