package zen

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/zenith-hosting/zen/internal/testutil"
)

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

func TestRenderWritesSSRDocumentToFiberResponse(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<main><h1>Hello</h1></main>`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:           true,
			viteURL:       "http://localhost:5173",
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: client,
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.Render(c, "Home", map[string]string{
			"title": "Hello",
		})
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")
	body := testutil.ReadBody(t, res)

	if res.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
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
			Dev:           true,
			viteURL:       "http://localhost:5173",
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: client,
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.RenderPage(c, "Home", map[string]string{
			"title": "Hello",
		})
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")

	if res.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}
	if client.req.Mode != "page" {
		t.Fatalf("expected page render mode, got %q", client.req.Mode)
	}
	if client.req.Page != "Home" {
		t.Fatalf("expected page Home, got %q", client.req.Page)
	}
}

func TestRenderPageUsesConfiguredDocumentTemplate(t *testing.T) {
	dir := t.TempDir()
	documentPath := filepath.Join(dir, "index.html")
	template := `<!doctype html>
<html lang="en">
<head>
<title><!--zen:title--></title>
<!--zen:head-->
<!--zen:base-->
<!--zen:meta-->
<!--zen:link-->
<!--zen:style-->
<!--zen:styles-->
<!--zen:script-->
</head>
<body class="custom-shell">
<div id="app"><!--zen:app--></div>
<!--zen:data-->
<!--zen:scripts-->
</body>
</html>`

	if err := os.WriteFile(documentPath, []byte(template), 0o644); err != nil {
		t.Fatal(err)
	}

	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<main><h1>Hello</h1></main>`,
			Head: `<meta name="description" content="From renderer">`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:           true,
			viteURL:       "http://localhost:5173",
			DocumentPath:  documentPath,
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: client,
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.RenderPage(c, "Home", map[string]string{"title": "Hello"},
			WithTitle("Custom <Title>"),
			Base(Href("/")),
			WithMeta(Name("description"), Content("From Go <unsafe>")),
			WithLink(Rel("canonical"), Href("https://example.com/?q=<unsafe>"), Attr("data-source", "go <unsafe>")),
			WithStyle(`body { color: red; }`),
			WithScript(Type("application/ld+json"), Text(`{"name":"Home"}`)),
		)
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")
	body := testutil.ReadBody(t, res)

	for _, want := range []string{
		`<body class="custom-shell">`,
		`<title>Custom &lt;Title&gt;</title>`,
		`<meta name="description" content="From renderer">`,
		`<base href="/">`,
		`<meta name="description" content="From Go &lt;unsafe&gt;">`,
		`<link rel="canonical" href="https://example.com/?q=&lt;unsafe&gt;" data-source="go &lt;unsafe&gt;">`,
		`<style>body { color: red; }</style>`,
		`<script type="application/ld+json">{"name":"Home"}</script>`,
		`<div id="app"><main><h1>Hello</h1></main></div>`,
		`http://localhost:5173/@vite/client`,
		`http://localhost:5173/.zen/entries/entry-client.tsx`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q\n%s", want, body)
		}
	}
}

