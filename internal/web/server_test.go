package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/statikowsky/mar/internal/store"
	"github.com/statikowsky/mar/internal/version"
)

func newTestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	s, err := store.Init(t.TempDir())
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	srv := httptest.NewServer(NewServer(s, "test-repo", "/tmp/my-project").Handler())
	t.Cleanup(func() {
		srv.Close()
		s.Close()
	})
	return srv, s
}

func get(t *testing.T, url string) (int, string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return resp.StatusCode, b.String()
}

func TestTaskFragmentOmitsLayout(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Modal me", "Body **here**.", "")

	fragCode, frag := get(t, srv.URL+"/task/T-1?fragment=1")
	if fragCode != 200 {
		t.Fatalf("fragment status = %d", fragCode)
	}
	if !strings.Contains(frag, "Modal me") {
		t.Errorf("fragment missing task title: %s", frag)
	}
	if strings.Contains(frag, "<html") || strings.Contains(frag, "EventSource") {
		t.Errorf("fragment should omit layout chrome: %s", frag)
	}

	fullCode, full := get(t, srv.URL+"/task/T-1")
	if fullCode != 200 || !strings.Contains(full, "<html") {
		t.Errorf("full task page should still render layout: status=%d", fullCode)
	}
}

func TestTaskFragmentNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	code, _ := get(t, srv.URL+"/task/T-99?fragment=1")
	if code != 404 {
		t.Errorf("status = %d, want 404", code)
	}
}

func postJSON(t *testing.T, url, body string) (int, string) {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return resp.StatusCode, b.String()
}

func requestJSON(t *testing.T, method, url, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(raw)
}

func TestScratchpadPageAndIndexLink(t *testing.T) {
	srv, _ := newTestServer(t)
	code, body := get(t, srv.URL+"/scratchpad")
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	for _, want := range []string{`<body class="scratch-page">`, `id="scratch-surface"`, `window.SCRATCHPAD_STATE`, `aria-label="Scratchpad notes"`, `new URLSearchParams(location.search).get("note")`, `linkedDocHref(note)`, `"?note=" + encodeURIComponent(note.id)`, `scratch-drag-handle`, `data-note-action="edit"`, `data-note-action="delete"`, `Create task`, `Create document`, `window.marIcon("pencil")`, `window.marIcon("list-todo")`, `window.marIcon("file-plus-2")`, `closest("[data-list-action]")`, `normalizeWheelDelta`, `isDiscreteWheel`, `scheduleViewSave`, `e.ctrlKey || e.metaKey`, `e.shiftKey`, `applyViewTransform`, `note-shortcut`, `⌘↵ saves`, `note-action-control`, `note-action-select`, `note-action-link`, `window.marIcon("file-text")`} {
		if !strings.Contains(body, want) {
			t.Errorf("scratchpad page missing %q:\n%s", want, body)
		}
	}
	_, css := get(t, srv.URL+"/static/style.css")
	for _, want := range []string{`.note-action-control`, `height: 28px`, `.note-action-select`, `.note-action-control:focus-visible`, `.note-action-control:disabled`, `.scratch-list-row .note-icon-button`, `box-sizing: border-box`} {
		if !strings.Contains(css, want) {
			t.Errorf("note action styles missing %q", want)
		}
	}
	_, index := get(t, srv.URL+"/")
	if !strings.Contains(index, `href="/scratchpad"`) {
		t.Errorf("index missing scratchpad link:\n%s", index)
	}
}

