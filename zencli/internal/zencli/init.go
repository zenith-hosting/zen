package zencli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const starterFrontendDir = "frontend"

const zenRuntimeSubdir = ".zen"

const zenRuntimeDir = starterFrontendDir + "/" + zenRuntimeSubdir

var requiredInitTools = []string{"go", "node", "pnpm"}

var postInitCommands = [][]string{
	{"go", "mod", "tidy"},
	{"pnpm", "--dir", "frontend", "install"},
	{"pnpm", "--dir", "frontend", "approve-builds", "--all"},
}

func runInit(ctx context.Context, root string, cfg Config, stdout io.Writer, stderr io.Writer) error {
	for _, tool := range requiredInitTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("zen init: required tool %q was not found in PATH", tool)
		}
	}

	existingProject, err := isExistingZenProject(root)
	if err != nil {
		return err
	}

	if existingProject {
		if err := writeZenRuntimeFiles(root, cfg.FrontendDir); err != nil {
			return err
		}
	} else {
		if err := writeStarterFiles(root); err != nil {
			return err
		}
	}

	for _, cmd := range postInitCommands {
		proc := ManagedProcess{
			Name: cmd[0],
			Command: ProcessCommand{
				Name: cmd[0],
				Args: cmd[1:],
				Dir:  root,
			},
			Stdout: stdout,
			Stderr: stderr,
		}

		if err := proc.Run(ctx); err != nil && cmd[len(cmd)-1] != "install" {
			return fmt.Errorf("zen init: failed to run %v: %w", cmd, err)
		}
	}

	return nil
}

func isExistingZenProject(root string) (bool, error) {
	_, err := os.Stat(filepath.Join(root, "zen.config.json"))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func writeStarterFiles(root string) error {
	for path := range starterFiles() {
		if isZenRuntimePath(path) {
			continue
		}

		fullPath := filepath.Join(root, path)

		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("zen init: %s already exists", path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	if err := writeZenRuntimeFiles(root, starterFrontendDir); err != nil {
		return err
	}

	for path, contents := range starterFiles() {
		if isZenRuntimePath(path) {
			continue
		}

		fullPath := filepath.Join(root, path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func writeZenRuntimeFiles(root, frontendDir string) error {
	if frontendDir == "" {
		frontendDir = starterFrontendDir
	}

	runtimeRoot := filepath.Join(root, frontendDir, zenRuntimeSubdir)
	if err := os.RemoveAll(runtimeRoot); err != nil {
		return err
	}

	for path, contents := range starterFiles() {
		if !isZenRuntimePath(path) {
			continue
		}

		fullPath := filepath.Join(root, zenRuntimeTargetPath(path, frontendDir))

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func isZenRuntimePath(relPath string) bool {
	return relPath == zenRuntimeDir || strings.HasPrefix(relPath, zenRuntimeDir+"/")
}

func zenRuntimeTargetPath(relPath, frontendDir string) string {
	rest := strings.TrimPrefix(relPath, starterFrontendDir+"/")
	return filepath.Join(frontendDir, filepath.FromSlash(rest))
}
