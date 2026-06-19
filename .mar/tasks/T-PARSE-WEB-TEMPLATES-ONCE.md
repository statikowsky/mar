---
title: Parse web templates once at startup instead of per request
status: active
created: "2026-06-12T19:19:04.232563Z"
updated: "2026-06-12T22:42:40.263715Z"
---
From project review. `render`/`renderFragment` (internal/web/server.go:67-84,116-124) call `template.Must(srv.tmpl.Clone())` + ParseFS on every request — wasteful and panics on error instead of returning 500; a mid-render failure can also append error text to a partial 200 body. Parse once in NewServer. Optional hardening while there: archive/delete POSTs have no CSRF/Origin check (body-less, so no preflight) — low risk for a localhost tool but cheap to add an Origin check.
