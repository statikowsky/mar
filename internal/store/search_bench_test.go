package store

import (
	"fmt"
	"strings"
	"testing"
)

// synthStore writes n docs and n tasks straight to disk (bypassing CreateDoc/
// CreateTask, whose per-call load() would make setup O(n²)). Each entity gets a
// ~1.4 KB markdown body; every 10th one contains "needle" so the term hits a
// known fraction.
func synthStore(b *testing.B, n int) *Store {
	b.Helper()
	dir := b.TempDir()
	s, err := Init(dir)
	if err != nil {
		b.Fatalf("Init: %v", err)
	}
	filler := strings.Repeat("lorem ipsum dolor sit amet consectetur adipiscing elit ", 25)
	now := nowStamp()
	taskCodes := make([]string, 0, n)
	for i := range n {
		code := fmt.Sprintf("T-%d", i)
		docBody, taskBody := filler, filler
		if i%10 == 0 {
			docBody += "\n\nThis one mentions the needle keyword."
			taskBody += "\n\nThis one mentions the needle keyword."
		}
		docBody += "\n\nRelated task: [[" + code + "]]"
		taskBody += fmt.Sprintf("\n\nRelated document: [[DOC-D%d]]", i)
		doc := &docEntity{meta: docMeta{Title: fmt.Sprintf("Doc title %d", i),
			Type: "design", Status: "active", Created: now, Updated: now, Tasks: []string{code}}, body: docBody}
		if err := s.writeDoc(fmt.Sprintf("DOC-D%d", i), doc); err != nil {
			b.Fatal(err)
		}
		task := &taskEntity{meta: taskMeta{Title: fmt.Sprintf("Task title %d", i),
			Status: "active", Created: now, Updated: now}, body: taskBody}
		if err := s.writeTask(code, task); err != nil {
			b.Fatal(err)
		}
		taskCodes = append(taskCodes, code)
	}
	board := boardFile{Columns: []boardColumn{
		{Name: "To do", Tasks: taskCodes}, {Name: "In progress"}, {Name: "Done"},
	}}
	if err := s.writeBoard(board); err != nil {
		b.Fatal(err)
	}
	return s
}

// BenchmarkReadViews measures the full fresh-load path behind the board,
// document, and task pages. The synthesized entities have links, backlinks,
// and document-task associations so each view performs its normal assembly.
func BenchmarkReadViews(b *testing.B) {
	for _, n := range []int{100, 1000} {
		s := synthStore(b, n)
		b.Run(fmt.Sprintf("board/%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if _, err := s.BoardView(); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("document/%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if _, err := s.DocumentView("DOC-D0"); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("task/%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if _, err := s.TaskView("T-0"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkSearch measures the cold path a real `mar search` pays: each call
// reloads and reparses the whole store from disk, then scans. Measurement only
// — no thresholds. Run with: go test -bench=Search ./internal/store/
func BenchmarkSearch(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			s := synthStore(b, n)
			b.ResetTimer()
			for range b.N {
				if _, err := s.Search("needle", SearchOpts{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
