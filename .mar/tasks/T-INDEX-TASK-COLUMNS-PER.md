---
title: Index task columns per store snapshot
status: active
created: "2026-07-12T20:04:27.801585Z"
updated: "2026-07-12T20:11:03.204931Z"
---
Build a task-code-to-column index once per loaded snapshot and use it when converting task entities. Keep board repair semantics unchanged; prioritize this only after read-path benchmarking confirms it matters.
