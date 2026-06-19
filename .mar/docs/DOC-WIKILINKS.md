---
title: Wiki-style links and backlinks between documents
type: design
status: active
created: "2026-06-14T12:03:36.69452387Z"
updated: "2026-06-14T12:32:03.39663421Z"
tasks:
    - T-ADD-WIKI-STYLE-LINKS
---
## Goal

Support inline `[[CODE]]` wiki-links in document bodies and a backlinks
("Referenced by") view, turning MAR's doc layer into an agent-maintained
knowledge base (the LLM Wiki pattern). Tracks T-ADD-WIKI-STYLE-LINKS.

## Decisions [[DOC-WEBSITE]]

1. **Backlinks are derived, not stored.** Parse `[[...]]` from bodies at
   render/query time and build the reverse index by scanning docs. The existing
   typed-link store (internal/store/link.go) stays for explicit doc<->task
   links only. Rationale: wiki-links live in prose and change on every edit;
   syncing them into the link table means write-time bookkeeping for no gain.
   A computed index is always consistent with the body.

2. **Syntax:** `[[CODE]]` and `[[CODE|label]]`. CODE resolves against both doc
   codes and task codes (`T-*`), giving doc<->task parity with one parser and
   one extra branch.

3. **Dangling links render as placeholders** ("red links") -- target doesn't
   exist yet. This is the point: link before the page exists.

4. **Links are by code; rename rewrites references.** Codes are the
   human-visible handle. On doc move/rename, rewrite `[[oldcode]]` ->
   `[[newcode]]` across bodies. ponytail: rewrite-on-rename, revisit with a
   stable-id scheme only if renames get frequent.

## Surfaces

- Web UI: render `[[CODE]]` as a real link (placeholder styling when dangling);
  show a "Referenced by" section on the doc page.
- `mar doc show` and `--json`: include backlinks.

## Out of scope

- Fuzzy/title-based matching (codes only).
- Stable-id link indirection.
- Cross-repo links.
