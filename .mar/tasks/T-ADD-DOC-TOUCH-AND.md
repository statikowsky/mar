---
title: Add doc touch and updated timestamp ergonomics
status: active
created: "2026-06-27T12:10:34.380692135Z"
updated: "2026-06-27T12:10:34.380692135Z"
---
Editing MAR docs requires remembering to run `mar doc edit --updated 2026-06-27` after each content change. Add a low-friction way to keep document `updated` metadata honest.

Desired outcomes:
- A direct `mar doc touch DOC-X [--updated DATE]` command, or equivalent, for metadata-only timestamp updates.
- Consider auto-touching docs from a commit hook or a mar-managed workflow so edits do not rely on memory.
- Decide how explicit timestamp edits interact with normal `doc edit`, generated docs, and archived docs.

Related sibling issue: explicit `--updated 2026-06-27` currently normalizes to midnight (`T00:00:00Z`), while created timestamps usually carry full timestamp precision. See the separate timestamp normalization task.
