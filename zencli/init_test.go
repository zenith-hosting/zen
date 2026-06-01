package zencli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitWritesConfigAndAirFile(t *testing.T) {
	dir := t.TempDir()

	err := runInit(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configRaw, err := os.ReadFile(filepath.Join(dir, "zen.config.json"))
	if err != nil {
		t.Fatalf("expected zen.config.json: %v", err)
	}

	if !strings.Contains(string(configRaw), `"frontendDir": "frontend"`) {
		t.Fatalf("config missing frontend dir: %s", configRaw)
	}

	airRaw, err := os.ReadFile(filepath.Join(dir, ".air.toml"))
	if err != nil {
		t.Fatalf("expected .air.toml: %v", err)
	}

	if !strings.Contains(string(airRaw), "cmd = \"go build -o ./tmp/zen-app .\"") {
		t.Fatalf("air config missing build command: %s", airRaw)
	}
}