func TestPagesUseSharedActionIcons(t *testing.T) {
	srv, s := newTestServer(t)
	doc, err := s.CreateDoc("ICONS", "Icons", "design", "Body")
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.CreateTask("Icons", "Body", "")
	if err != nil {
		t.Fatal(err)
	}
	pages := []struct {
		path  string
		wants []string
	}{
		{"/", []string{`data-mar-icon="file-plus-2"`, `data-mar-icon="archive"`}},
		{"/board", []string{`data-mar-icon="square-plus"`, `data-mar-icon="x"`}},
		{"/doc/" + doc.Code, []string{`data-mar-icon="arrow-left"`, `data-mar-icon="sticky-note"`, `data-mar-icon="pencil"`, `data-mar-icon="archive"`, `data-mar-icon="save"`}},
		{"/task/" + task.Code, []string{`data-mar-icon="arrow-left"`, `data-mar-icon="archive"`}},
		{"/scratchpad", []string{`data-mar-icon="undo-2"`, `data-mar-icon="redo-2"`, `data-mar-icon="copy"`, `data-mar-icon="zoom-in"`, `data-mar-icon="scan"`}},
	}
	for _, page := range pages {
		code, body := get(t, srv.URL+page.path)
		if code != http.StatusOK {
			t.Fatalf("GET %s status = %d", page.path, code)
		}
		for _, want := range page.wants {
			if !strings.Contains(body, want) {
				t.Errorf("GET %s missing %q", page.path, want)
			}
		}
		if !strings.Contains(body, `window.marHydrateIcons`) {
			t.Errorf("GET %s missing shared icon hydrator", page.path)
		}
	}
	code, fragment := get(t, srv.URL+"/task/"+task.Code+"?fragment=1")
	if code != http.StatusOK || !strings.Contains(fragment, `data-mar-icon="save"`) {
		t.Errorf("task modal editor missing shared save icon")
	}
}

func TestScratchpadCreateAndSaveRoutes(t *testing.T) {
	srv, _ := newTestServer(t)
	code, body := postJSON(t, srv.URL+"/scratchpad/note", `{"text":"Idea","x":20,"y":30,"color":"blue"}`)
	if code != http.StatusCreated {
		t.Fatalf("create status = %d: %s", code, body)
	}
	var created store.Scratchpad
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatal(err)
	}
	if created.Revision != 1 || len(created.Notes) != 1 || created.Notes[0].ID != "S-1" {
		t.Fatalf("created = %+v", created)
	}
	created.Notes[0].Text = "Changed"
	payload, _ := json.Marshal(map[string]any{"revision": created.Revision, "notes": created.Notes})
	code, body = requestJSON(t, http.MethodPut, srv.URL+"/scratchpad", string(payload))
	if code != http.StatusOK {
		t.Fatalf("save status = %d: %s", code, body)
	}
	code, _ = requestJSON(t, http.MethodPut, srv.URL+"/scratchpad", string(payload))
	if code != http.StatusConflict {
		t.Fatalf("stale save status = %d, want 409", code)
	}
}

func TestDocumentPageIncludesAnchoredNotesRail(t *testing.T) {
	srv, s := newTestServer(t)
	doc, err := s.CreateDoc("ANNOTATED", "Annotated", "design", "## Setup\n\nInstall mar here.")
	if err != nil {
		t.Fatal(err)
	}
	code, body := get(t, srv.URL+"/doc/"+doc.Code)
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	for _, want := range []string{`class="doc-annotation-gutter"`, `data-tooltip="Create note"`, `id="doc-notes-rail"`, `data-doc-code="DOC-ANNOTATED"`, `if (associated().length) openRail(false)`, `positionNotes()`, `positionNotesRail()`, `gutter.getBoundingClientRect().right`, `--doc-notes-left`, `new URLSearchParams(location.search).get("note")`, `scrollIntoView({ behavior: "smooth", block: "center" })`, `doc-note-target`, `note-action-link`, `note-icon-button danger`, `window.marIcon("sticky-note")`, `window.marIcon("save")`, `"save": '<path`, `Save changes`, `Create note`, `Saving…`, `note-shortcut`, `⌘↵ saves`, `data-note-status`, `hasDirtyNotes()`, `Discard unsaved note changes?`, `beforeunload`, `deleteRow(row`} {
		if !strings.Contains(body, want) {
			t.Errorf("document page missing %q", want)
		}
	}

	code, body = postJSON(t, srv.URL+"/scratchpad/note", `{"text":"Check install","docs":[{"code":"DOC-ANNOTATED","anchor":{"block":"setup-1","quote":"Install mar here."}}]}`)
	if code != http.StatusCreated {
		t.Fatalf("create status = %d: %s", code, body)
	}
	pad, err := s.Scratchpad()
	if err != nil {
		t.Fatal(err)
	}
	if got := pad.Notes[0].Docs[0].Anchor.Block; got != "setup-1" {
		t.Fatalf("anchor block = %q", got)
	}
}

