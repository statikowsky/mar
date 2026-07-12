package store

import (
	"errors"
	"testing"
)

func TestLinkAndLookup(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "Auth", "design", "")
	tk, _ := s.CreateTask("Wire auth", "", "")
	if err := s.Link(d.Code, tk.Code); err != nil {
		t.Fatalf("Link: %v", err)
	}

	tasks, err := s.TasksForDoc(d.Code)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Code != tk.Code {
		t.Errorf("TasksForDoc = %v", codes(tasks))
	}

	docs, err := s.DocsForTask(tk.Code)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || docs[0].Code != d.Code {
		t.Errorf("DocsForTask = %d docs", len(docs))
	}
}

func TestDocumentAndTaskViewsUseRelatedData(t *testing.T) {
	s := newTestStore(t)
	doc, err := s.CreateDoc("auth", "Auth", "design", "See [[T-WIRE]].")
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.CreateTaskWithCode("WIRE", "Wire auth", "See [[AUTH]].", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Link(doc.Code, task.Code); err != nil {
		t.Fatal(err)
	}

	docView, err := s.DocumentView(doc.Code)
	if err != nil {
		t.Fatal(err)
	}
	if docView.Doc.Code != doc.Code || len(docView.Tasks) != 1 || docView.Tasks[0].Code != task.Code {
		t.Errorf("document view = %+v", docView)
	}
	if code, kind, ok := docView.Resolve(task.Code); !ok || code != task.Code || kind != "task" {
		t.Errorf("document resolver = %q %q %v", code, kind, ok)
	}
	if len(docView.Backlinks) != 1 || docView.Backlinks[0].Code != task.Code {
		t.Errorf("document backlinks = %+v", docView.Backlinks)
	}

	taskView, err := s.TaskView(task.Code)
	if err != nil {
		t.Fatal(err)
	}
	if taskView.Task.Code != task.Code || len(taskView.Docs) != 1 || taskView.Docs[0].Code != doc.Code {
		t.Errorf("task view = %+v", taskView)
	}
	if code, kind, ok := taskView.Resolve(doc.Code); !ok || code != doc.Code || kind != "doc" {
		t.Errorf("task resolver = %q %q %v", code, kind, ok)
	}
}

func TestValidateDocCodes(t *testing.T) {
	s := newTestStore(t)
	doc, err := s.CreateDoc("auth", "Auth", "design", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ValidateDocCodes([]string{"auth", doc.Code}); err != nil {
		t.Fatalf("ValidateDocCodes: %v", err)
	}
	if err := s.ValidateDocCodes([]string{doc.Code, "missing"}); err == nil {
		t.Fatal("ValidateDocCodes missing doc = nil error")
	} else {
		var validation *DocValidationError
		if !errors.As(err, &validation) || validation.Code != "missing" {
			t.Errorf("validation error = %v, want missing code", err)
		}
	}
}

func TestTaskCarriesColumnName(t *testing.T) {
	s := newTestStore(t)
	tk, _ := s.CreateTask("x", "", "")
	if tk.Column != "To do" {
		t.Errorf("Column = %q, want To do", tk.Column)
	}
}

func TestCodeLookups(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("auth", "Auth", "design", "")
	tk, _ := s.CreateTask("Wire auth", "", "")
	s.Link(d.Code, tk.Code)

	docCodes, err := s.DocCodesForTask(tk.Code)
	if err != nil {
		t.Fatal(err)
	}
	if len(docCodes) != 1 || docCodes[0] != d.Code {
		t.Errorf("DocCodesForTask = %v, want [%s]", docCodes, d.Code)
	}

	taskCodes, err := s.TaskCodesForDoc(d.Code)
	if err != nil {
		t.Fatal(err)
	}
	if len(taskCodes) != 1 || taskCodes[0] != tk.Code {
		t.Errorf("TaskCodesForDoc = %v, want [%s]", taskCodes, tk.Code)
	}
}

func TestLinkIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("a", "A", "design", "")
	tk, _ := s.CreateTask("t", "", "")
	if err := s.Link(d.Code, tk.Code); err != nil {
		t.Fatal(err)
	}
	if err := s.Link(d.Code, tk.Code); err != nil {
		t.Fatalf("second Link should be idempotent: %v", err)
	}
	tasks, _ := s.TasksForDoc(d.Code)
	if len(tasks) != 1 {
		t.Errorf("want 1 link, got %d", len(tasks))
	}
}

func TestUnlink(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("a", "A", "design", "")
	tk, _ := s.CreateTask("t", "", "")
	s.Link(d.Code, tk.Code)
	if err := s.Unlink(d.Code, tk.Code); err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.TasksForDoc(d.Code)
	if len(tasks) != 0 {
		t.Errorf("want 0 links after unlink, got %d", len(tasks))
	}
}

func TestDeletingDocClearsLinks(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("a", "A", "design", "")
	tk, _ := s.CreateTask("t", "", "")
	s.Link(d.Code, tk.Code)
	if err := s.DeleteDoc(d.Code); err != nil {
		t.Fatal(err)
	}
	docs, _ := s.DocsForTask(tk.Code)
	if len(docs) != 0 {
		t.Errorf("links should be cleared on doc delete, got %d", len(docs))
	}
}

func TestBacklinks(t *testing.T) {
	s := newTestStore(t)
	target, _ := s.CreateDoc("auth", "Auth", "design", "")
	// A doc and a task both reference [[AUTH]]; one doc doesn't.
	s.CreateDoc("login", "Login", "design", "Built on [[AUTH]].")
	s.CreateDoc("other", "Other", "design", "Unrelated.")
	tk, _ := s.CreateTask("Wire it", "See [[AUTH|the doc]].", "")

	got, err := s.Backlinks(target.Code)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"DOC-LOGIN": "doc", tk.Code: "task"}
	if len(got) != len(want) {
		t.Fatalf("Backlinks = %v, want %v", got, want)
	}
	for _, b := range got {
		if want[b.Code] != b.Kind {
			t.Errorf("backlink %s kind = %q, want %q", b.Code, b.Kind, want[b.Code])
		}
	}
}

