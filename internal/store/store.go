package store

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

const (
	dirName        = ".mar"
	boardName      = "board.yml"
	scratchpadName = "scratchpad.yml"
	tasksDir       = "tasks"
	docsDir        = "docs"
	lockName       = ".lock"
)

var ErrNotFound = errors.New("not found")

type Doc struct {
	Code      string `json:"code"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	Body      string `json:"body"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Task struct {
	Code      string `json:"code"`
	Title     string `json:"title"`
	Body      string `json:"body,omitempty"`
	Column    string `json:"column"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Column struct {
	Name string `json:"name"`
}

type Store struct {
	dir string
}

// Open expects the path to an existing .mar directory containing board.yml.
func Open(dir string) (*Store, error) {
	if _, err := os.Stat(filepath.Join(dir, boardName)); err != nil {
		return nil, fmt.Errorf("no mar store at %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

func Init(dir string) (*Store, error) {
	marDir := filepath.Join(dir, dirName)
	if _, err := os.Stat(filepath.Join(marDir, boardName)); err == nil {
		return nil, fmt.Errorf("store already exists at %s", filepath.Join(marDir, boardName))
	}
	for _, d := range []string{marDir, filepath.Join(marDir, tasksDir), filepath.Join(marDir, docsDir)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("create %s: %w", d, err)
		}
	}
	s := &Store{dir: marDir}
	board := boardFile{Columns: []boardColumn{{Name: "To do"}, {Name: "In progress"}, {Name: "Done"}}}
	if err := s.writeBoard(board); err != nil {
		return nil, err
	}
	return s, nil
}

// Discover walks up from startDir looking for a .mar store and returns the
// .mar directory path.
func Discover(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", startDir, err)
	}
	for {
		marDir := filepath.Join(dir, dirName)
		if _, err := os.Stat(filepath.Join(marDir, boardName)); err == nil {
			return marDir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no mar store found (run 'mar init')")
		}
		dir = parent
	}
}

func (s *Store) Close() error { return nil }

func (s *Store) withLock(fn func() error) error {
	fl := flock.New(filepath.Join(s.dir, lockName))
	if err := fl.Lock(); err != nil {
		return fmt.Errorf("lock store: %w", err)
	}
	defer fl.Unlock()
	return fn()
}

func nowStamp() string { return time.Now().UTC().Format(time.RFC3339Nano) }

// normalizeDate accepts a YYYY-MM-DD date or a full RFC3339 timestamp and
// returns a canonical RFC3339Nano UTC stamp.
func normalizeDate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t.UTC().Format(time.RFC3339Nano), nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC().Format(time.RFC3339Nano), nil
	}
	return "", fmt.Errorf("invalid date %q: use YYYY-MM-DD or RFC3339", raw)
}

type taskEntity struct {
	meta taskMeta
	body string
}

type docEntity struct {
	meta docMeta
	body string
}

type data struct {
	board   boardFile
	tasks   map[string]*taskEntity
	docs    map[string]*docEntity
	columns map[string]string
}

func (s *Store) load() (*data, error) {
	raw, err := os.ReadFile(filepath.Join(s.dir, boardName))
	if err != nil {
		return nil, fmt.Errorf("read board: %w", err)
	}
	board, err := parseBoardFile(raw)
	if err != nil {
		return nil, err
	}
	d := &data{board: board, tasks: map[string]*taskEntity{}, docs: map[string]*docEntity{}}
	if err := s.loadDir(tasksDir, func(code string, raw []byte) error {
		meta, body, err := parseTaskFile(raw)
		if err != nil {
			return fmt.Errorf("task %s: %w", code, err)
		}
		d.tasks[code] = &taskEntity{meta: meta, body: body}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := s.loadDir(docsDir, func(code string, raw []byte) error {
		meta, body, err := parseDocFile(raw)
		if err != nil {
			return fmt.Errorf("doc %s: %w", code, err)
		}
		d.docs[code] = &docEntity{meta: meta, body: body}
		return nil
	}); err != nil {
		return nil, err
	}
	d.repair()
	return d, nil
}

func (s *Store) loadDir(sub string, fn func(code string, raw []byte) error) error {
	entries, err := os.ReadDir(filepath.Join(s.dir, sub))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", sub, err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(s.dir, sub, name))
		if err != nil {
			return fmt.Errorf("read %s/%s: %w", sub, name, err)
		}
		if err := fn(strings.TrimSuffix(name, ".md"), raw); err != nil {
			return err
		}
	}
	return nil
}

// repair reconciles board membership with the files on disk: codes listed
// without a file are dropped, duplicate codes keep their first occurrence,
// task files in no column are appended to the first column, and doc links
// to missing tasks are dropped. Repairs are in-memory; they persist the
// next time the affected file is written.
func (d *data) repair() {
	seen := map[string]bool{}
	for i := range d.board.Columns {
		col := &d.board.Columns[i]
		kept := col.Tasks[:0]
		for _, code := range col.Tasks {
			if _, ok := d.tasks[code]; ok && !seen[code] {
				kept = append(kept, code)
				seen[code] = true
			}
		}
		col.Tasks = kept
	}
	if len(d.board.Columns) > 0 {
		var orphans []string
		for code := range d.tasks {
			if !seen[code] {
				orphans = append(orphans, code)
			}
		}
		sort.Strings(orphans)
		first := &d.board.Columns[0]
		first.Tasks = append(first.Tasks, orphans...)
	}
	for _, doc := range d.docs {
		kept := doc.meta.Tasks[:0]
		for _, code := range doc.meta.Tasks {
			if _, ok := d.tasks[code]; ok {
				kept = append(kept, code)
			}
		}
		doc.meta.Tasks = kept
	}
	d.rebuildColumnIndex()
}

func (d *data) rebuildColumnIndex() {
	d.columns = make(map[string]string, len(d.tasks))
	for _, col := range d.board.Columns {
		for _, code := range col.Tasks {
			d.columns[code] = col.Name
		}
	}
}

func (d *data) columnOf(code string) string {
	return d.columns[code]
}

func (d *data) findColumn(name string) int {
	for i, c := range d.board.Columns {
		if c.Name == name {
			return i
		}
	}
	return -1
}

func (d *data) removeFromBoard(code string) {
	for i := range d.board.Columns {
		col := &d.board.Columns[i]
		for j, t := range col.Tasks {
			if t == code {
				col.Tasks = append(col.Tasks[:j], col.Tasks[j+1:]...)
				return
			}
		}
	}
}

func (d *data) task(code string) (Task, error) {
	e, ok := d.tasks[code]
	if !ok {
		return Task{}, fmt.Errorf("task %s: %w", code, ErrNotFound)
	}
	return Task{Code: code, Title: e.meta.Title, Body: e.body, Column: d.columnOf(code),
		Status: e.meta.Status, CreatedAt: e.meta.Created, UpdatedAt: e.meta.Updated}, nil
}

// findTask normalizes a user-supplied task code ("5" or "t-5" -> "T-5") and
// looks it up, returning the canonical code and its entity. This mirrors how
// GetDoc normalizes doc codes, so every task entry point accepts the same
// flexible forms. An unparseable code is a validation error, not not-found.
func (d *data) findTask(rawCode string) (string, *taskEntity, error) {
	code, err := normalizeTaskCode(rawCode)
	if err != nil {
		return "", nil, err
	}
	e, ok := d.tasks[code]
	if !ok {
		return "", nil, fmt.Errorf("task %s: %w", code, ErrNotFound)
	}
	return code, e, nil
}

func (d *data) doc(code string) (Doc, error) {
	e, ok := d.docs[code]
	if !ok {
		return Doc{}, fmt.Errorf("doc %s: %w", code, ErrNotFound)
	}
	return Doc{Code: code, Title: e.meta.Title, Type: e.meta.Type, Body: e.body,
		Status: e.meta.Status, CreatedAt: e.meta.Created, UpdatedAt: e.meta.Updated}, nil
}

func (s *Store) writeTask(code string, e *taskEntity) error {
	raw, err := marshalTaskFile(e.meta, e.body)
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(s.dir, tasksDir, code+".md"), raw)
}

func (s *Store) writeDoc(code string, e *docEntity) error {
	raw, err := marshalDocFile(e.meta, e.body)
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(s.dir, docsDir, code+".md"), raw)
}

func (s *Store) writeBoard(b boardFile) error {
	raw, err := marshalBoardFile(b)
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(s.dir, boardName), raw)
}

func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return fmt.Errorf("temp file for %s: %w", path, err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return fmt.Errorf("chmod %s: %w", path, err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("rename into %s: %w", path, err)
	}
	return nil
}

// DataVersion hashes the names, sizes, and mtimes of every store file so
// the web UI can poll for external changes.
func (s *Store) DataVersion() (int64, error) {
	h := fnv.New64a()
	add := func(rel string) error {
		info, err := os.Stat(filepath.Join(s.dir, rel))
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s|%d|%d;", rel, info.Size(), info.ModTime().UnixNano())
		return nil
	}
	if err := add(boardName); err != nil {
		return 0, fmt.Errorf("data version: %w", err)
	}
	if err := add(scratchpadName); err != nil {
		return 0, fmt.Errorf("data version: %w", err)
	}
	for _, sub := range []string{tasksDir, docsDir} {
		entries, err := os.ReadDir(filepath.Join(s.dir, sub))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return 0, fmt.Errorf("data version: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			if err := add(filepath.Join(sub, e.Name())); err != nil {
				return 0, fmt.Errorf("data version: %w", err)
			}
		}
	}
	return int64(h.Sum64()), nil
}
