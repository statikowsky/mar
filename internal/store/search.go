package store

import (
	"fmt"
	"sort"
	"strings"
)

// SearchResult is one doc or task that matched a search term. Field is "title"
// or "body" depending on where the first match landed; Snippet is set only for
// body matches. Type is the doc type (docs only); Column is the board column
// (tasks only).
type SearchResult struct {
	Kind    string `json:"kind"` // doc | task
	Code    string `json:"code"`
	Title   string `json:"title"`
	Field   string `json:"field"` // title | body
	Snippet string `json:"snippet,omitempty"`
	Type    string `json:"type,omitempty"`
	Status  string `json:"status"`
	Column  string `json:"column,omitempty"`
}

// SearchOpts narrows a search. Docs/Tasks both false means search both kinds.
// Status defaults to "active". Type filters docs by type (and, since tasks have
// no type, excludes tasks when set).
type SearchOpts struct {
	Docs   bool
	Tasks  bool
	Status string // active | archived | all
	Type   string
}

// Search scans titles and bodies of active docs and tasks for a case-insensitive
// substring, newest matching the term. Frontmatter is never searched: it scans
// parsed store objects, not raw files. ponytail: linear scan over the in-memory
// store; an index is unwarranted until repositories are far larger (see
// DOC-SEARCH).
func (s *Store) Search(term string, opts SearchOpts) ([]SearchResult, error) {
	term = strings.TrimSpace(term)
	if term == "" {
		return nil, fmt.Errorf("search term required")
	}
	status := opts.Status
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "archived" && status != "all" {
		return nil, fmt.Errorf("invalid status %q: use active, archived, or all", status)
	}
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	lowerTerm := strings.ToLower(term)
	bothKinds := opts.Docs == opts.Tasks
	statusOK := func(st string) bool { return status == "all" || st == status }

	var out []SearchResult
	if bothKinds || opts.Docs {
		for code, e := range d.docs {
			if !statusOK(e.meta.Status) || (opts.Type != "" && e.meta.Type != opts.Type) {
				continue
			}
			if field, snip, ok := matchEntity(e.meta.Title, e.body, lowerTerm); ok {
				out = append(out, SearchResult{Kind: "doc", Code: code, Title: e.meta.Title,
					Field: field, Snippet: snip, Type: e.meta.Type, Status: e.meta.Status})
			}
		}
	}
	if (bothKinds || opts.Tasks) && opts.Type == "" {
		for code, e := range d.tasks {
			if !statusOK(e.meta.Status) {
				continue
			}
			if field, snip, ok := matchEntity(e.meta.Title, e.body, lowerTerm); ok {
				out = append(out, SearchResult{Kind: "task", Code: code, Title: e.meta.Title,
					Field: field, Snippet: snip, Status: e.meta.Status, Column: d.columnOf(code)})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Field != out[j].Field {
			return out[i].Field == "title" // title hits before body hits
		}
		return out[i].Code < out[j].Code
	})
	return out, nil
}

// matchEntity reports the first field (title before body) containing lowerTerm,
// with a snippet for body matches.
func matchEntity(title, body, lowerTerm string) (field, snippet string, ok bool) {
	if strings.Contains(strings.ToLower(title), lowerTerm) {
		return "title", "", true
	}
	if i := strings.Index(strings.ToLower(body), lowerTerm); i >= 0 {
		return "body", makeSnippet(body, i, len(lowerTerm)), true
	}
	return "", "", false
}

// makeSnippet returns a whitespace-collapsed window around the match, with "..."
// where it was trimmed. ponytail: byte window, so a multibyte rune at an edge
// may render garbled — acceptable for a preview.
func makeSnippet(body string, idx, termLen int) string {
	const before, after = 30, 60
	start := max(idx-before, 0)
	end := min(idx+termLen+after, len(body))
	s := strings.Join(strings.Fields(body[start:end]), " ")
	if start > 0 {
		s = "..." + s
	}
	if end < len(body) {
		s += "..."
	}
	return s
}
