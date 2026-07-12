---
title: Benchmark read paths before adding store cache
status: active
created: "2026-07-12T20:04:20.329895Z"
updated: "2026-07-12T20:04:20.329895Z"
---
Add representative benchmarks for board, document, and task read paths before considering an in-memory cache. Measure cold and warm behavior at realistic repository sizes, and specify cache invalidation for external CLI or Git edits.
