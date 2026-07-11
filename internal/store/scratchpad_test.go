package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestScratchpadStartsEmptyWithoutCreatingFile(t *testing.T) {
	s := newTestStore(t)
	pad, err := s.Scratchpad()
	if err != nil {
		t.Fatal(err)
	}
	if pad.Schema != 1 || pad.Revision != 0 || len(pad.Notes) != 0 {
		t.Fatalf("Scratchpad = %+v, want empty schema 1 revision 0", pad)
	}
	if _, err := os.Stat(filepath.Join(s.dir, scratchpadName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("empty read created scratchpad file: %v", err)
	}
}

func TestScratchpadCreatePersistsStableIDsAndDefaults(t *testing.T) {
	s := newTestStore(t)
	first, err := s.CreateScratchNote("First idea", 20, 30, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.CreateScratchNote("Second idea", -40, 80, 320, "blue")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "S-1" || first.Width != defaultScratchNoteWidth || first.Color != "neutral" {
		t.Errorf("first note = %+v", first)
	}
	if second.ID != "S-2" || second.Z != 2 {
		t.Errorf("second note = %+v", second)
	}

	reopened, err := Open(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	pad, err := reopened.Scratchpad()
	if err != nil {
		t.Fatal(err)
	}
	if pad.Revision != 2 || len(pad.Notes) != 2 || pad.Notes[1].Text != "Second idea" {
		t.Fatalf("reopened Scratchpad = %+v", pad)
	}
}

func TestSaveScratchpadRejectsStaleRevision(t *testing.T) {
	s := newTestStore(t)
	note, err := s.CreateScratchNote("Original", 0, 0, 240, "yellow")
	if err != nil {
		t.Fatal(err)
	}
	pad, err := s.Scratchpad()
	if err != nil {
		t.Fatal(err)
	}
	note.Text = "Changed"
	updated, err := s.SaveScratchpad(pad.Revision, []ScratchNote{note})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Revision != pad.Revision+1 || updated.Notes[0].Text != "Changed" {
		t.Fatalf("updated = %+v", updated)
	}
	if _, err := s.SaveScratchpad(pad.Revision, []ScratchNote{note}); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale SaveScratchpad error = %v, want ErrConflict", err)
	}
}

func TestScratchpadUpdateDeleteAndValidation(t *testing.T) {
	s := newTestStore(t)
	note, err := s.CreateScratchNote("Idea", 1, 2, 240, "green")
	if err != nil {
		t.Fatal(err)
	}
	note.Text = "Updated idea"
	note.Width = 300
	updated, err := s.UpdateScratchNote(note)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Text != "Updated idea" || updated.Width != 300 {
		t.Errorf("updated = %+v", updated)
	}
	if _, err := s.UpdateScratchNote(ScratchNote{ID: "S-99", Text: "missing", Width: 240}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing update error = %v", err)
	}
	if _, err := s.CreateScratchNote("", 0, 0, 240, "neutral"); err == nil {
		t.Fatal("empty note should fail")
	}
	if _, err := s.CreateScratchNote("bad color", 0, 0, 240, "orange"); err == nil {
		t.Fatal("unknown color should fail")
	}
	if err := s.DeleteScratchNote(note.ID); err != nil {
		t.Fatal(err)
	}
	pad, err := s.Scratchpad()
	if err != nil {
		t.Fatal(err)
	}
	if len(pad.Notes) != 0 {
		t.Fatalf("notes after delete = %+v", pad.Notes)
	}
}

func TestScratchpadDataVersionChangesOnWrite(t *testing.T) {
	s := newTestStore(t)
	before, err := s.DataVersion()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateScratchNote("Idea", 0, 0, 240, "neutral"); err != nil {
		t.Fatal(err)
	}
	after, err := s.DataVersion()
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("DataVersion did not change after scratchpad write")
	}
}
