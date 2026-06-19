---
title: Drop custom directives in favor of standard Markdown
type: design
status: archived
created: "2026-06-12T19:25:56.182722Z"
updated: "2026-06-12T20:43:53.647763Z"
tasks:
    - T-CLIENT-SIDE-MARKDOWN-PREVIEW
    - T-DROP-CUSTOM-DIRECTIVES-GFM
    - T-REMOVE-MAR-MIGRATE-DIRECTIVES
---
# Drop custom directives in favor of standard Markdown

## Problem

MAR bodies support three custom fenced directives (`::: callout`, `::: card`,
`::: phase-block`) parsed by a line-based preprocessor before goldmark. They
cost more than they give:

- They block client-side rendering: both inline editors round-trip previews
  through the server solely because directives only render server-side, and
  the planned WYSIWYG editor (T-WYSIWYG-THREE-COLUMN-MARKDOWN) would have to
  replicate or special-case them.
- The preprocessor is the project's one real rendering bug: it does not track
  code fences, so a literal `:::` example inside a code block swallows the
  rest of the block (T-DIRECTIVE-PARSER-SWALLOWS-INSIDE).
- Bodies stop being portable Markdown, which contradicts MAR's premise.

Usage is small: 4 of 12 docs, 10 blocks total (callout plain/good, card with
title, phase-block p1/p2 with title).

## Decision

Remove the directive system. Map each directive to standard Markdown:

| Old | New |
| --- | --- |
| `::: callout` | `> [!NOTE]` blockquote (GFM alert) |
| `::: callout good` | `> [!TIP]` |
| `::: callout warn` | `> [!WARNING]` |
| `::: card "Title"` | `#### Title` + body |
| `::: phase-block pN "Title"` | `### Title` + body |

GFM alerts (`[!NOTE]`, `[!TIP]`, `[!IMPORTANT]`, `[!WARNING]`, `[!CAUTION]`)
are GitHub-native, so exported bodies render correctly outside MAR. The
phase-block pill variant (p1–p4/skip) is dropped; it was purely visual.

## Rendering

A small goldmark AST transformer (`internal/render/alerts.go`) detects a
blockquote whose first line is exactly `[!TYPE]`, removes the marker node, and
sets `class="alert alert-<type>"` on the blockquote (goldmark's default
renderer emits attributes when present). The label ("Note", "Warning", …) is
drawn by CSS `::before` per class, reusing the existing semantic accent
variables. `.callout`, `.card`, and `.phase-block` CSS is deleted; `.pill`
stays (used by board/doc templates).

`RenderMarkdown` collapses to a single goldmark pass — no segment splitting.
A stray `:::` line in old content renders as harmless literal text.

## Content migration

New command: `mar migrate directives [--dry-run]`.

- A fence-aware legacy parser in `internal/migrate` rewrites directive blocks
  in a Markdown string per the table above (fence-aware so the old parser's
  code-block bug is not reproduced during migration).
- `store.RewriteBodies(fn)` applies the rewrite to every doc and task body
  (all statuses, including archived) in one transaction and reports how many
  rows changed. `updated_at` is bumped on rewritten rows.
- Idempotent: a second run changes nothing.
- `--dry-run` lists the codes that would change without writing.

The HTML importer (`mar doc import`) stops emitting directives: `div.callout`
becomes a GFM alert blockquote, `div.card` / `div.phase-block` become a
heading plus body.

## Migrating other projects on older MAR versions

Other repos have `.mar/mar.db` files whose bodies contain directives. Options
considered:

1. **Auto-rewrite on `Open`** (like schema migration). Zero-touch, but
   silently mutates user content on first contact with a new binary; not
   every project commits its db, so it may be unrecoverable. Rejected.
2. **Render-time shim** (keep translating directives at render). Zero-touch
   but keeps the parser forever and the content never actually migrates.
   Rejected.
3. **Explicit `mar migrate directives`** (chosen). After upgrading the
   binary, run it once per project. Until then directives render as literal
   `:::` text — ugly but lossless and obvious. The command is idempotent and
   safe to run anywhere; `--dry-run` previews. Documented in README and
   `mar guide` so agents pick it up.

> [!NOTE]
> The schema `migrate()` that runs on every `Open` stays reserved for
> structural changes. Content rewrites are always explicit commands.

## Out of scope / follow-ups

- Client-side preview in the inline editors (now unblocked) — separate task.
- T-DIRECTIVE-PARSER-SWALLOWS-INSIDE is resolved by removal (the migration
  parser is fence-aware).
- The legacy spec/plan files under docs/superpowers still mention directives;
  they are covered by T-MIGRATE-LEGACY-DOCS-SUPERPOWERS.


## Addendum (2026-06-12): migration tooling removed

Only one project ever used directives and its store has been migrated, so the
one-shot `mar migrate directives` command, `internal/migrate`, and
`store.RewriteBodies` were removed the same day they shipped. Any stray
legacy `:::` text in an old store still renders as harmless literal
Markdown; if a forgotten store ever turns up, the rewriter can be recovered
from git history (commit 6c02538).
