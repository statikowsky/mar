package store

import (
	"fmt"
	"sort"

	"github.com/statikowsky/mar/internal/render"
)

// CodeResolver maps a raw wiki-link target ("WIKILINKS", "5", "T-5") to the
// canonical code and kind ("doc"/"task") of the entity it names, or ok=false
// when no such entity exists.
type CodeResolver func(raw string) (code, kind string, ok bool)

// Backlink is an entity whose body wiki-links to a given target.
type Backlink struct {
	Code  string `json:"code"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
}

// resolveRef canonicalizes a wiki-link target against the loaded data, trying a
// doc match before a task match. Unknown targets return ok=false.
func (d *data) resolveRef(raw string) (code, kind string, ok bool) {
	if full, err := normalizeDocCode(raw); err == nil {
		if _, exists := d.docs[full]; exists {
			return full, "doc", true
		}
	}
	if full, err := normalizeTaskCode(raw); err == nil {
		if _, exists := d.tasks[full]; exists {
			return full, "task", true
		}
	}
	return "", "", false
}

// Resolver returns a wiki-link resolver over a single store snapshot. Callers
// rebuild it per render. ponytail: reparses on each call; fine for a personal
// wiki, add a cached index if doc counts get large.
func (s *Store) Resolver() (CodeResolver, error) {
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	return d.resolveRef, nil
}

// Backlinks returns the docs and tasks whose body wiki-links ([[...]]) to the
// target code, sorted by code. ponytail: scans and parses every body per call;
// add a derived index if it gets slow.
func (s *Store) Backlinks(rawCode string) ([]Backlink, error) {
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	target, targetKind, ok := d.resolveRef(rawCode)
	if !ok {
		return nil, fmt.Errorf("%s: %w", rawCode, ErrNotFound)
	}
	var out []Backlink
	add := func(code, title, kind, body string) {
		if code == target && kind == targetKind {
			return // a body linking to itself is not a backlink
		}
		for _, ref := range render.Refs(body) {
			if rc, rk, ok := d.resolveRef(ref); ok && rc == target && rk == targetKind {
				out = append(out, Backlink{Code: code, Title: title, Kind: kind})
				return
			}
		}
	}
	for code, e := range d.docs {
		add(code, e.meta.Title, "doc", e.body)
	}
	for code, e := range d.tasks {
		add(code, e.meta.Title, "task", e.body)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, nil
}

func (s *Store) Link(docCode, taskCode string) error {
	full, err := normalizeDocCode(docCode)
	if err != nil {
		return err
	}
	return s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		doc, ok := d.docs[full]
		if !ok {
			return fmt.Errorf("doc %s: %w", full, ErrNotFound)
		}
		taskCode, _, err := d.findTask(taskCode)
		if err != nil {
			return err
		}
		for _, c := range doc.meta.Tasks {
			if c == taskCode {
				return nil
			}
		}
		doc.meta.Tasks = append(doc.meta.Tasks, taskCode)
		sort.Strings(doc.meta.Tasks)
		return s.writeDoc(full, doc)
	})
}

func (s *Store) Unlink(docCode, taskCode string) error {
	full, err := normalizeDocCode(docCode)
	if err != nil {
		return err
	}
	return s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		doc, ok := d.docs[full]
		if !ok {
			return fmt.Errorf("doc %s: %w", full, ErrNotFound)
		}
		taskCode, _, err := d.findTask(taskCode)
		if err != nil {
			return err
		}
		kept := doc.meta.Tasks[:0]
		removed := false
		for _, c := range doc.meta.Tasks {
			if c == taskCode {
				removed = true
				continue
			}
			kept = append(kept, c)
		}
		if !removed {
			return nil
		}
		doc.meta.Tasks = kept
		return s.writeDoc(full, doc)
	})
}

func (s *Store) TaskCodesForDoc(docCode string) ([]string, error) {
	full, err := normalizeDocCode(docCode)
	if err != nil {
		return nil, err
	}
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	doc, ok := d.docs[full]
	if !ok {
		return nil, fmt.Errorf("doc %s: %w", full, ErrNotFound)
	}
	codes := append([]string{}, doc.meta.Tasks...)
	sort.Strings(codes)
	return codes, nil
}

func (s *Store) DocCodesForTask(taskCode string) ([]string, error) {
	taskCode, err := normalizeTaskCode(taskCode)
	if err != nil {
		return nil, err
	}
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	var codes []string
	for docCode, doc := range d.docs {
		for _, c := range doc.meta.Tasks {
			if c == taskCode {
				codes = append(codes, docCode)
				break
			}
		}
	}
	sort.Strings(codes)
	return codes, nil
}

func (s *Store) TasksForDoc(docCode string) ([]Task, error) {
	codes, err := s.TaskCodesForDoc(docCode)
	if err != nil {
		return nil, err
	}
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, len(codes))
	for _, code := range codes {
		t, err := d.task(code)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *Store) DocsForTask(taskCode string) ([]Doc, error) {
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	taskCode, _, err = d.findTask(taskCode)
	if err != nil {
		return nil, err
	}
	codes, err := s.DocCodesForTask(taskCode)
	if err != nil {
		return nil, err
	}
	docs := make([]Doc, 0, len(codes))
	for _, code := range codes {
		doc, err := d.doc(code)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}
