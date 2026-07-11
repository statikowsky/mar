package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	scratchpadSchema        = 2
	legacyScratchpadSchema  = 1
	defaultScratchNoteWidth = 260
	minScratchNoteWidth     = 160
	maxScratchNoteWidth     = 800
	maxScratchCoordinate    = 1_000_000
	maxScratchTextLength    = 20_000
)

var (
	ErrConflict   = errors.New("conflict")
	scratchIDRe   = regexp.MustCompile(`^S-([1-9][0-9]*)$`)
	scratchLinkRe = regexp.MustCompile(`^(?:T|DOC)-[A-Z0-9]+(?:-[A-Z0-9]+)*$`)
	scratchColor  = map[string]bool{
		"neutral": true, "blue": true, "green": true,
		"yellow": true, "red": true, "purple": true,
	}
)

type ScratchNote struct {
	ID        string          `json:"id" yaml:"id"`
	Text      string          `json:"text" yaml:"text"`
	X         int             `json:"x" yaml:"x"`
	Y         int             `json:"y" yaml:"y"`
	Width     int             `json:"width" yaml:"width"`
	Color     string          `json:"color" yaml:"color"`
	Z         int             `json:"z" yaml:"z"`
	Link      string          `json:"link,omitempty" yaml:"link,omitempty"`
	Docs      []ScratchDocRef `json:"docs,omitempty" yaml:"docs,omitempty"`
	CreatedAt string          `json:"created_at" yaml:"created"`
	UpdatedAt string          `json:"updated_at" yaml:"updated"`
}

type ScratchDocRef struct {
	Code   string         `json:"code" yaml:"code"`
	Anchor *ScratchAnchor `json:"anchor,omitempty" yaml:"anchor,omitempty"`
}

type ScratchAnchor struct {
	Block string `json:"block" yaml:"block"`
	Quote string `json:"quote,omitempty" yaml:"quote,omitempty"`
}

type Scratchpad struct {
	Schema   int           `json:"schema" yaml:"version"`
	Revision int64         `json:"revision" yaml:"revision"`
	Notes    []ScratchNote `json:"notes" yaml:"notes"`
}

type scratchpadFile struct {
	Schema   int           `yaml:"version"`
	Revision int64         `yaml:"revision"`
	NextNote int           `yaml:"next_note"`
	Notes    []ScratchNote `yaml:"notes"`
}

func emptyScratchpadFile() scratchpadFile {
	return scratchpadFile{Schema: scratchpadSchema, NextNote: 1, Notes: []ScratchNote{}}
}

func (s *Store) readScratchpad() (scratchpadFile, error) {
	raw, err := os.ReadFile(filepath.Join(s.dir, scratchpadName))
	if errors.Is(err, os.ErrNotExist) {
		return emptyScratchpadFile(), nil
	}
	if err != nil {
		return scratchpadFile{}, fmt.Errorf("read scratchpad: %w", err)
	}
	var f scratchpadFile
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return scratchpadFile{}, fmt.Errorf("parse scratchpad: %w", err)
	}
	if f.Schema != legacyScratchpadSchema && f.Schema != scratchpadSchema {
		return scratchpadFile{}, fmt.Errorf("unsupported scratchpad version %d", f.Schema)
	}
	if f.NextNote < 1 {
		f.NextNote = nextScratchID(f.Notes)
	}
	if f.Notes == nil {
		f.Notes = []ScratchNote{}
	}
	if err := validateScratchNotes(f.Notes); err != nil {
		return scratchpadFile{}, fmt.Errorf("parse scratchpad: %w", err)
	}
	return f, nil
}

func (s *Store) writeScratchpad(f scratchpadFile) error {
	f.Schema = scratchpadSchema
	raw, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal scratchpad: %w", err)
	}
	if err := writeFileAtomic(filepath.Join(s.dir, scratchpadName), raw); err != nil {
		return fmt.Errorf("write scratchpad: %w", err)
	}
	return nil
}

func publicScratchpad(f scratchpadFile) Scratchpad {
	return Scratchpad{Schema: f.Schema, Revision: f.Revision, Notes: append([]ScratchNote{}, f.Notes...)}
}

