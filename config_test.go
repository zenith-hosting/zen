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
