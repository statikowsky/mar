---
title: Wikilink validation and backlink inspection
type: analysis
status: active
created: "2026-06-27T12:12:08.715161824Z"
updated: "2026-06-27T12:12:08.715161824Z"
tasks:
    - T-ADD-WIKILINK-LINT-AND
---
# Analysis

## Problem

MAR supports inline wiki-links such as `[[DOC-CODE]]`, `[[T-CODE]]`, and labelled variants. This encourages a good workflow: link freely, including to docs or tasks that do not exist yet.

The missing piece is validation and inspection. A dangling link can mean any of three things:

- a typo,
- a code that was renamed or deleted,
- an intentional future page.

Right now those cases look the same, and there is no command that audits them across the repository. Likewise, users need an easy way to ask "what links here?" without falling back to raw grep.

## Existing context

The older completed wiki-link work established rendering and backlinks as a concept. This follow-up is about operational tooling: linting, focused backlink commands, and machine-readable output suitable for agents.

Related prior design: [[DOC-WIKILINKS]].

## Proposed command surface

### `mar doc lint`

Scan all active docs and tasks for inline wiki-links and report unresolved targets.

Suggested output fields for `--json`:

- source code (`DOC-X` or `T-X`)
- source kind (`doc` or `task`)
- target code as written
- normalized target code, if parseable
- line number, when available
- label, if the link used `[[CODE|label]]`
- status: `dangling`, `invalid-code`, or `ok` when verbose

Plain output should be grep-friendly and grouped by source.

Default exit code should probably be 0 with findings printed, because future links are valid workflow. Add `--strict` to exit non-zero on dangling or invalid links for CI/pre-commit use.

### `mar backlink CODE`

A top-level command avoids forcing this to be doc-only. It should accept doc or task codes and return every inline wiki-link that targets the code.

Alternative: add `mar doc backlinks DOC-X` and `mar task backlinks T-X`. That is more explicit but splits a graph operation across command groups.

Recommendation: `mar backlink CODE [--json]`, with doc/task show continuing to include backlink summaries.

## Resolution rules

- Normalize target codes the same way `GetDoc`/`GetTask` do where possible.
- Resolve both docs and tasks.
- Ignore links inside fenced code blocks, matching user expectations for examples.
- Preserve dangling links as first-class findings; do not auto-create docs/tasks.

## Rename implications

If doc/task recoding or deletion leaves stale wiki-links, lint should catch them. A future enhancement could offer `mar doc move --rewrite-links` or a dedicated rewrite command, but linting should come first.

## Tests

- Parser tests for valid links, labelled links, invalid code text, and fenced code blocks.
- Store/CLI tests where one doc links to an existing doc, one links to an existing task, and one link is dangling.
- `mar doc lint --json` reports source, target, and line information.
- `mar doc lint --strict` exits non-zero on unresolved links.
- Backlink command returns references from both docs and tasks.
