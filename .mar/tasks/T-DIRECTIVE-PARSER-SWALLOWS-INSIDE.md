---
title: 'Directive parser swallows ::: inside fenced code blocks'
status: active
created: "2026-06-12T19:19:04.152448Z"
updated: "2026-06-12T19:37:36.648127Z"
---
From project review. `splitSegments` (internal/render/directive.go:31) does not track ``` fence state, so a line like `::: callout` inside a code fence is parsed as a directive opener and the rest of the code block is swallowed up to the next `:::`. Docs that show MAR's own directive syntax in code samples render corrupted.


Resolved by removal: the directive system was dropped (DOC-DIRECTIVES, T-DROP-CUSTOM-DIRECTIVES-GFM). The replacement `mar migrate directives` rewriter is fence-aware, so the bug no longer exists anywhere.
