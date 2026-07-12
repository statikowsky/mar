package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/statikowsky/mar/internal/store"
)

func (srv *Server) handleScratchpad(w http.ResponseWriter, r *http.Request) {
	pad, err := srv.store.Scratchpad()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	raw, err := json.Marshal(pad)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	srv.render(w, r, "scratchpad", map[string]any{
		"Title":          "Scratchpad",
		"BodyClass":      "scratch-page",
		"ScratchpadJSON": template.JS(raw),
	})
}

func (srv *Server) handleScratchpadData(w http.ResponseWriter, _ *http.Request) {
	pad, err := srv.store.Scratchpad()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusOK, pad)
}

func (srv *Server) handleCreateScratchNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text  string                `json:"text"`
		X     int                   `json:"x"`
		Y     int                   `json:"y"`
		Width int                   `json:"width"`
		Color string                `json:"color"`
		Docs  []store.ScratchDocRef `json:"docs"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	codes := make([]string, len(req.Docs))
	for i, ref := range req.Docs {
		codes[i] = ref.Code
	}
	if err := srv.store.ValidateDocCodes(codes); err != nil {
		ref := ""
		var validation *store.DocValidationError
		if errors.As(err, &validation) {
			ref = validation.Code
		}
		http.Error(w, fmt.Sprintf("scratch document %s: %v", ref, err), http.StatusBadRequest)
		return
	}
	note, err := srv.store.CreateScratchNote(req.Text, req.X, req.Y, req.Width, req.Color)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Docs) > 0 {
		note.Docs = req.Docs
		if _, err := srv.store.UpdateScratchNote(note); err != nil {
			_ = srv.store.DeleteScratchNote(note.ID)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	pad, err := srv.store.Scratchpad()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusCreated, pad)
}

func (srv *Server) handleSaveScratchpad(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Revision int64               `json:"revision"`
		Notes    []store.ScratchNote `json:"notes"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	pad, err := srv.store.SaveScratchpad(req.Revision, req.Notes)
	if errors.Is(err, store.ErrConflict) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSONResponse(w, http.StatusOK, pad)
}

func (srv *Server) handlePromoteScratchNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Kind string `json:"kind"`
		Code string `json:"code"`
		Type string `json:"type"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	pad, err := srv.store.Scratchpad()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var note *store.ScratchNote
	for i := range pad.Notes {
		if strings.EqualFold(pad.Notes[i].ID, r.PathValue("id")) {
			note = &pad.Notes[i]
			break
		}
	}
	if note == nil {
		http.NotFound(w, r)
		return
	}
	title, body := scratchTitleBody(note.Text)
	var code string
	switch req.Kind {
	case "task":
		task, err := srv.store.CreateTask(title, body, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		code = task.Code
	case "doc":
		doc, err := srv.store.CreateDoc(req.Code, title, req.Type, body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		code = doc.Code
	default:
		http.Error(w, "kind must be task or doc", http.StatusBadRequest)
		return
	}
	note.Link = code
	if _, err := srv.store.UpdateScratchNote(*note); err != nil {
		http.Error(w, fmt.Sprintf("created %s but could not link scratch note: %v", code, err), http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusCreated, map[string]string{"code": code})
}

func scratchTitleBody(text string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(text), "\n", 2)
	title := strings.TrimSpace(parts[0])
	if len(parts) == 1 {
		return title, ""
	}
	return title, strings.TrimSpace(parts[1])
}

func writeJSONResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
