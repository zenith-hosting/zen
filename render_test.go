package zen

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func mustRenderPage(t *testing.T, r *Renderer, url, page string, props any, options ...RenderOption) Response {
	t.Helper()
	response, err := r.RenderPage(context.Background(), url, page, props, options...)
	if err != nil {
		t.Fatalf("render page: %v", err)
	}
	return response
}

func mustRenderIsland(t *testing.T, r *Renderer, url, island string, props any) Response {
	t.Helper()
	response, err := r.RenderIsland(context.Background(), url, island, props)
	if err != nil {
		t.Fatalf("render island: %v", err)
	}
	return response
}

func TestNewRendererAppliesDefaults(t *testing.T) {
	r, err := New(Config{
		Dev: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.config.viteURL != "http://localhost:5173" {
		t.Fatalf("expected default vite url, got %q", r.config.viteURL)
	}

	if r.config.renderURL != "http://localhost:5173/__zen/render" {
		t.Fatalf("expected default render url, got %q", r.config.renderURL)
	}

	if r.ssr == nil {
		t.Fatal("expected renderer to create ssr client")
	}
}

func TestNewRendererRejectsInvalidProductionConfig(t *testing.T) {
	_, err := New(Config{
		Dev: false,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewRendererRejectsNegativeTimeout(t *testing.T) {
	_, err := New(Config{Dev: true, RenderTimeout: -time.Second})
	if err == nil {
		t.Fatal("expected timeout validation error")
	}
}

func TestRenderReturnsSSRDocument(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<main><h1>Hello</h1></main>`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:          true,
			viteURL:      "http://localhost:5173",
			DefaultTitle: "Zen",
		},
		ssr: client,
	}

	res, err := r.RenderPage(context.Background(), "/", "Home", map[string]string{
		"title": "Hello",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	body := string(res.Body)

	if res.Status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Status)
	}
	if res.ContentType != "text/html; charset=utf-8" {
		t.Fatalf("unexpected content type %q", res.ContentType)
	}
	if !strings.Contains(body, `<main><h1>Hello</h1></main>`) {
		t.Fatalf("body missing ssr html: %s", body)
	}
	if !strings.Contains(body, `"page":"Home"`) {
		t.Fatalf("body missing hydration page: %s", body)
	}
	if !strings.Contains(body, `http://localhost:5173/@vite/client`) {
		t.Fatalf("body missing Vite dev client: %s", body)
	}
}

func TestRenderPageSendsPageModeToRenderer(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<main><h1>Hello</h1></main>`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:          true,
			viteURL:      "http://localhost:5173",
			DefaultTitle: "Zen",
		},
		ssr: client,
	}

	res := mustRenderPage(t, r, "/users?active=true", "Home", map[string]string{
		"title": "Hello",
	})

	if res.Status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Status)
	}
	if client.req.Mode != "page" {
		t.Fatalf("expected page render mode, got %q", client.req.Mode)
	}
	if client.req.Page != "Home" {
		t.Fatalf("expected page Home, got %q", client.req.Page)
	}
	if client.req.URL != "/users?active=true" {
		t.Fatalf("expected original URL, got %q", client.req.URL)
	}
	if client.req.IdentifierPrefix != "zen-page-" {
		t.Fatalf("expected page identifier prefix, got %q", client.req.IdentifierPrefix)
	}
	if !strings.Contains(string(res.Body), `"identifierPrefix":"zen-page-"`) {
		t.Fatalf("body missing page identifier prefix: %s", res.Body)
	}
}

func TestRenderPageUsesHardcodedDocument(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<main><h1>Hello</h1></main>`,
			Head: `<meta name="description" content="From renderer">`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:          true,
			viteURL:      "http://localhost:5173",
			DefaultTitle: "Zen",
		},
		ssr: client,
	}

	res := mustRenderPage(t, r, "/", "Home", map[string]string{"title": "Hello"},
		WithTitle("Custom <Title>"),
		WithStatus(http.StatusCreated),
		WithBase(Href("/")),
		WithMeta(Name("description"), Content("From Go <unsafe>")),
		WithLink(Rel("canonical"), Href("https://example.com/?q=<unsafe>"), Attr("data-source", "go <unsafe>")),
		WithStyle(`body { color: red; }`),
		WithScript(Type("application/ld+json"), Text(`{"name":"Home"}`)),
	)
	body := string(res.Body)

	if res.Status != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Status)
	}
	for _, want := range []string{
		`<body>`,
		`<title>Custom &lt;Title&gt;</title>`,
		`<meta name="description" content="From renderer">`,
		`<base href="/">`,
		`<meta name="description" content="From Go &lt;unsafe&gt;">`,
		`<link rel="canonical" href="https://example.com/?q=&lt;unsafe&gt;" data-source="go &lt;unsafe&gt;">`,
		`<style>body { color: red; }</style>`,
		`<script type="application/ld+json">{"name":"Home"}</script>`,
		`<div id="app"><main><h1>Hello</h1></main></div>`,
		`http://localhost:5173/@vite/client`,
		`http://localhost:5173/.zen/entries/react-refresh.mjs`,
		`http://localhost:5173/.zen/entries/entry-client.tsx`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q\n%s", want, body)
		}
	}
}

