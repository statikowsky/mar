---
title: Unify remaining Scratchpad note controls
status: active
created: "2026-07-12T16:41:19.281021Z"
updated: "2026-07-12T16:41:19.281021Z"
---
## Confirmed mismatches

- Canvas note color and Task/Doc promotion controls use the legacy 10px, 3px × 6px rule while navigation and delete use the 28px note-action component.
- List-view Edit, Duplicate, color, Task/Doc controls use a separate 12px, 5px × 9px rule; navigation and delete remain 28px/11px.
- Legacy controls do not share the note-action focus-visible outline, muted/default colors, disabled state, or danger treatment.
- The Scratchpad toolbar can remain a distinct toolbar size, but all controls inside a note card or list-row action group should use one component contract.

## Acceptance criteria

- Use a shared 28px-height note action style for buttons, links, and selects in canvas cards, list rows, and document note cards.
- Keep semantic variants: primary save/create, secondary edit/navigation/promotion/color, dangerous delete.
- Normalize typography, radius, padding, icon size, hover, focus-visible, and disabled states.
- Add regression coverage for the shared classes/components on both Scratchpad views and document notes.
