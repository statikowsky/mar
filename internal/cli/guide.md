# mar — agent guide

mar is a local, per-directory Markdown documentation repository plus a kanban
board. You drive it from the CLI; humans browse it with `mar serve`. The store
is plain text under `./.mar/` — `board.yml` (columns and card order),
`tasks/*.md`, and `docs/*.md` with YAML frontmatter — discovered by walking up
from the current directory (like `git`). Run commands from the project root;
commit `.mar/` like any other source.

## Workflow

- **Check the board first.** Run `mar board show` before starting work to see
  the current tasks and columns.
- **Track work as tasks.** When you discover a bug, gap, or follow-up, add a
  task (`mar task create`) instead of only mentioning it.
- **Move cards as you go.** Move a task to `In progress` when you start it and
  to `Done` when it is finished, tested, and committed. Archive finished cards
  with `mar task archive` to keep the board clean.
- **Store specs and plans as mar docs**, not as loose Markdown files in the
  tree: `mar doc create --type design|plan ...`. Link them to their task with
  `mar doc link`.
- **Every command accepts `--json`** for structured output. On failure a
  command prints `{"error": "..."}` to stderr and exits non-zero. Prefer
  `--json` when parsing output.

## Codes

- Tasks have a code like `T-WIRE-AUTH` (auto-slugged from the title, capped to
  a few words) or `T-42` (pass `--code 42`). Most commands accept the code or
  a number.
- Docs have a code like `DOC-AUTH`. Pass `--code AUTH`; the `DOC-` prefix is
  added for you.

## Linking

- **Explicit links** tie a doc to a task: `mar doc link` / `mar task link`.
  They show as "tasks" on a doc and "docs" on a task.
- **Wiki-links** connect any doc or task to any other from inside a body: write
  `[[DOC-CODE]]`, `[[T-CODE]]`, or `[[CODE|label]]`. They render as real links
  in `mar serve`; a target that doesn't exist yet renders as a muted
  placeholder ("red link"), so you can link before the page exists.
- **Backlinks** are derived from those wiki-links: `mar doc show` (and the web
  doc page) list every doc and task whose body links to it under
  "Referenced by". `mar backlink CODE` is the focused query. Nothing to
  maintain — edit a body and the backlinks follow.
- **Lint** finds dangling wiki-links: `mar doc lint` reports `[[...]]` targets
  that don't resolve. Dangling links are fine by default (link before the page
  exists); add `--strict` to fail CI on them.

## Commands

Each command takes `--json`. Single-letter aliases shown in parentheses.

### init (i)

    mar init                       # create ./.mar/ (board.yml, tasks/, docs/)
        -> {"initialized": true, "path": ".mar/"}

### task (t)

    mar task create --title T [--column C] [--code X] [--body -|FILE] [--after T | --before T | --first | --last | --index N]
        -> the created task object
    mar task list [--column C] [--status active|archived]
        -> [task, ...]   (default: active only; each task includes "column")
    mar task show T-CODE           -> task object (+ "docs": linked doc codes)
    mar task edit T-CODE [--title ...] [--body -|FILE] [--created D] [--updated D]
    mar task move T-CODE [--column C] [--after T | --before T | --first | --last | --index N]
        # no placement flag = top; task create defaults to bottom
    mar task archive T-CODE        -> {"archived": true, "code": ...}
    mar task unarchive T-CODE      -> {"unarchived": true, "code": ...}
    mar task rm T-CODE --force     -> {"deleted": true, "code": ...}
    mar task link T-CODE DOC-CODE  -> {"linked": true, ...}
    mar task unlink T-CODE DOC-CODE

### doc (d)

    mar doc create --code X --title T --type TYPE [--body -|FILE]
        types: design analysis plan report board reference tooling
    mar doc import FILE.html --code X --type TYPE   # HTML -> Markdown
    mar doc list [--type T] [--status active|archived]
    mar doc show DOC-CODE [--render md]   -> doc object (+ "tasks", "backlinks")
    mar doc edit DOC-CODE [--title ...] [--type ...] [--body -|FILE]
    mar doc move DOC-CODE --code NEWCODE  # rename/recode
    mar doc archive DOC-CODE / mar doc unarchive DOC-CODE
    mar doc rm DOC-CODE --force
    mar doc link DOC-CODE T-CODE / mar doc unlink DOC-CODE T-CODE
    mar doc lint [--strict]               -> unresolved [[wiki-links]] by source
        # exit 0 with findings by default; --strict exits non-zero on any

### backlink

    mar backlink CODE                     -> docs and tasks that [[link]] to CODE

### search

    mar search TERM [--docs] [--tasks] [--type TYPE] [--status active|archived|all]
        -> [{kind, code, title, field, snippet, type|column, status}, ...]
        # case-insensitive substring over titles and bodies; active only by default

### board (b)

    mar board show                 # active cards only, archived listed separately
        -> [{name, tasks: [...]}, ...]

### column (c)

    mar column add NAME [--after C | --before C]
    mar column move NAME (--before C | --after C)
    mar column rename OLD NEW
    mar column rm NAME [--force]

### serve (s)

    mar serve [--port N] [--no-open]   # browse docs + board in a browser

### version

    mar version                    # also: mar --version, -v

### guide (g)

    mar guide                      # this guide (use --json for {"guide": "..."})
