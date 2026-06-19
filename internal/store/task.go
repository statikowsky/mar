package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var taskCodeRe = regexp.MustCompile(`^[A-Z0-9-]+$`)

func normalizeTaskCode(raw string) (string, error) {
	c := strings.ToUpper(strings.TrimSpace(raw))
	c = strings.TrimPrefix(c, "T-")
	if c == "" || !taskCodeRe.MatchString(c) {
		return "", fmt.Errorf("invalid task code %q: use letters, digits, and hyphens", raw)
	}
	return "T-" + c, nil
}

// slugMaxWords caps the number of title words in an auto-generated task code
// so long titles don't yield unwieldy codes. Explicit --code bypasses this.
const slugMaxWords = 4

func slugifyTitle(title string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range strings.ToUpper(title) {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevHyphen = false
		default:
			if !prevHyphen {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}
	words := strings.Split(strings.Trim(b.String(), "-"), "-")
	if len(words) > slugMaxWords {
		words = words[:slugMaxWords]
	}
	return strings.Join(words, "-")
}

// autoTaskCode derives a code from the title slug (T-WIRE-AUTH-LOGIN),
// adding a numeric suffix on collision. An empty slug falls back to
// max numeric code + 1.
func autoTaskCode(d *data, title string) string {
	slug := slugifyTitle(title)
	if slug == "" {
		return nextNumericCode(d)
	}
	base := "T-" + slug
	candidate := base
	for n := 2; ; n++ {
		if _, exists := d.tasks[candidate]; !exists {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", base, n)
	}
}

func nextNumericCode(d *data) string {
	max := 0
	for code := range d.tasks {
		if n, err := strconv.Atoi(strings.TrimPrefix(code, "T-")); err == nil && n > max {
			max = n
		}
	}
	return "T-" + strconv.Itoa(max+1)
}

func (s *Store) CreateTask(title, body, columnName string) (Task, error) {
	return s.CreateTaskWithCode("", title, body, columnName)
}

// CreateTaskWithCode creates a task with a caller-supplied code. An empty
// code auto-derives from the title slug; a supplied code is normalized to
// T-<CODE> and must be unique.
func (s *Store) CreateTaskWithCode(rawCode, title, body, columnName string) (Task, error) {
	var out Task
	err := s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		if len(d.board.Columns) == 0 {
			return errors.New("board has no columns")
		}
		idx := 0
		if columnName != "" {
			idx = d.findColumn(columnName)
			if idx == -1 {
				return fmt.Errorf("unknown column %q: %w", columnName, ErrNotFound)
			}
		}
		var code string
		if rawCode == "" {
			code = autoTaskCode(d, title)
		} else {
			code, err = normalizeTaskCode(rawCode)
			if err != nil {
				return err
			}
			if _, exists := d.tasks[code]; exists {
				return fmt.Errorf("task %s already exists", code)
			}
		}
		now := nowStamp()
		e := &taskEntity{meta: taskMeta{Title: title, Status: "active", Created: now, Updated: now}, body: body}
		if err := s.writeTask(code, e); err != nil {
			return err
		}
		col := &d.board.Columns[idx]
		col.Tasks = append(col.Tasks, code)
		if err := s.writeBoard(d.board); err != nil {
			return err
		}
		out = Task{Code: code, Title: title, Body: body, Column: col.Name,
			Status: "active", CreatedAt: now, UpdatedAt: now}
		return nil
	})
	return out, err
}

func (s *Store) GetTask(code string) (Task, error) {
	d, err := s.load()
	if err != nil {
		return Task{}, err
	}
	canonical, _, err := d.findTask(code)
	if err != nil {
		return Task{}, err
	}
	return d.task(canonical)
}

