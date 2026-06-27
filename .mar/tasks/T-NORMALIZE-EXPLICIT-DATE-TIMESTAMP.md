---
title: Normalize explicit date timestamp precision
status: active
created: "2026-06-27T12:11:08.148612016Z"
updated: "2026-06-27T12:11:08.148612016Z"
---
Fix inconsistent timestamp normalization for explicit date inputs.

Observed behavior:
- `mar doc edit DOC-X --updated 2026-06-27` writes `2026-06-27T00:00:00Z`.
- Automatically-created timestamps generally use RFC3339Nano precision.

Decide and implement a consistent timestamp policy for `--created` / `--updated` on docs and tasks. Date-only input should either intentionally mean start-of-day with documented precision, or be normalized to the same canonical timestamp shape used elsewhere.
