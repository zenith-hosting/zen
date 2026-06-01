package zen

import "testing"

func TestConfigWithDefaultsDev(t *testing.T) {
	cfg := Config{
		Dev: true,
	}

	got := cfg.withDefaults()

	if got.ViteURL != "http://localhost:5173" {
		t.Fatalf("expected default ViteURL, got %q", got.ViteURL)
	}
	if got.AppElementID != "app" {
		t.Fatalf("expected app element id app, got %q", got.AppElementID)
	}
	if got.DataElementID != "__ZEN_DATA__" {
		t.Fatalf("expected data element id __ZEN_DATA__, got %q", got.DataElementID)
	}
}

func TestConfigValidateProductionRequiresPaths(t *testing.T) {
	cfg := Config{
		Dev: false,
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
