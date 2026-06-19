---
title: Per-OS binary distribution via GitHub Actions
type: design
status: archived
created: "2026-06-11T09:38:09.063327Z"
updated: "2026-06-11T09:49:45.900608Z"
tasks:
    - T-DISTRIBUTE-PER-OS-BINARIES
---
# Per-OS binary distribution via GitHub Actions

Publish prebuilt `mar` binaries for each OS/arch on a GitHub Release, triggered
by pushing a version tag. Users download a binary instead of needing a Go
toolchain. `go install` remains as the alternative for Go users.

## Why this is easy here

> **Update:** SQLite has since been removed (see [[DOC-STORAGE]]); the store is
> plain files. The pure-Go, no-cgo property below still holds — mar now has no
> cgo dependency at all — so the cross-compile story is unchanged.

MAR uses `modernc.org/sqlite` (pure-Go SQLite) with no cgo anywhere, so every
target cross-compiles from one machine with just GOOS/GOARCH — no C toolchains,
no Docker, no per-OS runners.

## Targets (standard 5)

- darwin/arm64, darwin/amd64
- linux/amd64, linux/arm64
- windows/amd64

## 1. `make release` target — the build (one source of truth)

A Makefile target that cross-compiles all targets into `dist/`, archives each,
and writes checksums. Runs identically locally and in CI.

- Per target: `GOOS=… GOARCH=… CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)"
  -o mar[.exe]`. CGO off is safe (pure-Go SQLite) and yields static binaries.
- Package: `tar.gz` for darwin/linux, `zip` for windows. Each archive also
  includes `README.md` and `LICENSE` when present.
- Naming: `mar_<version>_<os>_<arch>.(tar.gz|zip)`.
- `dist/checksums.txt`: sha256 of every archive.
- `VERSION` uses the existing `git describe` default but is overridable
  (`make release VERSION=v0.1.0`) so CI passes the tag.
- `dist/` is gitignored; `make clean` removes it.

## 2. GitHub Actions workflow (.github/workflows/release.yml)

- **Trigger:** push of a tag matching `v*` (e.g. `v0.1.0`).
- **Job:** ubuntu-latest. `actions/checkout` with `fetch-depth: 0` (full history
  so `git describe` resolves), `actions/setup-go`, then
  `make release VERSION=${{ github.ref_name }}`.
- **Publish:** create/update the GitHub Release for the tag and upload `dist/*`
  (archives + checksums) via `softprops/action-gh-release`. Needs
  `permissions: contents: write`.
- One ubuntu runner builds all targets (cross-compile is free) — no build
  matrix.

Nothing auto-publishes until a `v*` tag is pushed.

## 3. Docs

- **README:** add an Install section leading with prebuilt-binary download
  (link to Releases; one-liner per OS to extract + put on PATH). Keep
  `go install …@latest` as the alternative for Go users.
- **Website plan (DOC-WEBSITE):** note prebuilt download as the primary path for
  non-Go users; the install copy-box can keep `go install`.

## 4. Verification

Build plumbing, so no Go unit tests apply. Verify by:

- `make release VERSION=v0.0.0-test` locally → 5 archives + `checksums.txt` in
  `dist/`, each archive contains a `mar` binary.
- Extract the native archive (darwin/arm64), run `./mar version` → prints
  `v0.0.0-test`.
- `make check` still green.
- The workflow is validated by cutting a real tag after merge (done by the user;
  no tags pushed automatically).

## Out of scope

- Homebrew/Scoop taps, Linux packages, Docker images (a later GoReleaser
  migration could add these without changing the tag-based release ritual).
- Code signing / notarization of the macOS binary.
