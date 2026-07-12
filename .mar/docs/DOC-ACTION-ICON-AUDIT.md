---
title: Action icon audit
type: report
status: active
created: "2026-07-12T17:34:20.333436Z"
updated: "2026-07-12T17:34:20.333436Z"
tasks:
    - T-ADD-CONSISTENT-ICONS-TO
---
# Executive summary

mar has a coherent local Lucide-compatible icon language, but it is applied mainly to theme controls and note actions. Repeated actions elsewhere remain text-only, and static templates do not have a shared icon-rendering mechanism. Icons should be added where they improve scanning and reinforce an action that recurs across screens; labels should remain visible except for universally understood compact controls.

The highest-value pass is to cover create, edit, save, archive/restore, delete, duplicate, undo/redo, view switching, and modal/rail close. Do not add icons to every link, select option, filter, or settings choice.

# Current icon system

- `window.marIcon(name)` returns inline, aria-hidden, 16px SVG using Lucide-compatible geometry.
- Registered icons: file-text, sticky-note, trash-2, grip, pencil, list-todo, file-plus-2, and save.
- Theme, color-scheme, and accessibility controls use separate inline SVG markup.
- Note actions use consistent icon-and-label or icon-only components.
- Static server-rendered actions cannot call `marIcon` directly, encouraging duplicated SVG or text-only controls.

# Recommended icon additions

## Priority 1 — repeated CRUD actions

| Surface | Action | Recommended Lucide icon | Treatment |
| --- | --- | --- | --- |
| Documents list | New document | file-plus-2 | Icon + label |
| Documents list/detail | Archive | archive | Icon + label |
| Archived documents | Unarchive | archive-restore | Icon + label |
| Archived documents | Delete | trash-2 | Icon + label, danger |
| Document detail | Edit | pencil | Icon + label |
| Document edit/new | Save/Create | save / file-plus-2 | Icon + label, primary |
| Board | New card | square-plus | Icon + label, primary |
| Archived tasks | Unarchive | archive-restore | Icon + label |
| Task detail | Edit | pencil | Icon + label |
| Task detail/list | Archive/Delete | archive / trash-2 | Icon + label |
| Task edit/new | Save/Create | save / square-plus | Icon + label, primary |

These actions repeat across screens and currently require users to reread labels. Consistent icons would create useful visual landmarks.

## Priority 2 — Scratchpad toolbar and list actions

| Action | Recommended Lucide icon | Notes |
| --- | --- | --- |
| Add note | sticky-note or message-square-plus | Keep label |
| Canvas/List view | layout-grid / list | Change icon and label together when toggled |
| Undo/Redo | undo-2 / redo-2 | Keep labels; disabled state remains explicit |
| Duplicate | copy | Add to toolbar and list-row Duplicate |
| Delete | trash-2 | Keep label in toolbar; icon-only remains appropriate inside notes |
| Zoom out/in | zoom-out / zoom-in | Replace typographic minus/plus with SVG; retain aria-label |
| Reset 100% | rotate-ccw or focus | Prefer label `100%`; icon is optional |
| Fit | scan | Keep label |
| Reload remote | refresh-cw | Conflict recovery action |
| Keep my version | upload or save | Use icon + label; wording carries the risk and must remain |

## Priority 3 — note rail and overlays

| Action | Recommended Lucide icon | Notes |
| --- | --- | --- |
| Notes toggle | sticky-note | Icon + count/label |
| Add note | message-square-plus | Icon + label |
| Close note rail/modal | x | Icon-only is acceptable with aria-label and tooltip |
| Document annotation gutter | message-square-plus | Keep accessible name and tooltip; icon could replace the visually empty target |

## Priority 4 — navigation and copy affordances

- Replace literal left-arrow characters in Back links with a shared arrow-left icon, keeping the text.
- Consider adding copy beside document/task codes using `copy`; keep the code text because it is the value being copied.
- Dashboard Board and Scratchpad cards may use layout-dashboard and sticky-note as larger navigational illustrations, but this is optional and should be treated separately from action buttons.

# Where icons should not be added

- Type and color selects: the native disclosure indicator and text value are sufficient.
- Search, type filters, and sortable table headings unless a specific compact affordance needs clarification.
- Theme/accessibility menu options: checked radio state already communicates selection.
- Document/task title links, backlinks, and linked-item lists: repeated icons would add clutter without clarifying the destination.
- Cancel buttons: text is clear and visually secondary; an X can be confused with closing the whole surface.
- Archive status tags, timestamps, and other metadata.

# Interaction and accessibility rules

- Keep visible labels for CRUD, destructive, conflict-resolution, and unfamiliar actions.
- Reserve icon-only controls for close, delete within a compact note row, and zoom controls with strong aria-label/title support.
- Icons must remain `aria-hidden`; the control owns the accessible name.
- Use one 16px default icon size and the established 28px compact note target; page-level actions can retain their existing larger target.
- Preserve semantic emphasis: primary create/save, secondary edit/navigation, danger delete, muted disabled.
- Test icon registration separately from icon invocation so an empty SVG cannot regress unnoticed.

# Implementation architecture

Before broad adoption, add a reusable static-template mechanism rather than copying SVG paths into each Go template. A small `data-mar-icon="name"` placeholder hydrated by the existing registry, or a Go template helper backed by the same registry, would let static and dynamic controls share one source of truth.

The registry should fail visibly in development or tests for unknown names. The recent Save button issue occurred because the caller requested an icon that was not registered; checking both the requested name and non-empty registered path closes that gap.

# Suggested rollout

1. Add the shared static icon rendering mechanism and registry validation.
2. Apply Priority 1 CRUD icons across document, task, and board templates.
3. Apply Priority 2 Scratchpad toolbar/list icons and verify responsive wrapping.
4. Apply note rail/overlay icons, then evaluate navigation/card embellishments separately.
5. Run keyboard, screen-reader-name, light/dark, high-contrast, desktop, and mobile checks.