func (s *Store) Scratchpad() (Scratchpad, error) {
	var out Scratchpad
	err := s.withLock(func() error {
		f, err := s.readScratchpad()
		if err != nil {
			return err
		}
		out = publicScratchpad(f)
		return nil
	})
	return out, err
}

func (s *Store) SaveScratchpad(expectedRevision int64, notes []ScratchNote) (Scratchpad, error) {
	var out Scratchpad
	notes = append([]ScratchNote{}, notes...)
	err := s.withLock(func() error {
		f, err := s.readScratchpad()
		if err != nil {
			return err
		}
		if f.Revision != expectedRevision {
			return fmt.Errorf("scratchpad revision is %d, expected %d: %w", f.Revision, expectedRevision, ErrConflict)
		}
		existing := make(map[string]ScratchNote, len(f.Notes))
		for _, note := range f.Notes {
			existing[note.ID] = note
		}
		now := nowStamp()
		for i := range notes {
			old, ok := existing[notes[i].ID]
			if !ok {
				n, valid := scratchNumber(notes[i].ID)
				if !valid || n >= f.NextNote {
					return fmt.Errorf("scratch note %s must be created before saving", notes[i].ID)
				}
				continue
			}
			notes[i].CreatedAt = old.CreatedAt
			if scratchNoteChanged(old, notes[i]) {
				notes[i].UpdatedAt = now
			} else {
				notes[i].UpdatedAt = old.UpdatedAt
			}
		}
		if err := validateScratchNotes(notes); err != nil {
			return err
		}
		f.Notes = append([]ScratchNote{}, notes...)
		f.Schema = scratchpadSchema
		f.NextNote = max(f.NextNote, nextScratchID(notes))
		f.Revision++
		if err := s.writeScratchpad(f); err != nil {
			return err
		}
		out = publicScratchpad(f)
		return nil
	})
	return out, err
}

func scratchNoteChanged(a, b ScratchNote) bool {
	return a.Text != b.Text || a.X != b.X || a.Y != b.Y || a.Width != b.Width ||
		a.Color != b.Color || a.Z != b.Z || a.Link != b.Link || !reflect.DeepEqual(a.Docs, b.Docs)
}

func (s *Store) CreateScratchNote(text string, x, y, width int, color string) (ScratchNote, error) {
	var out ScratchNote
	err := s.withLock(func() error {
		f, err := s.readScratchpad()
		if err != nil {
			return err
		}
		if width == 0 {
			width = defaultScratchNoteWidth
		}
		if color == "" {
			color = "neutral"
		}
		now := nowStamp()
		out = ScratchNote{ID: fmt.Sprintf("S-%d", f.NextNote), Text: text, X: x, Y: y,
			Width: width, Color: color, Z: nextScratchZ(f.Notes), CreatedAt: now, UpdatedAt: now}
		if err := validateScratchNote(out); err != nil {
			return err
		}
		f.NextNote++
		f.Revision++
		f.Notes = append(f.Notes, out)
		return s.writeScratchpad(f)
	})
	return out, err
}

func (s *Store) UpdateScratchNote(note ScratchNote) (ScratchNote, error) {
	var out ScratchNote
	err := s.withLock(func() error {
		f, err := s.readScratchpad()
		if err != nil {
			return err
		}
		index := -1
		for i := range f.Notes {
			if f.Notes[i].ID == note.ID {
				index = i
				break
			}
		}
		if index < 0 {
			return fmt.Errorf("scratch note %s: %w", note.ID, ErrNotFound)
		}
		note.CreatedAt = f.Notes[index].CreatedAt
		note.UpdatedAt = nowStamp()
		if err := validateScratchNote(note); err != nil {
			return err
		}
		f.Notes[index] = note
		f.Revision++
		if err := s.writeScratchpad(f); err != nil {
			return err
		}
		out = note
		return nil
	})
	return out, err
}

