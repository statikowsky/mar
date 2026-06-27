---
title: Explicit date timestamp precision
type: analysis
status: active
created: "2026-06-27T12:11:25.58892246Z"
updated: "2026-06-27T12:11:25.58892246Z"
tasks:
    - T-NORMALIZE-EXPLICIT-DATE-TIMESTAMP
---
# Analysis

## Problem

Explicit date edits currently produce a different timestamp shape than ordinary MAR-generated timestamps. A command such as `mar doc edit DOC-X --updated 2026-06-27` normalizes to `2026-06-27T00:00:00Z`, while creation and automatic update paths generally write RFC3339Nano timestamps such as `2026-06-27T11:56:45.238205589Z`.

This is not data loss if a date-only input is meant to be date-only, but the storage format has no date-only type. The result is an apparent precision mismatch in frontmatter and JSON output.

## Scope

This affects both docs and tasks because they share date-edit surfaces:

- `mar doc edit --created/--updated`
- `mar task edit --created/--updated`
- any future `mar doc touch --updated`

## Options

### Preserve date-only as midnight UTC

Keep the current semantic meaning: a user who supplies only a date gets the start of that UTC day. Document it explicitly and update tests to assert the exact canonical form.

Pros: simple, predictable, no fake precision.
Cons: frontmatter mixes `...Z` and `...123Z`, which looks accidental.

### Normalize all stored timestamps to seconds

Change `nowStamp()` and date parsing to use RFC3339 seconds only. This makes all timestamps visually consistent and avoids meaningless nanosecond precision for human-authored docs.

Pros: clean diffs, stable formatting, enough precision for MAR.
Cons: changes existing format style and may touch many tests.

### Store date-only input with current time on that date

For `--updated 2026-06-27`, combine the supplied date with the current time. This preserves nanosecond shape but invents information the user did not provide.

Not recommended.

## Recommendation

Prefer normalizing all MAR-written timestamps to RFC3339 seconds. MAR is a human-facing, git-tracked Markdown repository; nanosecond precision adds churn and little value. If that change is too broad, keep midnight UTC but make it explicitly documented and test-covered.

## Acceptance criteria

- Shared date normalization tests cover date-only and full RFC3339 input.
- Docs and tasks use the same behavior.
- `mar guide` and help text describe what date-only input means.
- Existing created/updated metadata is not bulk-rewritten unless a migration is intentionally added.
