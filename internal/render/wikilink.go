package render

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Resolver maps a wiki-link target code to its URL and whether it exists. A nil
// resolver (plain RenderMarkdown) renders every link with a conventional URL
// and never marks one missing.
type Resolver func(code string) (url string, exists bool)

var resolverKey = parser.NewContextKey()

// wikiLink is an inline [[CODE]] / [[CODE|label]] reference. URL/Missing are
// filled in at parse time from the Resolver in the parser context, so the
// renderer stays a pure leaf write.
type wikiLink struct {
	ast.BaseInline
	Code, Label, URL string
	Missing          bool
}

var kindWikiLink = ast.NewNodeKind("WikiLink")

func (n *wikiLink) Kind() ast.NodeKind { return kindWikiLink }

func (n *wikiLink) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Code": n.Code, "Label": n.Label}, nil)
}

// wikiLinkParser parses [[CODE]] and [[CODE|label]]. Registered before the
// standard link parser; returning nil falls through to it for normal [text](url).
type wikiLinkParser struct{}

func (wikiLinkParser) Trigger() []byte { return []byte{'['} }

func (wikiLinkParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 5 || line[1] != '[' {
		return nil
	}
	end := bytes.Index(line, []byte("]]"))
	if end < 3 {
		return nil
	}
	inner := line[2:end]
	if bytes.ContainsAny(inner, "[]\n") {
		return nil
	}
	code := inner
	var label []byte
	if i := bytes.IndexByte(inner, '|'); i >= 0 {
		code, label = inner[:i], inner[i+1:]
	}
	code = bytes.TrimSpace(code)
	if len(code) == 0 {
		return nil
	}
	block.Advance(end + 2)

	n := &wikiLink{Code: string(code), Label: string(bytes.TrimSpace(label))}
	if n.Label == "" {
		n.Label = n.Code
	}
	if resolve, ok := pc.Get(resolverKey).(Resolver); ok && resolve != nil {
		if url, exists := resolve(n.Code); exists {
			n.URL = url
		} else {
			n.Missing = true
		}
	}
	return n
}

// defaultURL is the fallback target for a wiki-link the resolver did not place
// (nil resolver, or a missing target rendered as a navigable "red link").
func defaultURL(code string) string {
	if strings.HasPrefix(strings.ToUpper(code), "T-") {
		return "/task/" + code
	}
	return "/doc/" + code
}

type wikiLinkRenderer struct{}

func (wikiLinkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindWikiLink, renderWikiLink)
}

func renderWikiLink(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*wikiLink)
	url := n.URL
	if url == "" {
		url = defaultURL(n.Code)
	}
	class := "wikilink"
	if n.Missing {
		class = "wikilink wikilink-missing"
	}
	_, _ = w.WriteString(`<a href="`)
	_, _ = w.Write(util.EscapeHTML(util.URLEscape([]byte(url), false)))
	_, _ = w.WriteString(`" class="` + class + `">`)
	_, _ = w.Write(util.EscapeHTML([]byte(n.Label)))
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

// Refs returns the raw target codes of every wiki-link in source, in document
// order with duplicates kept. Parsing (not regex) makes it fence- and
// code-span-safe: [[X]] inside `code` or a fenced block is not a reference.
func Refs(source string) []string {
	doc := md.Parser().Parse(text.NewReader([]byte(source)))
	var refs []string
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if wl, ok := n.(*wikiLink); ok {
				refs = append(refs, wl.Code)
			}
		}
		return ast.WalkContinue, nil
	})
	return refs
}
