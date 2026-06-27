---
title: Doc touch and updated timestamp ergonomics
type: analysis
status: active
created: "2026-06-27T12:10:56.190447403Z"
updated: "2026-06-27T12:10:56.190447403Z"
tasks:
    - T-ADD-DOC-TOUCH-AND
---
# Analysis

## Problem

MAR documents carry `updated` metadata in frontmatter, but keeping it accurate is currently an extra manual step. After editing a doc, the user has to remember a second command such as `mar doc edit DOC-X --updated 2026-06-27`. That is easy to miss, especially because MAR encourages docs as the canonical place for plans and design notes.

The friction creates two bad outcomes: stale metadata that makes docs look older than their content, or noisy edits where a user has to reopen a doc just to bump a date.

## Candidate approaches

### `mar doc touch DOC-X`

A dedicated command is the smallest useful primitive. It should update only the document's `updated` field, defaulting to now, and optionally accept `--updated DATE` for backdating or cleanup.

Pros:
- Easy to teach and script.
- No hidden behavior.
- Works for docs edited outside MAR by a normal editor.

Cons:
- Still requires remembering a second command unless wrapped by tooling.

### Auto-touch during `mar doc edit`

`doc edit` already bumps `updated` when changing title/type/body. This covers CLI edits, but not editor-based changes to `.mar/docs/*.md`.

This should remain true and should be covered by tests, but it is not sufficient by itself.

### Git hook or mar hook

A pre-commit hook could detect modified `.mar/docs/*.md` files and touch their `updated` timestamps before commit. That addresses hand-edited docs, which is exactly where this problem appears.

Pros:
- Best ergonomics once installed.
- Makes stale updated metadata harder to commit.

Cons:
- Hooks are local and opt-in unless MAR installs/manages them.
- A hook that edits files during commit can surprise users and requires clear output.

### `mar doc lint` warning

A linter could report docs whose body changed without a corresponding timestamp bump. This is useful if hook mutation feels too magical. It depends on git history or index state, so it is a separate layer from the core store.

## Recommendation

Implement `mar doc touch DOC-X [--updated DATE]` first. It is the primitive that both humans and hooks need. Then add either a documented optional hook installer or a `mar doc lint --timestamps` check that can run in CI or pre-commit.

The command should preserve `created`, title, type, body, links, and status; only `updated` changes. It should support `--json` with the updated doc object, matching `doc edit`.

## Test notes

- Touch without `--updated` changes `updated_at` and leaves body/title/type/status untouched.
- Touch with `--updated 2026-06-27` uses the shared date normalization behavior.
- Unknown doc returns `ErrNotFound` through the existing JSON error envelope.
- Archived docs can be touched unless a stronger policy is chosen; metadata maintenance is not content editing.

## Related

See [[T-NORMALIZE-EXPLICIT-DATE-PRECISION]] for the timestamp precision inconsistency that this command will expose more often.
