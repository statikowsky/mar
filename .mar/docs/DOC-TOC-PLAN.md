---
title: Document table of contents implementation plan
type: plan
status: active
created: "2026-07-11T17:24:10.586763Z"
updated: "2026-07-11T17:24:10.586763Z"
tasks:
    - T-ADD-DOCUMENT-TABLE-OF
---
## Goal

Add an automatically generated table of contents to document pages. On wide screens it appears in a sticky left sidebar; on narrow screens it flows above the document. The current section is highlighted while scrolling.

## Design

- Enable stable automatic IDs on rendered Markdown headings.
- Build the outline from `h2` and `h3` elements inside the document body only.
- Render an accessible `nav` beside the document and populate it progressively in the browser from the already-rendered HTML.
- Keep the sidebar sticky with its own overflow for long outlines.
- Track scrolling with a lightweight passive scroll handler and set `aria-current` on the active link.
- Hide the navigation when the document has no eligible headings.
- Switch to an in-flow panel above the document at narrow viewport widths.
- Hide the reading outline while the inline editor is open so it cannot describe stale preview content.

## Verification

- Add renderer coverage proving headings receive stable, unique IDs.
- Add web response coverage for the document outline shell and document-page layout class.
- Run `make check`.
