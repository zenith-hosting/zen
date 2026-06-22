//go:build !windows

package zencli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestManagedProcessStopsDescendantsOnContextCancel(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "child.pid")
	ctx, cancel := context.WithCancel(context.Background())
	var out strings.Builder

	proc := ManagedProcess{
		Name: "test",
		Command: ProcessCommand{
			Name: "sh",
			Args: []string{"-c", fmt.Sprintf("sleep 30 & echo $! > %s; wait", strconv.Quote(pidPath))},
		},
		Stdout: &out,
		Stderr: &out,
	}

	errs := make(chan error, 1)
	go func() {
		errs <- proc.Run(ctx)
	}()

	pid := waitForPIDFile(t, pidPath)
	defer killProcessIfRunning(pid)

	cancel()

	select {
	case err := <-errs:
		if err == nil {
			t.Fatal("expected cancellation to stop the process")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for managed process to stop")
	}

	if err := waitForProcessExit(pid, 2*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestRunManagedProcessesStopsSiblingsBeforeReturning(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "child.pid")
	var out strings.Builder

	plan := []NamedProcessCommand{
		{
			Name: "fast",
			Command: ProcessCommand{
				Name: "sh",
				Args: []string{"-c", "sleep 0.1; exit 0"},
			},
		},
		{
			Name: "slow",
			Command: ProcessCommand{
				Name: "sh",
				Args: []string{"-c", fmt.Sprintf("sleep 30 & echo $! > %s; wait", strconv.Quote(pidPath))},
			},
		},
	}

	errs := make(chan error, 1)
	go func() {
		errs <- runManagedProcesses(context.Background(), plan, &out, &out, func(context.Context) error {
			waitForPIDFile(t, pidPath)
			return nil
		})
	}()

	pid := waitForPIDFile(t, pidPath)
	defer killProcessIfRunning(pid)

	select {
	case err := <-errs:
		if err != nil {
			t.Fatalf("expected nil error when first process exits cleanly, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for supervisor to stop sibling processes")
	}

	if err := waitForProcessExit(pid, 2*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestRunManagedProcessesCancelsWhileAfterStartIsWaiting(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "child.pid")
	var out strings.Builder

	plan := []NamedProcessCommand{
		{
			Name: "fast",
			Command: ProcessCommand{
				Name: "sh",
				Args: []string{"-c", "sleep 0.1; exit 0"},
			},
		},
		{
			Name: "slow",
			Command: ProcessCommand{
				Name: "sh",
				Args: []string{"-c", fmt.Sprintf("sleep 30 & echo $! > %s; wait", strconv.Quote(pidPath))},
			},
		},
	}

	errs := make(chan error, 1)
	go func() {
		errs <- runManagedProcesses(context.Background(), plan, &out, &out, func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
	}()

	pid := waitForPIDFile(t, pidPath)
	defer killProcessIfRunning(pid)

	select {
	case err := <-errs:
		if err != nil {
			t.Fatalf("expected nil error when the first process exits cleanly, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for supervisor to cancel while startup hook was blocked")
	}

	if err := waitForProcessExit(pid, 2*time.Second); err != nil {
		t.Fatal(err)
	}
}

func waitForPIDFile(t *testing.T, pidPath string) int {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(pidPath)
		if err == nil {
			pidText := strings.TrimSpace(string(data))
			if pidText == "" {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			pid, err := strconv.Atoi(pidText)
			if err != nil {
				t.Fatalf("parse pid %q: %v", pidText, err)
			}

			return pid
		}

		if !os.IsNotExist(err) {
			t.Fatalf("read pid file: %v", err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for pid file %s", pidPath)
	return 0
}

func processExists(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func killProcessIfRunning(pid int) {
	if processExists(pid) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
}

func waitForProcessExit(pid int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			return nil
		}

		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("expected process %d to be stopped", pid)
}