func (s *Store) ListTasks(columnName, status string) ([]Task, error) {
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	if columnName != "" && d.findColumn(columnName) == -1 {
		return nil, fmt.Errorf("unknown column %q: %w", columnName, ErrNotFound)
	}
	var tasks []Task
	for _, c := range d.board.Columns {
		if columnName != "" && c.Name != columnName {
			continue
		}
		for _, code := range c.Tasks {
			t, err := d.task(code)
			if err != nil {
				return nil, err
			}
			if status != "" && t.Status != status {
				continue
			}
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

// ArchivedTasks returns all archived tasks across columns, ordered by
// column then board position.
func (s *Store) ArchivedTasks() ([]Task, error) {
	return s.ListTasks("", "archived")
}

func (s *Store) setTaskStatus(code, status string) error {
	return s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		canonical, e, err := d.findTask(code)
		if err != nil {
			return err
		}
		e.meta.Status = status
		e.meta.Updated = nowStamp()
		return s.writeTask(canonical, e)
	})
}

func (s *Store) ArchiveTask(code string) error   { return s.setTaskStatus(code, "archived") }
func (s *Store) UnarchiveTask(code string) error { return s.setTaskStatus(code, "active") }

func (s *Store) EditTask(code string, title, body *string) (Task, error) {
	var out Task
	err := s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		canonical, e, err := d.findTask(code)
		if err != nil {
			return err
		}
		if title != nil {
			e.meta.Title = *title
		}
		if body != nil {
			e.body = *body
		}
		e.meta.Updated = nowStamp()
		if err := s.writeTask(canonical, e); err != nil {
			return err
		}
		out, err = d.task(canonical)
		return err
	})
	return out, err
}

// SetTaskDates overrides created and/or updated with caller-supplied dates
// (YYYY-MM-DD or RFC3339). Nil leaves a field unchanged. Unlike EditTask
// this does not auto-bump updated, so historical timestamps survive.
func (s *Store) SetTaskDates(code string, created, updated *string) (Task, error) {
	var out Task
	err := s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		canonical, e, err := d.findTask(code)
		if err != nil {
			return err
		}
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
		if err := s.writeTask(canonical, e); err != nil {
			return err
		}
		out, err = d.task(canonical)
		return err
	})
	return out, err
}

// MoveTask moves a task to columnName (empty keeps its column) and places
// it after afterCode, or at the top of the column when afterCode is nil.
func (s *Store) MoveTask(code, columnName string, afterCode *string) (Task, error) {
	var out Task
	err := s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		code, e, err := d.findTask(code)
		if err != nil {
			return err
		}
		var after string
		if afterCode != nil {
			after, _, err = d.findTask(*afterCode)
			if err != nil {
				return err
			}
		}
		target := d.columnOf(code)
		if columnName != "" {
			target = columnName
		}
		ti := d.findColumn(target)
		if ti == -1 {
			return fmt.Errorf("unknown column %q: %w", target, ErrNotFound)
		}
		d.removeFromBoard(code)
		col := &d.board.Columns[ti]
		insert := 0
		if afterCode != nil {
			ai := -1
			for j, t := range col.Tasks {
				if t == after {
					ai = j
					break
				}
			}
			if ai == -1 {
				return fmt.Errorf("task %s is not in the target column", after)
			}
			insert = ai + 1
		}
		col.Tasks = append(col.Tasks, "")
		copy(col.Tasks[insert+1:], col.Tasks[insert:])
		col.Tasks[insert] = code
		e.meta.Updated = nowStamp()
		if err := s.writeTask(code, e); err != nil {
			return err
		}
		if err := s.writeBoard(d.board); err != nil {
			return err
		}
		out, err = d.task(code)
		return err
	})
	return out, err
}

func (s *Store) DeleteTask(code string) error {
	return s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		code, _, err := d.findTask(code)
		if err != nil {
			return err
		}
		for docCode, doc := range d.docs {
			kept := doc.meta.Tasks[:0]
			for _, c := range doc.meta.Tasks {
				if c != code {
					kept = append(kept, c)
				}
			}
			if len(kept) != len(doc.meta.Tasks) {
				doc.meta.Tasks = kept
				if err := s.writeDoc(docCode, doc); err != nil {
					return err
				}
			}
		}
		if err := os.Remove(filepath.Join(s.dir, tasksDir, code+".md")); err != nil {
			return fmt.Errorf("delete task %s: %w", code, err)
		}
		d.removeFromBoard(code)
		return s.writeBoard(d.board)
	})
}
