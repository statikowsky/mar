---
title: Scratchpad whiteboard design
type: design
status: active
created: "2026-07-11T17:53:44.761362Z"
updated: "2026-07-11T17:53:44.761362Z"
tasks:
    - T-SPECIFY-SCRATCHPAD-WHITEBOARD
---
## Summary

Add a project-level **Scratchpad**: a large spatial surface for quickly capturing and arranging ideas as text notes. A user opens it, double-clicks anywhere, types, and moves on. Notes can be repositioned, grouped visually, and promoted into mar tasks or documents when an idea becomes real work.

The Scratchpad should feel closer to sticky notes on an infinite desk than to a general-purpose drawing application. The first release should favor fast text capture, keyboard fluency, local persistence, and integration with mar over freehand drawing or elaborate diagramming.

## Product principles

1. **Capture in two seconds.** Opening the surface and adding an idea should require no form or naming step.
2. **Spatial, but still text-native.** Position carries meaning, while note contents remain selectable, searchable, accessible text.
3. **Local-first and inspectable.** Scratchpad data lives under `.mar/`, is safe to commit, and has no cloud dependency.
4. **A path from vague to concrete.** An idea can become a task or document without copying it manually.
5. **Progressive complexity.** The initial tool is excellent for notes; drawing, connectors, and multiple boards can come later.

## Recommended product shape

Start with **one Scratchpad per mar repository**, available from the main navigation at `/scratchpad`. This avoids a board-picker ceremony before capture. The storage format should leave room for named boards later, but multiple scratchpads are not needed in the first release.

Use regular HTML elements positioned on a transformed plane rather than rendering notes into an HTML `<canvas>`. DOM notes provide native text selection, editing, focus, screen-reader semantics, theme support, and simpler testing. The surface can still pan and zoom like a whiteboard.

## MVP features

### Capture and edit

- Double-click empty space to create a note at that position and immediately focus its editor.
- A visible **Add note** button and keyboard shortcut provide non-pointer alternatives.
- Notes support multiline plain text. Render URLs and mar wiki-links as links after editing, but do not introduce rich-text controls.
- `Cmd/Ctrl+Enter` finishes editing; `Escape` cancels a newly empty note or restores existing content.
- Double-click a note or press `Enter` while it is selected to edit it.
- Autosave after a short debounce, with a small `Saving…`, `Saved`, or `Save failed` status.

### Arrange

- Drag a note to move it.
- Resize a note from a visible handle, with sensible minimum and maximum widths.
- Select one note by clicking it; `Shift+click` adds or removes notes from the selection.
- Drag a marquee on empty space to select several notes.
- Move a multi-selection together.
- Duplicate and delete selected notes from the toolbar or keyboard.
- Choose from a small theme-aware color palette: neutral, blue, green, yellow, red, and purple.
- New or interacted-with notes come to the front. Persist an integer stacking order rather than arbitrary CSS values.

### Navigate the surface

- Drag empty space to pan; hold `Space` while dragging so pan remains available when starting over a note.
- Mouse wheel or trackpad scroll pans. `Cmd/Ctrl+wheel` and toolbar controls zoom around the pointer.
- Provide zoom in, zoom out, reset to 100%, and **Fit all notes** controls.
- Keep a finite safe coordinate range internally while presenting the surface as effectively infinite.
- Remember the last viewport locally in the browser. Note data is project state; each user's viewport is not and should not create repository churn.

### Turn ideas into mar work

- **Create task from note** opens the normal task creation flow with the first line as the title and remaining text as the body.
- **Create document from note** opens document creation with the note text as the initial body.
- After creation, retain the note and show a small link to the resulting `T-*` or `DOC-*` item. Do not silently delete the source idea.
- Render existing `[[T-*]]` and `[[DOC-*]]` references inside notes using mar's normal wiki-link behavior.

### Recovery and safety

- Undo and redo note creation, text edits, moves, resize, color changes, duplication, and deletion for the current browser session.
- Warn before leaving only when a save request is still pending or has failed.
- Server writes remain atomic and use the existing mar store lock.
- If the store changes externally, reload when there are no local changes; otherwise show a conflict banner with **Reload remote** and **Keep my version** choices.

## Keyboard and accessibility

The spatial view must not be the only usable representation.

- Every note is focusable and exposes its text, color label, linked item, and selection state.
- Arrow keys move focus between notes using approximate spatial direction; `Tab` reaches toolbar controls.
- With a note selected, `Cmd/Ctrl+D` duplicates, `Delete` removes, `Enter` edits, and arrow keys nudge it. `Shift+arrow` uses a larger nudge.
- `Cmd/Ctrl+Z` and `Cmd/Ctrl+Shift+Z` undo and redo.
- Provide a **List view** that presents every note in creation order with edit, color, promote, and delete actions. This is both an accessibility fallback and useful on small screens.
- Announce save failures, creation, deletion, and promotion through an ARIA live region.
- Respect reduced-motion, theme, contrast, text-size, and font preferences already supported by mar.