func (s *Store) DeleteScratchNote(id string) error {
	return s.withLock(func() error {
		f, err := s.readScratchpad()
		if err != nil {
			return err
		}
		kept := f.Notes[:0]
		found := false
		for _, note := range f.Notes {
			if note.ID == id {
				found = true
				continue
			}
			kept = append(kept, note)
		}
		if !found {
			return fmt.Errorf("scratch note %s: %w", id, ErrNotFound)
		}
		f.Notes = kept
		f.Revision++
		return s.writeScratchpad(f)
	})
}

func validateScratchNotes(notes []ScratchNote) error {
	seen := map[string]bool{}
	for _, note := range notes {
		if seen[note.ID] {
			return fmt.Errorf("duplicate scratch note %s", note.ID)
		}
		seen[note.ID] = true
		if err := validateScratchNote(note); err != nil {
			return err
		}
	}
	return nil
}

func validateScratchNote(note ScratchNote) error {
	if !scratchIDRe.MatchString(note.ID) {
		return fmt.Errorf("invalid scratch note id %q", note.ID)
	}
	if strings.TrimSpace(note.Text) == "" {
		return fmt.Errorf("scratch note text must not be empty")
	}
	if len(note.Text) > maxScratchTextLength {
		return fmt.Errorf("scratch note text exceeds %d bytes", maxScratchTextLength)
	}
	if note.X < -maxScratchCoordinate || note.X > maxScratchCoordinate ||
		note.Y < -maxScratchCoordinate || note.Y > maxScratchCoordinate {
		return fmt.Errorf("scratch note coordinates are out of range")
	}
	if note.Width < minScratchNoteWidth || note.Width > maxScratchNoteWidth {
		return fmt.Errorf("scratch note width must be between %d and %d", minScratchNoteWidth, maxScratchNoteWidth)
	}
	if !scratchColor[note.Color] {
		return fmt.Errorf("invalid scratch note color %q", note.Color)
	}
	if note.Z < 0 {
		return fmt.Errorf("scratch note z must not be negative")
	}
	if note.CreatedAt != "" {
		if _, err := time.Parse(time.RFC3339Nano, note.CreatedAt); err != nil {
			return fmt.Errorf("invalid scratch note created timestamp: %w", err)
		}
	}
	if note.UpdatedAt != "" {
		if _, err := time.Parse(time.RFC3339Nano, note.UpdatedAt); err != nil {
			return fmt.Errorf("invalid scratch note updated timestamp: %w", err)
		}
	}
	if note.Link != "" && !scratchLinkRe.MatchString(note.Link) {
		return fmt.Errorf("invalid scratch note link %q", note.Link)
	}
	seenDocs := map[string]bool{}
	for i := range note.Docs {
		note.Docs[i].Code = strings.ToUpper(strings.TrimSpace(note.Docs[i].Code))
		if !strings.HasPrefix(note.Docs[i].Code, "DOC-") || !scratchLinkRe.MatchString(note.Docs[i].Code) {
			return fmt.Errorf("invalid scratch document code %q", note.Docs[i].Code)
		}
		if seenDocs[note.Docs[i].Code] {
			return fmt.Errorf("duplicate scratch document %s", note.Docs[i].Code)
		}
		seenDocs[note.Docs[i].Code] = true
		if anchor := note.Docs[i].Anchor; anchor != nil {
			anchor.Block = strings.TrimSpace(anchor.Block)
			anchor.Quote = strings.TrimSpace(anchor.Quote)
			if anchor.Block == "" || len(anchor.Block) > 300 || len(anchor.Quote) > 500 {
				return fmt.Errorf("invalid scratch document anchor for %s", note.Docs[i].Code)
			}
		}
	}
	return nil
}

func nextScratchID(notes []ScratchNote) int {
	next := 1
	for _, note := range notes {
		n, ok := scratchNumber(note.ID)
		if !ok {
			continue
		}
		next = max(next, n+1)
	}
	return next
}

func scratchNumber(id string) (int, bool) {
	match := scratchIDRe.FindStringSubmatch(id)
	if len(match) != 2 {
		return 0, false
	}
	n, err := strconv.Atoi(match[1])
	return n, err == nil
}

func nextScratchZ(notes []ScratchNote) int {
	z := 0
	for _, note := range notes {
		z = max(z, note.Z)
	}
	return z + 1
}
