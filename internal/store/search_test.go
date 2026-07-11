package store

import (
	"strings"
	"testing"
)

func resultByCode(rs []SearchResult, code string) (SearchResult, bool) {
	for _, r := range rs {
		if r.Code == code {
			return r, true
		}
	}
	return SearchResult{}, false
}

func TestSearchMatchesTitlesAndBodies(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("auth", "Authentication design", "design", "Bodies are stored as markdown.")
	s.CreateDoc("other", "Unrelated", "analysis", "Talks about widgets and gizmos.")
	s.CreateTask("Wire authentication", "", "")            // title match
	s.CreateTask("Misc", "Needs the markdown parser.", "") // body match

	got, err := s.Search("markdown", SearchOpts{})
	if err != nil {
		t.Fatal(err)
	}
	// DOC-AUTH (body), and the misc task (body) match "markdown".
	if len(got) != 2 {
		t.Fatalf("Search(markdown) = %+v, want 2 results", got)
	}
	d, ok := resultByCode(got, "DOC-AUTH")
	if !ok || d.Kind != "doc" || d.Field != "body" || d.Type != "design" {
		t.Errorf("DOC-AUTH result = %+v", d)
	}
	if !strings.Contains(strings.ToLower(d.Snippet), "markdown") {
		t.Errorf("snippet should contain the term: %q", d.Snippet)
	}
}

func TestSearchCaseInsensitiveTitle(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("auth", "Authentication Design", "design", "")
	got, err := s.Search("AUTHENTICATION", SearchOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := resultByCode(got, "DOC-AUTH"); !ok || r.Field != "title" {
		t.Errorf("case-insensitive title match failed: %+v", got)
	}
}

func TestSearchEmptyTermErrors(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Search("  ", SearchOpts{}); err == nil {
		t.Error("empty term should error, not return everything")
	}
}

func TestSearchSnippetExcludesFrontmatter(t *testing.T) {
	s := newTestStore(t)
	// "design" appears in the type frontmatter; searching it must match only
	// real body/title text, never the YAML.
	s.CreateDoc("a", "A doc", "design", "The body never says that word.")
	got, _ := s.Search("status:", SearchOpts{})
	if len(got) != 0 {
		t.Errorf("frontmatter must not be searchable, got %+v", got)
	}
}

func TestSearchFiltersByKind(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("shared", "shared term doc", "design", "")
	s.CreateTask("shared term task", "", "")

	if got, _ := s.Search("shared", SearchOpts{Docs: true}); len(got) != 1 || got[0].Kind != "doc" {
		t.Errorf("--docs should return only docs: %+v", got)
	}
	if got, _ := s.Search("shared", SearchOpts{Tasks: true}); len(got) != 1 || got[0].Kind != "task" {
		t.Errorf("--tasks should return only tasks: %+v", got)
	}
	if got, _ := s.Search("shared", SearchOpts{}); len(got) != 2 {
		t.Errorf("no kind flag should return both: %+v", got)
	}
}

func TestSearchStatusFilter(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("a", "archive me term", "design", "")
	s.ArchiveDoc(d.Code)

	if got, _ := s.Search("term", SearchOpts{}); len(got) != 0 {
		t.Errorf("archived excluded by default, got %+v", got)
	}
	if got, _ := s.Search("term", SearchOpts{Status: "archived"}); len(got) != 1 {
		t.Errorf("--status archived should find it, got %+v", got)
	}
	if got, _ := s.Search("term", SearchOpts{Status: "all"}); len(got) != 1 {
		t.Errorf("--status all should find it, got %+v", got)
	}
	if _, err := s.Search("term", SearchOpts{Status: "bogus"}); err == nil {
		t.Error("invalid status should error")
	}
}

func TestSearchTypeFilterIsDocScoped(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("d1", "term design", "design", "")
	s.CreateDoc("d2", "term analysis", "analysis", "")
	s.CreateTask("term task", "", "")

	got, _ := s.Search("term", SearchOpts{Type: "design"})
	if len(got) != 1 || got[0].Code != "DOC-D1" {
		t.Errorf("--type design should match only the design doc, got %+v", got)
	}
}

func TestSearchOrdersTitleBeforeBody(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("bbody", "Nothing here", "design", "the keyword lives in the body")
	s.CreateDoc("atitle", "keyword in title", "design", "")

	got, _ := s.Search("keyword", SearchOpts{})
	if len(got) != 2 || got[0].Field != "title" {
		t.Errorf("title hit should sort before body hit, got %+v", got)
	}
}
