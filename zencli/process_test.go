package zencli

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestShellCommandUsesPortableShell(t *testing.T) {
	cmd := shellCommand("echo hello")

	if cmd.Name != "sh" {
		t.Fatalf("expected sh, got %q", cmd.Name)
	}

	wantArgs := []string{"-c", "echo hello"}
	if !reflect.DeepEqual(cmd.Args, wantArgs) {
		t.Fatalf("expected args %#v, got %#v", wantArgs, cmd.Args)
	}
}

func TestRendererCommands(t *testing.T) {
	cfg := Config{
		FrontendDir:      "frontend",
		DevRendererPort:  5173,
		ProdRendererPort: 4174,
	}

	dev := devRendererCommand(cfg)
	if dev.Name != "node" {
		t.Fatalf("expected node, got %q", dev.Name)
	}

	if !contains(dev.Args, "js/dev-renderer.mjs") {
		t.Fatalf("expected dev renderer script in args, got %#v", dev.Args)
	}

	prod := prodRendererCommand(cfg)
	if prod.Name != "node" {
		t.Fatalf("expected node, got %q", prod.Name)
	}

	if !contains(prod.Args, "js/prod-renderer.mjs") {
		t.Fatalf("expected prod renderer script in args, got %#v", prod.Args)
	}
}

func TestManagedProcessRunsCommand(t *testing.T) {
	var out strings.Builder

	proc := ManagedProcess{
		Name: "test",
		Command: ProcessCommand{
			Name: "sh",
			Args: []string{"-c", "echo hello"},
		},
		Stdout: &out,
		Stderr: &out,
	}

	err := proc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "[test] hello") {
		t.Fatalf("expected prefixed output, got %q", out.String())
	}
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}

	return false
}
