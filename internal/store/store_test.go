package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestInitCreatesStoreAndSeedsColumns(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(filepath.Join(dir, ".mar", "board.yml")); err != nil {
		t.Fatalf("board.yml not created: %v", err)
	}

	cols, err := s.ListColumns()
	if err != nil {
		t.Fatalf("ListColumns: %v", err)
	}
	want := []string{"To do", "In progress", "Done"}
	if len(cols) != len(want) {
		t.Fatalf("got %d columns, want %d", len(cols), len(want))
	}
	for i, c := range cols {
		if c.Name != want[i] {
			t.Errorf("column %d = %q, want %q", i, c.Name, want[i])
		}
	}
}

func TestInitTwiceFails(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	s.Close()
	if _, err := Init(dir); err == nil {
		t.Fatal("second Init should fail, got nil")
	}
}

func TestDiscoverWalksUp(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	s.Close()

	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	found, err := Discover(sub)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	want := filepath.Join(dir, ".mar")
	if found != want {
		t.Errorf("Discover = %q, want %q", found, want)
	}
}

func TestDiscoverNotFound(t *testing.T) {
	if _, err := Discover(t.TempDir()); err == nil {
		t.Fatal("expected error when no store exists")
	}
}

func TestDataVersionChangesOnWrite(t *testing.T) {
	dir := t.TempDir()
	observer, err := Init(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer observer.Close()

	v1, err := observer.DataVersion()
	if err != nil {
		t.Fatal(err)
	}

	writerPath, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	writer, err := Open(writerPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.CreateDoc("a", "A", "design", ""); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	v2, err := observer.DataVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v1 == v2 {
		t.Errorf("DataVersion did not change after external write: %d -> %d", v1, v2)
	}
}

func TestReopenDoesNotReseedEmptiedColumns(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	for _, name := range []string{"To do", "In progress", "Done"} {
		if err := s.RemoveColumn(name, true); err != nil {
			t.Fatalf("RemoveColumn %q: %v", name, err)
		}
	}
	s.Close()

	marDir, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(marDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer reopened.Close()

	cols, err := reopened.ListColumns()
	if err != nil {
		t.Fatalf("ListColumns: %v", err)
	}
	if len(cols) != 0 {
		t.Errorf("reopen re-seeded an emptied board: got %d columns, want 0", len(cols))
	}
}

func TestRepairOrphanTaskFileJoinsFirstColumn(t *testing.T) {
	s := newTestStore(t)
	raw, err := marshalTaskFile(taskMeta{Title: "Orphan", Status: "active",
		Created: "2026-01-02T03:04:05Z", Updated: "2026-01-02T03:04:05Z"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.dir, tasksDir, "T-99.md"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	tk, err := s.GetTask("T-99")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if tk.Column != "To do" {
		t.Errorf("orphan column = %q, want To do", tk.Column)
	}
}

func TestRepairDropsDanglingBoardCode(t *testing.T) {
	s := newTestStore(t)
	tk, err := s.CreateTask("Doomed", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(s.dir, tasksDir, tk.Code+".md")); err != nil {
		t.Fatal(err)
	}
	board, err := s.Board()
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	for _, col := range board {
		for _, bt := range col.Tasks {
			if bt.Code == tk.Code {
				t.Errorf("dangling code %s still on board", tk.Code)
			}
		}
	}
}

func TestRepairBuildsTaskColumnIndex(t *testing.T) {
	s := newTestStore(t)
	task, err := s.CreateTask("Indexed", "", "")
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.load()
	if err != nil {
		t.Fatal(err)
	}
	if got := d.columnOf(task.Code); got != "To do" {
		t.Errorf("columnOf(%s) = %q, want To do", task.Code, got)
	}
}

func TestExternalEditVisibleWithoutReopen(t *testing.T) {
	s := newTestStore(t)
	tk, err := s.CreateTask("Watch me", "", "")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.dir, tasksDir, tk.Code+".md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(raw), "title: Watch me", "title: Hand edited", 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetTask(tk.Code)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Hand edited" {
		t.Errorf("title = %q, want hand edit visible", got.Title)
	}
}
