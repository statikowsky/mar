---
title: MAR landing page — single-page website plan
type: plan
status: active
created: "2026-06-10T22:36:52.546919Z"
updated: "2026-06-11T09:48:46.210515Z"
---
# MAR landing page — single-page website plan

A user-oriented, single-page marketing site for MAR. Goal: a developer lands,
understands what MAR is in seconds, sees it working, and can install it
immediately.

## Hero

**Tagline:**

> Your project's docs and kanban board, in one local folder. Browse in the
> browser, drive from the CLI — local-first and built for both humans and
> agents.

Keep "kanban board" (not "task board") — it is accurate (columns, cards,
reorder) and the right term for a developer audience. Use it consistently
everywhere.

## Sections (top to bottom)

### 1. Primary actions (directly under the tagline)

- **Primary: download a prebuilt binary.** A prominent "Download for <your OS>"
  button (auto-detect OS/arch, link to the latest GitHub Release asset), with a
  small "all platforms" link to the releases page. No Go toolchain needed —
  this is the lowest-friction path and should lead.
- A secondary one-click **copy box** for Go users:
  `go install github.com/statikowsky/mar@latest`.
- Buttons: **GitHub** (primary) and a jump-link to **Quick start**.

Prebuilt binaries are published per OS/arch (macOS arm64/amd64, Linux
amd64/arm64, Windows amd64) on each GitHub Release, with checksums.

### 2. The 30-second demo (centerpiece)

Two side-by-side panels embodying the tagline:

- **Left — drive from the CLI:** a short, real, copy-pasteable terminal
  transcript: `mar init`, `mar task create`, `mar task move`, `mar doc create`.
- **Right — browse in the browser:** a screenshot or short looping GIF of the
  real board UI (with the archived section) and a doc page with a rendered
  callout.

A developer tool converts on "show me," so this section does the heavy lifting.

### 3. Three pillars (proof for the tagline's claims)

One concrete card each:

- **Local-first:** everything lives in one `.mar/` folder next to your code. No
  accounts, no cloud, no lock-in — it's just SQLite in your repo. (Frame as
  local-first, not "privacy as a feature.")
- **CLI-first, browser-friendly:** every command is scriptable; `mar serve`
  gives a live-reloading web view; run it in many projects at once.
- **Built for agents:** every command speaks `--json` with consistent keys, and
  `mar guide` hands an AI agent the whole workflow in one call. This is a
  genuine differentiator — lean into it.

### 4. What's inside (feature strip)

Compact grid — MAR is two products in one folder:

- **Docs:** Markdown or import HTML→MD; types (design/plan/report/…);
  callout / card / phase-block directives; syntax highlighting.
- **Kanban:** custom columns, reorder, archive/restore, auto task codes, link
  cards to docs.
- **One web view:** board + docs + live reload; project path shown per
  instance; light/dark theme.

### 5. Quick start

The real commands, verbatim from the README: prebuilt-binary download (lead),
`go install` (Go users), then `mar init`, `mar serve`, `mar guide`. People
scroll here to confirm it is real and trivial.

### 6. The agent angle (dedicated mini-section)

Its own block because it is the edge and timely. Show actual `mar guide --json`
or `mar task show --json` output (with the `column` and `docs` fields).
Headline e.g. "Your AI pair drives it too." Mention `mar init` scaffolding
`AGENTS.md` once that ships, so agents auto-discover the workflow.

### 7. Footer

GitHub, the `go install` line again, version/license, "local-first · no
telemetry."

## Deliberately left out

- No pricing / signup / "get started free" — it is a local tool; those signals
  undercut the local-first message.
- No long feature essays — let the demo and command snippets carry it.
- No fake testimonials or logo walls.

## Priorities

If only two things ship below the hero: the **copy-able install line** and the
**side-by-side CLI + browser demo**. "See it working and try it in 10 seconds"
converts far better than prose for a developer tool.

## Open questions / notes

- Where does the site live? (Static page; could be served by MAR itself, a
  GitHub Pages site, or a standalone HTML file.) Decide before implementation.
- The "Download for <your OS>" button needs to resolve to the right release
  asset — either client-side OS/arch detection linking to the GitHub Releases
  asset, or a short redirect. Prebuilt binaries are produced by the release
  workflow (see DOC-DIST).
- The `go install` fallback depends on the module path being network-resolvable
  (handled by the github.com/statikowsky/mar rename).

