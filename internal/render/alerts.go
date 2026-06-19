package render

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var alertVariants = map[string]string{
	"[!NOTE]":      "note",
	"[!TIP]":       "tip",
	"[!IMPORTANT]": "important",
	"[!WARNING]":   "warning",
	"[!CAUTION]":   "caution",
}

// alertTransformer turns a blockquote whose first line is exactly a GFM alert
// marker ([!NOTE], [!WARNING], ...) into <blockquote class="alert alert-...">
// with the marker line removed.
type alertTransformer struct{}

func (alertTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	var quotes []*ast.Blockquote
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if bq, ok := n.(*ast.Blockquote); ok && entering {
			quotes = append(quotes, bq)
		}
		return ast.WalkContinue, nil
	})
	for _, bq := range quotes {
		para, ok := bq.FirstChild().(*ast.Paragraph)
		if !ok || para.Lines().Len() == 0 {
			continue
		}
		first := para.Lines().At(0)
		variant, ok := alertVariants[strings.TrimSpace(string(first.Value(source)))]
		if !ok {
			continue
		}
		stripMarkerLine(para, first)
		if para.ChildCount() == 0 {
			bq.RemoveChild(bq, para)
		}
		bq.SetAttributeString("class", []byte("alert alert-"+variant))
	}
}

func stripMarkerLine(para *ast.Paragraph, line text.Segment) {
	child := para.FirstChild()
	for child != nil {
		next := child.NextSibling()
		t, ok := child.(*ast.Text)
		if !ok || t.Segment.Start >= line.Stop {
			break
		}
		para.RemoveChild(para, child)
		child = next
	}
}
