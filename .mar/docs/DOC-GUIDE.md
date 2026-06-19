---
title: mar guide — one-call agent onboarding
type: design
status: archived
created: "2026-06-10T22:07:53.721464Z"
updated: "2026-06-10T22:10:47.174835Z"
tasks:
    - T-ADD-MAR-GUIDE-COMMAND
---
# mar guide — one-call agent onboarding

A `mar guide` command that prints MAR's agent workflow plus a compact command
cheatsheet, so an agent onboards in a single call instead of multi-turn `--help`
exploration.

## Motivation

An agent's use of MAR depends on discovering its workflow. `mar init` creates
only `.mar/mar.db` — a fresh agent has no signal about conventions. `mar guide`
gives a single authoritative reference an agent (or human) can read in one call.

## Command

- New top-level `mar guide` subcommand, alias `g`.
- No store access — prints static embedded content, so it works before
  `mar init` (an agent can learn the tool first).
- Supports `--json`.

## Content

A single embedded Markdown document covering:

- **Workflow:** check the board first; dogfood tasks; move cards through
  columns; store specs/plans as MAR docs (not loose files); the universal
  `--json` contract; archive cards when done.
- **Command cheatsheet:** every command grouped (init, task, doc, board,
  column, serve, version, guide) with key flags and the JSON shape returned.
  Compact and scannable.

## Source of truth & wiring

- Guide text lives in `internal/cli/guide.md`, embedded via `//go:embed
  guide.md`. Editing the guide is a Markdown edit, no Go change.
- `newGuideCmd()`: plain prints the Markdown; `--json` emits
  `{"guide": "<markdown>"}` via the existing `printJSON`. Registered on root
  alongside `version`, alias `g`.
- The guide is the canonical, richest copy of the workflow conventions.
  `AGENTS.md` points to it ("run `mar guide` for the full reference") rather
  than duplicating the command list, so the two do not drift.

## Output format decision

Markdown by default (token-efficient, readable by humans and agents, matches
MAR's doc ethos), with `--json` wrapping it as `{"guide": "<markdown>"}` to
honor the universal `--json` contract without maintaining two content
representations.

## Testing (TDD)

`make check` (fmt + vet + test-race) green before commit.

- `mar guide` exits 0 and prints non-empty Markdown containing anchor strings:
  a workflow marker ("check the board"), the `--json` contract mention, and
  representative commands (`task create`, `doc create`, `board show`).
- `mar guide --json` returns valid JSON with a non-empty `guide` field
  containing those same anchors.
- `mar g` (alias) behaves identically.
- Coverage guard: every top-level command name (derived from the root command's
  registered subcommands) appears in the guide, so adding a command without
  documenting it fails the test.
