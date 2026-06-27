package store

import (
	"errors"
	"testing"
)

func TestCreateTaskAutoNumbersAndDefaultsToFirstColumn(t *testing.T) {
	s := newTestStore(t)
	t1, err := s.CreateTask("First", "", "")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if t1.Code != "T-FIRST" {
		t.Errorf("Code = %q, want T-FIRST", t1.Code)
	}
	t2, _ := s.CreateTask("Second", "", "")
	if t2.Code != "T-SECOND" {
		t.Errorf("Code = %q, want T-SECOND", t2.Code)
	}
	cols, _ := s.ListColumns()
	if t1.Column != cols[0].Name {
		t.Errorf("default column = %q, want %q", t1.Column, cols[0].Name)
	}
}

func TestNewTaskIsActive(t *testing.T) {
	s := newTestStore(t)
	tk, _ := s.CreateTask("First", "", "")
	got, _ := s.GetTask(tk.Code)
	if got.Status != "active" {
		t.Errorf("Status = %q, want active", got.Status)
	}
}

func TestArchiveAndUnarchiveTask(t *testing.T) {
	s := newTestStore(t)
	tk, _ := s.CreateTask("First", "", "")
	if err := s.ArchiveTask(tk.Code); err != nil {
		t.Fatalf("ArchiveTask: %v", err)
	}
	if got, _ := s.GetTask(tk.Code); got.Status != "archived" {
		t.Errorf("Status = %q, want archived", got.Status)
	}
	if err := s.UnarchiveTask(tk.Code); err != nil {
		t.Fatalf("UnarchiveTask: %v", err)
	}
	if got, _ := s.GetTask(tk.Code); got.Status != "active" {
		t.Errorf("Status = %q, want active", got.Status)
	}
}

func TestArchiveTaskNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.ArchiveTask("T-NOPE"); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if err := s.UnarchiveTask("T-NOPE"); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestBoardExcludesArchivedTasks(t *testing.T) {
	s := newTestStore(t)
	s.CreateTaskWithCode("1", "Active", "", "")
	old, _ := s.CreateTaskWithCode("2", "Old", "", "")
	s.ArchiveTask(old.Code)

	board, _ := s.Board()
	for _, col := range board {
		for _, tk := range col.Tasks {
			if tk.Code == "T-2" {
				t.Errorf("archived task T-2 should not appear on the board")
			}
		}
	}

	archived, err := s.ArchivedTasks()
	if err != nil {
		t.Fatalf("ArchivedTasks: %v", err)
	}
	if len(archived) != 1 || archived[0].Code != "T-2" {
		t.Errorf("ArchivedTasks = %v, want [T-2]", archived)
	}
}

func TestListTasksFiltersByStatus(t *testing.T) {
	s := newTestStore(t)
	s.CreateTaskWithCode("1", "Active", "", "")
	old, _ := s.CreateTaskWithCode("2", "Old", "", "")
	s.ArchiveTask(old.Code)

	active, _ := s.ListTasks("", "active")
	if len(active) != 1 || active[0].Code != "T-1" {
		t.Errorf("active = %v, want [T-1]", active)
	}
	arch, _ := s.ListTasks("", "archived")
	if len(arch) != 1 || arch[0].Code != "T-2" {
		t.Errorf("archived = %v, want [T-2]", arch)
	}
	all, _ := s.ListTasks("", "")
	if len(all) != 2 {
		t.Errorf("all = %v, want 2 tasks", all)
	}
}

func TestCreateTaskInNamedColumn(t *testing.T) {
	s := newTestStore(t)
	tk, err := s.CreateTask("X", "", "In progress")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if tk.Column != "In progress" {
		t.Errorf("Column = %q, want In progress", tk.Column)
	}
}

func TestCreateTaskWithPlacementBeforeFirst(t *testing.T) {
	s := newTestStore(t)
	first, _ := s.CreateTask("First", "", "")
	if _, err := s.CreateTaskWithPlacement("Pinned", "", "To do", Placement{Mode: PlacementBefore, Code: first.Code}); err != nil {
		t.Fatalf("CreateTaskWithPlacement: %v", err)
	}
	tasks, _ := s.ListTasks("To do", "")
	if !eq(codes(tasks), []string{"T-PINNED", "T-FIRST"}) {
		t.Errorf("order = %v, want [T-PINNED T-FIRST]", codes(tasks))
	}
}

