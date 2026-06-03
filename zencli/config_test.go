package zencli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigAppliesDefaults(t *testing.T) {
	dir := t.TempDir()

	cfg, err := loadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ZenConfigVersion != 1 {
		t.Fatalf("expected zen config version 1, got %d", cfg.ZenConfigVersion)
	}

	if cfg.AppCommand != "go run ." {
		t.Fatalf("expected default app command, got %q", cfg.AppCommand)
	}

	if cfg.AirCommand != "go tool air" {
		t.Fatalf("expected default air command, got %q", cfg.AirCommand)
	}

	if cfg.FrontendDir != "frontend" {
		t.Fatalf("expected default frontend dir, got %q", cfg.FrontendDir)
	}

	if cfg.DevRendererPort != 5173 {
		t.Fatalf("expected dev renderer port 5173, got %d", cfg.DevRendererPort)
	}

	if cfg.ProdRendererPort != 4174 {
		t.Fatalf("expected prod renderer port 4174, got %d", cfg.ProdRendererPort)
	}
}

func TestLoadConfigReadsZenConfigJSON(t *testing.T) {
	dir := t.TempDir()

	raw := `{
		"appCommand": "go run ./cmd/app",
		"airCommand": "go tool air -c .air.toml",
		"frontendDir": "web",
		"devRendererPort": 5273,
		"prodRendererPort": 4274,
		"binaryPath": "./bin/server"
	}`

	if err := os.WriteFile(filepath.Join(dir, "zen.config.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppCommand != "go run ./cmd/app" {
		t.Fatalf("expected configured app command, got %q", cfg.AppCommand)
	}

	if cfg.FrontendDir != "web" {
		t.Fatalf("expected frontend dir web, got %q", cfg.FrontendDir)
	}

	if cfg.BinaryPath != "./bin/server" {
		t.Fatalf("expected binary path, got %q", cfg.BinaryPath)
	}
}
