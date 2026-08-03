package zen

import (
	"path/filepath"
	"testing"
	"time"
)

func TestConfigWithDefaultsDev(t *testing.T) {
	cfg := Config{
		Dev: true,
	}

	got := cfg.withDefaults()

	if got.viteURL != "http://localhost:5173" {
		t.Fatalf("expected default ViteURL, got %q", got.viteURL)
	}

	if got.renderURL != "http://localhost:5173/__zen/render" {
		t.Fatalf("expected default RenderURL, got %q", got.renderURL)
	}

	if got.RenderTimeout != 5*time.Second {
		t.Fatalf("expected render timeout 5s, got %s", got.RenderTimeout)
	}
	if got.FrontendDir != "frontend" {
		t.Fatalf("expected default frontend dir, got %q", got.FrontendDir)
	}
	if got.DevRendererPort != 5173 {
		t.Fatalf("expected default dev renderer port, got %d", got.DevRendererPort)
	}
	if got.ProdRendererPort != 4174 {
		t.Fatalf("expected default prod renderer port, got %d", got.ProdRendererPort)
	}
}

func TestConfigWithDefaultsUsesPublicFields(t *testing.T) {
	dir := t.TempDir()

	dev := Config{
		Dev:             true,
		ProjectRoot:     dir,
		FrontendDir:     "web",
		DevRendererPort: 5273,
	}.withDefaults()

	if dev.viteURL != "http://localhost:5273" {
		t.Fatalf("expected configured dev vite URL, got %q", dev.viteURL)
	}
	if dev.renderURL != "http://localhost:5273/__zen/render" {
		t.Fatalf("expected configured dev render URL, got %q", dev.renderURL)
	}
	prod := Config{
		ProjectRoot:      dir,
		FrontendDir:      "web",
		ProdRendererPort: 4274,
	}.withDefaults()

	if prod.renderURL != "http://127.0.0.1:4274/__zen/render" {
		t.Fatalf("expected configured prod render URL, got %q", prod.renderURL)
	}
	if prod.clientDist != filepath.Join(dir, "web", "dist", "client") {
		t.Fatalf("expected client dist from frontend dir, got %q", prod.clientDist)
	}
	if prod.manifest != filepath.Join(dir, "web", "dist", "client", ".vite", "manifest.json") {
		t.Fatalf("expected manifest from frontend dir, got %q", prod.manifest)
	}
}

func TestConfigWithDefaultsProduction(t *testing.T) {
	cfg := Config{
		Dev: false,
	}

	got := cfg.withDefaults()

	if got.renderURL != "http://127.0.0.1:4174/__zen/render" {
		t.Fatalf("expected default render URL, got %q", got.renderURL)
	}

	if got.RenderTimeout != 5*time.Second {
		t.Fatalf("expected render timeout 5s, got %s", got.RenderTimeout)
	}
}
