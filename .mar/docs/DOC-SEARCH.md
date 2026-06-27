---
title: Unified search command
type: analysis
status: active
created: "2026-06-27T12:12:47.793706254Z"
updated: "2026-06-27T14:08:53.507177443Z"
tasks:
    - T-ADD-UNIFIED-SEARCH-COMMAND
---
# Analysis

## Problem

Finding related MAR knowledge currently requires leaving MAR: `mar doc list | grep`, `mar task list | grep`, or raw `grep -r .mar/`. That works in a pinch, but it loses structure, misses useful filters, and forces agents to parse storage files instead of using the CLI contract.

A first-class `mar search <term>` would likely become one of the most-used commands because MAR is both a docs repository and a task board.

## Goals

- Search across doc titles, doc bodies, task titles, and task bodies.
- Return enough context to decide what to open next.
- Preserve MAR identity in output: kind, code, title, type/status/column where relevant.
- Support `--json` so agents do not scrape text output.
- Work well on the current file-backed store without adding an index up front.
- Leverage fast local search tools when available, without making them required for correctness.

## Proposed command

`mar search TERM [--docs] [--tasks] [--status active|archived|all] [--type TYPE] [--json]`

Default behavior:

- Search active docs and active tasks.
- Case-insensitive substring search.
- Return matches ordered by rough usefulness: title hits before body hits, then code/title order.

Plain output sketch:

```text
DOC-STORAGE  doc/design  title  Git-friendly storage: replacing the binary SQLite store
T-ADD-WIKILINK-LINT-AND  task/To do  body  ...there is no mar doc lint...
```

JSON result shape:

```json
[
  {
    "kind": "doc",
    "code": "DOC-STORAGE",
    "title": "Git-friendly storage: replacing the binary SQLite store",
    "field": "body",
    "snippet": "...",
    "type": "analysis",
    "status": "active"
  }
]
```

## Local tool strategy

Search should have a pure Go baseline and may use local tools as accelerators. Correctness belongs to MAR's parsed store model; external tools are optional implementation details.

Recommended tiering:

1. `rg` / ripgrep when present in `PATH`. It is the best local accelerator: fast, common on developer machines, supports `--json`, line numbers, context, smart case, and globs.
2. `git grep` when inside a git worktree and `rg` is unavailable. It is fast and usually available with git, but it mainly searches tracked files unless invoked carefully, so it can miss newly-created untracked MAR docs/tasks.
3. Pure Go scan over parsed MAR docs/tasks. This must always exist as the correctness fallback and may be fast enough for most repositories.
4. Plain `grep` only as a last-resort fallback if we need it. Its output is less portable and harder to parse robustly than `rg --json` or MAR's own scanner.

Important caveat: raw file tools see YAML frontmatter and file paths, not MAR entities. If `rg` or `git grep` is used, MAR should map matched paths back to parsed docs/tasks and classify matches as title/body/frontmatter. Frontmatter matches should either be filtered out or intentionally reported as metadata matches; they should not be mistaken for body text.

A pragmatic first implementation can ship the pure Go scan only, then add `rg --json` as an accelerator once the result model is stable. If an accelerator disagrees with the parsed model, the parsed model wins.

## Implementation notes

The repository is small, and the file store already loads docs/tasks into memory. A linear scan is the right correctness baseline. FTS or an inverted index would add complexity without clear value until repositories become much larger.

Snippet generation should be simple and deterministic: find the first match in the chosen field, trim to a fixed character window, collapse whitespace for plain output, and preserve enough context in JSON.

If local-tool acceleration is added, keep it behind a small interface so tests can exercise the Go searcher without depending on host binaries. Tool detection should be runtime-only; missing `rg` is not an error.

## Filters

Useful initial filters:

- `--docs` and `--tasks` to restrict kinds.
- `--type` for docs.
- `--status active|archived|all`, default active.
- Possibly `--case-sensitive` later; default should be case-insensitive.

## Edge cases

- Empty term should return a validation error, not every document.
- Markdown frontmatter should not be searched as body text through raw file grep; use store objects.
- Archived items should be omitted by default but searchable when explicitly requested.
- Bodyless tasks should still be searchable by title.
- Untracked MAR files may not appear in `git grep`; do not rely on it as the only backend.

## Tests

- CLI JSON and plain output for doc title, doc body, task title, and task body matches.
- Filters for docs/tasks/status/type.
- Empty search term error.
- Case-insensitive matching by default.
- Snippet includes the matched term and does not include YAML frontmatter.
- Search works when no external search binary is present.
- If local-tool acceleration is implemented, tests cover path-to-entity mapping and frontmatter filtering with a fake backend or fixture.
