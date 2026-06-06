package zencli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitWritesCompleteStarterProject(t *testing.T) {
	dir := t.TempDir()
	withFakeInitTools(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runInit(t.Context(), dir, &stdout, &stderr)
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
	withFakeInitTools(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runInit(t.Context(), dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "frontend", "src", "pages", "Home.tsx")); err != nil {
		t.Fatalf("expected nested frontend page to exist: %v", err)
	}
}

func TestRunInitRefusesToOverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	withFakeInitTools(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	existing := filepath.Join(dir, "main.go")
	if err := os.WriteFile(existing, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runInit(t.Context(), dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected overwrite protection error")
	}

	if !strings.Contains(err.Error(), "main.go already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitRunsPostInitSetup(t *testing.T) {
	dir := t.TempDir()
	logPath := withFakeInitTools(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runInit(t.Context(), dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected post-init commands to run: %v", err)
	}

	want := strings.Join([]string{
		"go mod tidy",
		"pnpm --dir frontend install",
		"pnpm --dir frontend approve-builds --all",
		"",
	}, "\n")

	if string(raw) != want {
		t.Fatalf("unexpected post-init commands:\nwant:\n%s\ngot:\n%s", want, string(raw))
	}
}

func withFakeInitTools(t *testing.T) string {
	t.Helper()

	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(t.TempDir(), "post-init.log")
	script := "#!/bin/sh\nprintf '%s %s\\n' \"$(basename \"$0\")\" \"$*\" >> \"$ZEN_INIT_LOG\"\n"

	for _, tool := range []string{"go", "node", "pnpm"} {
		if err := os.WriteFile(filepath.Join(binDir, tool), []byte(script), 0o755); err != nil {
			t.Fatalf("write fake %s: %v", tool, err)
		}
	}

	oldPath := os.Getenv("PATH")
	oldLog := os.Getenv("ZEN_INIT_LOG")

	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("ZEN_INIT_LOG", logPath); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = os.Setenv("PATH", oldPath)
		_ = os.Setenv("ZEN_INIT_LOG", oldLog)
	})

	return logPath
}
