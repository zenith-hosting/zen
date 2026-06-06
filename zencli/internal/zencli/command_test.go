package zencli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunWithoutArgsPrintsUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()

	for _, want := range []string{"zen init", "zen dev", "zen build", "zen start"} {
		if !strings.Contains(got, want) {
			t.Fatalf("usage missing %q in %q", want, got)
		}
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"whatever"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), `unknown command "whatever"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitCommandCreatesStarter(t *testing.T) {
	dir := t.TempDir()
	withFakeInitTools(t)

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(old)
	}()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err = Run([]string{"init"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat("frontend/src/pages/Home.tsx"); err != nil {
		t.Fatalf("expected starter project file: %v", err)
	}
}
