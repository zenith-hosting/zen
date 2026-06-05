package zen

import (
	"strings"
	"testing"
)

func TestRenderDocumentIncludesSSRHTMLAndHydrationData(t *testing.T) {
	doc := renderDocument(documentInput{
		Title:         "Home",
		AppElementID:  "app",
		DataElementID: "__ZEN_DATA__",
		HTML:          `<main><h1>Hello</h1></main>`,
		HydrationJSON: `{"page":"Home","props":{"title":"Hello"}}`,
		Styles:        []string{"/assets/app.css"},
		Scripts:       []string{"/assets/entry-client.js"},
		DevScripts:    nil,
	})

	for _, want := range []string{
		"<!doctype html>",
		"<title>Home</title>",
		`<link rel="stylesheet" href="/assets/app.css">`,
		`<div id="app"><main><h1>Hello</h1></main></div>`,
		`<script id="__ZEN_DATA__" type="application/json">{"page":"Home","props":{"title":"Hello"}}</script>`,
		`<script type="module" src="/assets/entry-client.js"></script>`,
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("document missing %q\n%s", want, doc)
		}
	}
}

func TestRenderDocumentEscapesTitle(t *testing.T) {
	doc := renderDocument(documentInput{
		Title:         `<script>alert(1)</script>`,
		AppElementID:  "app",
		DataElementID: "__ZEN_DATA__",
		HTML:          "",
		HydrationJSON: `{"page":"Home","props":{}}`,
	})

	if strings.Contains(doc, "<title><script>") {
		t.Fatalf("title was not escaped: %s", doc)
	}
	if !strings.Contains(doc, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("escaped title missing: %s", doc)
	}
}