func TestScratchpadPromotesNoteToTaskAndDocument(t *testing.T) {
	srv, s := newTestServer(t)
	note, err := s.CreateScratchNote("Ship scratchpad\nDetailed body", 0, 0, 260, "yellow")
	if err != nil {
		t.Fatal(err)
	}
	code, body := postJSON(t, srv.URL+"/scratchpad/note/"+note.ID+"/promote", `{"kind":"task"}`)
	if code != http.StatusCreated {
		t.Fatalf("task promotion status = %d: %s", code, body)
	}
	pad, _ := s.Scratchpad()
	if len(pad.Notes) != 1 || !strings.HasPrefix(pad.Notes[0].Link, "T-") {
		t.Fatalf("task promotion link = %+v", pad.Notes)
	}

	note, err = s.CreateScratchNote("Scratch reference\nDocument body", 0, 0, 260, "neutral")
	if err != nil {
		t.Fatal(err)
	}
	code, body = postJSON(t, srv.URL+"/scratchpad/note/"+note.ID+"/promote", `{"kind":"doc","code":"SCRATCH-REF","type":"reference"}`)
	if code != http.StatusCreated {
		t.Fatalf("doc promotion status = %d: %s", code, body)
	}
	doc, err := s.GetDoc("DOC-SCRATCH-REF")
	if err != nil || doc.Title != "Scratch reference" || doc.Body != "Document body\n" {
		t.Fatalf("promoted doc = %+v, err = %v", doc, err)
	}
}

func TestStaticCSSLoadsVendoredInter(t *testing.T) {
	srv, _ := newTestServer(t)
	code, css := get(t, srv.URL+"/static/style.css")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	for _, want := range []string{
		"@font-face",
		`font-family: "Inter"`,
		"url(\"/static/fonts/InterVariable.woff2\") format(\"woff2\")",
		`--ui-font: "Inter", -apple-system`,
		"font: 15px/1.55 var(--ui-font)",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("style.css missing %q", want)
		}
	}

	resp, err := http.Get(srv.URL + "/static/fonts/InterVariable.woff2")
	if err != nil {
		t.Fatalf("GET font: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("font status = %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read font: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("font response is empty")
	}
}

func TestMoveTaskRouteMovesToColumn(t *testing.T) {
	srv, s := newTestServer(t)
	if _, err := s.CreateTaskWithCode("1", "Drag me", "", ""); err != nil {
		t.Fatal(err)
	}
	code, _ := postJSON(t, srv.URL+"/task/T-1/move", `{"column":"Done"}`)
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	board, err := s.Board()
	if err != nil {
		t.Fatal(err)
	}
	for _, col := range board {
		if col.Name == "Done" {
			if len(col.Tasks) != 1 || col.Tasks[0].Code != "T-1" {
				t.Errorf("Done = %v, want [T-1]", col.Tasks)
			}
		}
		if col.Name == "To do" && len(col.Tasks) != 0 {
			t.Errorf("To do should be empty, got %v", col.Tasks)
		}
	}
}

func TestMoveTaskRouteReordersAfter(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "a", "", "")
	s.CreateTaskWithCode("2", "b", "", "")
	s.CreateTaskWithCode("3", "c", "", "")
	code, _ := postJSON(t, srv.URL+"/task/T-3/move", `{"column":"To do","after":"T-1"}`)
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	tasks, _ := s.ListTasks("To do", "")
	got := []string{tasks[0].Code, tasks[1].Code, tasks[2].Code}
	want := []string{"T-1", "T-3", "T-2"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestMoveTaskRouteUnknownCode(t *testing.T) {
	srv, _ := newTestServer(t)
	code, _ := postJSON(t, srv.URL+"/task/T-99/move", `{"column":"Done"}`)
	if code != 404 {
		t.Errorf("status = %d, want 404", code)
	}
}

func TestMoveTaskRouteUnknownColumn(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "x", "", "")
	code, _ := postJSON(t, srv.URL+"/task/T-1/move", `{"column":"Nope"}`)
	if code != 400 {
		t.Errorf("status = %d, want 400", code)
	}
}

func TestArchiveDocRoute(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	code, _ := postJSON(t, srv.URL+"/doc/DOC-AUTH/archive", "")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	d, _ := s.GetDoc("DOC-AUTH")
	if d.Status != "archived" {
		t.Errorf("status = %q, want archived", d.Status)
	}
}

func TestUnarchiveDocRoute(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	s.ArchiveDoc("DOC-AUTH")
	code, _ := postJSON(t, srv.URL+"/doc/DOC-AUTH/unarchive", "")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	d, _ := s.GetDoc("DOC-AUTH")
	if d.Status != "active" {
		t.Errorf("status = %q, want active", d.Status)
	}
}

