package store

// Interface is the storage contract. *Store (file-backed) is the only
// implementation today; a server-backed store would be a second one.
type Interface interface {
	Close() error
	DataVersion() (int64, error)
	Board() ([]BoardColumn, error)
	ListColumns() ([]Column, error)
	AddColumn(name, afterName string) (Column, error)
	AddColumnBefore(name, beforeName string) (Column, error)
	MoveColumn(name, target string, before bool) error
	RenameColumn(oldName, newName string) error
	RemoveColumn(name string, force bool) error
	CreateTask(title, body, columnName string) (Task, error)
	CreateTaskWithCode(rawCode, title, body, columnName string) (Task, error)
	GetTask(code string) (Task, error)
	ListTasks(columnName, status string) ([]Task, error)
	ArchivedTasks() ([]Task, error)
	EditTask(code string, title, body *string) (Task, error)
	SetTaskDates(code string, created, updated *string) (Task, error)
	MoveTask(code, columnName string, afterCode *string) (Task, error)
	ArchiveTask(code string) error
	UnarchiveTask(code string) error
	DeleteTask(code string) error
	CreateDoc(code, title, docType, body string) (Doc, error)
	GetDoc(code string) (Doc, error)
	ListDocs(docType, status string) ([]Doc, error)
	EditDoc(code string, title, docType, body *string) (Doc, error)
	SetDocDates(code string, created, updated *string) (Doc, error)
	RecodeDoc(oldCode, newCode string) (Doc, error)
	ArchiveDoc(code string) error
	UnarchiveDoc(code string) error
	DeleteDoc(code string) error
	Link(docCode, taskCode string) error
	Unlink(docCode, taskCode string) error
	TasksForDoc(docCode string) ([]Task, error)
	DocsForTask(taskCode string) ([]Doc, error)
	DocCodesForTask(taskCode string) ([]string, error)
	TaskCodesForDoc(docCode string) ([]string, error)
	Resolver() (CodeResolver, error)
	Backlinks(rawCode string) ([]Backlink, error)
}

var _ Interface = (*Store)(nil)
