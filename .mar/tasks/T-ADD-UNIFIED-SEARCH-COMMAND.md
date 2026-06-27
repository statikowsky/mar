---
title: Add unified search command
status: active
created: "2026-06-27T12:12:26.187534752Z"
updated: "2026-06-27T12:12:26.187534752Z"
---
Add a single `mar search <term>` command across docs and tasks.

Current workaround is `mar doc list | grep` plus raw `grep -r .mar/`, which loses MAR metadata and forces users/agents outside the CLI.

Desired scope:
- Search document and task titles and bodies.
- Return source kind, code, title, field/body snippet, and match context.
- Support `--json` for agents.
- Consider filters such as `--docs`, `--tasks`, `--type`, `--status`, and case-sensitive/insensitive behavior.
