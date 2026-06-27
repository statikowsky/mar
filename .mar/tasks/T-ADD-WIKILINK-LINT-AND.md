---
title: Add wikilink lint and backlink inspection
status: active
created: "2026-06-27T12:11:40.030439796Z"
updated: "2026-06-27T12:11:40.030439796Z"
---
Add validation and inspection tooling for inline `[[...]]` links.

Current gap:
- MAR encourages linking to future docs/tasks, but there is no `mar doc lint` or equivalent to list dangling `[[...]]` targets.
- Typos, renamed codes, and intentionally-not-yet-created targets are indistinguishable.
- Backlinks exist conceptually in rendering/doc show, but there is no focused CLI workflow for "what links to this doc/task?" across inline wiki-links.

Desired outcomes:
- A lint command that reports dangling wiki-links with source doc/task and target code.
- A backlink command or `doc show`/`task show` enhancement for reverse references.
- Clear handling for intentional future links, possibly via severity, allowlist, or status output rather than hard failure by default.
