---
title: Multiple terminal color schemes
type: design
status: archived
created: "2026-06-12T11:53:32.779132Z"
updated: "2026-06-12T20:44:09.136149Z"
tasks:
    - T-ADD-MULTIPLE-TERMINAL-COLOR
---
# Multiple terminal color schemes!

Add popular terminal color schemes alongside the existing **Light** and
**Catppuccin Mocha** themes, selectable from the web UI. Purely presentational:
new CSS variable blocks plus a scheme picker. No server data model or DB changes.

## Background

Today theming is a binary light↔dark toggle:
- A `theme` cookie holds `light`, `dark`, or empty (= follow the OS).
- The server (`themeFromRequest` in `internal/web/server.go`) validates the
  cookie and renders it as `data-theme` on `<html>` so there is no flash on load.
- A sun/moon toggle button (`layout.gohtml`) flips `data-theme` instantly and
  writes the cookie.
- `style.css` defines `:root` (light) and `:root[data-theme="dark"]`
  (Catppuccin Mocha) variable blocks, plus a `prefers-color-scheme: dark`
  fallback that mirrors the dark block for the follow-OS case.

All UI colors already flow through CSS variables (`--bg`, `--surface`, `--fg`,
`--accent`, semantic accents, doc-type tag colors), so adding a scheme is just a
new variable block — no structural CSS changes.

## Scheme set

Eight named schemes, each a fixed palette:

- **Light** (existing) — id `light`
- **Catppuccin Mocha** (existing dark) — id `dark`
- **Gruvbox Light** — id `gruvbox-light`
- **Gruvbox Dark** — id `gruvbox-dark`
- **Solarized Light** — id `solarized-light`
- **Solarized Dark** — id `solarized-dark`
- **Dracula** (dark) — id `dracula`
- **Nord** (dark) — id `nord`

The existing ids `light` and `dark` are kept as-is for backward compatibility
with cookies already set in users' browsers. New schemes get descriptive ids.

Each scheme is one `:root[data-theme="<id>"]` block in `style.css` overriding the
full existing variable set (backgrounds, surfaces, text, lines, accent +
accent-soft, the semantic accents warn/good/skip/danger, and the doc-type tag
bg/fg pairs). The `prefers-color-scheme` fallback block is unchanged (follow-OS
still resolves to Light or Catppuccin Mocha).

## Selection model

- Single source of truth for the server render is the **`theme` cookie** = the
  active scheme id, or empty = **Follow OS** (resolved client-side to `light` or
  `dark` exactly as today).
- `themeFromRequest` is widened from accepting only `light`/`dark` to accepting
  any id in the known scheme set; unknown values fall back to "" (follow OS).
- Each scheme is tagged **light** or **dark** in a small map, mirrored in JS, so
  the sun/moon toggle knows which schemes are "light side" vs "dark side".

## UI: toggle + cog

- **Keep the sun/moon toggle** as a one-click light↔dark switch.
- **Add a cog button** next to it. Clicking it opens a small popover listing all
  schemes grouped under **Light** / **Dark**, plus a **Follow OS** entry; the
  active scheme is checked. Selecting one sets the `theme` cookie and updates
  `data-theme` instantly (no reload).
- The toggle remembers the last-chosen scheme **per mode** in `localStorage`
  (`theme-light` / `theme-dark`, defaulting to `light` / `dark`). Flipping
  dark→light→dark returns you to e.g. Gruvbox Dark rather than a generic default.
  Picking a scheme from the cog updates the matching per-mode key.
  `localStorage` (not a cookie) is sufficient because it only drives toggle
  clicks, not the initial server render.

## Files touched

- `internal/web/static/style.css` — six new `:root[data-theme="<id>"]` variable
  blocks (Gruvbox light/dark, Solarized light/dark, Dracula, Nord).
- `internal/web/server.go` — widen `themeFromRequest` to validate against the
  known scheme-id set.
- `internal/web/templates/layout.gohtml` — cog button + scheme popover markup;
  picker JS (open/close, select → cookie + `data-theme`); update the existing
  toggle JS to use the per-mode `localStorage` memory and the light/dark map.
- `internal/web/server_test.go` — `themeFromRequest` accepts each known id and
  rejects unknown values (→ "").

## Out of scope

- No server data model or DB changes.
- No per-document or per-board theme overrides; the theme is global per browser.
- No custom/user-defined palettes; the scheme set is fixed in CSS.

## Testing

- `themeFromRequest` returns the id for each known scheme cookie, "" for unknown
  or missing.
- Manual: cog popover lists all schemes grouped by light/dark, selecting applies
  instantly and survives reload (cookie); toggle flips between the last-used
  light and dark schemes.
