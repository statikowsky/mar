package convert

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func HTMLToMarkdown(source string) (string, error) {
	root, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}
	var c converter
	c.block(root)
	return strings.TrimSpace(c.out.String()) + "\n", nil
}

type converter struct {
	out strings.Builder
}

// DocumentTitle returns the <title> text, falling back to the first <h1>, or "".
func DocumentTitle(source string) string {
	root, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return ""
	}
	var title, h1 string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.DataAtom == atom.Title && title == "" {
				title = strings.TrimSpace(textOf(n))
			}
			if n.DataAtom == atom.H1 && h1 == "" {
				h1 = strings.TrimSpace(textOf(n))
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(root)
	if title != "" {
		return title
	}
	return h1
}

func classOf(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "class" {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, name string) bool {
	for _, f := range strings.Fields(classOf(n)) {
		if f == name {
			return true
		}
	}
	return false
}

func (c *converter) writeBlock(s string) {
	s = strings.Trim(s, "\n")
	if s == "" {
		return
	}
	if c.out.Len() > 0 {
		c.out.WriteString("\n\n")
	}
	c.out.WriteString(s)
}

// block walks children of n, emitting block-level Markdown. Consecutive loose
// inline children (text or inline elements not wrapped in a block) are gathered
// and flushed as a single paragraph.
func (c *converter) block(n *html.Node) {
	var inlineRun strings.Builder
	flush := func() {
		if s := collapseSpaces(inlineRun.String()); s != "" {
			c.writeBlock(s)
		}
		inlineRun.Reset()
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if isInlineContent(child) {
			inlineRun.WriteString(inline(child))
			continue
		}
		flush()
		c.blockNode(child)
	}
	flush()
}

// isInlineContent reports whether a node should be treated as inline flow
// (text, or an inline element) rather than a block.
func isInlineContent(n *html.Node) bool {
	switch n.Type {
	case html.TextNode:
		return true
	case html.ElementNode:
		switch n.DataAtom {
		case atom.Strong, atom.B, atom.Em, atom.I, atom.Code, atom.A, atom.Span, atom.Br, atom.Sub, atom.Sup, atom.Small:
			return true
		}
	}
	return false
}

func (c *converter) blockNode(n *html.Node) {
	switch n.Type {
	case html.ElementNode:
		switch n.DataAtom {
		case atom.Script, atom.Style, atom.Head, atom.Title:
			return
		case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
			level := int(n.Data[1] - '0')
			c.writeBlock(strings.Repeat("#", level) + " " + c.inlineChildren(n))
			return
		case atom.P:
			c.writeBlock(c.inlineChildren(n))
			return
		case atom.Ul:
			c.writeBlock(c.list(n, false))
			return
		case atom.Ol:
			c.writeBlock(c.list(n, true))
			return
		case atom.Pre:
			c.writeBlock(c.codeBlock(n))
			return
		case atom.Table:
			c.writeBlock(c.table(n))
			return
		case atom.Blockquote:
			c.writeBlock("> " + strings.TrimSpace(c.inlineChildren(n)))
			return
		case atom.Hr:
			c.writeBlock("---")
			return
		case atom.Div:
			if directive := c.directive(n); directive != "" {
				c.writeBlock(directive)
				return
			}
			c.block(n) // transparent wrapper
			return
		}
	}
	// Default: descend into anything else (body, html, sections, unknown tags).
	if n.Type == html.ElementNode || n.Type == html.DocumentNode {
		c.block(n)
	}
}

// directive reverse-maps reports-style component divs to standard Markdown:
// callout to a GFM alert blockquote, card and phase-block to a heading plus
// body. Returns "" if n is not a recognized component.
func (c *converter) directive(n *html.Node) string {
	switch {
	case hasClass(n, "callout"):
		marker := "[!NOTE]"
		for _, f := range strings.Fields(classOf(n)) {
			switch f {
			case "good":
				marker = "[!TIP]"
			case "warn":
				marker = "[!WARNING]"
			}
		}
		return quoteAlert(marker, strings.TrimSpace(c.innerMarkdown(n)))
	case hasClass(n, "card"):
		title := firstHeadingText(n)
		body := strings.TrimSpace(c.innerMarkdownSkipHeading(n))
		if title == "" {
			return body
		}
		return "#### " + title + "\n\n" + body
	case hasClass(n, "phase-block"):
		title := phaseTitle(n)
		if title == "" {
			title = "P1"
			for _, f := range strings.Fields(classOf(n)) {
				if strings.HasPrefix(f, "p") && len(f) == 2 || f == "skip" {
					title = strings.ToUpper(f)
				}
			}
		}
		body := strings.TrimSpace(c.innerMarkdownSkipPhaseHeader(n))
		return "### " + title + "\n\n" + body
	}
	return ""
}

func quoteAlert(marker, content string) string {
	quoted := []string{"> " + marker}
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "" {
			quoted = append(quoted, ">")
			continue
		}
		quoted = append(quoted, "> "+line)
	}
	return strings.Join(quoted, "\n")
}

