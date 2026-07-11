---
title: Contextual scratchpad notes in documents
type: design
status: active
created: "2026-07-11T22:34:57.226168Z"
updated: "2026-07-11T23:02:13.144045Z"
tasks:
    - T-ADD-ANCHORED-SCRATCHPAD-RAIL
    - T-SCRATCHPAD-IN-DOCS
---
## Summary

Add a narrow annotation gutter immediately after the document's reading column and an optional **Notes** rail beyond it. Clicking the gutter opens the rail and starts an unsaved scratch note at the touched passage. A separate Notes action opens the rail without creating a note.

The rail is a contextual view of the repository Scratchpad, not a second note system and not a compressed infinite canvas. Notes created there remain ordinary `S-N` scratch notes, appear on `/scratchpad`, and carry an explicit association with the current document.

## Recommendation

Use three distinct concepts:

1. **Panel visibility is personal UI state.** Store whether the rail is open in browser `localStorage`, namespaced by repository and document code. Do not put this in document frontmatter: opening a panel should not edit shared project data, bump the document timestamp, or enable it for every user.
2. **Note association is shared project state.** Add an optional `docs` list to each scratch note. This is separate from `link`, which currently records the task or document created by promotion.
3. **Document anchors are semantic, not pixel coordinates.** Store the nearest rendered block identity plus a short text quote. The clicked screen position is only used to choose that anchor; it is never persisted.
4. **Spatial position remains Scratchpad state.** The doc rail aligns anchored notes with document passages when space permits. The full Scratchpad continues to use each note's `x`, `y`, `width`, and `z`.

This gives the user notes “to the right” while preserving the existing Scratchpad's purpose and avoiding a cramped whiteboard inside a sidebar.

## User experience

### Closed by default

- Add a **Notes** action beside Edit and Archive on active document pages.
- The button includes the associated-note count when nonzero, for example `Notes · 3`.
- The rail starts closed for a document that has never been enabled in this browser.
- Toggling it updates `aria-expanded` and moves focus to the rail heading when opened.
- Remember the choice per repository and doc in `localStorage`. Failure to access local storage only makes the choice session-local.
- Archived documents may show existing notes in a read-only rail, but do not offer Add, edit, delete, or promotion actions there.

### Annotation gutter

- Render a slim vertical gutter directly beside the right edge of the reading column on pointer-capable layouts.
- Clicking or tapping it finds the nearest rendered block, opens the Notes rail, inserts an unsaved draft aligned with that block, and focuses its editor.
- Existing anchored notes render as accessible markers in the gutter. Activating a marker opens and focuses that note without creating another.
- The gutter has an explicit accessible name and keyboard-operable anchor buttons; it must not rely on a precise pointer gesture.
- Empty drafts are cancelled rather than persisted. Escape cancels a new draft and returns focus to the triggering gutter position.
- The ordinary Notes action remains available to browse notes without creating one.

### Open rail

The right rail contains:

- a **Notes** heading, count, close button, and **Add note** button;
- associated notes ordered by most recently updated first;
- inline multiline editing with the existing scratchpad colors;
- promote to task/document, open the promoted item, delete, and **Open in Scratchpad** actions;
- visible `S-N` identity and save state;
- an empty state: “No notes for this document yet.”

**Add note** inserts a focused unanchored editor at the top. A gutter-created draft carries the selected passage anchor. Saving creates a normal scratch note associated with this doc. Empty new notes are cancelled rather than persisted. `Cmd/Ctrl+Enter` saves and `Escape` cancels. Existing notes autosave after the same short debounce as the full Scratchpad.

Anchored notes are ordered by document position, with unanchored notes after them ordered by most recently updated. On wide screens cards should align with their passage when possible without overlapping; on compact screens show anchor context such as “Near: The right rail contains…” instead of preserving vertical alignment.

The first release should not include drag/reorder, resize, pan/zoom, multi-select, undo/redo, or an “attach existing note” picker in the rail. Those controls belong to the full Scratchpad. A link from each rail card opens `/scratchpad?note=S-N`, where the full view selects and brings that note into view.

