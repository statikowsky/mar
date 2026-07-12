package store

import "testing"

func colNames(cols []Column) []string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.Name
	}
	return out
}

func TestAddColumnAtEnd(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AddColumn("Review", ""); err != nil {
		t.Fatal(err)
	}
	cols, _ := s.ListColumns()
	got := colNames(cols)
	if got[len(got)-1] != "Review" {
		t.Errorf("last column = %q, want Review", got[len(got)-1])
	}
}

func TestBoardViewIncludesLinkedDocsAndArchivedTasks(t *testing.T) {
	s := newTestStore(t)
	active, err := s.CreateTaskWithCode("ACTIVE", "Active", "", "")
	if err != nil {
		t.Fatal(err)
	}
	archived, err := s.CreateTaskWithCode("ARCHIVED", "Archived", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateDoc("PLAN", "Plan", "design", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Link("DOC-PLAN", active.Code); err != nil {
		t.Fatal(err)
	}
	if err := s.ArchiveTask(archived.Code); err != nil {
		t.Fatal(err)
	}

	view, err := s.BoardView()
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Columns) == 0 || len(view.Columns[0].Tasks) != 1 || view.Columns[0].Tasks[0].Code != active.Code {
		t.Fatalf("active board tasks = %+v, want %s", view.Columns, active.Code)
	}
	if got := view.DocCodesByTask[active.Code]; !eqStr(got, []string{"DOC-PLAN"}) {
		t.Errorf("linked docs = %v, want [DOC-PLAN]", got)
	}
	if len(view.ArchivedTasks) != 1 || view.ArchivedTasks[0].Code != archived.Code {
		t.Errorf("archived tasks = %+v, want %s", view.ArchivedTasks, archived.Code)
	}
}

func TestAddColumnAfter(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AddColumn("Review", "To do"); err != nil {
		t.Fatal(err)
	}
	cols, _ := s.ListColumns()
	if colNames(cols)[1] != "Review" {
		t.Errorf("columns = %v, want Review in position 2", colNames(cols))
	}
}

func TestAddDuplicateColumnFails(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AddColumn("To do", ""); err == nil {
		t.Fatal("expected duplicate-name error")
	}
}

func TestRenameColumn(t *testing.T) {
	s := newTestStore(t)
	if err := s.RenameColumn("To do", "Backlog"); err != nil {
		t.Fatal(err)
	}
	cols, _ := s.ListColumns()
	if colNames(cols)[0] != "Backlog" {
		t.Errorf("columns = %v", colNames(cols))
	}
}

func TestRemoveEmptyColumn(t *testing.T) {
	s := newTestStore(t)
	s.AddColumn("Review", "")
	if err := s.RemoveColumn("Review", false); err != nil {
		t.Fatal(err)
	}
	cols, _ := s.ListColumns()
	for _, c := range cols {
		if c.Name == "Review" {
			t.Fatal("Review should be removed")
		}
	}
}

func TestRemoveNonEmptyColumnRequiresForce(t *testing.T) {
	s := newTestStore(t)
	s.CreateTask("a", "", "To do")
	if err := s.RemoveColumn("To do", false); err == nil {
		t.Fatal("expected non-empty column to require force")
	}
	if err := s.RemoveColumn("To do", true); err != nil {
		t.Fatalf("force remove: %v", err)
	}
}

func TestAddColumnBeforeFirst(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AddColumnBefore("Under consideration", "To do"); err != nil {
		t.Fatal(err)
	}
	cols, _ := s.ListColumns()
	if colNames(cols)[0] != "Under consideration" {
		t.Errorf("columns = %v, want Under consideration first", colNames(cols))
	}
}

func TestAddColumnBeforeMiddle(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AddColumnBefore("Review", "Done"); err != nil {
		t.Fatal(err)
	}
	cols, _ := s.ListColumns()
	got := colNames(cols)
	// To do, In progress, Review, Done
	if got[2] != "Review" || got[3] != "Done" {
		t.Errorf("columns = %v, want Review just before Done", got)
	}
}

func TestAddColumnBeforeUnknown(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AddColumnBefore("X", "Nonexistent"); err == nil {
		t.Fatal("expected unknown-column error")
	}
}

func TestMoveColumnToFront(t *testing.T) {
	s := newTestStore(t)
	// default: To do, In progress, Done -> move Done before To do
	if err := s.MoveColumn("Done", "To do", true); err != nil {
		t.Fatal(err)
	}
	cols, _ := s.ListColumns()
	if !eqStr(colNames(cols), []string{"Done", "To do", "In progress"}) {
		t.Errorf("columns = %v", colNames(cols))
	}
}

func TestMoveColumnAfter(t *testing.T) {
	s := newTestStore(t)
	// move To do after In progress -> In progress, To do, Done
	if err := s.MoveColumn("To do", "In progress", false); err != nil {
		t.Fatal(err)
	}
	cols, _ := s.ListColumns()
	if !eqStr(colNames(cols), []string{"In progress", "To do", "Done"}) {
		t.Errorf("columns = %v", colNames(cols))
	}
}

func TestMoveColumnUnknown(t *testing.T) {
	s := newTestStore(t)
	if err := s.MoveColumn("Nope", "To do", true); err == nil {
		t.Fatal("expected unknown-column error")
	}
}

func eqStr(a, b []string) bool {
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