func TestCreateTaskWithPlacementFirstLastAfterAndIndex(t *testing.T) {
	s := newTestStore(t)
	a, _ := s.CreateTask("a", "", "")
	b, _ := s.CreateTaskWithPlacement("b", "", "To do", Placement{Mode: PlacementFirst})
	c, _ := s.CreateTaskWithPlacement("c", "", "To do", Placement{Mode: PlacementAfter, Code: a.Code})
	d, _ := s.CreateTaskWithPlacement("d", "", "To do", Placement{Mode: PlacementIndex, Index: 2})
	e, _ := s.CreateTaskWithPlacement("e", "", "To do", Placement{Mode: PlacementLast})
	tasks, _ := s.ListTasks("To do", "")
	want := []string{b.Code, d.Code, a.Code, c.Code, e.Code}
	if !eq(codes(tasks), want) {
		t.Errorf("order = %v, want %v", codes(tasks), want)
	}
}

func TestCreateTaskWithPlacementRejectsInvalidIndex(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateTaskWithPlacement("x", "", "To do", Placement{Mode: PlacementIndex, Index: 2}); err == nil {
		t.Fatal("expected invalid index error")
	}
}

func TestCreateTaskWithPlacementTargetMustBeInColumn(t *testing.T) {
	s := newTestStore(t)
	target, _ := s.CreateTask("target", "", "Done")
	if _, err := s.CreateTaskWithPlacement("x", "", "To do", Placement{Mode: PlacementAfter, Code: target.Code}); err == nil {
		t.Fatal("expected target-column error")
	}
}

func TestCreateTaskWithPlacementTargetMustBeActive(t *testing.T) {
	s := newTestStore(t)
	target, _ := s.CreateTask("target", "", "")
	if err := s.ArchiveTask(target.Code); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTaskWithPlacement("x", "", "To do", Placement{Mode: PlacementAfter, Code: target.Code}); err == nil {
		t.Fatal("expected inactive-target error")
	}
}

func TestCreateTaskUnknownColumnFails(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateTask("X", "", "Nonexistent"); err == nil {
		t.Fatal("expected unknown-column error")
	}
}

