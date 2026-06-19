package render

import (
	"reflect"
	"strings"
	"testing"
)

func TestWikiLinkResolved(t *testing.T) {
	resolve := func(code string) (string, bool) {
		if code == "AUTH" {
			return "/doc/DOC-AUTH", true
		}
		return "", false
	}
	html, err := RenderMarkdownLinks("See [[AUTH]] and [[GHOST]].", resolve)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `<a href="/doc/DOC-AUTH" class="wikilink">AUTH</a>`) {
		t.Errorf("resolved link wrong: %s", html)
	}
	// Missing target: red link, still navigable to the conventional URL.
	if !strings.Contains(html, `class="wikilink wikilink-missing">GHOST</a>`) {
		t.Errorf("missing link wrong: %s", html)
	}
}

func TestWikiLinkLabelAndTask(t *testing.T) {
	html, err := RenderMarkdown("[[T-5|the fix]]")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `href="/task/T-5"`) || !strings.Contains(html, `>the fix</a>`) {
		t.Errorf("label/task link wrong: %s", html)
	}
}

func TestWikiLinkNotInCodeSpanOrFence(t *testing.T) {
	for _, src := range []string{"`[[AUTH]]`", "```\n[[AUTH]]\n```"} {
		html, err := RenderMarkdown(src)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(html, "wikilink") {
			t.Errorf("code content should not become a wiki-link: %s", html)
		}
		if !strings.Contains(html, "[[AUTH]]") {
			t.Errorf("literal [[AUTH]] should survive: %s", html)
		}
	}
}

func TestWikiLinkFallsThroughToNormalLink(t *testing.T) {
	html, err := RenderMarkdown("[text](http://example.com)")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `href="http://example.com"`) || strings.Contains(html, "wikilink") {
		t.Errorf("normal link broken: %s", html)
	}
}

func TestRefs(t *testing.T) {
	got := Refs("[[A]] then [[B|x]] and `[[C]]` then [[A]]")
	want := []string{"A", "B", "A"} // code-span [[C]] excluded; duplicates kept
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Refs = %v, want %v", got, want)
	}
}
