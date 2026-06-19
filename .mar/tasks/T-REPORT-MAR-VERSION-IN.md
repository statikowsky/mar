---
title: Report mar version in CLI, web UI, and on server start
status: active
created: "2026-06-10T21:26:37.504186Z"
updated: "2026-06-10T21:39:13.380921Z"
---
Surface a mar version string in three places:

- **CLI:** a `mar version` command (and/or `--version` flag) printing the
  version; support `--json`.
- **Web UI:** show the version somewhere unobtrusive (e.g. footer or alongside
  the project-path banner) on the served pages.
- **Server start:** log/print the version when `mar serve` boots.

Decide a single source of truth for the version (build-time ldflags var vs a
constant) in the design. Prerequisite: write a design doc first.