func TestGetTaskNotFound(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.GetTask("T-99"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestTaskLookupsNormalizeCode(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateTaskWithCode("5", "Wire auth", "", ""); err != nil {
		t.Fatal(err)
	}
	// GetTask resolves the canonical code, a bare number, and a lowercase
	// prefixed form alike — matching GetDoc.
	for _, in := range []string{"T-5", "5", "t-5", " t-5 "} {
		got, err := s.GetTask(in)
		if err != nil {
			t.Errorf("GetTask(%q): %v", in, err)
			continue
		}
		if got.Code != "T-5" {
			t.Errorf("GetTask(%q).Code = %q, want T-5", in, got.Code)
		}
	}
	// An unparseable code is a validation error, not a bare not-found.
	if _, err := s.GetTask("bad code!"); err == nil {
		t.Error("GetTask(invalid) should return an error")
	} else if errors.Is(err, ErrNotFound) {
		t.Errorf("GetTask(invalid) = %v, want validation error not ErrNotFound", err)
	}
}

func TestTaskMutationsNormalizeCode(t *testing.T) {
	s := newTestStore(t)
	s.CreateTaskWithCode("5", "Target", "", "")
	s.CreateTaskWithCode("6", "Other", "", "")

	title := "Edited"
	if _, err := s.EditTask("5", &title, nil); err != nil {
		t.Errorf("EditTask(bare): %v", err)
	}
	if err := s.ArchiveTask("t-5"); err != nil {
		t.Errorf("ArchiveTask(lowercase): %v", err)
	}
	if err := s.UnarchiveTask("5"); err != nil {
		t.Errorf("UnarchiveTask(bare): %v", err)
	}
	created := "2026-01-01"
	if _, err := s.SetTaskDates("t-5", &created, nil); err != nil {
		t.Errorf("SetTaskDates(lowercase): %v", err)
	}
	// MoveTask normalizes both the moved code and the --after code.
	after := "5"
	if _, err := s.MoveTask("6", "To do", &after); err != nil {
		t.Errorf("MoveTask(bare, after bare): %v", err)
	}
	if err := s.DeleteTask("t-6"); err != nil {
		t.Errorf("DeleteTask(lowercase): %v", err)
	}
}

func TestLinkNormalizesTaskCode(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("auth", "Auth", "design", "")
	s.CreateTaskWithCode("5", "Wire", "", "")
	if err := s.Link("auth", "t-5"); err != nil {
		t.Fatalf("Link(lowercase task): %v", err)
	}
	docs, err := s.DocsForTask("5")
	if err != nil {
		t.Fatalf("DocsForTask(bare): %v", err)
	}
	if len(docs) != 1 || docs[0].Code != "DOC-AUTH" {
		t.Errorf("DocsForTask(bare) = %v, want [DOC-AUTH]", docs)
	}
	codes, err := s.DocCodesForTask("T-5")
	if err != nil || len(codes) != 1 {
		t.Errorf("DocCodesForTask = %v, %v", codes, err)
	}
	if err := s.Unlink("auth", "5"); err != nil {
		t.Errorf("Unlink(bare): %v", err)
	}
}

func TestTaskSeqDerivedFromMaxNumericCode(t *testing.T) {
	s := newTestStore(t)
	// Symbol-only titles use the numeric fallback, derived from the highest
	// numeric code present (file store has no persistent counter).
	t1, _ := s.CreateTask("!", "", "")
	if t1.Code != "T-1" {
		t.Fatalf("first fallback code = %q, want T-1", t1.Code)
	}
	t2, _ := s.CreateTask("!", "", "")
	if t2.Code != "T-2" {
		t.Errorf("Code = %q, want T-2", t2.Code)
	}
	// Deleting the highest numeric task frees its number for reuse.
	if err := s.DeleteTask(t2.Code); err != nil {
		t.Fatal(err)
	}
	t3, _ := s.CreateTask("!", "", "")
	if t3.Code != "T-2" {
		t.Errorf("Code = %q, want T-2 (derived from max)", t3.Code)
	}
}

func codes(tasks []Task) []string {
	out := make([]string, len(tasks))
	for i, t := range tasks {
		out[i] = t.Code
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestListTasksOrderedByPosition(t *testing.T) {
	s := newTestStore(t)
	s.CreateTask("a", "", "")
	s.CreateTask("b", "", "")
	s.CreateTask("c", "", "")
	tasks, err := s.ListTasks("To do", "")
	if err != nil {
		t.Fatal(err)
	}
	if !eq(codes(tasks), []string{"T-A", "T-B", "T-C"}) {
		t.Errorf("order = %v", codes(tasks))
	}
}

func TestEditTask(t *testing.T) {
	s := newTestStore(t)
	tk, _ := s.CreateTask("old", "", "")
	nt, nb := "new", "notes"
	got, err := s.EditTask(tk.Code, &nt, &nb)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "new" || got.Body != "notes" {
		t.Errorf("got %+v", got)
	}
}

func TestMoveTaskToFrontOfColumn(t *testing.T) {
	s := newTestStore(t)
	s.CreateTask("a", "", "")
	s.CreateTask("b", "", "")
	c, _ := s.CreateTask("c", "", "")
	if _, err := s.MoveTask(c.Code, "To do", nil); err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.ListTasks("To do", "")
	if !eq(codes(tasks), []string{"T-C", "T-A", "T-B"}) {
		t.Errorf("order = %v, want [T-C T-A T-B]", codes(tasks))
	}
}

func TestMoveTaskAfterAnother(t *testing.T) {
	s := newTestStore(t)
	a, _ := s.CreateTask("a", "", "")
	s.CreateTask("b", "", "")
	c, _ := s.CreateTask("c", "", "")
	after := a.Code
	if _, err := s.MoveTask(c.Code, "To do", &after); err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.ListTasks("To do", "")
	if !eq(codes(tasks), []string{"T-A", "T-C", "T-B"}) {
		t.Errorf("order = %v, want [T-A T-C T-B]", codes(tasks))
	}
}

func TestMoveTaskBeforeAnother(t *testing.T) {
	s := newTestStore(t)
	s.CreateTask("a", "", "")
	b, _ := s.CreateTask("b", "", "")
	c, _ := s.CreateTask("c", "", "")
	if _, err := s.MoveTaskWithPlacement(c.Code, "To do", Placement{Mode: PlacementBefore, Code: b.Code}); err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.ListTasks("To do", "")
	if !eq(codes(tasks), []string{"T-A", "T-C", "T-B"}) {
		t.Errorf("order = %v, want [T-A T-C T-B]", codes(tasks))
	}
}

func TestMoveTaskLastAndIndex(t *testing.T) {
	s := newTestStore(t)
	a, _ := s.CreateTask("a", "", "")
	b, _ := s.CreateTask("b", "", "")
	c, _ := s.CreateTask("c", "", "")
	if _, err := s.MoveTaskWithPlacement(a.Code, "To do", Placement{Mode: PlacementLast}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.MoveTaskWithPlacement(c.Code, "To do", Placement{Mode: PlacementIndex, Index: 2}); err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.ListTasks("To do", "")
	if !eq(codes(tasks), []string{b.Code, c.Code, a.Code}) {
		t.Errorf("order = %v, want [%s %s %s]", codes(tasks), b.Code, c.Code, a.Code)
	}
}

func TestMoveTaskRejectsInvalidIndex(t *testing.T) {
	s := newTestStore(t)
	a, _ := s.CreateTask("a", "", "")
	if _, err := s.MoveTaskWithPlacement(a.Code, "To do", Placement{Mode: PlacementIndex, Index: 2}); err == nil {
		t.Fatal("expected invalid index error")
	}
}

func TestMoveTaskToOtherColumn(t *testing.T) {
	s := newTestStore(t)
	a, _ := s.CreateTask("a", "", "")
	moved, err := s.MoveTask(a.Code, "Done", nil)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Column != "Done" {
		t.Errorf("Column = %q, want Done", moved.Column)
	}
	if got, _ := s.ListTasks("To do", ""); len(got) != 0 {
		t.Errorf("To do should be empty, got %v", codes(got))
	}
}

func TestEditTaskNotFound(t *testing.T) {
	s := newTestStore(t)
	nt := "x"
	if _, err := s.EditTask("T-99", &nt, nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestMoveTaskNotFound(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.MoveTask("T-99", "Done", nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestCreateTaskWithCode(t *testing.T) {
	s := newTestStore(t)
	tk, err := s.CreateTaskWithCode("39", "Manual code", "", "")
	if err != nil {
		t.Fatalf("CreateTaskWithCode: %v", err)
	}
	if tk.Code != "T-39" {
		t.Errorf("Code = %q, want T-39", tk.Code)
	}
}

func TestCreateTaskWithCodeAcceptsPrefixed(t *testing.T) {
	s := newTestStore(t)
	tk, err := s.CreateTaskWithCode("t-7", "x", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Code != "T-7" {
		t.Errorf("Code = %q, want T-7 (normalized)", tk.Code)
	}
}

func TestCreateTaskWithCodeRejectsDuplicate(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateTaskWithCode("5", "a", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTaskWithCode("T-5", "b", "", ""); err == nil {
		t.Fatal("expected duplicate-code error")
	}
}

func TestCreateTaskWithCodeRejectsInvalid(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateTaskWithCode("has space", "x", "", ""); err == nil {
		t.Fatal("expected invalid-code error")
	}
}

func TestCreateTaskWithCodeAdvancesAutoCounter(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateTaskWithCode("100", "manual", "", ""); err != nil {
		t.Fatal(err)
	}
	// A symbol-only title slugs to empty, so it uses the numeric seq fallback,
	// which must have advanced past the manual T-100.
	auto, err := s.CreateTask("!!!", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if auto.Code != "T-101" {
		t.Errorf("auto code = %q, want T-101 (counter advanced past manual T-100)", auto.Code)
	}
}

func TestCreateTaskSlugifiesTitle(t *testing.T) {
	s := newTestStore(t)
	tk, err := s.CreateTask("Wire auth login", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Code != "T-WIRE-AUTH-LOGIN" {
		t.Errorf("Code = %q, want T-WIRE-AUTH-LOGIN", tk.Code)
	}
}

func TestCreateTaskSlugCollisionGetsSuffix(t *testing.T) {
	s := newTestStore(t)
	a, _ := s.CreateTask("Same title", "", "")
	b, _ := s.CreateTask("Same title", "", "")
	if a.Code != "T-SAME-TITLE" {
		t.Errorf("first = %q, want T-SAME-TITLE", a.Code)
	}
	if b.Code != "T-SAME-TITLE-2" {
		t.Errorf("second = %q, want T-SAME-TITLE-2", b.Code)
	}
}

func TestCreateTaskSlugStripsPunctuationAndCollapses(t *testing.T) {
	s := newTestStore(t)
	tk, _ := s.CreateTask("  Fix /fetch: timeout (again)!  ", "", "")
	if tk.Code != "T-FIX-FETCH-TIMEOUT-AGAIN" {
		t.Errorf("Code = %q, want T-FIX-FETCH-TIMEOUT-AGAIN", tk.Code)
	}
}

func TestCreateTaskSlugCapsLongTitles(t *testing.T) {
	s := newTestStore(t)
	tk, err := s.CreateTask("Resolve column id to name in task JSON", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Code != "T-RESOLVE-COLUMN-ID-TO" {
		t.Errorf("Code = %q, want T-RESOLVE-COLUMN-ID-TO (capped at 4 words)", tk.Code)
	}
}

func TestCreateTaskEmptyTitleSlugFallsBackToSeq(t *testing.T) {
	s := newTestStore(t)
	tk, err := s.CreateTask("!!!", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Code != "T-1" {
		t.Errorf("Code = %q, want numeric fallback T-1", tk.Code)
	}
}

func TestSetTaskDates(t *testing.T) {
	s := newTestStore(t)
	tk, _ := s.CreateTask("x", "", "")
	created := "2026-05-26"
	updated := "2026-06-03T12:30:00Z"
	got, err := s.SetTaskDates(tk.Code, &created, &updated)
	if err != nil {
		t.Fatalf("SetTaskDates: %v", err)
	}
	if got.CreatedAt[:10] != "2026-05-26" {
		t.Errorf("CreatedAt = %q", got.CreatedAt)
	}
	if got.UpdatedAt[:10] != "2026-06-03" {
		t.Errorf("UpdatedAt = %q", got.UpdatedAt)
	}
}
