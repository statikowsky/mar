---
title: 'Git-friendly storage: replacing the binary SQLite store'
type: analysis
status: archived
created: "2026-06-12T21:32:27.311635Z"
updated: "2026-06-12T22:47:12.146276Z"
tasks:
    - T-12
    - T-13
    - T-14
    - T-15
    - T-16
---
## Problem

The canonical store is `.mar/mar.db`, a binary SQLite file committed to the repo. Consequences:

- Every board/doc change commits as `1 file changed, 0 insertions(+), 0 deletions(-)` — invisible in `git log -p`, unreviewable in a PR.
- A merge conflict on `mar.db` is unresolvable; the only options are ours/theirs, silently losing one side's changes.
- The `-shm`/`-wal` sidecars are gitignored now, but the DB itself still churns as an opaque blob on every mutation.

For a tool whose pitch is "a Markdown documentation repository living inside a git repo," history should tell the story in text.

## What the database actually holds

Current schema (5 tables), with live data as of 2026-06-12 (50 tasks, 13 docs, 23 links):

| Table | Columns | Notes |
| --- | --- | --- |
| `meta` | key, value | Two rows: `task_seq=11`, `seeded=1` |
| `columns` | id, name (unique), position (int) | 3 rows: To do / In progress / Done |
| `docs` | id, code (unique), title, type, body, status, created_at, updated_at | Body is Markdown, 2.5–5.5 KB each |
| `tasks` | id, code (unique), title, body, column_id, position (REAL), status, created_at, updated_at | Position is fractional-ranked float |
| `doc_tasks` | doc_id, task_id | Many-to-many doc↔task links |

Observations that shape the design:

- **Codes are already natural keys.** `code` is unique on both docs and tasks; the integer `id`s exist only for FK plumbing. A text format can drop ids entirely and key everything by code.
- **The float ranking is already degrading.** Repeated insert-at-top has driven positions down to `1.8e-09` (doubling toward float-precision exhaustion). Any text representation should encode order *as order* (an ordered list), not as a number — which fixes this bug for free.
- **The data is tiny.** ~65 entities, < 100 KB of Markdown total. Whole-store load-into-memory is sub-millisecond at 100× this size; nothing here needs an index on disk.
- **`meta.task_seq` is the only non-entity state.** It can be derived (max numeric task code + 1) instead of stored, which also removes a merge-conflict hotspot.

## Options

### Option A — Markdown files as the canonical store (file-per-entity)

```
.mar/
  board.yml            # columns + per-column ordered task codes
  tasks/T-39.md        # frontmatter + body
  docs/DOC-STORAGE.md  # frontmatter + body
```

Task file:

```markdown
---
title: Parse web templates once at startup
status: active
created: 2026-06-10T14:03:22Z
updated: 2026-06-11T09:41:05Z
---
Body markdown here…
```

Board file (replaces `columns`, `tasks.column_id`, `tasks.position`):

```yaml
columns:
  - name: To do
    tasks: [T-7, T-9, T-39]
  - name: In progress
    tasks: []
  - name: Done
    tasks: [T-11, T-5]
```

Doc frontmatter carries `type` plus the links (`tasks: [T-1, T-2]`), replacing `doc_tasks`. Links live on the doc side only — one owner, no two-place sync.

Data mapping: ids dropped (code is the key); `position` floats become list order in `board.yml`; `task_seq` derived from max existing `T-<n>`; `seeded` becomes "directory exists"; timestamps move to frontmatter.

- ✓ Diffs are the actual change: a moved card is a one-line move in `board.yml`; an edited doc is a Markdown diff.
- ✓ Merge conflicts are human-resolvable (and mostly avoided: two tasks edited concurrently touch two different files).
- ✓ Docs are readable/editable in any editor, on GitHub, in `grep` — the store *is* the documentation repo.
- ✓ Kills the float-position bug structurally.
- ✗ Largest implementation: rewrite `internal/store` (~1,200 non-test lines) from SQL to file I/O.
- ✗ Needs explicit atomicity: write-temp-then-rename per file, plus a `.mar/lock` flock around multi-file mutations (move = board.yml + task timestamp). Note: SQLite isn't saving us much today — T-MAKE-STORE-MUTATIONS-TRANSACTIONAL exists because current mutations aren't transactional anyway.
- ✗ Cross-entity invariants (unique codes, link targets exist, every active task in exactly one column list) move from FK constraints into validation code; want a `mar fsck`-style check on load.

### Option B — Single structured text file (`.mar/mar.yml` or `.json`)

All entities in one deterministic, pretty-printed document.

- ✓ Simplest write path (serialize whole store atomically); trivially transactional.
- ✗ JSON embeds Markdown bodies as escaped single-line strings — diffs unreadable, which forfeits the main goal. YAML block scalars are readable but fragile (indentation-sensitive bodies, `norway problem`).
- ✗ One file = every concurrent change conflicts with every other.
- ✗ Docs aren't browsable as plain Markdown.

Dismissed: it fixes "binary" but not "reviewable."

### Option C — Keep SQLite canonical, commit a deterministic text mirror

On every mutation, export `.mar/tasks/*.md` + `board.yml` (same shapes as Option A); gitignore `mar.db`; rebuild the DB from the mirror when missing or stale (compare a content hash).

