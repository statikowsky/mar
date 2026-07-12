---
title: Document note save-state interaction
type: analysis
status: active
created: "2026-07-12T15:17:59.388699Z"
updated: "2026-07-12T15:17:59.388699Z"
tasks:
    - T-CLARIFY-DOCUMENT-NOTE-SAVE
---
# Current behavior

Document-rail notes are always rendered as textareas. Existing notes persist only when Save is clicked; new drafts are created only when Save is clicked. The button is always enabled and visually uses the generic Scratchpad button rule, while the adjacent navigation and delete actions use the newer note-action components. Document-note edits are not included in the page's dirty-state unload warning.

Scratchpad canvas edits follow a mostly automatic model: committed edits schedule a debounced save and expose global Saving/Saved status. This makes the two surfaces communicate persistence differently.

# Recommendation

Keep explicit save in the document rail for now. Saving the entire scratchpad state on every note keystroke would introduce overlapping revision writes across multiple open note cards and obscure conflicts. Make the state explicit instead:

- Existing note starts clean with a disabled `Save changes` button.
- Input that differs from the last persisted value enables a primary-styled `Save changes` button with a Save icon.
- While saving, disable it and show `Saving…`; after success return to disabled `Saved` briefly, then `Save changes`.
- A new draft uses `Create note`, disabled while empty, rather than `Save`.
- Cmd/Ctrl+Enter triggers the enabled action.
- Closing the rail, opening Scratchpad, deleting the note, or leaving the page while dirty should ask before discarding changes.
- Failed saves retain the dirty state and expose an inline error/status rather than relying only on alert.

The Save control should share the same height, radius, typography, icon sizing, hover, and focus treatment as note navigation and delete controls. It should use the accent-filled primary variant because it is the main commit action; the other actions remain secondary or dangerous. Matching does not require all three actions to have identical emphasis.