## Mobile behavior

The MVP is desktop-first but must remain usable on touch devices:

- Tap **Add note**, then tap a location or create at the viewport center.
- One-finger drag moves a selected note; dragging empty space pans.
- Pinch zooms the surface.
- A bottom action bar exposes edit, color, promote, duplicate, and delete.
- List view is the recommended compact-screen mode and should be easy to switch to.

## Persistence model

Recommended initial file: `.mar/scratchpad.yml`.

```yaml
version: 1
next_note: 4
notes:
  - id: S-1
    text: Explore offline search
    x: 240
    y: 180
    width: 260
    color: yellow
    z: 1
    link: T-EXPLORE-OFFLINE-SEARCH
    created: "2026-07-11T18:00:00Z"
    updated: "2026-07-11T18:05:00Z"
```

One YAML file keeps a board snapshot atomic and makes multi-note moves one write. Spatial editing naturally causes noisy diffs, so optimizing for one-file-per-note diffs is less valuable than consistent snapshots. Store integer coordinates and widths; derive height from content. Do not persist browser zoom or pan.

Use stable `S-N` IDs so notes can be addressed from the CLI and eventually referenced elsewhere. Include a schema version from the start. Unknown future fields should be preserved if practical or rejected with a clear version error rather than discarded.

## CLI and HTTP surface

The browser is primary, but core operations should remain agent-friendly:

```text
mar scratch show
mar scratch add --text - [--x N --y N --color C]
mar scratch edit S-1 [--text - --x N --y N --width N --color C]
mar scratch rm S-1 --force
mar scratch promote S-1 --task|--doc [creation flags]
```

All commands support `--json`. A bulk browser endpoint should save an ordered set of changes with the store data version it was based on. The server rejects a stale version instead of overwriting external edits.

## Deliberately outside the MVP

- Freehand pen/highlighter strokes.
- Shapes, arrows, connectors, and snap-to-grid.
- Images, file attachments, embeds, or pasted screenshots.
- Frames, groups, templates, and presentation mode.
- Multiple named scratchpads.
- Real-time multi-user collaboration, cursors, or comments.
- AI clustering or automatic rewriting.
- Export to PNG/PDF.

These features would pull the product toward Miro or Excalidraw before mar has validated the simpler idea-capture loop.

## Follow-on phases

### Phase 2: structure

- Connectors between notes.
- Frames with titles and optional background colors.
- Group/ungroup and align/distribute commands.
- Search and filter notes.
- Export selected notes to a Markdown document.

### Phase 3: richer boards

- Multiple named scratchpads.
- Images and attachments stored under `.mar/assets/`.
- Shapes and freehand strokes.
- PNG/SVG export and presentation mode.

## MVP acceptance criteria

- A new repository can open an empty Scratchpad without creating data until the first note is saved.
- A user can add, edit, move, resize, recolor, duplicate, multi-select, and delete notes.
- Pan, zoom, reset, and fit-all work with mouse, trackpad, keyboard controls, and touch where applicable.
- Refreshing or restarting mar preserves all note content and layout.
- Autosave state and failures are visible, and stale updates cannot silently overwrite external changes.
- A note can create and link to a mar task or document.
- The list view supports every destructive/content-changing note operation without spatial gestures.
- Core note operations are exposed through `mar scratch` with JSON output.
- Existing docs, task board, themes, accessibility preferences, and live reload continue to work.

## Suggested implementation slices

1. Store schema and tested CRUD/version-conflict operations.
2. `mar scratch` CLI commands and JSON contracts.
3. Read-only spatial and list views with navigation entry.
4. Create/edit/autosave and error states.
5. Move/resize/color/delete/duplicate plus undo/redo.
6. Pan/zoom/fit and multi-selection.
7. Task/document promotion and wiki-links.
8. Touch, accessibility audit, and end-to-end browser tests.

## Decisions to validate with a prototype

- Whether creation should be single-click in an active text tool or double-click with the normal select tool. Recommendation: double-click plus an explicit Add note mode.
- Whether Markdown rendering inside compact notes helps more than it distracts. Recommendation: start with plain text plus link detection.
- Whether the single YAML snapshot remains comfortable once a board reaches hundreds of notes. Recommendation: set a soft performance target of 500 notes and revisit storage only with evidence.
