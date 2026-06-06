package zen

import (
	"strings"
	"testing"
)

func TestRenderDocumentTemplateInjectsSlots(t *testing.T) {
	template := `<!doctype html>
<html lang="en">
<head>
<title><!--zen:title--></title>
<!--zen:head-->
<!--zen:styles-->
</head>
<body class="font-sans">
<div id="app"><!--zen:app--></div>
<!--zen:data-->
<!--zen:scripts-->
</body>
</html>`

	doc, err := renderDocumentTemplate(template, documentInput{
		Title:         "Home <unsafe>",
		Head:          `<meta name="description" content="Hello">`,
		DataElementID: "__ZEN_DATA__",
		HTML:          `<main><h1>Hello</h1></main>`,
		HydrationJSON: `{"page":"Home","props":{"title":"Hello"}}`,
		Styles:        []string{"/assets/app.css"},
		Scripts:       []string{"/assets/entry-client.js"},
		DevScripts:    []string{"http://localhost:5173/@vite/client"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		`<title>Home &lt;unsafe&gt;</title>`,
		`<meta name="description" content="Hello">`,
		`<link rel="stylesheet" href="/assets/app.css">`,
		`<div id="app"><main><h1>Hello</h1></main></div>`,
		`<script id="__ZEN_DATA__" type="application/json">{"page":"Home","props":{"title":"Hello"}}</script>`,
		`<script type="module" src="http://localhost:5173/@vite/client"></script>`,
		`<script type="module" src="/assets/entry-client.js"></script>`,
		`<body class="font-sans">`,
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("document missing %q\n%s", want, doc)
		}
	}
}

func TestRenderDocumentTemplateRequiresSlots(t *testing.T) {
	template := `<!doctype html><html><body><!--zen:app--></body></html>`

	_, err := renderDocumentTemplate(template, documentInput{
		Title:         "Home",
		DataElementID: "__ZEN_DATA__",
		HTML:          `<main></main>`,
		HydrationJSON: `{}`,
	})
	if err == nil {
		t.Fatal("expected missing slot error")
	}

	if !strings.Contains(err.Error(), `missing required document slot <!--zen:title-->`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultDocumentTemplateContainsRequiredSlots(t *testing.T) {
	for _, slot := range requiredDocumentSlots() {
		if !strings.Contains(defaultDocumentTemplate, slot) {
			t.Fatalf("default template missing slot %s", slot)
		}
	}
}
