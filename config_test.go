package zen

import (
	"testing"
	"time"
)

func TestConfigWithDefaultsDev(t *testing.T) {
	cfg := Config{
		Dev: true,
	}

	got := cfg.withDefaults()

	if got.ViteURL != "http://localhost:5173" {
		t.Fatalf("expected default ViteURL, got %q", got.ViteURL)
	}

	if got.RenderURL != "http://localhost:5173/__zen/render" {
		t.Fatalf("expected default RenderURL, got %q", got.RenderURL)
	}

	if got.AppElementID != "app" {
		t.Fatalf("expected app element id app, got %q", got.AppElementID)
	}

	if got.DataElementID != "__ZEN_DATA__" {
		t.Fatalf("expected data element id __ZEN_DATA__, got %q", got.DataElementID)
	}

	if got.RenderTimeout != 5*time.Second {
		t.Fatalf("expected render timeout 5s, got %s", got.RenderTimeout)
	}
}

func TestConfigWithDefaultsProduction(t *testing.T) {
	cfg := Config{
		Dev:       false,
		RenderURL: "http://127.0.0.1:4174/__zen/render",
	}

	got := cfg.withDefaults()

	if got.RenderURL != "http://127.0.0.1:4174/__zen/render" {
		t.Fatalf("expected configured RenderURL, got %q", got.RenderURL)
	}

	if got.RenderTimeout != 5*time.Second {
		t.Fatalf("expected render timeout 5s, got %s", got.RenderTimeout)
	}
}

func TestConfigValidateProductionRequiresPaths(t *testing.T) {
	cfg := Config{
		Dev:       false,
		RenderURL: "http://127.0.0.1:4174/__zen/render",
	}

	err := cfg.withDefaults().validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConfigValidateRequiresRenderURL(t *testing.T) {
	cfg := Config{
		Dev:       true,
		ViteURL:   "http://localhost:5173",
		RenderURL: " ",
	}

	err := cfg.withDefaults().validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConfigValidateDevRequiresViteURL(t *testing.T) {
	cfg := Config{
		Dev:     true,
		ViteURL: " ",
	}

	err := cfg.withDefaults().validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}
