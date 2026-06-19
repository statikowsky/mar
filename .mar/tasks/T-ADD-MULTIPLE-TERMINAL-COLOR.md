---
title: Add multiple terminal color schemes
status: active
created: "2026-06-12T11:53:04.9895Z"
updated: "2026-06-12T12:03:10.434245Z"
---
Add popular terminal color schemes alongside the existing Light and Catppuccin
Mocha themes: Gruvbox (Light/Dark), Solarized (Light/Dark), Dracula, and Nord.

Keep the sun/moon light/dark toggle, and add a cog button next to it that opens
a popover for picking any scheme (grouped Light/Dark, plus Follow OS). Each
scheme is a fixed palette defined as a `:root[data-theme="<id>"]` CSS variable
block; no structural CSS or DB changes.

**Prerequisite:** design doc DOC-THEMES (approved).
