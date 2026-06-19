package convert

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdownBasics(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"heading", "<h2>Title</h2>", "## Title"},
		{"paragraph", "<p>Hello world.</p>", "Hello world."},
		{"bold", "<p>a <strong>b</strong> c</p>", "a **b** c"},
		{"italic", "<p>a <em>b</em> c</p>", "a *b* c"},
		{"inline code", "<p>run <code>go test</code></p>", "run `go test`"},
		{"link", `<p><a href="https://x.io">x</a></p>`, "[x](https://x.io)"},
		{"unordered list", "<ul><li>one</li><li>two</li></ul>", "- one\n- two"},
		{"ordered list", "<ol><li>one</li><li>two</li></ol>", "1. one\n2. two"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTMLToMarkdown(tt.html)
			if err != nil {
				t.Fatalf("HTMLToMarkdown: %v", err)
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("got %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

func TestHTMLToMarkdownCodeBlock(t *testing.T) {
	got, err := HTMLToMarkdown("<pre><code>func main() {}\n</code></pre>")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "```") || !strings.Contains(got, "func main() {}") {
		t.Errorf("code block not fenced: %q", got)
	}
}

func TestHTMLToMarkdownTable(t *testing.T) {
	html := "<table><tr><th>A</th><th>B</th></tr><tr><td>1</td><td>2</td></tr></table>"
	got, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"| A | B |", "| --- | --- |", "| 1 | 2 |"} {
		if !strings.Contains(got, want) {
			t.Errorf("table missing %q in:\n%s", want, got)
		}
	}
}

func TestHTMLToMarkdownCalloutBecomesAlert(t *testing.T) {
	got, err := HTMLToMarkdown(`<div class="callout warn"><p>Watch <strong>out</strong>.</p></div>`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "> [!WARNING]") {
		t.Errorf("callout not mapped to alert: %q", got)
	}
	if !strings.Contains(got, "> Watch **out**.") {
		t.Errorf("callout inner markdown not quoted: %q", got)
	}
}

func TestHTMLToMarkdownCardBecomesHeading(t *testing.T) {
	got, err := HTMLToMarkdown(`<div class="card"><h4>My card</h4><p>Body.</p></div>`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "#### My card") {
		t.Errorf("card not mapped to heading: %q", got)
	}
	if !strings.Contains(got, "Body.") {
		t.Errorf("card body missing: %q", got)
	}
	if strings.Contains(got, ":::") {
		t.Errorf("directive syntax must not be emitted: %q", got)
	}
}

func TestHTMLToMarkdownPhaseBlockBecomesHeading(t *testing.T) {
	html := `<div class="phase-block p2"><div class="phase-header"><span class="pill p2">P2</span><h3>Phase 2 — Next</h3></div><ul><li>step</li></ul></div>`
	got, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "### Phase 2 — Next") {
		t.Errorf("phase-block not mapped to heading: %q", got)
	}
	if !strings.Contains(got, "- step") {
		t.Errorf("phase-block inner list missing: %q", got)
	}
	// The synthesized pill must NOT leak into the body as literal text.
	if strings.Contains(got, "P2P2") || strings.Contains(got, "<span") {
		t.Errorf("pill markup leaked into output: %q", got)
	}
}

func TestHTMLToMarkdownLooseInlineText(t *testing.T) {
	// Inline content directly inside a block (no <p> wrapper) must be kept.
	got, err := HTMLToMarkdown(`<div class="callout"><strong>TL;DR.</strong> Auth belongs in <code>go-store</code>.</div>`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "**TL;DR.**") || !strings.Contains(got, "Auth belongs in `go-store`.") {
		t.Errorf("loose inline content dropped: %q", got)
	}
}

func TestHTMLToMarkdownStripsScriptStyle(t *testing.T) {
	got, err := HTMLToMarkdown("<style>.x{}</style><p>keep</p><script>alert(1)</script>")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "alert") || strings.Contains(got, ".x{}") {
		t.Errorf("script/style not stripped: %q", got)
	}
	if !strings.Contains(got, "keep") {
		t.Errorf("body content dropped: %q", got)
	}
}

func TestDocumentTitle(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"title tag", "<html><head><title>My Report — sub</title></head><body><h1>H</h1></body></html>", "My Report — sub"},
		{"falls back to h1", "<body><h1>Heading One</h1><p>x</p></body>", "Heading One"},
		{"empty when neither", "<body><p>x</p></body>", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DocumentTitle(tt.html); got != tt.want {
				t.Errorf("DocumentTitle = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTMLToMarkdownFullDocument(t *testing.T) {
	html := `<!DOCTYPE html><html><head><title>T</title><style>x</style></head>
<body><h1>Doc</h1><p>Intro.</p></body></html>`
	got, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "# Doc") || !strings.Contains(got, "Intro.") {
		t.Errorf("full document not converted: %q", got)
	}
	if strings.Contains(got, "<title>") {
		t.Errorf("head markup leaked: %q", got)
	}
}