// innerMarkdown converts a node's children as block content into Markdown.
func (c *converter) innerMarkdown(n *html.Node) string {
	var sub converter
	sub.block(n)
	return sub.out.String()
}

func (c *converter) innerMarkdownSkipHeading(n *html.Node) string {
	var sub converter
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && isHeading(child) {
			continue
		}
		sub.blockNode(child)
	}
	return sub.out.String()
}

func (c *converter) innerMarkdownSkipPhaseHeader(n *html.Node) string {
	var sub converter
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.DataAtom == atom.Div && hasClass(child, "phase-header") {
			continue
		}
		sub.blockNode(child)
	}
	return sub.out.String()
}

func isHeading(n *html.Node) bool {
	switch n.DataAtom {
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		return true
	}
	return false
}

func firstHeadingText(n *html.Node) string {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && isHeading(child) {
			return strings.TrimSpace(textOf(child))
		}
	}
	return ""
}

// phaseTitle finds the <h3> inside a phase-header (skipping the pill span).
func phaseTitle(n *html.Node) string {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.DataAtom == atom.Div && hasClass(child, "phase-header") {
			for h := child.FirstChild; h != nil; h = h.NextSibling {
				if h.Type == html.ElementNode && isHeading(h) {
					return strings.TrimSpace(textOf(h))
				}
			}
		}
	}
	return ""
}

func (c *converter) list(n *html.Node, ordered bool) string {
	var b strings.Builder
	i := 1
	for li := n.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.DataAtom != atom.Li {
			continue
		}
		marker := "- "
		if ordered {
			marker = fmt.Sprintf("%d. ", i)
		}
		b.WriteString(marker + strings.TrimSpace(c.inlineChildren(li)) + "\n")
		i++
	}
	return b.String()
}

func (c *converter) codeBlock(n *html.Node) string {
	code := textOf(n)
	code = strings.TrimRight(code, "\n")
	return "```\n" + code + "\n```"
}

func (c *converter) table(n *html.Node) string {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		for ch := node.FirstChild; ch != nil; ch = ch.NextSibling {
			if ch.Type == html.ElementNode && ch.DataAtom == atom.Tr {
				var cells []string
				for cell := ch.FirstChild; cell != nil; cell = cell.NextSibling {
					if cell.Type == html.ElementNode && (cell.DataAtom == atom.Td || cell.DataAtom == atom.Th) {
						cells = append(cells, strings.TrimSpace(c.inlineChildren(cell)))
					}
				}
				if len(cells) > 0 {
					rows = append(rows, cells)
				}
			} else {
				walk(ch)
			}
		}
	}
	walk(n)
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("| " + strings.Join(rows[0], " | ") + " |\n")
	sep := make([]string, len(rows[0]))
	for i := range sep {
		sep[i] = "---"
	}
	b.WriteString("| " + strings.Join(sep, " | ") + " |\n")
	for _, r := range rows[1:] {
		b.WriteString("| " + strings.Join(r, " | ") + " |\n")
	}
	return b.String()
}

// inlineChildren renders a node's children as inline Markdown (no block breaks).
func (c *converter) inlineChildren(n *html.Node) string {
	var b strings.Builder
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		b.WriteString(inline(child))
	}
	return collapseSpaces(b.String())
}

func inline(n *html.Node) string {
	switch n.Type {
	case html.TextNode:
		return n.Data
	case html.ElementNode:
		switch n.DataAtom {
		case atom.Strong, atom.B:
			return "**" + inlineOf(n) + "**"
		case atom.Em, atom.I:
			return "*" + inlineOf(n) + "*"
		case atom.Code:
			return "`" + textOf(n) + "`"
		case atom.A:
			return "[" + inlineOf(n) + "](" + attr(n, "href") + ")"
		case atom.Br:
			return "\n"
		case atom.Script, atom.Style:
			return ""
		default:
			return inlineOf(n)
		}
	}
	return ""
}

func inlineOf(n *html.Node) string {
	var b strings.Builder
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		b.WriteString(inline(child))
	}
	return b.String()
}

func textOf(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for ch := node.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(n)
	return b.String()
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func collapseSpaces(s string) string {
	var b strings.Builder
	var prevSpace bool
	for _, r := range s {
		if r == '\n' {
			b.WriteRune(r)
			prevSpace = false
			continue
		}
		if r == ' ' || r == '\t' {
			if !prevSpace {
				b.WriteRune(' ')
			}
			prevSpace = true
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return strings.TrimSpace(b.String())
}