### Responsive layout

On wide screens, use a three-region document layout:

```text
table of contents | document | notes rail
```

- Only allocate the right column while the rail is open.
- Target a 300–340 px sticky rail and keep the reading column at least 620 px wide.
- When the left outline is absent, the document and notes can remain centered as one unit.
- With both outline and notes visible, allow the doc page to use more viewport width rather than squeezing the existing 820 px reading measure into the current 1270 px body cap.
- At widths where a 620 px reading column plus rail no longer fits, use a right-side drawer over the page.
- On compact screens, use a full-width sheet below the document header. Do not place a narrow note column beside the text.

The rail is sticky below the viewport top, scrolls independently when necessary, respects reduced motion, and uses the existing theme, contrast, text-size, and font variables.

## Data model

Extend `ScratchNote` with an optional list:

```yaml
- id: S-12
  text: Verify this claim
  x: 640
  y: 220
  width: 260
  color: yellow
  z: 12
  docs:
    - code: DOC-SCRATCHPAD-DOCS
      anchor:
        block: open-rail
        quote: The right rail contains
  created: "2026-07-12T08:00:00Z"
  updated: "2026-07-12T08:02:00Z"
```

Use `docs`, not a single `doc`, because the same observation can apply to more than one document and can have a different anchor in each. Normalize codes, reject duplicate codes and nonexistent/non-document codes when associating through normal commands, and preserve an association if a doc is temporarily archived.

The optional anchor contains a stable rendered block key and a normalized quote excerpt. Prefer an existing heading id; otherwise derive a deterministic block key from the nearest heading and block ordinal. When rendering, match the block key first and use the quote to relocate the anchor if edits changed block ordinals. If neither matches, retain the association as unanchored and visibly label it “Original passage not found.”

Do not overload `link`. A note may be associated with `DOC-A` while its promotion link points to `T-FOLLOW-UP`; these relationships have different meanings and lifecycles.

### Schema compatibility

Adding a field without changing the schema is unsafe because an older mar binary can read an unknown YAML field and later drop it during a full scratchpad save. Introduce scratchpad schema version 2:

- the new binary reads version 1 and treats missing `docs` as empty;
- the first scratchpad mutation writes version 2;
- an old binary rejects version 2 instead of silently destroying associations;
- tests cover reading v1, writing v2, and preserving `docs` through every mutation.

### Initial spatial placement

A note created in the rail still needs coordinates for `/scratchpad`. Place it in the first free slot in a simple grid to the right of the current note bounds, with the normal default width and next z-index. This is deterministic and prevents every rail-created note from landing on top of `(0, 0)`. The user can arrange it later in the full view.

## Server and command surface

Prefer extending the existing scratchpad contract rather than adding per-doc note storage:

- `GET /doc/{code}` loads the scratchpad once and passes notes whose `docs` contains the document code.
- `POST /scratchpad/note` accepts optional `docs` and validates them.
- `PUT /scratchpad` preserves and validates `docs` during bulk saves.
- Existing optimistic revision checks continue to protect edits from multiple tabs and CLI changes.
- The doc rail must keep the complete scratchpad snapshot in client state when using the bulk endpoint; it must never submit only the filtered notes and thereby delete unrelated notes.

CLI parity:

```text
mar scratch add --text - --doc DOC-CODE
mar scratch edit S-N --doc DOC-CODE
mar scratch edit S-N --remove-doc DOC-CODE
```

`mar scratch show --json` naturally exposes `docs`. A later `--doc DOC-CODE` filter is useful but is not required for the first web integration.

## Document lifecycle

- **Archive:** keep associations. Notes remain project memory and appear read-only on the archived doc.
- **Unarchive:** restores normal rail editing without migration.
- **Recode:** rewrite matching scratch-note `docs` entries while holding the existing store lock. Treat the recode and association rewrite as one logical operation; test failure behavior so it cannot leave a half-rewritten project.
- **Delete:** remove the deleted code from `docs` but preserve each note. A note with no remaining associations remains on the repository Scratchpad.
- **Promotion:** leave `docs` unchanged and set `link` as today.