func TestRenderIslandWritesHydratableFragment(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<button>Count 0</button>`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:          true,
			viteURL:      "http://localhost:5173",
			DefaultTitle: "Zen",
		},
		ssr: client,
	}

	res := mustRenderIsland(t, r, "/counter", "Counter", map[string]int{
		"count": 0,
	})
	body := string(res.Body)

	if res.Status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Status)
	}
	if strings.Contains(body, "<!doctype html>") {
		t.Fatalf("island render should not return a full document: %s", body)
	}
	for _, want := range []string{
		`data-zen-island-root`,
		`data-zen-island="Counter"`,
		`<button>Count 0</button>`,
		`"island":"Counter"`,
		`"props":{"count":0}`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("island fragment missing %q\n%s", want, body)
		}
	}
	for _, unwanted := range []string{"<link", `<script type="module"`} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("island fragment includes page asset %q\n%s", unwanted, body)
		}
	}
	if client.req.Mode != "island" {
		t.Fatalf("expected island render mode, got %q", client.req.Mode)
	}
	if client.req.Island != "Counter" {
		t.Fatalf("expected island Counter, got %q", client.req.Island)
	}
	if client.req.URL != "/counter" {
		t.Fatalf("expected original URL, got %q", client.req.URL)
	}
	firstPrefix := client.req.IdentifierPrefix
	if firstPrefix == "" {
		t.Fatal("expected island identifier prefix")
	}
	if !strings.Contains(body, `"identifierPrefix":"`+firstPrefix+`"`) {
		t.Fatalf("island fragment missing identifier prefix: %s", body)
	}

	mustRenderIsland(t, r, "/counter", "Counter", map[string]int{"count": 1})
	if client.req.IdentifierPrefix == firstPrefix {
		t.Fatalf("expected unique island identifier prefixes, got %q", firstPrefix)
	}
}

func TestRenderInjectsProductionManifestAssets(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<main>Production</main>`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:          false,
			DefaultTitle: "Zen",
		},
		ssr: client,
		manifest: viteManifest{
			".zen/entries/entry-client.tsx": {
				File: "assets/entry-client.abc123.js",
				CSS:  []string{"assets/app.def456.css"},
			},
		},
	}

	res := mustRenderPage(t, r, "/", "Home", map[string]string{
		"title": "Production",
	})
	body := string(res.Body)

	if !strings.Contains(body, `<link rel="stylesheet" href="/assets/app.def456.css">`) {
		t.Fatalf("body missing production css: %s", body)
	}
	if !strings.Contains(body, `<script type="module" src="/assets/entry-client.abc123.js"></script>`) {
		t.Fatalf("body missing production script: %s", body)
	}
	if strings.Contains(body, "/@vite/client") {
		t.Fatalf("production body should not include vite dev client: %s", body)
	}
}

func TestNewRendererCreatesProductionHTTPSSRClient(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")

	err := os.WriteFile(manifestPath, []byte(`{
		".zen/entries/entry-client.tsx": {
			"file": "assets/entry-client.js"
		}
	}`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	r, err := New(Config{
		Dev:        false,
		clientDist: dir,
		manifest:   manifestPath,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.ssr == nil {
		t.Fatal("expected production ssr client")
	}
}

func TestRenderReturnsErrorWhenSSRClientMissing(t *testing.T) {
	r := &Renderer{
		config: Config{
			Dev:          true,
			viteURL:      "http://localhost:5173",
			DefaultTitle: "Zen",
		},
		ssr: nil,
	}

	_, err := r.RenderPage(context.Background(), "/", "Home", map[string]string{})
	if err == nil {
		t.Fatal("expected missing ssr client error")
	}
}

func TestRenderInlineStylesReplacesProductionStylesheetLinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets/app.css"), []byte("body{color:red}"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := &Renderer{
		config: Config{InlineStyles: true, clientDist: dir},
		ssr:    &fakeSSRClient{res: ssrResponse{HTML: "<main>Home</main>"}},
		manifest: viteManifest{".zen/entries/entry-client.tsx": {
			File: "assets/app.js",
			CSS:  []string{"assets/app.css"},
		}},
	}

	body := string(mustRenderPage(t, r, "/", "Home", nil).Body)
	if !strings.Contains(body, "<style>body{color:red}</style>") || strings.Contains(body, `<link rel="stylesheet"`) {
		t.Fatalf("expected inline CSS without stylesheet link: %s", body)
	}
}

func TestRenderInlineStylesLeavesDevelopmentAssetsAlone(t *testing.T) {
	r := &Renderer{
		config: Config{Dev: true, InlineStyles: true, viteURL: "http://localhost:5173"},
		ssr:    &fakeSSRClient{res: ssrResponse{HTML: "<main>Home</main>"}},
	}

	body := string(mustRenderPage(t, r, "/", "Home", nil).Body)
	if !strings.Contains(body, "http://localhost:5173/@vite/client") || strings.Contains(body, "<style>") {
		t.Fatalf("expected unchanged development assets: %s", body)
	}
}
