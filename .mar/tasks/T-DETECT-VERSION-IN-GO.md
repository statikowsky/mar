---
title: Detect version in Go-installed binaries
status: active
created: "2026-07-16T14:26:28.889589Z"
updated: "2026-07-16T14:28:21.372198Z"
---
Make binaries installed with go install report their embedded module version instead of dev. Preserve Makefile ldflag versions and dev behavior for local/test builds.
