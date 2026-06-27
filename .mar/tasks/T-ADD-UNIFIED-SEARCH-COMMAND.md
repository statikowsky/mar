---
title: Add unified search command
status: active
created: "2026-06-27T12:12:26.187534752Z"
updated: "2026-06-27T14:09:02.00160645Z"
---
Add a single `mar search <term>` command across docs and tasks.

Current workaround is `mar doc list | grep` plus raw `grep -r .mar/`, which loses MAR metadata and forces users/agents outside the CLI.

Desired scope:
- Search document and task titles and bodies.
- Return source kind, code, title, field/body snippet, and match context.
- Support `--json` for agents.
- Consider filters such as `--docs`, `--tasks`, `--type`, `--status`, and case-sensitive/insensitive behavior.
- Use a pure Go scan over parsed MAR objects as the correctness baseline, but consider optional local-tool acceleration when available: `rg --json` first, then `git grep`, with raw grep only as a last resort.
- Ensure any local-tool backend maps file matches back to MAR docs/tasks and does not treat YAML frontmatter as body text.

See [[DOC-SEARCH]].
