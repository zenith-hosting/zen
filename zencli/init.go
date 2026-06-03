package zencli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var requiredTools = []string{"go", "node", "pnpm"}

var initCmds = [][]string{
	{"go", "mod", "tidy"},
	{"pnpm", "--dir", "frontend", "install"},
	{"pnpm", "--dir", "frontend", "approve-builds", "--all"},
}

func runInit(ctx context.Context, root string, stdout, stderr io.Writer) error {
	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("zen init: required tool %q was not found in PATH", tool)
		}
	}

	for path := range starterFiles() {
		fullPath := filepath.Join(root, path)

		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("zen init: %s already exists", path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	for path, contents := range starterFiles() {
		fullPath := filepath.Join(root, path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			return err
		}
	}

	for _, cmd := range initCmds {
		proc := ManagedProcess{
			Name:    cmd[0],
			Command: ProcessCommand{Name: cmd[0], Args: cmd[1:]},
			Stdout:  stdout,
			Stderr:  stderr,
		}

		if err := proc.Run(ctx); err != nil {
			return fmt.Errorf("zen init: failed to run %v: %w", cmd, err)
		}
	}

	return nil
}
