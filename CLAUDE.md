# CLAUDE.md

See [AGENTS.md](./AGENTS.md) for working guidelines in this repository.

Key points:

- Do not create new branches unless explicitly instructed; work on the current
  branch.
- Use mar itself (`bin/mar`) to track and manage work — check the board before
  starting, add tasks for bugs/follow-ups, and move tasks through the columns.
- Always store design docs, specs, and plans as mar documents
  (`bin/mar doc create --type design|plan ...`), not as Markdown files under
  `docs/`. Do not leave spec/plan files committed to the repo.
