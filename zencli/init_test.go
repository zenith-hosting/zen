package zencli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitWritesCompleteStarterProject(t *testing.T) {
	dir := t.TempDir()

	err := runInit(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for path, expected := range starterFiles() {
		raw, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatalf("expected %s to be written: %v", path, err)
		}

		if string(raw) != expected {
			t.Fatalf("unexpected contents for %s", path)
		}
	}
}

func TestRunInitCreatesNestedDirectories(t *testing.T) {
	dir := t.TempDir()

	err := runInit(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "frontend", "src", "pages", "Home.tsx")); err != nil {
		t.Fatalf("expected nested frontend page to exist: %v", err)
	}
}

func TestRunInitRefusesToOverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()

	existing := filepath.Join(dir, "main.go")
	if err := os.WriteFile(existing, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runInit(dir)
	if err == nil {
		t.Fatal("expected overwrite protection error")
	}

	if !strings.Contains(err.Error(), "main.go already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}
