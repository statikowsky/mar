---
title: Auto-generate task codes as title slugs (T-DESC)
status: archived
created: "2026-06-09T21:24:26.148456Z"
updated: "2026-06-10T21:19:05.087614Z"
---
Change auto-numbered task codes from T-1, T-2... to a slug derived from the
title: "Wire auth login" -> T-WIRE-AUTH-LOGIN (uppercased, non-alphanumerics to
hyphens, collapsed/trimmed). Append -2, -3... on collision. Empty/symbol-only
titles fall back to the numeric T-<seq> code. Existing tasks keep their codes;
--code override still wins. Requested while working on T-2.
