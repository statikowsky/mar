---
title: Note controls across docs and Scratchpad
type: analysis
status: active
created: "2026-07-12T15:06:05.198414Z"
updated: "2026-07-12T15:06:05.198414Z"
tasks:
    - T-UNIFY-NOTE-CONTROLS-ACROSS
---
# Findings

The document note rail exposes a plain text “Open in Scratchpad” link beside Save. The Scratchpad uses a plain “← Back to DOC-X” link, while note actions use small bordered buttons. The link has no component-level styling, icon, consistent label, or equivalent placement between the two surfaces.

Scratchpad notes are already movable by dragging the card and resizable through a corner handle. Document-rail notes are automatically positioned beside their anchors, so “reposition” there would mean changing the document anchor rather than moving a card.

The app already embeds outline SVG icons locally for theme and accessibility controls. A Lucide-style icon language is therefore compatible, but adding the Lucide runtime or CDN is unnecessary; a small local SVG partial/sprite avoids a dependency and works offline.

# Recommendation

Use one compact note action row on both surfaces:

- A secondary icon-and-label link: `Open in Scratchpad` with a sticky-note or panels icon, and `Open document` with file-text plus arrow-up-right. Keep visible text; icon-only navigation is less discoverable.
- An icon-only `Trash 2` button at the trailing edge, using the danger color, an accessible label/title, and confirmation or undo. Add it to document-rail cards and individual Scratchpad cards; retain bulk Delete in the Scratchpad toolbar.
- Show a `Grip` or `Move` icon as a drag affordance on Scratchpad cards, not as a button. The whole card can remain draggable.
- Do not add a generic reposition icon to document notes yet. If re-anchoring is desired, define an explicit `Re-anchor` action with a crosshair/locate icon and a mode where the user selects a new document block.

Create a reusable icon-button/link component style with consistent 28–32px targets, focus-visible treatment, hover background, muted default color, and danger variant. Inline local Lucide-compatible SVGs are preferable to importing the library at runtime.

# Scope notes

List view should receive the same navigation and delete treatment. Keyboard deletion and bulk toolbar actions remain available. Tests should cover accessible labels, note deletion persistence, link destinations, and re-anchor behavior only if that separate feature is included.
