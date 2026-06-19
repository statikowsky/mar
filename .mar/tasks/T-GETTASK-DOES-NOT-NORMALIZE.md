---
title: GetTask does not normalize codes the way GetDoc does
status: active
created: "2026-06-12T19:19:04.172345Z"
updated: "2026-06-12T22:29:27.034006Z"
---
From project review. `GetDoc("auth")` resolves to DOC-AUTH, but `GetTask("t-5")` or `GetTask("5")` returns not-found because `normalizeTaskCode` is only applied on create (internal/store/task.go:109, vs doc.go:71). Every task subcommand (show/edit/move/archive/rm) inherits the asymmetry. Same for DeleteTask (task.go:217).
