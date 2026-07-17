---
title: Serve doc-relative images (.mar/docs) in the web UI
status: active
created: "2026-07-17T14:17:04.109383Z"
updated: "2026-07-17T14:17:04.15408Z"
---
Doc bodies can reference images relatively (e.g. `![mockup](img/foo.png)` with
the file at `.mar/docs/img/foo.png`), but `mar serve` has no route for them —
the browser resolves the src against the doc page URL (`/doc/img/foo.png`) and
gets a 404, so images silently don't render.

Fix: `GET /doc/{path...}` serving files from the store's docs dir
(`http.ServeFileFS` over `os.DirFS`), added alongside the more specific
`GET /doc/{code}` page route. Needs a `DocsDir()` accessor on the store.
Found by embedding engine-rendered mockup PNGs in a motorik design doc.
