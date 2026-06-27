# Agent guidelines for mar

Guidance for any AI agent (Claude, etc.) working in this repository.

## Git workflow

- **Do not create new branches unless explicitly instructed.** Work on the
  current branch (normally `main`). When the user wants isolation they will say
  so; otherwise commit directly to the current branch.
- Commit when work is complete and verified. Only push or open PRs when asked.

## Use mar to manage work

This project is `mar` — a local Markdown documentation repository and kanban
board. Dogfood it: track work here with the `mar` CLI instead of an ad-hoc list.

**For the full agent workflow and command cheatsheet, run `bin/mar guide`** (or
`bin/mar guide --json`). It is the canonical reference; the essentials are
below.

Build the binary with `make build` (produces `bin/mar`), then:

- **See the board:** `bin/mar board show`
- **Add a task:** `bin/mar task create --title "..." [--column "To do"] [--body -] [--first|--last|--after T-M|--before T-M|--index N]`
- **Move / reorder:** `bin/mar task move T-N --column "In progress" [--after T-M|--before T-M|--first|--last|--index N]`
  (also supports `--before`, `--first`, `--last`, and `--index`; no placement flag places it at the top of the column)
- **Mark done:** `bin/mar task move T-N --column "Done"`
- **Docs:** `bin/mar doc create|list|show|edit|...` (see `README.md`)

Workflow expectations:

- Before starting work, check `bin/mar board show` for the current task list.
- When you discover a bug, gap, or follow-up, **add it as a task** rather than
  only mentioning it in chat.
- Move a task to `In progress` when you start it and to `Done` when it is
  finished, tested, and committed.

### Docs and plans live in mar, not in files

- **Always store design docs, specs, and implementation plans as mar
  documents** (`bin/mar doc create --type design|plan ...`), not as Markdown
  files under `docs/` or elsewhere in the tree. mar is the documentation
  repository — use it.
- Body can come from a file or stdin (`--body -`); write the content to a temp
  file or pipe it in, create the doc, then remove the temp file. Do not leave
  spec/plan files committed to the repo.
- Link docs to the tasks they describe with `bin/mar doc link DOC-X T-N`.
- The `docs/` directory is not the place for working specs/plans; reserve it
  for anything the build itself requires.

## Quality bar

- Follow the Go conventions already in the codebase: explicit error wrapping
  with `%w`, no comments except minimal godoc, self-documenting names,
  parameterized SQL, table-driven tests.
- Use TDD for features and bug fixes: write the failing test first.
- Before claiming work complete: `make check` (fmt + vet + test-race) must pass.