## Implementation shape

Extract the reusable non-spatial note-card behavior from the large inline Scratchpad script before wiring the rail. Share card rendering, inline editing, colors, promotion, persistence state, conflict handling, and link rendering. Keep plane gestures, selection, history, pan, and zoom exclusive to the full Scratchpad controller.

Avoid copying a second version of scratchpad save/conflict logic into `doc.gohtml`; duplicate state machines will drift quickly. A small embedded static module loaded by both pages is sufficient and does not require a browser dependency.

## Accessibility

- The toggle is a real button with `aria-controls` and `aria-expanded`.
- The rail is an `aside` labelled “Notes for DOC-X”.
- Announce note creation, save failure, conflict, promotion, and deletion through a polite live region.
- Keep every operation available by keyboard without spatial gestures.
- On drawer/sheet layouts, move focus into the panel on open, close with Escape, return focus to the toggle, and prevent focus from disappearing behind the overlay.
- Do not use color as the only note label.

## Conflict and live-reload behavior

The rail shares the current scratchpad revision model. On HTTP 409, stop autosaving and show the existing Reload remote / Keep my version choice inside the rail. “Keep my version” must merge the edited note into the freshly loaded complete scratchpad rather than overwriting unrelated remote notes with a stale array.

The current full Scratchpad overwrite path should be audited at the same time because whole-snapshot replacement is risky when only one note was changed. A note-scoped patch endpoint would make both surfaces safer, but it can be a follow-up if the first implementation retains complete snapshots and tests merging explicitly.

## Alternatives considered

### Embed the full spatial canvas on the right

Rejected for the first version. Pan, zoom, selection, resizing, and a useful canvas require more width than a reading rail provides. It would also compete with document scrolling and keyboard focus.

### Store notes in document frontmatter or body

Rejected. It creates a second note model, makes quick notes bump the document's updated timestamp, and removes them from the repository Scratchpad and CLI workflow.

### Use `link: DOC-X` as the association

Rejected. `link` already means the item created by promotion and permits only one target. Reuse would prevent a note from being contextual to one doc and promoted to a different task.

### Make the rail open state shared project metadata

Rejected. Panel visibility is a personal viewing preference, not project knowledge. It should not create git churn.

## Acceptance criteria

- Document pages render with no notes rail column until the user opens it.
- A gutter click opens the rail with a focused, unsaved draft anchored to the nearest document block; no pixel coordinate is stored.
- Existing anchored-note markers open their notes and keyboard users can perform the same operation.
- Open/closed state is remembered for that repository and document without changing `.mar/`.
- A user can add, edit, recolor, promote, and delete an associated scratch note from the rail, and orphaned anchors degrade to clearly labelled unanchored notes.
- Rail notes are ordinary stable `S-N` notes visible on `/scratchpad`; unrelated notes never appear in the doc rail.
- A note may be associated with multiple docs and may independently link to a promoted task or doc.
- Existing version-1 scratchpad files migrate without data loss, and older binaries fail safely on version 2.
- Opening both the table of contents and notes rail keeps the document readable on wide screens; smaller screens use a drawer or sheet.
- Archive, recode, delete, optimistic conflicts, keyboard navigation, themes, and accessibility preferences have explicit test coverage.
- `make check` passes and desktop/mobile browser checks cover the closed, open, empty, populated, conflict, and archived states.

## Suggested delivery slices

1. Version-2 data model, migration, doc association validation, lifecycle behavior, and tests.
2. Shared note-card/persistence module extracted from the existing Scratchpad UI.
3. Closed-by-default doc toggle, wide rail, create/edit/color/promote/delete, and handler tests.
4. Drawer/sheet responsive behavior, focus management, conflict merging, and browser verification.
5. Deep-link `/scratchpad?note=S-N`, CLI `--doc` parity, documentation, and final checks.
