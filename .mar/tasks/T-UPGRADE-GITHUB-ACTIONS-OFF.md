---
title: Upgrade GitHub Actions off Node 20
status: active
created: "2026-07-11T17:48:17.168985Z"
updated: "2026-07-11T17:48:17.168985Z"
---
The v0.4.0 release workflow succeeded but GitHub warned that actions/checkout@v4, actions/setup-go@v5, and softprops/action-gh-release@v2 target deprecated Node.js 20 and are being forced onto Node.js 24. Upgrade to supported action versions and verify both CI and release workflows.
