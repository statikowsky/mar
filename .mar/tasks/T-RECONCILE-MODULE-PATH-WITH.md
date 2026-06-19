---
title: Reconcile module path with GitHub remote
status: active
created: "2026-06-10T21:51:20.79258Z"
updated: "2026-06-10T21:56:51.434104Z"
---
The Go module path is `github.com/jopa/mar` (see go.mod) but the repo actually
lives at `github.com/statikowsky/mar`. Because they differ, network install
(`go install github.com/jopa/mar@latest`) cannot resolve — installs only work
from a local checkout via `make install`.

Reconcile so `go install` works over the network:

- Rename the module to `github.com/statikowsky/mar` in go.mod and update all
  internal import paths (internal/..., main.go) and the Makefile ldflags `-X`
  target accordingly.
- Update README install instructions.
- Then the go-store docs (readcube/go-store) can point at a real `go install`.

Found while documenting MAR usage in readcube/go-store.
