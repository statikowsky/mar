package store

// DocValidationError identifies the input code that failed batch validation.
type DocValidationError struct {
	Code string
	Err  error
}

func (e *DocValidationError) Error() string { return e.Err.Error() }

func (e *DocValidationError) Unwrap() error { return e.Err }

// IndexView contains all data rendered by the document index from one store
// snapshot.
type IndexView struct {
	ActiveDocs   []Doc
	ArchivedDocs []Doc
}

func (s *Store) IndexView() (IndexView, error) {
	d, err := s.load()
	if err != nil {
		return IndexView{}, err
	}
	active, err := d.listDocs("", "active")
	if err != nil {
		return IndexView{}, err
	}
	archived, err := d.listDocs("", "archived")
	if err != nil {
		return IndexView{}, err
	}
	return IndexView{ActiveDocs: active, ArchivedDocs: archived}, nil
}

// DocumentView contains all store-backed data rendered by a document page
// from one snapshot.
type DocumentView struct {
	Doc       Doc
	Tasks     []Task
	Backlinks []Backlink
	Resolve   CodeResolver
}

func (s *Store) DocumentView(code string) (DocumentView, error) {
	d, err := s.load()
	if err != nil {
		return DocumentView{}, err
	}
	full, err := normalizeDocCode(code)
	if err != nil {
		return DocumentView{}, err
	}
	doc, err := d.doc(full)
	if err != nil {
		return DocumentView{}, err
	}
	tasks, err := d.tasksForDoc(full)
	if err != nil {
		return DocumentView{}, err
	}
	backlinks, err := d.backlinks(full)
	if err != nil {
		return DocumentView{}, err
	}
	return DocumentView{Doc: doc, Tasks: tasks, Backlinks: backlinks, Resolve: d.resolveRef}, nil
}

// TaskView contains all store-backed data rendered by a task page from one
// snapshot.
type TaskView struct {
	Task    Task
	Docs    []Doc
	Resolve CodeResolver
}

func (s *Store) TaskView(code string) (TaskView, error) {
	d, err := s.load()
	if err != nil {
		return TaskView{}, err
	}
	canonical, _, err := d.findTask(code)
	if err != nil {
		return TaskView{}, err
	}
	task, err := d.task(canonical)
	if err != nil {
		return TaskView{}, err
	}
	docs, err := d.docsForTask(canonical)
	if err != nil {
		return TaskView{}, err
	}
	return TaskView{Task: task, Docs: docs, Resolve: d.resolveRef}, nil
}

func (s *Store) ValidateDocCodes(codes []string) error {
	d, err := s.load()
	if err != nil {
		return err
	}
	for _, code := range codes {
		full, err := normalizeDocCode(code)
		if err != nil {
			return &DocValidationError{Code: code, Err: err}
		}
		if _, err := d.doc(full); err != nil {
			return &DocValidationError{Code: code, Err: err}
		}
	}
	return nil
}
