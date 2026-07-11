package render

import (
	"bytes"
	"fmt"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(chromahtml.WithClasses(false)),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
		parser.WithASTTransformers(util.Prioritized(alertTransformer{}, 100)),
		// Before the standard link parser (priority 200) so [[CODE]] wins over [text](url).
		parser.WithInlineParsers(util.Prioritized(wikiLinkParser{}, 100)),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		renderer.WithNodeRenderers(util.Prioritized(wikiLinkRenderer{}, 100)),
	),
)

func RenderMarkdown(source string) (string, error) {
	return RenderMarkdownLinks(source, nil)
}

// RenderMarkdownLinks renders source with resolve mapping [[CODE]] wiki-links to
// URLs and existence (for "red link" styling). A nil resolve renders links with
// conventional URLs and never marks any missing.
func RenderMarkdownLinks(source string, resolve Resolver) (string, error) {
	var buf bytes.Buffer
	ctx := parser.NewContext()
	ctx.Set(resolverKey, resolve)
	if err := md.Convert([]byte(source), &buf, parser.WithContext(ctx)); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return buf.String(), nil
}