- ✓ Smallest diff to current code: store layer untouched, add an exporter + importer.
- ✓ Same git story as Option A.
- ✗ Two sources of truth. Every bypass (crash between commit and export, hand-edited mirror, merge resolved in the mirror) needs reconciliation logic; "rebuild when stale" must be bulletproof or the board silently forks.
- ✗ Permanent double-write cost on every mutation; the mirror format must be maintained forever anyway — at which point it *is* the store with extra steps.

Honest framing: C is A with a cache. Worth it only if SQLite features (FTS, complex queries) are on the roadmap; at current scale they aren't, and search over ~100 KB is a linear scan.

### Option D — Commit a SQL text dump (`sqlite3 .dump`)

- ✓ One-line change to ship.
- ✗ Row-per-line INSERTs with whole Markdown bodies embedded on one line: technically text, practically still unreviewable; merges still near-impossible. Dismissed.

## What we give up by dropping SQLite

Ranked by how much it actually hurts; the first three are real costs that price into Option A's step 2.

1. **Multi-process concurrency control — the biggest real loss.** `mar serve` holds the DB open while CLI commands mutate it; WAL mode + `busy_timeout(5000)` makes that safe with zero visible code, and journaling means a crashed process can't leave a torn write. The file store must rebuild this: a `.mar/lock` flock around mutations, write-temp-then-rename per file, and a *multi-file* consistency story — a task move touches `board.yml` and the task file, and a crash between the two renames produces a state SQLite never could. Mitigation: a write-ordering discipline ("board.yml last; on load, a task absent from every column list appends to its column") plus fsck-on-load.
2. **Cheap cross-process change detection.** `internal/web/events.go` polls `PRAGMA data_version` — a one-page read answering "did another process write?" — to drive the web UI's live refresh. Files have no equivalent primitive; replacement is an mtime/content-hash scan of `.mar/` or fsnotify, more code with edge cases (mtime granularity, rename-style editor writes).
3. **Declarative integrity.** `UNIQUE` codes, foreign keys, and `ON DELETE CASCADE` on `doc_tasks` make invalid states unrepresentable today. In files, "two files claim T-9," "board.yml references a deleted task," and dangling doc links are all representable — and reachable via hand-edits or git merges, not just bugs — so load-time validation is mandatory, not optional.

Theoretical losses that don't bite at MAR's scale:

- **Transactions:** barely used today — only the `task_seq` increment and column renumbering run in a `tx`; everything else is autocommit (hence T-MAKE-STORE-MUTATIONS-TRANSACTIONAL). Deriving the sequence from max code removes the main user.
- **Query power / indexes:** every store query is a single-table SELECT with filter + ORDER BY; in-memory slice sort at ~65 entities.
- **FTS5:** relevant to future search, but linear scan over <100 KB Markdown is instant; FTS earns its keep around thousands of large docs.
- **Scale headroom:** rewriting `board.yml` per move and loading everything at start is fine to thousands of tasks; far beyond the design point.

Net: we trade free concurrency/crash-safety and a free change signal for a lockfile, atomic renames, an fsck, and an mtime watcher.

## Non-goals (recorded for later)

**Hosted real-time multi-user MAR.** Multi-user *through git* — branches, PRs, merges over the board — is exactly what Option A optimizes for and is the supported collaboration mode. Multi-user *outside* git (a team hitting one live shared board, no repo) is explicitly out of scope for this design, and choosing files now does not mortgage it:

- That mode changes the architecture before it changes storage: a long-running server owns the data and clients go through an API (auth, sessions, concurrent writers, live updates). Today's access model — CLI and `mar serve` both opening the store off the local filesystem — doesn't survive that transition under *any* storage engine, so keeping SQLite today buys no head start.
- If it ever materializes, the right store there is relational (SQLite behind one server process, Postgres beyond) implemented as a second backend behind the `Store` interface from step 2 — an addition, not an unwinding of the file store.
- The data-model decisions made here transfer: codes as natural keys, links owned by one side, ordering expressed as an operation ("move after X") rather than a stored float all map cleanly onto a relational schema.

## Recommendation

**Option A.** The dataset is small enough that SQLite buys nothing we use (we don't use transactions properly today, don't use FTS, don't need indexes), and the file-per-entity layout is the only option where the git history *is* the product story. Option C is the fallback if rewrite risk feels too high, but it carries permanent dual-store complexity for a temporary saving.

Suggested shape of the work (each its own task when scheduled):

1. Define the on-disk format + parser/serializer (frontmatter via `gopkg.in/yaml.v3`, golden-file tests for determinism: stable key order, LF endings, trailing newline).
2. Reimplement `internal/store.Store` over the file layout behind the existing method set, with flock + atomic rename; add load-time validation (`fsck`). While reimplementing every method anyway, extract the method set into a `Store` interface — web and CLI already hold a `*store.Store` and only call methods (`server.go:23`, `root.go:77`), so this is mechanical now and keeps the storage contract open for other backends later (see Non-goals).
3. `mar migrate store`: one-shot export from `mar.db`, renumbering slug task codes optionally deferred; keep read-only fallback that detects a legacy `mar.db` and tells the user to migrate.
4. Update `mar init`, `.gitignore` guidance, README, and `mar guide`.

Open questions:

- Renumber legacy slug codes (`T-CREATE-CARDS-IN-THE` → `T-12`) during migration, or keep them? Renumbering breaks prose references in existing doc bodies.
- Should `updated` timestamps stay? They make every edit a 2-hunk diff (body + frontmatter). Proposal: keep `created` in frontmatter, derive "updated" from git when available, drop it from the format.
- Archived entities: same directories with `status: archived` (recommended — preserves history locality) vs. an `archive/` subtree.