func TestDeleteDocRouteRequiresArchived(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	// Active doc cannot be deleted: 409 guard.
	code, _ := postJSON(t, srv.URL+"/doc/DOC-AUTH/delete", "")
	if code != 409 {
		t.Fatalf("delete active status = %d, want 409", code)
	}
	if _, err := s.GetDoc("DOC-AUTH"); err != nil {
		t.Errorf("active doc should survive delete attempt: %v", err)
	}
	// Once archived, delete succeeds.
	s.ArchiveDoc("DOC-AUTH")
	code, _ = postJSON(t, srv.URL+"/doc/DOC-AUTH/delete", "")
	if code != 200 {
		t.Fatalf("delete archived status = %d, want 200", code)
	}
	if _, err := s.GetDoc("DOC-AUTH"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("doc should be gone after delete")
	}
}

func TestDocLifecycleRoutesUnknownCode(t *testing.T) {
	srv, _ := newTestServer(t)
	for _, action := range []string{"archive", "unarchive", "delete"} {
		code, _ := postJSON(t, srv.URL+"/doc/DOC-NOPE/"+action, "")
		if code != 404 {
			t.Errorf("%s unknown code status = %d, want 404", action, code)
		}
	}
}

func TestArchiveTaskRoute(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Card", "", "")
	code, _ := postJSON(t, srv.URL+"/task/T-1/archive", "")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	tk, _ := s.GetTask("T-1")
	if tk.Status != "archived" {
		t.Errorf("status = %q, want archived", tk.Status)
	}
}

func TestUnarchiveTaskRoute(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Card", "", "")
	s.ArchiveTask("T-1")
	code, _ := postJSON(t, srv.URL+"/task/T-1/unarchive", "")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	tk, _ := s.GetTask("T-1")
	if tk.Status != "active" {
		t.Errorf("status = %q, want active", tk.Status)
	}
}

func TestDeleteTaskRouteRequiresArchived(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Card", "", "")
	// Active card cannot be deleted: 409 guard.
	code, _ := postJSON(t, srv.URL+"/task/T-1/delete", "")
	if code != 409 {
		t.Fatalf("delete active status = %d, want 409", code)
	}
	if _, err := s.GetTask("T-1"); err != nil {
		t.Errorf("active card should survive delete attempt: %v", err)
	}
	s.ArchiveTask("T-1")
	code, _ = postJSON(t, srv.URL+"/task/T-1/delete", "")
	if code != 200 {
		t.Fatalf("delete archived status = %d, want 200", code)
	}
	if _, err := s.GetTask("T-1"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("card should be gone after delete")
	}
}

func TestTaskLifecycleRoutesUnknownCode(t *testing.T) {
	srv, _ := newTestServer(t)
	for _, action := range []string{"archive", "unarchive", "delete"} {
		code, _ := postJSON(t, srv.URL+"/task/T-99/"+action, "")
		if code != 404 {
			t.Errorf("%s unknown code status = %d, want 404", action, code)
		}
	}
}

func TestBoardNoArchivedSectionWhenNoneArchived(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Active", "", "")
	_, body := get(t, srv.URL+"/board")
	if strings.Contains(body, "Archived") {
		t.Errorf("board should omit Archived section when none archived:\n%s", body)
	}
}

func TestBoardShowsArchivedSection(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Active", "", "")
	s.CreateTaskWithCode("2", "Old", "", "")
	s.ArchiveTask("T-2")
	_, body := get(t, srv.URL+"/board")
	if !strings.Contains(body, "Archived") {
		t.Errorf("board missing Archived section:\n%s", body)
	}
	// Archived card is not a draggable board card but appears with actions.
	if !strings.Contains(body, `data-code="T-2" data-action="unarchive"`) ||
		!strings.Contains(body, `data-code="T-2" data-action="delete"`) {
		t.Errorf("archived card row missing unarchive/delete actions:\n%s", body)
	}
}

func TestTaskFragmentShowsActions(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Card", "", "")
	_, body := get(t, srv.URL+"/task/T-1?fragment=1")
	if !strings.Contains(body, `data-code="T-1" data-action="archive"`) {
		t.Errorf("active task fragment missing archive action:\n%s", body)
	}
	s.ArchiveTask("T-1")
	_, body = get(t, srv.URL+"/task/T-1?fragment=1")
	if !strings.Contains(body, `data-code="T-1" data-action="unarchive"`) ||
		!strings.Contains(body, `data-code="T-1" data-action="delete"`) {
		t.Errorf("archived task fragment missing unarchive/delete actions:\n%s", body)
	}
}

