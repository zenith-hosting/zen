package zen

import (
	"strings"
	"testing"
)

func TestRenderDocument(t *testing.T) {
	doc := renderDocument(documentInput{
		Title:            "Home <unsafe>",
		HeadHTML:         `<meta name="description" content="Hello">`,
		BaseElements:     []headElement{newHeadElement("base", Href("/"))},
		MetaElements:     []headElement{newHeadElement("meta", Property("og:title"), Content("Home <unsafe>"))},
		LinkElements:     []headElement{newHeadElement("link", Rel("canonical"), Href("https://example.com/?q=<unsafe>"))},
		HeadScripts:      []headElement{newHeadElement("script", Type("application/ld+json"), Text(`{"name":"Home"}`))},
		CustomCSS:        `body { color: red; }`,
		AppHTML:          `<main><h1>Hello</h1></main>`,
		HydrationJSON:    `{"page":"Home","props":{"title":"Hello"}}`,
		StylesheetURLs:   []string{"/assets/app.css"},
		ModuleScriptURLs: []string{"http://localhost:5173/@vite/client", "/assets/entry-client.js"},
	})

	for _, want := range []string{
		`<title>Home &lt;unsafe&gt;</title>`,
		`<meta name="description" content="Hello">`,
		`<base href="/">`,
		`<meta property="og:title" content="Home &lt;unsafe&gt;">`,
		`<link rel="canonical" href="https://example.com/?q=&lt;unsafe&gt;">`,
		`<style>body { color: red; }</style>`,
		`<script type="application/ld+json">{"name":"Home"}</script>`,
		`<link rel="stylesheet" href="/assets/app.css">`,
		`<div id="app"><main><h1>Hello</h1></main></div>`,
		`<script id="__ZEN_DATA__" type="application/json">{"page":"Home","props":{"title":"Hello"}}</script>`,
		`<script type="module" src="http://localhost:5173/@vite/client"></script>`,
		`<script type="module" src="/assets/entry-client.js"></script>`,
		`<body>`,
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("document missing %q\n%s", want, doc)
		}
	}
}

func TestHeadElementTagsEscapesAttributesAndSkipsUnsafeNames(t *testing.T) {
	html := headElementTags([]headElement{
		newHeadElement("meta",
			Attr("data-value", `A "quoted" <value>`),
			Attr(`bad name="x`, "ignored"),
		),
	})

	want := `<meta data-value="A &#34;quoted&#34; &lt;value&gt;">`
	if strings.TrimSpace(html) != want {
		t.Fatalf("expected %q, got %q", want, strings.TrimSpace(html))
	}
	if strings.Contains(html, "bad name") {
		t.Fatalf("unsafe attribute name should be skipped: %s", html)
	}
}