func TestRendererDocumentTemplateReportsMissingFile(t *testing.T) {
	r := &Renderer{
		config: Config{
			DocumentPath: filepath.Join(t.TempDir(), "missing.html"),
		},
	}

	_, err := r.documentTemplate()
	if err == nil {
		t.Fatal("expected missing document template error")
	}
	if !strings.Contains(err.Error(), "zen: missing document template") {
		t.Fatalf("unexpected error: %v", err)
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
			Dev:           true,
			viteURL:       "http://localhost:5173",
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: client,
	}

	app := fiber.New()
	app.Get("/counter", func(c fiber.Ctx) error {
		return r.RenderIsland(c, "Counter", map[string]int{
			"count": 0,
		})
	})

	res := testutil.PerformRequest(t, app, "GET", "/counter", "")
	body := testutil.ReadBody(t, res)

	if res.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
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
		`http://localhost:5173/@vite/client`,
		`http://localhost:5173/.zen/entries/entry-client.tsx`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("island fragment missing %q\n%s", want, body)
		}
	}
	if client.req.Mode != "island" {
		t.Fatalf("expected island render mode, got %q", client.req.Mode)
	}
	if client.req.Island != "Counter" {
		t.Fatalf("expected island Counter, got %q", client.req.Island)
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
			Dev:           false,
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: client,
		manifest: viteManifest{
			".zen/entries/entry-client.tsx": {
				File: "assets/entry-client.abc123.js",
				CSS:  []string{"assets/app.def456.css"},
			},
		},
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.Render(c, "Home", map[string]string{
			"title": "Production",
		})
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")
	body := testutil.ReadBody(t, res)

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

func TestRenderIslandInjectsProductionManifestAssets(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<button>Count 0</button>`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:           false,
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: client,
		manifest: viteManifest{
			".zen/entries/entry-client.tsx": {
				File: "assets/entry-client.abc123.js",
				CSS:  []string{"assets/app.def456.css"},
			},
		},
	}

	app := fiber.New()
	app.Get("/counter", func(c fiber.Ctx) error {
		return r.RenderIsland(c, "Counter", map[string]int{
			"count": 0,
		})
	})

	res := testutil.PerformRequest(t, app, "GET", "/counter", "")
	body := testutil.ReadBody(t, res)

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
			Dev:           true,
			viteURL:       "http://localhost:5173",
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: nil,
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.Render(c, "Home", map[string]string{})
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")

	if res.StatusCode == fiber.StatusOK {
		t.Fatal("expected non-200 status when renderer has no ssr client")
	}
}

func TestRenderReturnsRendererHTTPErrorThroughFiber(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		_ = json.NewEncoder(w).Encode(httpRendererErrorResponse{
			Error: httpRendererError{
				Message: "renderer exploded",
			},
		})
	}))
	defer server.Close()

	r := &Renderer{
		config: Config{
			Dev:           true,
			viteURL:       "http://localhost:5173",
			renderURL:     server.URL,
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: newHTTPSSRClient(httpSSRClientConfig{
			RenderURL: server.URL,
			Timeout:   time.Second,
		}),
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.Render(c, "Home", map[string]string{})
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")

	if res.StatusCode == fiber.StatusOK {
		t.Fatal("expected non-200 response")
	}
}

func TestRenderInlineStylesEmitsStyleTagInsteadOfLink(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	css := `.marquee{display:flex}`
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.def456.css"), []byte(css), 0o644); err != nil {
		t.Fatalf("write css: %v", err)
	}

	client := &fakeSSRClient{
		res: ssrResponse{HTML: `<main>Public</main>`},
	}

	r := &Renderer{
		config: Config{
			Dev:           false,
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
			clientDist:    dir,
		},
		ssr: client,
		manifest: viteManifest{
			".zen/entries/entry-client.tsx": {
				File: "assets/entry-client.abc123.js",
				CSS:  []string{"assets/app.def456.css"},
			},
		},
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.Render(c, "Public", map[string]string{}, WithInlineStyles())
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")
	body := testutil.ReadBody(t, res)

	if !strings.Contains(body, "<style>"+css+"</style>") {
		t.Fatalf("body missing inlined css: %s", body)
	}
	if strings.Contains(body, `<link rel="stylesheet"`) {
		t.Fatalf("body should not include a render-blocking stylesheet link: %s", body)
	}
	// The JS entry still loads normally.
	if !strings.Contains(body, `<script type="module" src="/assets/entry-client.abc123.js"></script>`) {
		t.Fatalf("body missing entry script: %s", body)
	}
}

func TestRenderInlineStylesIgnoredInDev(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{HTML: `<main>Public</main>`},
	}

	r := &Renderer{
		config: Config{
			Dev:           true,
			viteURL:       "http://localhost:5173",
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: client,
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.Render(c, "Public", map[string]string{}, WithInlineStyles())
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")
	body := testutil.ReadBody(t, res)

	// Dev injects CSS via the Vite client; inlining is a no-op and must not break.
	if strings.Contains(body, "<style>") && !strings.Contains(body, "@vite/client") {
		t.Fatalf("dev render should keep vite client injection, got: %s", body)
	}
	if !strings.Contains(body, `http://localhost:5173/@vite/client`) {
		t.Fatalf("dev body missing vite client: %s", body)
	}
}
