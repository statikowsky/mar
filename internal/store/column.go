package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type BoardColumn struct {
	Column
	Tasks []Task `json:"tasks"`
}

// BoardView contains all data rendered on the board page from one store
// snapshot.
type BoardView struct {
	Columns        []BoardColumn
	ArchivedTasks  []Task
	DocCodesByTask map[string][]string
}

func (s *Store) Board() ([]BoardColumn, error) {
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	return d.boardColumns()
}

func (s *Store) BoardView() (BoardView, error) {
	d, err := s.load()
	if err != nil {
		return BoardView{}, err
	}
	columns, err := d.boardColumns()
	if err != nil {
		return BoardView{}, err
	}
	archivedTasks, err := d.archivedTasks()
	if err != nil {
		return BoardView{}, err
	}
	return BoardView{Columns: columns, ArchivedTasks: archivedTasks, DocCodesByTask: d.docCodesByTask()}, nil
}

func (d *data) boardColumns() ([]BoardColumn, error) {
	board := make([]BoardColumn, len(d.board.Columns))
	for i, c := range d.board.Columns {
		bc := BoardColumn{Column: Column{Name: c.Name}}
		for _, code := range c.Tasks {
			t, err := d.task(code)
			if err != nil {
				return nil, err
			}
			if t.Status != "active" {
				continue
			}
			t.Body = ""
			bc.Tasks = append(bc.Tasks, t)
		}
		board[i] = bc
	}
	return board, nil
}

func (d *data) archivedTasks() ([]Task, error) {
	var tasks []Task
	for _, c := range d.board.Columns {
		for _, code := range c.Tasks {
			t, err := d.task(code)
			if err != nil {
				return nil, err
			}
			if t.Status == "archived" {
				tasks = append(tasks, t)
			}
		}
	}
	return tasks, nil
}

func (d *data) docCodesByTask() map[string][]string {
	codesByTask := make(map[string][]string)
	for docCode, doc := range d.docs {
		for _, taskCode := range doc.meta.Tasks {
			codesByTask[taskCode] = append(codesByTask[taskCode], docCode)
		}
	}
	for taskCode := range codesByTask {
		sort.Strings(codesByTask[taskCode])
	}
	return codesByTask
}

func (s *Store) ListColumns() ([]Column, error) {
	d, err := s.load()
	if err != nil {
		return nil, err
	}
	cols := make([]Column, len(d.board.Columns))
	for i, c := range d.board.Columns {
		cols[i] = Column{Name: c.Name}
	}
	return cols, nil
}

func (d *data) insertColumn(idx int, name string) {
	cols := append(d.board.Columns, boardColumn{})
	copy(cols[idx+1:], cols[idx:])
	cols[idx] = boardColumn{Name: name, Tasks: []string{}}
	d.board.Columns = cols
}

func (s *Store) AddColumn(name, afterName string) (Column, error) {
	err := s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		if d.findColumn(name) != -1 {
			return fmt.Errorf("column %q already exists", name)
		}
		idx := len(d.board.Columns)
		if afterName != "" {
			i := d.findColumn(afterName)
			if i == -1 {
				return fmt.Errorf("unknown column %q: %w", afterName, ErrNotFound)
			}
			idx = i + 1
		}
		d.insertColumn(idx, name)
		return s.writeBoard(d.board)
	})
	return Column{Name: name}, err
}

// AddColumnBefore inserts a new column immediately before beforeName.
func (s *Store) AddColumnBefore(name, beforeName string) (Column, error) {
	err := s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		if d.findColumn(name) != -1 {
			return fmt.Errorf("column %q already exists", name)
		}
		i := d.findColumn(beforeName)
		if i == -1 {
			return fmt.Errorf("unknown column %q: %w", beforeName, ErrNotFound)
		}
		d.insertColumn(i, name)
		return s.writeBoard(d.board)
	})
	return Column{Name: name}, err
}

// MoveColumn reorders an existing column to sit before (before=true) or
// after (before=false) the target column.
func (s *Store) MoveColumn(name, target string, before bool) error {
	if name == target {
		return fmt.Errorf("cannot move column %q relative to itself", name)
	}
	return s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		i := d.findColumn(name)
		if i == -1 {
			return fmt.Errorf("column %q: %w", name, ErrNotFound)
		}
		moved := d.board.Columns[i]
		rest := append(append([]boardColumn{}, d.board.Columns[:i]...), d.board.Columns[i+1:]...)
		targetIdx := -1
		for j, c := range rest {
			if c.Name == target {
				targetIdx = j
				break
			}
		}
		if targetIdx == -1 {
			return fmt.Errorf("column %q: %w", target, ErrNotFound)
		}
		insertAt := targetIdx
		if !before {
			insertAt = targetIdx + 1
		}
		ordered := make([]boardColumn, 0, len(d.board.Columns))
		ordered = append(ordered, rest[:insertAt]...)
		ordered = append(ordered, moved)
		ordered = append(ordered, rest[insertAt:]...)
		d.board.Columns = ordered
		return s.writeBoard(d.board)
	})
}

func (s *Store) RenameColumn(oldName, newName string) error {
	return s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		i := d.findColumn(oldName)
		if i == -1 {
			return fmt.Errorf("column %q: %w", oldName, ErrNotFound)
		}
		if oldName != newName && d.findColumn(newName) != -1 {
			return fmt.Errorf("column %q already exists", newName)
		}
		d.board.Columns[i].Name = newName
		return s.writeBoard(d.board)
	})
}

func (s *Store) RemoveColumn(name string, force bool) error {
	return s.withLock(func() error {
		d, err := s.load()
		if err != nil {
			return err
		}
		i := d.findColumn(name)
		if i == -1 {
			return fmt.Errorf("unknown column %q: %w", name, ErrNotFound)
		}
		doomed := d.board.Columns[i].Tasks
		if len(doomed) > 0 && !force {
			return fmt.Errorf("column %q has %d task(s); pass --force to delete them too", name, len(doomed))
		}
		gone := map[string]bool{}
		for _, code := range doomed {
			gone[code] = true
		}
		for docCode, doc := range d.docs {
			kept := doc.meta.Tasks[:0]
			for _, code := range doc.meta.Tasks {
				if !gone[code] {
					kept = append(kept, code)
				}
			}
			if len(kept) != len(doc.meta.Tasks) {
				doc.meta.Tasks = kept
				if err := s.writeDoc(docCode, doc); err != nil {
					return err
				}
			}
		}
		for _, code := range doomed {
			if err := os.Remove(filepath.Join(s.dir, tasksDir, code+".md")); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("delete task %s: %w", code, err)
			}
		}
		d.board.Columns = append(d.board.Columns[:i], d.board.Columns[i+1:]...)
		return s.writeBoard(d.board)
	})
}
