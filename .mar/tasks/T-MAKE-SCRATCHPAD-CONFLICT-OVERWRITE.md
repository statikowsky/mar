---
title: Make scratchpad conflict overwrite merge safe
status: active
created: "2026-07-11T22:35:08.60295Z"
updated: "2026-07-11T22:35:08.60295Z"
---
The current Scratchpad persists the complete notes array. After a revision conflict, the “Keep my version” path can overwrite unrelated remote note changes with the stale local snapshot.

Define and implement a safe conflict strategy before or alongside [[DOC-SCRATCHPAD-DOCS]]: preferably note-scoped mutations, or an explicit three-way/note-level merge that preserves unrelated remote additions and edits. Cover concurrent edits to different notes, the same note, remote deletion, and remote creation.
