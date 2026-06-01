package zen

import "testing"

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