func TestIndexListsDocs(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth design", "design", "")
	code, body := get(t, srv.URL+"/")
	if code != 200 {
		t.Fatalf("status = %d", code)
	}
	if !strings.Contains(body, "DOC-AUTH") || !strings.Contains(body, "Auth design") {
		t.Errorf("index missing doc: %s", body)
	}
}

func TestIndexNoArchivedSectionWhenNoneArchived(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	_, body := get(t, srv.URL+"/")
	if strings.Contains(body, "Archived") {
		t.Errorf("index should omit Archived section when none archived:\n%s", body)
	}
	// Active doc row offers an Archive action.
	if !strings.Contains(body, `data-code="DOC-AUTH" data-action="archive"`) {
		t.Errorf("active row missing archive action:\n%s", body)
	}
}

func TestIndexShowsArchivedSection(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	s.CreateDoc("old", "Old", "design", "")
	s.ArchiveDoc("DOC-OLD")
	_, body := get(t, srv.URL+"/")
	if !strings.Contains(body, "Archived") {
		t.Errorf("index missing Archived section:\n%s", body)
	}
	// Archived row offers unarchive + delete; active doc is not in the archived list.
	if !strings.Contains(body, `data-code="DOC-OLD" data-action="unarchive"`) ||
		!strings.Contains(body, `data-code="DOC-OLD" data-action="delete"`) {
		t.Errorf("archived row missing unarchive/delete actions:\n%s", body)
	}
}

func TestDocPageShowsActions(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	_, body := get(t, srv.URL+"/doc/DOC-AUTH")
	if !strings.Contains(body, `data-code="DOC-AUTH" data-action="archive"`) {
		t.Errorf("active doc page missing archive action:\n%s", body)
	}
	s.ArchiveDoc("DOC-AUTH")
	_, body = get(t, srv.URL+"/doc/DOC-AUTH")
	if !strings.Contains(body, `data-code="DOC-AUTH" data-action="unarchive"`) ||
		!strings.Contains(body, `data-code="DOC-AUTH" data-action="delete"`) {
		t.Errorf("archived doc page missing unarchive/delete actions:\n%s", body)
	}
}

func TestDocPageRenders(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "# Hello\n\n> [!TIP]\n> Great.")
	code, body := get(t, srv.URL+"/doc/DOC-AUTH")
	if code != 200 {
		t.Fatalf("status = %d", code)
	}
	if !strings.Contains(body, "<h1") || !strings.Contains(body, `<blockquote class="alert alert-tip">`) {
		t.Errorf("doc not rendered: %s", body)
	}
}

