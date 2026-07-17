package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/statikowsky/mar/internal/store"
	"github.com/statikowsky/mar/internal/version"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

type Server struct {
	store       store.Interface
	repo        string
	projectPath string
	projectName string
	pages       map[string]*template.Template
}

// pageNames are the content templates rendered into the layout. Each page file
// defines "content"; parsing them separately (over a shared base) keeps those
// definitions from colliding.
var pageNames = []string{"index", "board", "doc", "task", "docnew", "tasknew", "scratchpad"}

func NewServer(s store.Interface, repo, projectPath string) *Server {
	base := template.Must(template.ParseFS(templateFS, "templates/*.gohtml"))
	pages := make(map[string]*template.Template, len(pageNames))
	for _, name := range pageNames {
		pt := template.Must(base.Clone())
		template.Must(pt.ParseFS(templateFS, "templates/"+name+".gohtml"))
		pages[name] = pt
	}
	return &Server{store: s, repo: repo, projectPath: projectPath, projectName: filepath.Base(projectPath), pages: pages}
}

func (srv *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", srv.handleIndex)
	mux.HandleFunc("GET /board", srv.handleBoard)
	mux.HandleFunc("GET /scratchpad", srv.handleScratchpad)
	mux.HandleFunc("GET /scratchpad/data", srv.handleScratchpadData)
	mux.HandleFunc("POST /scratchpad/note", srv.handleCreateScratchNote)
	mux.HandleFunc("PUT /scratchpad", srv.handleSaveScratchpad)
	mux.HandleFunc("POST /scratchpad/note/{id}/promote", srv.handlePromoteScratchNote)
	mux.HandleFunc("GET /doc/new", srv.handleNewDocForm)
	mux.HandleFunc("GET /doc/{code}", srv.handleDoc)
	mux.HandleFunc("GET /doc/{path...}", srv.handleDocAsset)
	mux.HandleFunc("GET /task/new", srv.handleNewTaskForm)
	mux.HandleFunc("GET /task/{code}", srv.handleTask)
	mux.HandleFunc("POST /task", srv.handleCreateTask)
	mux.HandleFunc("POST /doc", srv.handleCreateDoc)
	mux.HandleFunc("POST /preview", srv.handlePreview)
	mux.HandleFunc("POST /task/{code}/move", srv.handleMoveTask)
	mux.HandleFunc("POST /task/{code}/edit", srv.handleEditTask)
	mux.HandleFunc("POST /task/{code}/archive", srv.handleArchiveTask)
	mux.HandleFunc("POST /task/{code}/unarchive", srv.handleUnarchiveTask)
	mux.HandleFunc("POST /task/{code}/delete", srv.handleDeleteTask)
	mux.HandleFunc("POST /doc/{code}/edit", srv.handleEditDoc)
	mux.HandleFunc("POST /doc/{code}/archive", srv.handleArchiveDoc)
	mux.HandleFunc("POST /doc/{code}/unarchive", srv.handleUnarchiveDoc)
	mux.HandleFunc("POST /doc/{code}/delete", srv.handleDeleteDoc)
	mux.HandleFunc("GET /events/version", srv.handleVersion)
	mux.HandleFunc("GET /favicon.ico", srv.handleFavicon)
	mux.Handle("GET /static/", http.FileServer(http.FS(staticFS)))
	return sameOriginOnly(mux)
}

// sameOriginOnly rejects cross-origin state-changing requests. mar serves an
// unauthenticated browser UI on localhost, so without this a malicious page in
// the same browser could POST to a mutating endpoint (CSRF / DNS-rebinding).
// Requests with no Origin header (CLI tools, same-origin navigations) are
// allowed; only a present, mismatched Origin is rejected.
func sameOriginOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if origin := r.Header.Get("Origin"); origin != "" {
				if u, err := url.Parse(origin); err != nil || u.Host != r.Host {
					http.Error(w, "cross-origin request forbidden", http.StatusForbidden)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (srv *Server) ServeListener(ln net.Listener) error {
	httpSrv := &http.Server{
		Handler:      srv.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return httpSrv.Serve(ln)
}

func (srv *Server) render(w http.ResponseWriter, r *http.Request, page string, data map[string]any) {
	if _, ok := data["ProjectName"]; !ok {
		data["ProjectName"] = srv.projectName
	}
	if _, ok := data["Version"]; !ok {
		data["Version"] = version.Display()
	}
	if _, ok := data["Theme"]; !ok {
		data["Theme"] = themeFromRequest(r)
	}
	srv.execute(w, page, "layout", data)
}

// execute renders the named template of a prebuilt page into a buffer before
// writing it, so a mid-render failure becomes a clean 500 rather than a
// partial 200 body with error text appended.
func (srv *Server) execute(w http.ResponseWriter, page, name string, data map[string]any) {
	pt, ok := srv.pages[page]
	if !ok {
		log.Printf("render: unknown page %q", page)
		http.Error(w, "render: unknown page", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := pt.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("render %s/%s: %v", page, name, err)
		http.Error(w, fmt.Sprintf("render: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// knownThemes is the set of scheme ids that may be rendered as data-theme.
// "light" and "dark" (Catppuccin Mocha) are the originals; the rest are the
// added terminal color schemes. Keep this in sync with the data-theme blocks in
// static/style.css and the THEMES list in templates/layout.gohtml.
var knownThemes = map[string]bool{
	"light":           true,
	"dark":            true,
	"github-light":    true,
	"gruvbox-light":   true,
	"gruvbox-dark":    true,
	"solarized-light": true,
	"solarized-dark":  true,
	"dracula":         true,
	"nord":            true,
}

// themeFromRequest returns the visitor's explicitly chosen scheme id if the
// theme cookie names a known scheme, or "" to mean follow the OS preference
// (resolved client-side).
func themeFromRequest(r *http.Request) string {
	c, err := r.Cookie("theme")
	if err != nil {
		return ""
	}
	if knownThemes[c.Value] {
		return c.Value
	}
	return ""
}

func (srv *Server) renderFragment(w http.ResponseWriter, page string, data map[string]any) {
	srv.execute(w, page, "content", data)
}
