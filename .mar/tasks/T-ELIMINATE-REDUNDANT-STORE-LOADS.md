---
title: Eliminate redundant store loads on doc and task pages
status: active
created: "2026-07-12T20:04:17.283541Z"
updated: "2026-07-12T20:10:46.402082Z"
---
Replace the chained read calls on document, task, and index pages with one-snapshot view methods. First merge the internal double-loads in TasksForDoc and DocsForTask; preserve fresh reads across external file edits.
