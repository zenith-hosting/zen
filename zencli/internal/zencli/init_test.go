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

	err := runInit(t.Context(), dir, defaultConfig(), &stdout, &stderr)
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

	err := runInit(t.Context(), dir, defaultConfig(), &stdout, &stderr)
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

	err := runInit(t.Context(), dir, defaultConfig(), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected overwrite protection error")
	}

	if !strings.Contains(err.Error(), "main.go already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitRunsPostInitSetupInExistingZenProject(t *testing.T) {
	dir := t.TempDir()
	logPath := withFakeInitTools(t)
	customMain := []byte("package main\n\nfunc main() {}\n")

	writeStarterProject(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "main.go"), customMain, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runInit(t.Context(), dir, defaultConfig(), &stdout, &stderr)
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

	raw, err = os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(customMain) {
		t.Fatal("expected existing project files to be left unchanged")
	}
}

func TestRunInitReplacesZenRuntimeInExistingProject(t *testing.T) {
	dir := t.TempDir()
	withFakeInitTools(t)

	writeStarterProject(t, dir)

	// A developer-owned file should be left untouched.
	customMain := []byte("package main\n\nfunc main() {}\n")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), customMain, 0o644); err != nil {
		t.Fatal(err)
	}

	// A stale Zen-managed file and a tampered runtime file should be replaced.
	stale := filepath.Join(dir, "frontend", ".zen", "renderers", "stale.mjs")
	if err := os.WriteFile(stale, []byte("// stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	renderer := filepath.Join(dir, "frontend", ".zen", "renderers", "dev-renderer.mjs")
	if err := os.WriteFile(renderer, []byte("// tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := runInit(t.Context(), dir, defaultConfig(), &stdout, &stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("expected stale runtime file to be removed, got err=%v", err)
	}

	raw, err := os.ReadFile(renderer)
	if err != nil {
		t.Fatalf("expected renderer to be rewritten: %v", err)
	}
	if string(raw) != starterFiles()["frontend/.zen/renderers/dev-renderer.mjs"] {
		t.Fatal("expected runtime file to be replaced with starter contents")
	}

	raw, err = os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(customMain) {
		t.Fatal("expected developer-owned files to be left unchanged")
	}
}

func TestRunInitRefreshesRuntimeUnderConfiguredFrontendDir(t *testing.T) {
	dir := t.TempDir()
	withFakeInitTools(t)

	// Mark this as an existing project with a non-default frontend dir.
	if err := os.WriteFile(filepath.Join(dir, "zen.config.json"), []byte(`{"frontendDir":"web"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := defaultConfig()
	cfg.FrontendDir = "web"

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := runInit(t.Context(), dir, cfg, &stdout, &stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "web", ".zen", "renderers", "dev-renderer.mjs"))
	if err != nil {
		t.Fatalf("expected runtime under configured frontendDir: %v", err)
	}
	if string(raw) != starterFiles()["frontend/.zen/renderers/dev-renderer.mjs"] {
		t.Fatal("expected runtime file to match starter contents")
	}

	if _, err := os.Stat(filepath.Join(dir, "frontend", ".zen")); !os.IsNotExist(err) {
		t.Fatalf("expected no runtime under default frontend dir, got err=%v", err)
	}
}

func TestRunInitReplacesExistingZenRuntimeInNewProject(t *testing.T) {
	dir := t.TempDir()
	withFakeInitTools(t)

	// A pre-existing .zen file should not trigger overwrite protection.
	renderer := filepath.Join(dir, "frontend", ".zen", "renderers", "dev-renderer.mjs")
	if err := os.MkdirAll(filepath.Dir(renderer), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(renderer, []byte("// tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := runInit(t.Context(), dir, defaultConfig(), &stdout, &stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(renderer)
	if err != nil {
		t.Fatalf("expected renderer to be written: %v", err)
	}
	if string(raw) != starterFiles()["frontend/.zen/renderers/dev-renderer.mjs"] {
		t.Fatal("expected pre-existing runtime file to be replaced")
	}
}

func TestRunInitRunsPostInitSetup(t *testing.T) {
	dir := t.TempDir()
	logPath := withFakeInitTools(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runInit(t.Context(), dir, defaultConfig(), &stdout, &stderr)
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

func writeStarterProject(t *testing.T, dir string) {
	t.Helper()

	for path, contents := range starterFiles() {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
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
