---
title: Remove mar migrate directives command
status: active
created: "2026-06-12T19:39:48.997684Z"
updated: "2026-06-12T19:40:30.643852Z"
---
The only project that used ::: directives has been migrated, so the one-shot migration tooling is no longer needed. Remove the migrate CLI command, internal/migrate, and store.RewriteBodies; update README, guide, and amend DOC-DIRECTIVES.
