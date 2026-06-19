---
title: Client-side Markdown preview in inline editors
status: active
created: "2026-06-12T19:37:15.249454Z"
updated: "2026-06-12T22:45:45.930837Z"
---
Now that custom directives are gone (DOC-DIRECTIVES), inline-editor previews no longer need a server round-trip for *directive* correctness — a client-side GFM renderer could in principle replace the debounced POST /preview in doc.gohtml and board.gohtml.

Decision: CLOSED — keep server-side preview, no change.

Rationale (evaluated 2026-06-13): the preview deliberately reuses render.RenderMarkdown so it matches the saved view exactly. Two server-only concerns remain after directives:
- GFM alerts ([!NOTE] → <blockquote class="alert alert-...">) via a custom goldmark AST transformer.
- Syntax highlighting via chroma with inline styles (WithClasses(false), github style).
A client-side renderer (marked/markdown-it + highlight.js/Prism) would need a reimplemented alert plugin kept in sync with alertVariants + CSS, and could not reproduce chroma's inline-styled output — so the preview would diverge from the real render, defeating its purpose. The only upside (dropping a network hop) is negligible on a localhost tool, and it would add a JS dependency to an app with no build step. Server round-trip is the correct design, not a workaround.
