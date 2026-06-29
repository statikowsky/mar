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
		body := filler
		if i%10 == 0 {
			body += "\n\nThis one mentions the needle keyword."
		}
		doc := &docEntity{meta: docMeta{Title: fmt.Sprintf("Doc title %d", i),
			Type: "design", Status: "active", Created: now, Updated: now}, body: body}
		if err := s.writeDoc(fmt.Sprintf("DOC-D%d", i), doc); err != nil {
			b.Fatal(err)
		}
		code := fmt.Sprintf("T-%d", i)
		task := &taskEntity{meta: taskMeta{Title: fmt.Sprintf("Task title %d", i),
			Status: "active", Created: now, Updated: now}, body: body}
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
