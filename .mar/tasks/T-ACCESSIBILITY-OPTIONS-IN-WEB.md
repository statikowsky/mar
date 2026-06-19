---
title: Accessibility options in web UI
status: active
created: "2026-06-16T19:25:25.158089Z"
updated: "2026-06-16T19:32:35.978881Z"
---
Add user-configurable accessibility settings to the web UI, persisted (cookie/localStorage, matching the theme toggle approach).

Requested:
- Text size (e.g. small/normal/large, or a scale slider)
- Font family choice (e.g. Inter / system / a dyslexia-friendly font like OpenDyslexic / monospace)
- Contrast (high-contrast theme variant)

Other ideas:
- Respect `prefers-reduced-motion` for any animations/transitions
- Respect `prefers-color-scheme` (already partly handled via theme toggle)
- Line-height / content-width (reading width) control
- Proper focus-visible outlines and full keyboard navigation
- Skip-to-content link and correct ARIA landmarks/labels

ponytail: lean on native CSS — `rem`-based sizing scaled by a root font-size variable, `@media (prefers-reduced-motion)`, CSS custom properties for the contrast variant. No JS framework, no settings backend.
