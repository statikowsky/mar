---
title: Unified search command
type: analysis
status: active
created: "2026-06-27T12:12:47.793706254Z"
updated: "2026-06-27T12:12:47.793706254Z"
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

## Implementation notes

The repository is small, and the file store already loads docs/tasks into memory. A linear scan is the right first implementation. FTS or an inverted index would add complexity without clear value until repositories become much larger.

Snippet generation should be simple and deterministic: find the first match in the chosen field, trim to a fixed character window, collapse whitespace for plain output, and preserve enough context in JSON.

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

## Tests

- CLI JSON and plain output for doc title, doc body, task title, and task body matches.
- Filters for docs/tasks/status/type.
- Empty search term error.
- Case-insensitive matching by default.
- Snippet includes the matched term and does not include YAML frontmatter.
