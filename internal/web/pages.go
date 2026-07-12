package web

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/statikowsky/mar/internal/render"
	"github.com/statikowsky/mar/internal/store"
)

func shortDate(stamp string) string {
	if len(stamp) >= 10 {
		return stamp[:10]
	}
	return stamp
}

type docRow struct {
	store.Doc
	Updated string
}

func (srv *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	view, err := srv.store.IndexView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	srv.render(w, r, "index", map[string]any{
		"Title":    srv.repo,
		"Repo":     srv.repo,
		"Docs":     toDocRows(view.ActiveDocs),
		"Archived": toDocRows(view.ArchivedDocs),
	})
}

func toDocRows(docs []store.Doc) []docRow {
	rows := make([]docRow, len(docs))
	for i, d := range docs {
		rows[i] = docRow{Doc: d, Updated: shortDate(d.UpdatedAt)}
	}
	return rows
}

type boardTask struct {
	store.Task
	DocCodes []string
}

type boardColumnView struct {
	Name  string
	Tasks []boardTask
}

func (srv *Server) handleBoard(w http.ResponseWriter, r *http.Request) {
	view, err := srv.store.BoardView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cols := make([]boardColumnView, len(view.Columns))
	for i, bc := range view.Columns {
		tasks := make([]boardTask, len(bc.Tasks))
		for j, tk := range bc.Tasks {
			tasks[j] = boardTask{Task: tk, DocCodes: view.DocCodesByTask[tk.Code]}
		}
		cols[i] = boardColumnView{Name: bc.Name, Tasks: tasks}
	}

	archived := make([]relatedTask, len(view.ArchivedTasks))
	for i, tk := range view.ArchivedTasks {
		archived[i] = relatedTask{Task: tk, ColumnName: tk.Column}
	}

	srv.render(w, r, "board", map[string]any{
		"Title":    "Board — " + srv.repo,
		"Columns":  cols,
		"Archived": archived,
	})
}

type relatedTask struct {
	store.Task
	ColumnName string
}

// wikiResolver builds a render.Resolver mapping [[CODE]] targets to web URLs,
// over a single store snapshot. Errors render plain links (no resolution).
func (srv *Server) wikiResolver() render.Resolver {
	resolve, err := srv.store.Resolver()
	if err != nil {
		return nil
	}
	return wikiResolver(resolve)
}

func wikiResolver(resolve store.CodeResolver) render.Resolver {
	return func(raw string) (string, bool) {
		code, kind, ok := resolve(raw)
		if !ok {
			return "", false
		}
		if kind == "task" {
			return "/task/" + code, true
		}
		return "/doc/" + code, true
	}
}

func (srv *Server) handleDoc(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	view, err := srv.store.DocumentView(code)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bodyHTML, err := render.RenderMarkdownLinks(view.Doc.Body, wikiResolver(view.Resolve))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	related := make([]relatedTask, len(view.Tasks))
	for i, t := range view.Tasks {
		related[i] = relatedTask{Task: t, ColumnName: t.Column}
	}
	pad, err := srv.store.Scratchpad()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rawPad, err := json.Marshal(pad)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	srv.render(w, r, "doc", map[string]any{
		"Title":          view.Doc.Title,
		"BodyClass":      "doc-page",
		"Doc":            view.Doc,
		"Updated":        shortDate(view.Doc.UpdatedAt),
		"Body":           template.HTML(bodyHTML),
		"Tasks":          related,
		"Backlinks":      view.Backlinks,
		"DocTypes":       store.DocTypes,
		"ScratchpadJSON": template.JS(rawPad),
	})
}

// taskViewData assembles the template data for a task's read view (rendered
// body, column name, linked docs). Shared by handleTask and handleEditTask so
// the modal can swap straight back to an up-to-date read view after a save.
func taskViewData(view store.TaskView) (map[string]any, error) {
	bodyHTML, err := render.RenderMarkdownLinks(view.Task.Body, wikiResolver(view.Resolve))
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"Title":      view.Task.Title,
		"Task":       view.Task,
		"ColumnName": view.Task.Column,
		"Body":       template.HTML(bodyHTML),
		"Docs":       view.Docs,
	}, nil
}

