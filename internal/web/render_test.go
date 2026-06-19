package web

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// failOnExecute renders some output and then errors, simulating a template
// whose execution fails partway through.
var failOnExecute = template.Must(template.New("x").Parse(
	`{{define "layout"}}BEFORE-{{call .Fail}}-AFTER{{end}}`))

func TestExecuteDoesNotWritePartialBodyOnError(t *testing.T) {
	srv := &Server{pages: map[string]*template.Template{"boom": failOnExecute}}
	rec := httptest.NewRecorder()
	data := map[string]any{"Fail": func() (string, error) { return "", fmt.Errorf("kaboom") }}

	srv.execute(rec, "boom", "layout", data)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "BEFORE") {
		t.Errorf("partial body leaked before the error: %q", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "render:") {
		t.Errorf("body = %q, want a render error message", rec.Body.String())
	}
}

func TestExecuteUnknownPage(t *testing.T) {
	srv := &Server{pages: map[string]*template.Template{}}
	rec := httptest.NewRecorder()
	srv.execute(rec, "nope", "layout", nil)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestSameOriginOnly(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := sameOriginOnly(next)

	cases := []struct {
		name   string
		method string
		origin string
		want   int
	}{
		{"cross-origin POST rejected", http.MethodPost, "http://evil.example", http.StatusForbidden},
		{"same-origin POST allowed", http.MethodPost, "http://127.0.0.1:7777", http.StatusOK},
		{"no-origin POST allowed", http.MethodPost, "", http.StatusOK},
		{"cross-origin GET allowed", http.MethodGet, "http://evil.example", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://127.0.0.1:7777/task/T-1/archive", nil)
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestCrossOriginMutationRejectedEndToEnd(t *testing.T) {
	srv, s := newTestServer(t)
	s.CreateTaskWithCode("1", "Guard me", "", "")

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/task/T-1/archive", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://evil.example")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("cross-origin archive status = %d, want 403", resp.StatusCode)
	}
	// The task must remain active — the mutation was blocked.
	tk, err := s.GetTask("T-1")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Status != "active" {
		t.Errorf("task status = %q, want active (mutation should be blocked)", tk.Status)
	}
}
