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

	if r.config.ViteURL != "http://localhost:5173" {
		t.Fatalf("expected default vite url, got %q", r.config.ViteURL)
	}

	if r.config.RenderURL != "http://localhost:5173/__zen/render" {
		t.Fatalf("expected default render url, got %q", r.config.RenderURL)
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
			ViteURL:       "http://localhost:5173",
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
			ViteURL:       "http://localhost:5173",
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

func TestRenderIslandWritesHydratableFragment(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<button>Count 0</button>`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:           true,
			ViteURL:       "http://localhost:5173",
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
		RenderURL:  "http://127.0.0.1:4174/__zen/render",
		ClientDist: dir,
		Manifest:   manifestPath,
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
			ViteURL:       "http://localhost:5173",
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
			ViteURL:       "http://localhost:5173",
			RenderURL:     server.URL,
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
