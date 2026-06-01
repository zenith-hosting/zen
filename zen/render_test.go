package zen

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/zenith/zen/internal/testutil"
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
			"src/entry-client.tsx": {
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

func TestNewRendererCreatesProductionSSRClient(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")

	err := os.WriteFile(manifestPath, []byte(`{
		"src/entry-client.tsx": {
			"file": "assets/entry-client.js"
		}
	}`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	r, err := New(Config{
		Dev:        false,
		ClientDist: dir,
		Manifest:   manifestPath,
		SSRCommand: []string{
			"node",
			"../js/ssr-worker.mjs",
			"--entry",
			"../js/fixtures/entry-server-ok.mjs",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.ssr == nil {
		t.Fatal("expected production ssr client")
	}
}

type closeTrackingSSRClient struct {
	closed bool
}

func (c *closeTrackingSSRClient) Render(ctx context.Context, req ssrRequest) (ssrResponse, error) {
	return ssrResponse{HTML: ""}, nil
}

func (c *closeTrackingSSRClient) Close() error {
	c.closed = true
	return nil
}

func TestRendererCloseClosesSSRClient(t *testing.T) {
	client := &closeTrackingSSRClient{}

	r := &Renderer{
		config: Config{
			Dev: true,
		},
		ssr: client,
	}

	err := r.Close()
	if err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	if !client.closed {
		t.Fatal("expected SSR client to be closed")
	}
}