func (srv *Server) handleTask(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	view, err := srv.store.TaskView(code)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := taskViewData(view)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("fragment") == "1" {
		data["Modal"] = true
		srv.renderFragment(w, "task", data)
		return
	}
	srv.render(w, r, "task", data)
}

func (srv *Server) handleArchiveDoc(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if err := srv.store.ArchiveDoc(code); errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) handleUnarchiveDoc(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if err := srv.store.UnarchiveDoc(code); errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) handleDeleteDoc(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	d, err := srv.store.GetDoc(code)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if d.Status != "archived" {
		http.Error(w, "archive the document before deleting it", http.StatusConflict)
		return
	}
	if err := srv.store.DeleteDoc(d.Code); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) handleArchiveTask(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if err := srv.store.ArchiveTask(code); errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) handleUnarchiveTask(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if err := srv.store.UnarchiveTask(code); errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	t, err := srv.store.GetTask(code)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if t.Status != "archived" {
		http.Error(w, "archive the card before deleting it", http.StatusConflict)
		return
	}
	if err := srv.store.DeleteTask(t.Code); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) handleMoveTask(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if _, err := srv.store.GetTask(code); errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req struct {
		Column string  `json:"column"`
		After  *string `json:"after"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	t, err := srv.store.MoveTask(code, req.Column, req.After)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// decodeJSON reads a small JSON request body into dst with unknown fields
// rejected, mirroring handleMoveTask. Returns false (after writing a 400) on a
// malformed body.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}

// writePreview renders body Markdown to HTML and writes it. Shared by the task
// and doc preview endpoints; resolves [[wiki-links]] so the preview matches the
// saved view (existing vs missing targets).
func (srv *Server) writePreview(w http.ResponseWriter, body string) {
	html, err := render.RenderMarkdownLinks(body, srv.wikiResolver())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// handlePreview renders posted body Markdown to HTML without touching the
// store. Shared by every create and edit form (the create forms have no code
// yet, so a per-code route does not fit).
func (srv *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Body string `json:"body"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	srv.writePreview(w, req.Body)
}

func (srv *Server) handleNewTaskForm(w http.ResponseWriter, r *http.Request) {
	srv.renderFragment(w, "tasknew", map[string]any{})
}

func (srv *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "title must not be empty", http.StatusBadRequest)
		return
	}
	// Empty column name => first column (To Do); code auto-generated from title.
	if _, err := srv.store.CreateTask(title, req.Body, ""); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (srv *Server) handleNewDocForm(w http.ResponseWriter, r *http.Request) {
	srv.render(w, r, "docnew", map[string]any{
		"Title":    "New document",
		"DocTypes": store.DocTypes,
	})
}

func (srv *Server) handleCreateDoc(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code  string `json:"code"`
		Title string `json:"title"`
		Type  string `json:"type"`
		Body  string `json:"body"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "title must not be empty", http.StatusBadRequest)
		return
	}
	// Pre-check distinguishes a duplicate (409) from a malformed code (400);
	// GetDoc normalizes and validates the code the same way CreateDoc does.
	// ponytail: benign TOCTOU race on localhost single-user; CreateDoc still guards.
	if _, err := srv.store.GetDoc(req.Code); err == nil {
		http.Error(w, "a document with that code already exists", http.StatusConflict)
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := srv.store.CreateDoc(req.Code, title, req.Type, req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"code": d.Code})
}

func (srv *Server) handleEditTask(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if _, err := srv.store.GetTask(code); errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "title must not be empty", http.StatusBadRequest)
		return
	}

	if _, err := srv.store.EditTask(code, &title, &req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	view, err := srv.store.TaskView(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := taskViewData(view)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data["Modal"] = true
	srv.renderFragment(w, "task", data)
}

func (srv *Server) handleEditDoc(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if _, err := srv.store.GetDoc(code); errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req struct {
		Title string `json:"title"`
		Type  string `json:"type"`
		Body  string `json:"body"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "title must not be empty", http.StatusBadRequest)
		return
	}

	// EditDoc validates the type against the known set and returns an error for
	// an unknown one, which we surface as a 400.
	if _, err := srv.store.EditDoc(code, &title, &req.Type, &req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