func TestDocPageIncludesReadingOutline(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "## Overview\n\n### Details")
	code, body := get(t, srv.URL+"/doc/DOC-AUTH")
	if code != 200 {
		t.Fatalf("status = %d", code)
	}
	for _, want := range []string{
		`<body class="doc-page">`,
		`class="doc-layout"`,
		`class="doc-outline"`,
		`aria-label="Table of contents"`,
		`class="doc-outline-list"`,
		`id="overview"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("doc page missing %q:\n%s", want, body)
		}
	}
}

func TestDocPageShowsCopyableCodeChip(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	code, body := get(t, srv.URL+"/doc/DOC-AUTH")
	if code != 200 {
		t.Fatalf("status = %d", code)
	}
	if !strings.Contains(body, `<button type="button" class="code-chip" data-copy-code="DOC-AUTH"`) {
		t.Errorf("doc page missing copyable code chip:\n%s", body)
	}
	if !strings.Contains(body, `navigator.clipboard.writeText`) {
		t.Errorf("doc page missing clipboard handler:\n%s", body)
	}
}

func TestDocPageNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	code, _ := get(t, srv.URL+"/doc/DOC-NOPE")
	if code != 404 {
		t.Errorf("status = %d, want 404", code)
	}
}

func TestBoardPageShowsTask(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTask("Wire it", "", "")
	code, body := get(t, srv.URL+"/board")
	if code != 200 {
		t.Fatalf("status = %d", code)
	}
	if !strings.Contains(body, "T-WIRE-IT") || !strings.Contains(body, "Wire it") {
		t.Errorf("board missing task: %s", body)
	}
}

func TestBoardShowsLinkedDocCode(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	s.CreateTaskWithCode("1", "Wire it", "", "")
	if err := s.Link("DOC-AUTH", "T-1"); err != nil {
		t.Fatal(err)
	}
	code, body := get(t, srv.URL+"/board")
	if code != 200 {
		t.Fatalf("status = %d", code)
	}
	if !strings.Contains(body, "DOC-AUTH") {
		t.Errorf("board card missing linked doc code: %s", body)
	}
}

func TestTaskPageRenders(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Wire auth", "Some **notes**.", "")
	code, body := get(t, srv.URL+"/task/T-1")
	if code != 200 {
		t.Fatalf("status = %d", code)
	}
	if !strings.Contains(body, "Wire auth") || !strings.Contains(body, "<strong>notes</strong>") {
		t.Errorf("task page not rendered: %s", body)
	}
}

func TestTaskPageShowsCopyableCodeChip(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Wire auth", "", "")
	code, body := get(t, srv.URL+"/task/T-1")
	if code != 200 {
		t.Fatalf("status = %d", code)
	}
	if !strings.Contains(body, `<button type="button" class="code-chip" data-copy-code="T-1"`) {
		t.Errorf("task page missing copyable code chip:\n%s", body)
	}
	if !strings.Contains(body, `navigator.clipboard.writeText`) {
		t.Errorf("task page missing clipboard handler:\n%s", body)
	}

	code, body = get(t, srv.URL+"/task/T-1?fragment=1")
	if code != 200 {
		t.Fatalf("fragment status = %d", code)
	}
	if !strings.Contains(body, `<button type="button" class="code-chip" data-copy-code="T-1"`) {
		t.Errorf("task fragment missing copyable code chip:\n%s", body)
	}
}

func TestTaskPageNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	code, _ := get(t, srv.URL+"/task/T-99")
	if code != 404 {
		t.Errorf("status = %d, want 404", code)
	}
}

func TestStaticCSSServed(t *testing.T) {
	srv, _ := newTestServer(t)
	code, body := get(t, srv.URL+"/static/style.css")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	if !strings.Contains(body, "--accent") {
		t.Errorf("style.css not served correctly: %.120s", body)
	}
}

func TestVersionEndpointReturnsJSON(t *testing.T) {
	srv, _ := newTestServer(t)
	code, body := get(t, srv.URL+"/events/version")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	var v struct {
		Version int64 `json:"version"`
	}
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("unmarshal %q: %v", body, err)
	}
}

func TestVersionEndpointReflectsStore(t *testing.T) {
	srv, s := newTestServer(t)
	want, err := s.DataVersion()
	if err != nil {
		t.Fatal(err)
	}
	_, body := get(t, srv.URL+"/events/version")
	var v struct {
		Version int64 `json:"version"`
	}
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("unmarshal %q: %v", body, err)
	}
	if v.Version != want {
		t.Errorf("endpoint version = %d, want %d (store DataVersion)", v.Version, want)
	}
}

func TestProjectNameShownOnPages(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	s.CreateTaskWithCode("1", "Wire it", "", "")
	for _, path := range []string{"/", "/board", "/doc/DOC-AUTH", "/task/T-1"} {
		code, body := get(t, srv.URL+path)
		if code != 200 {
			t.Fatalf("%s status = %d", path, code)
		}
		if !strings.Contains(body, "my-project") {
			t.Errorf("%s missing project name", path)
		}
	}
}

func TestVersionShownOnPages(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")
	_, body := get(t, srv.URL+"/")
	if !strings.Contains(body, "mar "+version.Version) {
		t.Errorf("index page missing version banner %q:\n%s", "mar "+version.Version, body)
	}
}

func TestThemeFromCookie(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "")

	cases := []struct {
		cookie string
		want   string // substring expected (or "" -> the no-attr form)
	}{
		{"dark", `data-theme="dark"`},
		{"light", `data-theme="light"`},
		{"github-light", `data-theme="github-light"`},
		{"gruvbox-light", `data-theme="gruvbox-light"`},
		{"gruvbox-dark", `data-theme="gruvbox-dark"`},
		{"solarized-light", `data-theme="solarized-light"`},
		{"solarized-dark", `data-theme="solarized-dark"`},
		{"dracula", `data-theme="dracula"`},
		{"nord", `data-theme="nord"`},
	}
	for _, tc := range cases {
		req, _ := http.NewRequest("GET", srv.URL+"/", nil)
		req.AddCookie(&http.Cookie{Name: "theme", Value: tc.cookie})
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET with cookie %s: %v", tc.cookie, err)
		}
		body := readAll(t, resp)
		resp.Body.Close()
		if !strings.Contains(body, tc.want) {
			t.Errorf("cookie %q: response missing %q", tc.cookie, tc.want)
		}
	}

	// An unknown scheme id falls back to the no-attr (follow-OS) form.
	req, _ := http.NewRequest("GET", srv.URL+"/", nil)
	req.AddCookie(&http.Cookie{Name: "theme", Value: "bogus"})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET with unknown cookie: %v", err)
	}
	unknownBody := readAll(t, resp)
	resp.Body.Close()
	if strings.Contains(unknownBody, "data-theme=") {
		t.Errorf("unknown scheme should yield no data-theme attribute:\n%s", unknownBody)
	}

	// No cookie -> no server-rendered data-theme (OS-driven client-side).
	_, body := get(t, srv.URL+"/")
	if strings.Contains(body, "data-theme=") {
		t.Errorf("no cookie should yield no data-theme attribute:\n%s", body)
	}
	// The toggle and scheme picker are always present.
	if !strings.Contains(body, "theme-toggle") {
		t.Errorf("theme toggle button missing from page")
	}
	if !strings.Contains(body, "theme-cog") {
		t.Errorf("theme scheme-picker cog missing from page")
	}
}

func readAll(t *testing.T, resp *http.Response) string {
	t.Helper()
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return b.String()
}

func TestEditTaskUpdatesAndReturnsFragment(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Old title", "old body", "")

	code, body := postJSON(t, srv.URL+"/task/T-1/edit",
		`{"title":"New title","body":"new **body**"}`)
	if code != 200 {
		t.Fatalf("edit status = %d: %s", code, body)
	}
	// Returns the re-rendered fragment (no layout chrome), reflecting the edit.
	if !strings.Contains(body, "New title") {
		t.Errorf("fragment missing new title: %s", body)
	}
	if strings.Contains(body, "<html") {
		t.Errorf("edit response should be a fragment, not a full page: %s", body)
	}
	if !strings.Contains(body, "<strong>body</strong>") {
		t.Errorf("fragment body not rendered from new markdown: %s", body)
	}

	tk, err := s.GetTask("T-1")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if tk.Title != "New title" || tk.Body != "new **body**\n" {
		t.Errorf("stored task not updated: %+v", tk)
	}
}

func TestEditTaskRejectsEmptyTitle(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Keep", "body", "")

	code, _ := postJSON(t, srv.URL+"/task/T-1/edit", `{"title":"   ","body":"x"}`)
	if code != http.StatusBadRequest {
		t.Errorf("empty title: status = %d, want 400", code)
	}
	if code, _ := postJSON(t, srv.URL+"/task/T-NOPE/edit",
		`{"title":"x","body":"y"}`); code != http.StatusNotFound {
		t.Errorf("unknown code: status = %d, want 404", code)
	}
}

func TestEditDocUpdatesTitleTypeBody(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "old")

	code, body := postJSON(t, srv.URL+"/doc/DOC-AUTH/edit",
		`{"title":"Login","type":"plan","body":"new"}`)
	if code != 200 {
		t.Fatalf("edit status = %d: %s", code, body)
	}
	d, err := s.GetDoc("DOC-AUTH")
	if err != nil {
		t.Fatalf("GetDoc: %v", err)
	}
	if d.Title != "Login" || d.Type != "plan" || d.Body != "new\n" {
		t.Errorf("stored doc not updated: %+v", d)
	}
}

func TestEditDocRejectsBadInput(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateDoc("auth", "Auth", "design", "old")

	if code, _ := postJSON(t, srv.URL+"/doc/DOC-AUTH/edit",
		`{"title":"","type":"design","body":"x"}`); code != http.StatusBadRequest {
		t.Errorf("empty title: status = %d, want 400", code)
	}
	if code, _ := postJSON(t, srv.URL+"/doc/DOC-AUTH/edit",
		`{"title":"ok","type":"bogus","body":"x"}`); code != http.StatusBadRequest {
		t.Errorf("invalid type: status = %d, want 400", code)
	}
	if code, _ := postJSON(t, srv.URL+"/doc/DOC-NOPE/edit",
		`{"title":"ok","type":"design","body":"x"}`); code != http.StatusNotFound {
		t.Errorf("unknown code: status = %d, want 404", code)
	}
}

func TestPreviewRendersWithoutMutating(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "T", "", "")
	s.CreateDoc("auth", "Auth", "design", "")

	before, err := s.DataVersion()
	if err != nil {
		t.Fatalf("DataVersion: %v", err)
	}

	code, body := postJSON(t, srv.URL+"/preview",
		`{"body":"## Heading\n\n> [!NOTE]\n> hi"}`)
	if code != 200 {
		t.Fatalf("preview status = %d: %s", code, body)
	}
	if !strings.Contains(body, "<h2") {
		t.Errorf("preview did not render markdown heading: %s", body)
	}

	after, err := s.DataVersion()
	if err != nil {
		t.Fatalf("DataVersion: %v", err)
	}
	if before != after {
		t.Errorf("preview mutated the store: version %d -> %d", before, after)
	}
}

func TestCreateTaskRoute(t *testing.T) {
	srv, s := newTestServer(t)
	code, body := postJSON(t, srv.URL+"/task", `{"title":"Wire it","body":"do **this**"}`)
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", code, body)
	}
	// Lands in the first column with an auto-generated code.
	board, err := s.Board()
	if err != nil {
		t.Fatal(err)
	}
	if len(board) == 0 || len(board[0].Tasks) != 1 || board[0].Tasks[0].Title != "Wire it" {
		t.Errorf("new card not in first column: %+v", board)
	}

	// Empty title is rejected.
	if code, _ := postJSON(t, srv.URL+"/task", `{"title":"   ","body":"x"}`); code != http.StatusBadRequest {
		t.Errorf("empty title: status = %d, want 400", code)
	}
}

func TestNewTaskFormFragment(t *testing.T) {
	srv, _ := newTestServer(t)
	code, body := get(t, srv.URL+"/task/new?fragment=1")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	if strings.Contains(body, "<html") {
		t.Errorf("create-card form should be a fragment, not a full page: %s", body)
	}
	if !strings.Contains(body, `data-create="task"`) {
		t.Errorf("fragment missing create form: %s", body)
	}
}

func TestCreateDocRoute(t *testing.T) {
	srv, s := newTestServer(t)
	code, body := postJSON(t, srv.URL+"/doc",
		`{"code":"AUTH","title":"Auth design","type":"design","body":"hi"}`)
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", code, body)
	}
	var resp struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal %q: %v", body, err)
	}
	if resp.Code != "DOC-AUTH" {
		t.Errorf("returned code = %q, want DOC-AUTH", resp.Code)
	}
	if _, err := s.GetDoc("DOC-AUTH"); err != nil {
		t.Errorf("doc not created: %v", err)
	}

	// Empty title and invalid type are 400.
	if code, _ := postJSON(t, srv.URL+"/doc",
		`{"code":"X","title":"","type":"design","body":""}`); code != http.StatusBadRequest {
		t.Errorf("empty title: status = %d, want 400", code)
	}
	if code, _ := postJSON(t, srv.URL+"/doc",
		`{"code":"Y","title":"ok","type":"bogus","body":""}`); code != http.StatusBadRequest {
		t.Errorf("invalid type: status = %d, want 400", code)
	}
	// Duplicate code is 409.
	if code, _ := postJSON(t, srv.URL+"/doc",
		`{"code":"AUTH","title":"Dup","type":"design","body":""}`); code != http.StatusConflict {
		t.Errorf("duplicate code: status = %d, want 409", code)
	}
}

func TestNewDocFormPage(t *testing.T) {
	srv, _ := newTestServer(t)
	code, body := get(t, srv.URL+"/doc/new")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	if !strings.Contains(body, "<html") || !strings.Contains(body, `id="doc-create"`) {
		t.Errorf("create-doc page missing layout or form: %s", body)
	}
}

func TestFaviconNotServedIndex(t *testing.T) {
	srv, _ := newTestServer(t)
	resp, err := http.Get(srv.URL + "/favicon.ico")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); strings.HasPrefix(ct, "text/html") {
		t.Errorf("favicon served HTML (matched the catch-all): %q", ct)
	}
}
