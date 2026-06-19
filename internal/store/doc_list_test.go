package store

import "testing"

func TestListDocsFilters(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("a", "A", "design", "")
	s.CreateDoc("b", "B", "analysis", "")
	d, _ := s.CreateDoc("c", "C", "design", "")
	if err := s.ArchiveDoc(d.Code); err != nil {
		t.Fatal(err)
	}

	active, err := s.ListDocs("", "active")
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 2 {
		t.Errorf("active docs = %d, want 2", len(active))
	}
	designs, err := s.ListDocs("design", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(designs) != 2 {
		t.Errorf("design docs = %d, want 2", len(designs))
	}
}
