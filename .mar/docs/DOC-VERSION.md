---
title: Report mar version (CLI, web UI, server start)
type: design
status: archived
created: "2026-06-10T21:34:13.410876Z"
updated: "2026-06-10T21:39:13.362314Z"
tasks:
    - T-REPORT-MAR-VERSION-IN
---
# Report mar version (CLI, web UI, server start)

Surface a single mar version string across the CLI, the web UI, and the server
startup line, from one source of truth.

## Background

`main.go` already declares `var version = "dev"` and the Makefile injects it
via `-X main.version=$(VERSION)` (from `git describe --tags --always --dirty`).
But that variable is never read anywhere — no command prints it, the web UI does
not show it, and `mar serve` startup omits it. Because Go cannot import package
`main`, the variable must move somewhere importable.

## 1. Version source — internal/version

New package `internal/version` with a single exported variable:

    package version
    var Version = "dev"

- Makefile ldflags change from `-X main.version=$(VERSION)` to
  `-X github.com/jopa/mar/internal/version.Version=$(VERSION)`.
- The unused `var version` in `main.go` is removed.
- CLI, web, and serve all read `version.Version`. Single source of truth.

## 2. CLI — command + flag

- **`mar version`** subcommand: prints `mar <version>`; with `--json` prints
  `{"version": "<version>"}`. Registered alongside the other subcommands.
- **Root `--version` / `-v`** flag via Cobra's built-in version support
  (`rootCmd.Version` + a version template) so `mar --version` and `mar -v`
  print `mar <version>`.

## 3. Web UI — banner

Append the version to the existing project-path banner so it reads
`/path/to/project · mar <version>`. The banner already flows through
`layout.gohtml` (the `ProjectPath` var) and `server.go`'s data injection; pass a
`Version` value the same way and render it inline. `NewServer` reads
`version.Version` directly — no new constructor parameter.

## 4. Server startup

`mar serve` currently prints `MAR serving <repo> at <url>`. Extend it to
`MAR <version> serving <repo> at <url>` so the version shows on boot.

## 5. Testing (TDD)

`make check` (fmt + vet + test-race) must pass before commit.

- **version pkg:** assert `Version` defaults to `"dev"` (the var exists and is
  exported; test binaries are not built with ldflags).
- **CLI:** `mar version --json` emits `{"version": ...}`; `mar version` prints
  the plain string; the `--version` flag prints it. Assertions compare against
  `version.Version` ("dev" under test) rather than a hardcoded number.
- **Web:** a served page's HTML contains the version string in the banner.
- **Serve:** the startup line includes the version.

## Out of scope

- A `/version` HTTP/JSON API endpoint (the existing `/events/version` is the
  unrelated live-reload data-version counter, not the build version).
- Changing how the version number itself is computed (git describe stays).