func TestBacklinksIgnoresCodeFence(t *testing.T) {
	s := newTestStore(t)
	target, _ := s.CreateDoc("auth", "Auth", "design", "")
	s.CreateDoc("doc", "Doc", "design", "```\n[[AUTH]]\n```")
	got, _ := s.Backlinks(target.Code)
	if len(got) != 0 {
		t.Errorf("fenced [[AUTH]] should not be a backlink, got %v", got)
	}
}

func TestLint(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("auth", "Auth", "design", "")
	tk, _ := s.CreateTask("Wire it", "", "")
	// One link resolves to a doc, one to the task, one is a dangling code, and
	// one is not a valid code at all. Only the bad two should be reported.
	s.CreateDoc("login", "Login", "design",
		"Built on [[AUTH]] and [["+tk.Code+"]].\nSee [[GHOST]] and [[not a code]].")
	// A fenced dangling link must be ignored.
	s.CreateTask("Other", "```\n[[ALSO-GHOST]]\n```", "")

	findings, err := s.Lint()
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 2 {
		t.Fatalf("Lint = %+v, want 2 findings", findings)
	}
	byTarget := map[string]LintFinding{}
	for _, f := range findings {
		byTarget[f.Target] = f
	}
	ghost := byTarget["GHOST"]
	if ghost.Status != "dangling" || ghost.Normalized != "DOC-GHOST" || ghost.Source != "DOC-LOGIN" || ghost.Line != 2 {
		t.Errorf("GHOST finding = %+v, want dangling DOC-GHOST in DOC-LOGIN line 2", ghost)
	}
	if bad := byTarget["not a code"]; bad.Status != "invalid-code" || bad.Normalized != "" {
		t.Errorf("invalid finding = %+v, want invalid-code with no normalized", bad)
	}
}

func TestLintSkipsArchivedSources(t *testing.T) {
	s := newTestStore(t)
	d, _ := s.CreateDoc("stale", "Stale", "design", "Links to [[GHOST]].")
	if got, _ := s.Lint(); len(got) != 1 {
		t.Fatalf("want 1 finding before archive, got %+v", got)
	}
	s.ArchiveDoc(d.Code)
	if got, _ := s.Lint(); len(got) != 0 {
		t.Errorf("archived source should not be linted, got %+v", got)
	}
}

func TestLintResolvesTaskTarget(t *testing.T) {
	s := newTestStore(t)
	tk, _ := s.CreateTask("Real", "", "")
	s.CreateDoc("doc", "Doc", "design", "Points to [["+tk.Code+"]].")
	if got, _ := s.Lint(); len(got) != 0 {
		t.Errorf("link to existing task should resolve, got %+v", got)
	}
}

func TestResolver(t *testing.T) {
	s := newTestStore(t)
	s.CreateDoc("auth", "Auth", "design", "")
	tk, _ := s.CreateTask("t", "", "")
	resolve, err := s.Resolver()
	if err != nil {
		t.Fatal(err)
	}
	if code, kind, ok := resolve("auth"); !ok || code != "DOC-AUTH" || kind != "doc" {
		t.Errorf("resolve(auth) = %q %q %v", code, kind, ok)
	}
	if code, kind, ok := resolve(tk.Code); !ok || code != tk.Code || kind != "task" {
		t.Errorf("resolve(%s) = %q %q %v", tk.Code, code, kind, ok)
	}
	if _, _, ok := resolve("GHOST"); ok {
		t.Errorf("resolve(GHOST) should be missing")
	}
}
