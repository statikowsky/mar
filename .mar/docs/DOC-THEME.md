---
title: Dark theme with cookie-persisted toggle
type: design
status: archived
created: "2026-06-10T22:27:54.325134Z"
updated: "2026-06-10T22:33:45.013329Z"
tasks:
    - T-DARK-THEME-CSS-PALETTE
    - T-THEME-TOGGLE-COOKIE-PERSISTENCE
---
# Dark theme with cookie-persisted toggle

Add a dark theme to the MAR web UI, inspired by a blue palette, with a toggle
that defaults to the OS preference and remembers an explicit choice via a
cookie.

## Palette

Dark theme inspired by:

- `#020817` near-black blue — page background
- `#03122f` very dark blue — surfaces, cards, code, table heads
- `#06376e` dark blue — borders, accent-soft
- `#0a5d9e` medium blue
- `#2286c4` bright blue — accent (links, strips)
- `#73bce4` light blue — muted text, link/strip highlight
- `#dff7ff` near-white blue — foreground text

## 1. Theme model & palette

Two themes driven by the existing CSS custom properties. The current `:root`
block stays as the **light** theme. A `[data-theme="dark"]` block overrides the
same variables with the blue palette:

- `--bg: #020817`; surfaces/cards `#03122f`
- `--fg: #dff7ff`; `--muted: #73bce4`
- `--accent: #2286c4`; `--accent-soft: #06376e`; links/strips lean on `#73bce4`
- `--code-bg` / `--table-head`: `#03122f`; `--line: #06376e`

The active theme is selected by a `data-theme` attribute on `<html>` (`light`
or `dark`). All variable-driven CSS adapts automatically.

## 2. Hardcoded colors

Some rules use literal colors that will not adapt: the `.typetag.*` doc-type
badges and the semantic `--warn/good/danger/skip` soft backgrounds are
light-tuned and would glow on a dark background.

- Promote the typetag colors to CSS variables so the dark block can retune them.
- Override the semantic `*-soft` values in the dark block to muted,
  dark-appropriate tints so callouts, pills, and phase-blocks read well.

## 3. Resolution order (default = follow OS, explicit choice remembered)

Three-state preference: `light`, `dark`, or unset (= follow OS).

- **Server side:** the handler reads a `theme` cookie. If it is `light` or
  `dark`, render `<html data-theme="...">` directly — correct on first paint, no
  flash.
- **No cookie (follow OS):** render with no `data-theme`. A tiny inline `<head>`
  script sets the attribute from `prefers-color-scheme` before first paint, and
  an `@media (prefers-color-scheme: dark)` CSS fallback keeps it correct with JS
  disabled.

## 4. Toggle + persistence

A small fixed top-right button (sun/moon icon), present on every served page.
Clicking it:

- flips `data-theme` on `<html>` immediately (no reload), and
- writes a `theme` cookie via `document.cookie` (~1 year, `SameSite=Lax`) so the
  server renders the chosen theme on subsequent loads.

No new endpoint and no store/db change — the cookie is written client-side and
read server-side. The cookie is per-browser and does not touch the git-tracked
`mar.db`.

## 5. Wiring

- `layout.gohtml` gains: the `data-theme` attribute driven by a `Theme`
  template value, the toggle button, the pre-paint OS script, and the toggle
  click script.
- `server.go`'s `render` resolves the theme from the request cookie via a small
  `themeFromRequest(r)` helper and injects it into the template `data` map.
  Because `render` currently has no request access, pass the resolved theme (or
  the request) through to it.

## 6. Testing

`make check` (fmt + vet + test-race) green before commit.

- Handler test: a request with `theme=dark` cookie yields response HTML
  containing `data-theme="dark"`; `theme=light` yields `data-theme="light"`; no
  cookie yields no `data-theme` attribute (OS-driven path).
- End-to-end browser check: toggle dark, reload, confirm persistence; verify
  callouts/pills/typetags/board read well in both themes.

## Out of scope

- Per-user server-side storage in the db (rejected: the db is git-tracked, so a
  viewer preference should not live there).
- More than two themes.
