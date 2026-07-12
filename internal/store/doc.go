package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// DocTypes lists the valid document types in display order (used for the doc
// editor's type selector). validDocTypes is derived from it for lookups.
var DocTypes = []string{
	"design", "analysis", "plan", "report", "board", "reference", "tooling",
}

var (
	docCodeRe     = regexp.MustCompile(`^[A-Z0-9-]+$`)
	validDocTypes = func() map[string]bool {
		m := make(map[string]bool, len(DocTypes))
		for _, t := range DocTypes {
			m[t] = true
		}
		return m
	}()
)

func normalizeDocCode(raw string) (string, error) {
	c := strings.ToUpper(strings.TrimSpace(raw))
	c = strings.TrimPrefix(c, "DOC-")
	if c == "" || !docCodeRe.MatchString(c) {
		return "", fmt.Errorf("invalid doc code %q: use letters, digits, and hyphens", raw)
	}
	return "DOC-" + c, nil
}

func validateDocType(t string) error {
	if !validDocTypes[t] {
		return fmt.Errorf("invalid doc type %q", t)
	}
	return nil
}

func (s *Store) CreateDoc(code, title, docType, body string) (Doc, error) {
	full, err := normalizeDocCode(code)
	if err != nil {
		return Doc{}, err
	}
	if err := validateDocType(docType); err != nil {
		return Doc{}, err
	}
	var out Doc
	err = s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		if _, exists := d.docs[full]; exists {
			return fmt.Errorf("doc %s already exists", full)
		}
		now := nowStamp()
		e := &docEntity{meta: docMeta{Title: title, Type: docType, Status: "active", Created: now, Updated: now}, body: body}
		if err := s.writeDoc(full, e); err != nil {
			return err
		}
		out = Doc{Code: full, Title: title, Type: docType, Body: body,
			Status: "active", CreatedAt: now, UpdatedAt: now}
		return nil
	})
	return out, err
}

func (s *Store) GetDoc(code string) (Doc, error) {
	full, err := normalizeDocCode(code)
	if err != nil {
		return Doc{}, err
	}
	d, err := s.load()
	if err != nil {
		return Doc{}, err
	}
	return d.doc(full)
}

func docTime(stamp string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, stamp)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (s *Store) ListDocs(docType, status string) ([]Doc, error) {
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	return d.listDocs(docType, status)
}

func (d *data) listDocs(docType, status string) ([]Doc, error) {
	var docs []Doc
	for code := range d.docs {
		doc, err := d.doc(code)
		if err != nil {
			return nil, err
		}
		if docType != "" && doc.Type != docType {
			continue
		}
		if status != "" && doc.Status != status {
			continue
		}
		docs = append(docs, doc)
	}
	sort.Slice(docs, func(i, j int) bool {
		ti, tj := docTime(docs[i].UpdatedAt), docTime(docs[j].UpdatedAt)
		if ti.Equal(tj) {
			return docs[i].Code < docs[j].Code
		}
		return ti.After(tj)
	})
	return docs, nil
}

func (s *Store) mutateDoc(code string, fn func(e *docEntity) error) (Doc, error) {
	full, err := normalizeDocCode(code)
	if err != nil {
		return Doc{}, err
	}
	var out Doc
	err = s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		e, ok := d.docs[full]
		if !ok {
			return fmt.Errorf("doc %s: %w", full, ErrNotFound)
		}
		if err := fn(e); err != nil {
			return err
		}
		if err := s.writeDoc(full, e); err != nil {
			return err
		}
		out, err = d.doc(full)
		return err
	})
	return out, err
}

func (s *Store) EditDoc(code string, title, docType, body *string) (Doc, error) {
	return s.mutateDoc(code, func(e *docEntity) error {
		if title != nil {
			e.meta.Title = *title
		}
		if docType != nil {
			if err := validateDocType(*docType); err != nil {
				return err
			}
			e.meta.Type = *docType
		}
		if body != nil {
			e.body = *body
		}
		e.meta.Updated = nowStamp()
		return nil
	})
}

// SetDocDates overrides created and/or updated with caller-supplied dates
// (YYYY-MM-DD or RFC3339). Nil leaves a field unchanged. Unlike EditDoc
// this does not auto-bump updated, so historical timestamps can be
// preserved.
func (s *Store) SetDocDates(code string, created, updated *string) (Doc, error) {
	return s.mutateDoc(code, func(e *docEntity) error {
		if created != nil {
			stamp, err := normalizeDate(*created)
			if err != nil {
				return err
			}
			e.meta.Created = stamp
		}
		if updated != nil {
			stamp, err := normalizeDate(*updated)
			if err != nil {
				return err
			}
			e.meta.Updated = stamp
		}
		return nil
	})
}

func (s *Store) RecodeDoc(oldCode, newCode string) (Doc, error) {
	oldFull, err := normalizeDocCode(oldCode)
	if err != nil {
		return Doc{}, err
	}
	newFull, err := normalizeDocCode(newCode)
	if err != nil {
		return Doc{}, err
	}
	var out Doc
	err = s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		e, ok := d.docs[oldFull]
		if !ok {
			return fmt.Errorf("doc %s: %w", oldFull, ErrNotFound)
		}
		if oldFull == newFull {
			out, err = d.doc(oldFull)
			return err
		}
		if _, exists := d.docs[newFull]; exists {
			return fmt.Errorf("doc %s already exists", newFull)
		}
		e.meta.Updated = nowStamp()
		if err := s.writeDoc(newFull, e); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(s.dir, docsDir, oldFull+".md")); err != nil {
			return fmt.Errorf("recode doc to %s: %w", newFull, err)
		}
		d.docs[newFull] = e
		out, err = d.doc(newFull)
		return err
	})
	return out, err
}

func (s *Store) setDocStatus(code, status string) error {
	_, err := s.mutateDoc(code, func(e *docEntity) error {
		e.meta.Status = status
		e.meta.Updated = nowStamp()
		return nil
	})
	return err
}

func (s *Store) ArchiveDoc(code string) error   { return s.setDocStatus(code, "archived") }
func (s *Store) UnarchiveDoc(code string) error { return s.setDocStatus(code, "active") }

func (s *Store) DeleteDoc(code string) error {
	full, err := normalizeDocCode(code)
	if err != nil {
		return err
	}
	return s.withLock(func() error {
		err := os.Remove(filepath.Join(s.dir, docsDir, full+".md"))
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("doc %s: %w", full, ErrNotFound)
		}
		if err != nil {
			return fmt.Errorf("delete doc %s: %w", full, err)
		}
		return nil
	})
}
