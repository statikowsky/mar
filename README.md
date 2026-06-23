# mar

Local-first docs and tasks for you and your agents.

A per-directory repository of Markdown docs plus a kanban board. Browse it in
your browser or drive it from the CLI; changes sync both ways. Data is plain
Markdown and YAML on disk: a single binary, no telemetry, no cloud, no
accounts. Agent friendly as every command speaks `--json`.

## Install

Download the archive for your OS from the
[latest release](https://github.com/statikowsky/mar/releases/latest), extract
it, and put `mar` (or `mar.exe`) on your `PATH`. Verify against
`checksums.txt`. Or, with a Go toolchain:

    go install github.com/statikowsky/mar@latest

On macOS the binary is unsigned, so Gatekeeper blocks the first launch. Clear
the quarantine flag once: `xattr -d com.apple.quarantine mar`.

## Quick start

    mar init      # create ./.mar/ (board.yml, tasks/, docs/)
    mar serve     # browse at http://127.0.0.1:7777 (--no-open for headless)
    mar guide     # agent workflow + full command cheatsheet (--json too)

`serve` uses port 7777, or a free port if it's taken; `--port N` forces one.
Each top-level command has a single-letter alias (`mar s`, `mar b show`,
`mar d list`), and every command accepts `--json`.

## Documents

    mar doc create --code AUTH --title "Auth design" --type design --body file.md
    mar doc import report.html --code AUTH --type design   # HTML -> Markdown
    mar doc list
    mar doc show DOC-AUTH

Bodies are GitHub-flavored Markdown — alerts (`> [!WARNING]`) and
syntax-highlighted code included. Link to other docs or tasks inline with
`[[DOC-CODE]]` / `[[T-CODE]]` wiki-links; each doc lists what references it
("Referenced by"). `--body` takes a file path or `-` for stdin.

## Board

    mar board show
    mar task create --title "Wire auth" --column "To do"   # -> T-WIRE-AUTH
    mar task move T-1 --column "In progress" --after T-3
    mar column add "Review" --after "In progress"
    mar doc link DOC-AUTH T-1                               # link doc <-> task

See `mar guide` or `<command> --help` for the full command surface.

## License

MIT — see [LICENSE](./LICENSE).
