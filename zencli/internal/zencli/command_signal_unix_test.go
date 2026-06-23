//go:build !windows

package zencli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestRunDevStopsManagedProcessesOnInterrupt(t *testing.T) {
	if os.Getenv("ZEN_TEST_RUN_DEV_INTERRUPT_HELPER") == "1" {
		runInterruptHelper(t, "dev")
		return
	}

	dir := t.TempDir()
	writeInterruptTestProject(t, dir, 49173)

	cmd := exec.Command(os.Args[0], "-test.run", "^TestRunDevStopsManagedProcessesOnInterrupt$")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"ZEN_TEST_RUN_DEV_INTERRUPT_HELPER=1",
		"ZEN_TEST_PID_DIR="+dir,
		"PATH="+filepath.Join(dir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	pids := []int{
		waitForPIDFile(t, filepath.Join(dir, "pnpm.pid")),
		waitForPIDFile(t, filepath.Join(dir, "node.pid")),
		waitForPIDFile(t, filepath.Join(dir, "go.pid")),
	}
	for _, pid := range pids {
		defer killProcessIfRunning(pid)
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("helper exited with error: %v\n%s", err, out.String())
		}
	case <-time.After(4 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("timed out waiting for interrupted zen dev to exit\n%s", out.String())
	}

	for _, pid := range pids {
		if err := waitForProcessExit(pid, 2*time.Second); err != nil {
			t.Fatal(err)
		}
	}
}

func writeInterruptTestProject(t *testing.T, dir string, port int) {
	t.Helper()

	mustMkdirAll(t, filepath.Join(dir, "frontend", "node_modules", "vite"))
	mustMkdirAll(t, filepath.Join(dir, "bin"))
	mustWriteFile(t, filepath.Join(dir, "frontend", "node_modules", "vite", "package.json"), "{}\n")
	mustWriteFile(t, filepath.Join(dir, "zen.config.json"), fmt.Sprintf(`{
  "zenConfigVersion": 1,
  "airCommand": "go tool air",
  "frontendDir": "frontend",
  "devRendererPort": %d
}
`, port))

	mustWriteFile(t, filepath.Join(dir, "bin", "pnpm"), `#!/bin/sh
echo $$ > "$ZEN_TEST_PID_DIR/pnpm.pid"
exec sleep 30
`)

	mustWriteFile(t, filepath.Join(dir, "bin", "go"), `#!/bin/sh
echo $$ > "$ZEN_TEST_PID_DIR/go.pid"
exec sleep 30
`)

	mustWriteFile(t, filepath.Join(dir, "bin", "node"), `#!/bin/sh
echo $$ > "$ZEN_TEST_PID_DIR/node.pid"
exec sleep 30
`)

	for _, name := range []string{"pnpm", "go", "node"} {
		if err := os.Chmod(filepath.Join(dir, "bin", name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunStartStopsManagedProcessesOnInterrupt(t *testing.T) {
	if os.Getenv("ZEN_TEST_RUN_START_INTERRUPT_HELPER") == "1" {
		runInterruptHelper(t, "start")
		return
	}

	dir := t.TempDir()
	writeInterruptTestStartProject(t, dir, 49174)

	cmd := exec.Command(os.Args[0], "-test.run", "^TestRunStartStopsManagedProcessesOnInterrupt$")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"ZEN_TEST_RUN_START_INTERRUPT_HELPER=1",
		"ZEN_TEST_PID_DIR="+dir,
		"PATH="+filepath.Join(dir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	pids := []int{
		waitForPIDFile(t, filepath.Join(dir, "node.pid")),
		waitForPIDFile(t, filepath.Join(dir, "app.pid")),
	}
	for _, pid := range pids {
		defer killProcessIfRunning(pid)
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("helper exited with error: %v\n%s", err, out.String())
		}
	case <-time.After(4 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("timed out waiting for interrupted zen start to exit\n%s", out.String())
	}

	for _, pid := range pids {
		if err := waitForProcessExit(pid, 2*time.Second); err != nil {
			t.Fatal(err)
		}
	}
}

func runInterruptHelper(t *testing.T, command string) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{command}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(%s): %v\nstdout:\n%s\nstderr:\n%s", command, err, stdout.String(), stderr.String())
	}
}

func writeInterruptTestStartProject(t *testing.T, dir string, port int) {
	t.Helper()

	mustMkdirAll(t, filepath.Join(dir, "frontend"))
	mustMkdirAll(t, filepath.Join(dir, "bin"))
	mustWriteFile(t, filepath.Join(dir, "zen.config.json"), fmt.Sprintf(`{
  "zenConfigVersion": 1,
  "frontendDir": "frontend",
  "prodRendererPort": %d,
  "binaryPath": "./bin/app"
}
`, port))

	mustWriteFile(t, filepath.Join(dir, "bin", "node"), `#!/bin/sh
echo $$ > "$ZEN_TEST_PID_DIR/node.pid"
exec sleep 30
`)

	mustWriteFile(t, filepath.Join(dir, "bin", "app"), `#!/bin/sh
echo $$ > "$ZEN_TEST_PID_DIR/app.pid"
exec sleep 30
`)

	for _, name := range []string{"node", "app"} {
		if err := os.Chmod(filepath.Join(dir, "bin", name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
