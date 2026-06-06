package zen

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigWithDefaultsDevUsesZenConfigDefaults(t *testing.T) {
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

	if got.AppElementID != "app" {
		t.Fatalf("expected app element id app, got %q", got.AppElementID)
	}

	if got.DataElementID != "__ZEN_DATA__" {
		t.Fatalf("expected data element id __ZEN_DATA__, got %q", got.DataElementID)
	}

	if got.DocumentPath != filepath.Join(".", "frontend", "index.html") {
		t.Fatalf("expected default document path, got %q", got.DocumentPath)
	}

	if got.RenderTimeout != 5*time.Second {
		t.Fatalf("expected render timeout 5s, got %s", got.RenderTimeout)
	}
}

func TestConfigWithDefaultsUsesZenConfigJSON(t *testing.T) {
	dir := t.TempDir()
	raw := `{
		"frontendDir": "web",
		"devRendererPort": 5273,
		"prodRendererPort": 4274
	}`

	if err := os.WriteFile(filepath.Join(dir, "zen.config.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	dev := Config{
		Dev:         true,
		ProjectRoot: dir,
	}.withDefaults()

	if dev.viteURL != "http://localhost:5273" {
		t.Fatalf("expected dev vite URL from zen.config.json, got %q", dev.viteURL)
	}
	if dev.renderURL != "http://localhost:5273/__zen/render" {
		t.Fatalf("expected dev render URL from zen.config.json, got %q", dev.renderURL)
	}
	if dev.DocumentPath != filepath.Join(dir, "web", "index.html") {
		t.Fatalf("expected dev document path from frontend dir, got %q", dev.DocumentPath)
	}

	prod := Config{
		Dev:         false,
		ProjectRoot: dir,
	}.withDefaults()

	if prod.renderURL != "http://127.0.0.1:4274/__zen/render" {
		t.Fatalf("expected prod render URL from zen.config.json, got %q", prod.renderURL)
	}
	if prod.clientDist != filepath.Join(dir, "web", "dist", "client") {
		t.Fatalf("expected client dist from frontend dir, got %q", prod.clientDist)
	}
	if prod.manifest != filepath.Join(dir, "web", "dist", "client", ".vite", "manifest.json") {
		t.Fatalf("expected manifest from frontend dir, got %q", prod.manifest)
	}
	if prod.DocumentPath != filepath.Join(dir, "web", "index.html") {
		t.Fatalf("expected prod document path from frontend dir, got %q", prod.DocumentPath)
	}
}

func TestConfigWithDefaultsKeepsDocumentPathOverride(t *testing.T) {
	cfg := Config{
		Dev:          true,
		DocumentPath: "custom/document.html",
	}

	got := cfg.withDefaults()

	if got.DocumentPath != "custom/document.html" {
		t.Fatalf("expected document path override, got %q", got.DocumentPath)
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

func TestConfigValidateProductionRequiresPaths(t *testing.T) {
	cfg := Config{
		Dev: false,
	}

	err := cfg.withDefaults().validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigValidateRequiresRenderURL(t *testing.T) {
	cfg := Config{
		Dev: true,
	}

	err := cfg.withDefaults().validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
