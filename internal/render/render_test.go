package render

import (
	"strings"
	"testing"
)

func TestRenderHeadingAndParagraph(t *testing.T) {
	html, err := RenderMarkdown("# Title\n\nHello **world**.")
	if err != nil {
		t.Fatalf("RenderMarkdown: %v", err)
	}
	if !strings.Contains(html, "<h1") || !strings.Contains(html, "Title") {
		t.Errorf("missing heading: %s", html)
	}
	if !strings.Contains(html, "<strong>world</strong>") {
		t.Errorf("missing bold: %s", html)
	}
}

func TestRenderHeadingsHaveStableUniqueIDs(t *testing.T) {
	html, err := RenderMarkdown("## Overview\n\n### Details\n\n## Overview")
	if err != nil {
		t.Fatalf("RenderMarkdown: %v", err)
	}
	for _, want := range []string{
		`<h2 id="overview">Overview</h2>`,
		`<h3 id="details">Details</h3>`,
		`<h2 id="overview-1">Overview</h2>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered headings missing %q: %s", want, html)
		}
	}
}

func TestRenderEscapesRawHTML(t *testing.T) {
	html, err := RenderMarkdown("<script>alert(1)</script>")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "<script>") {
		t.Errorf("raw HTML should be escaped: %s", html)
	}
}

func TestRenderCodeBlockHighlights(t *testing.T) {
	src := "```go\nfunc main() {}\n```"
	html, err := RenderMarkdown(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "<pre") {
		t.Errorf("expected highlighted pre block: %s", html)
	}
}

func TestRenderAlerts(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantClass string
		wantBody  string
	}{
		{"note", "> [!NOTE]\n> Plain note.", `<blockquote class="alert alert-note">`, "Plain note."},
		{"tip", "> [!TIP]\n> Good news.", `<blockquote class="alert alert-tip">`, "Good news."},
		{"important", "> [!IMPORTANT]\n> Read me.", `<blockquote class="alert alert-important">`, "Read me."},
		{"warning", "> [!WARNING]\n> Watch **out**.", `<blockquote class="alert alert-warning">`, "<strong>out</strong>"},
		{"caution", "> [!CAUTION]\n> Danger.", `<blockquote class="alert alert-caution">`, "Danger."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := RenderMarkdown(tt.src)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(html, tt.wantClass) {
				t.Errorf("missing alert class %q in: %s", tt.wantClass, html)
			}
			if !strings.Contains(html, tt.wantBody) {
				t.Errorf("missing body %q in: %s", tt.wantBody, html)
			}
			if strings.Contains(html, "[!") {
				t.Errorf("marker leaked into output: %s", html)
			}
		})
	}
}

func TestRenderAlertMarkerOnlyBlockquote(t *testing.T) {
	html, err := RenderMarkdown("> [!NOTE]")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `<blockquote class="alert alert-note">`) {
		t.Errorf("expected alert class: %s", html)
	}
	if strings.Contains(html, "[!NOTE]") {
		t.Errorf("marker leaked into output: %s", html)
	}
}

func TestRenderPlainBlockquoteUntouched(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"no marker", "> Just a quote."},
		{"unknown type", "> [!BOGUS]\n> Body."},
		{"marker with trailing text", "> [!NOTE] inline extra\n> Body."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := RenderMarkdown(tt.src)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(html, "<blockquote>") {
				t.Errorf("expected plain blockquote: %s", html)
			}
			if strings.Contains(html, "alert") {
				t.Errorf("should not be an alert: %s", html)
			}
		})
	}
}

func TestRenderAlertInsideCodeFenceUntouched(t *testing.T) {
	src := "```\n> [!NOTE]\n> not an alert\n```"
	html, err := RenderMarkdown(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "[!NOTE]") {
		t.Errorf("fenced marker should render literally: %s", html)
	}
	if strings.Contains(html, "alert-note") {
		t.Errorf("fenced content must not become an alert: %s", html)
	}
}

func TestRenderLegacyDirectiveIsLiteralText(t *testing.T) {
	src := "::: callout warn\nOld content.\n:::"
	html, err := RenderMarkdown(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "::: callout warn") {
		t.Errorf("legacy directive should render as literal text: %s", html)
	}
	if strings.Contains(html, "<div class=\"callout") {
		t.Errorf("directive must not render as component: %s", html)
	}
}

func TestRenderFencedDirectiveExampleIntact(t *testing.T) {
	src := "Before.\n\n```\n::: callout\nexample\n:::\n```\n\nAfter."
	html, err := RenderMarkdown(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Before.", "::: callout", "example", "After."} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in: %s", want, html)
		}
	}
}
