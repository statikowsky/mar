---
title: Add wiki-style links and backlinks between documents
status: active
created: "2026-06-13T11:28:21.488351022Z"
updated: "2026-06-14T12:14:08.628284363Z"
---
MAR's doc layer is increasingly used as a knowledge base (AGENTS.md mandates
storing all design docs, specs, and plans in MAR rather than under docs/). Today
documents can only be connected through explicit, typed doc<->task links
(internal/store/link.go); there is no free-form linking between documents and no
way to see what links *to* a given doc.

This gap stands out when comparing MAR to the "LLM Wiki" pattern (Karpathy, Apr
2026), where a directory of interlinked markdown entity pages connected by
`[[wiki-links]]` forms a compounding, agent-maintained knowledge base. MAR is
already being pushed toward that use case but lacks the inline-linking and
backlink graph that make it work.

Proposal:
- Support inline `[[DOC-CODE]]` (and optionally `[[DOC-CODE|label]]`) wiki-links
  inside document bodies, rendered as real links in the web UI.
- Parse these references at render/store time and maintain a backlinks index so
  each doc can show "Referenced by" (which docs/tasks link to it).
- Surface backlinks in `mar doc show` (and --json) and on the doc page in the
  web UI.
- Decide whether `[[...]]` also resolves task codes (T-*) for doc<->task parity
  with the existing explicit link model.

Open questions:
- Reuse the existing typed-link store, or add a derived/computed backlink index
  built from body parsing?
- How to handle dangling links (target doc doesn't exist yet) -- render as a
  placeholder, like wiki "red links"?
- Interaction with doc move/rename (codes change): rewrite references or keep
  links by stable id?

Found while analyzing MAR vs the LLM Wiki pattern.
