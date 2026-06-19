package store

import (
	"bytes"
	"testing"
)

func TestTaskFileRoundTrip(t *testing.T) {
	in := taskMeta{Title: "Fix: tricky title", Status: "active",
		Created: "2026-06-10T14:03:22.123456789Z", Updated: "2026-06-11T09:41:05.987654321Z"}
	body := "Line one.\n\nLine two.\n"
	raw, err := marshalTaskFile(in, body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	again, err := marshalTaskFile(in, body)
	if err != nil || !bytes.Equal(raw, again) {
		t.Fatalf("marshal not deterministic")
	}
	gotMeta, gotBody, err := parseTaskFile(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if gotMeta != in || gotBody != body {
		t.Errorf("round trip: got %+v %q", gotMeta, gotBody)
	}
}

func TestTaskFileGolden(t *testing.T) {
	raw, err := marshalTaskFile(taskMeta{Title: "Simple", Status: "active",
		Created: "2026-01-02T03:04:05Z", Updated: "2026-01-02T03:04:05Z"}, "Body.\n")
	if err != nil {
		t.Fatal(err)
	}
	want := "---\ntitle: Simple\nstatus: active\ncreated: \"2026-01-02T03:04:05Z\"\nupdated: \"2026-01-02T03:04:05Z\"\n---\nBody.\n"
	if string(raw) != want {
		t.Errorf("golden mismatch:\n%q\nwant\n%q", raw, want)
	}
}

func TestTaskFileEmptyBodyAndNewlineNormalization(t *testing.T) {
	m := taskMeta{Title: "X", Status: "active", Created: "2026-01-02T03:04:05Z", Updated: "2026-01-02T03:04:05Z"}
	raw, err := marshalTaskFile(m, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, body, err := parseTaskFile(raw); err != nil || body != "" {
		t.Errorf("empty body round trip: %q %v", body, err)
	}
	raw, err = marshalTaskFile(m, "no trailing newline")
	if err != nil {
		t.Fatal(err)
	}
	if _, body, _ := parseTaskFile(raw); body != "no trailing newline\n" {
		t.Errorf("body = %q, want trailing newline added", body)
	}
}

func TestDocFileRoundTripWithLinks(t *testing.T) {
	in := docMeta{Title: "Auth design", Type: "design", Status: "active",
		Created: "2026-01-02T03:04:05Z", Updated: "2026-01-02T03:04:05Z", Tasks: []string{"T-1", "T-2"}}
	raw, err := marshalDocFile(in, "Body with\n---\nseparator in it.\n")
	if err != nil {
		t.Fatal(err)
	}
	got, body, err := parseDocFile(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != in.Title || got.Type != in.Type || len(got.Tasks) != 2 || got.Tasks[0] != "T-1" {
		t.Errorf("meta = %+v", got)
	}
	if body != "Body with\n---\nseparator in it.\n" {
		t.Errorf("body = %q", body)
	}
}

func TestParseDefaultsStatusActive(t *testing.T) {
	m, _, err := parseTaskFile([]byte("---\ntitle: X\n---\n"))
	if err != nil || m.Status != "active" {
		t.Errorf("status = %q, err %v; want active", m.Status, err)
	}
}

func TestBoardFileRoundTrip(t *testing.T) {
	in := boardFile{Columns: []boardColumn{
		{Name: "To do", Tasks: []string{"T-1"}},
		{Name: "Done", Tasks: nil},
	}}
	raw, err := marshalBoardFile(in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := parseBoardFile(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Columns) != 2 || got.Columns[0].Tasks[0] != "T-1" || got.Columns[1].Tasks == nil {
		t.Errorf("board = %+v (nil Tasks must normalize to empty)", got)
	}
}

func TestSplitFrontmatterErrors(t *testing.T) {
	if _, _, err := splitFrontmatter([]byte("no frontmatter")); err == nil {
		t.Error("expected error for missing frontmatter")
	}
	if _, _, err := splitFrontmatter([]byte("---\ntitle: X\n")); err == nil {
		t.Error("expected error for unterminated frontmatter")
	}
}
