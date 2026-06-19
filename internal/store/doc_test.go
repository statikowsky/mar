package store

import (
	"errors"
	"testing"
)

func TestCreateDocNormalizesCodeAndPrefixes(t *testing.T) {
	s := newTestStore(t)
	d, err := s.CreateDoc("auth", "Auth design", "design", "# Hello")
	if err != nil {
		t.Fatalf("CreateDoc: %v", err)
	}
	if d.Code != "DOC-AUTH" {
		t.Errorf("Code = %q, want DOC-AUTH", d.Code)
	}
	if d.Status != "active" {
		t.Errorf("Status = %q, want active", d.Status)
	}
	if d.CreatedAt == "" || d.UpdatedAt == "" {
		t.Error("timestamps not set")
	}
}

func TestCreateDocRejectsDuplicateCode(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateDoc("auth", "A", "design", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateDoc("DOC-AUTH", "B", "design", ""); err == nil {
		t.Fatal("expected duplicate-code error")
	}
}

func TestCreateDocRejectsBadType(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateDoc("x", "X", "wishlist", ""); err == nil {
		t.Fatal("expected invalid-type error")
	}
}

func TestCreateDocRejectsBadCode(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateDoc("has space", "X", "design", ""); err == nil {
		t.Fatal("expected invalid-code error")
	}
}

func TestGetDocNotFound(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.GetDoc("DOC-NOPE"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestEditDocUpdatesFieldsAndTimestamp(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "A", "design", "old")
	newTitle, newType, newBody := "New", "plan", "new body"
	got, err := s.EditDoc(d.Code, &newTitle, &newType, &newBody)
	if err != nil {
		t.Fatalf("EditDoc: %v", err)
	}
	if got.Title != "New" || got.Type != "plan" || got.Body != "new body" {
		t.Errorf("fields not updated: %+v", got)
	}
	if got.UpdatedAt == d.UpdatedAt {
		t.Error("UpdatedAt should change")
	}
}

func TestRecodeDoc(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "A", "design", "")
	got, err := s.RecodeDoc(d.Code, "login")
	if err != nil {
		t.Fatalf("RecodeDoc: %v", err)
	}
	if got.Code != "DOC-LOGIN" {
		t.Errorf("Code = %q, want DOC-LOGIN", got.Code)
	}
	if _, err := s.GetDoc("DOC-AUTH"); !errors.Is(err, ErrNotFound) {
		t.Error("old code should no longer resolve")
	}
}

func TestArchiveDoc(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "A", "design", "")
	if err := s.ArchiveDoc(d.Code); err != nil {
		t.Fatalf("ArchiveDoc: %v", err)
	}
	got, _ := s.GetDoc(d.Code)
	if got.Status != "archived" {
		t.Errorf("Status = %q, want archived", got.Status)
	}
}

func TestUnarchiveDoc(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "A", "design", "")
	s.ArchiveDoc(d.Code)
	if err := s.UnarchiveDoc(d.Code); err != nil {
		t.Fatalf("UnarchiveDoc: %v", err)
	}
	got, _ := s.GetDoc(d.Code)
	if got.Status != "active" {
		t.Errorf("Status = %q, want active", got.Status)
	}
}

func TestUnarchiveDocNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.UnarchiveDoc("DOC-NOPE"); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteDoc(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "A", "design", "")
	if err := s.DeleteDoc(d.Code); err != nil {
		t.Fatalf("DeleteDoc: %v", err)
	}
	if _, err := s.GetDoc(d.Code); !errors.Is(err, ErrNotFound) {
		t.Error("doc should be gone")
	}
}

func TestDeleteDocNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.DeleteDoc("DOC-NOPE"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestSetDocDates(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "A", "design", "")
	created := "2026-05-26"
	updated := "2026-06-03"
	got, err := s.SetDocDates(d.Code, &created, &updated)
	if err != nil {
		t.Fatalf("SetDocDates: %v", err)
	}
	if got.CreatedAt[:10] != "2026-05-26" {
		t.Errorf("CreatedAt = %q, want 2026-05-26...", got.CreatedAt)
	}
	if got.UpdatedAt[:10] != "2026-06-03" {
		t.Errorf("UpdatedAt = %q, want 2026-06-03...", got.UpdatedAt)
	}
}

func TestSetDocDatesOnlyUpdated(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "A", "design", "")
	origCreated := d.CreatedAt
	updated := "2026-01-01"
	got, err := s.SetDocDates(d.Code, nil, &updated)
	if err != nil {
		t.Fatal(err)
	}
	if got.CreatedAt != origCreated {
		t.Errorf("CreatedAt changed: %q -> %q", origCreated, got.CreatedAt)
	}
	if got.UpdatedAt[:10] != "2026-01-01" {
		t.Errorf("UpdatedAt = %q", got.UpdatedAt)
	}
}

func TestSetDocDatesRejectsBadDate(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "A", "design", "")
	bad := "not-a-date"
	if _, err := s.SetDocDates(d.Code, &bad, nil); err == nil {
		t.Fatal("expected invalid-date error")
	}
}
