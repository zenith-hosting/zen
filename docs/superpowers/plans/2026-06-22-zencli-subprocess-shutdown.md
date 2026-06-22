# Zen CLI Subprocess Shutdown Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `zen dev` and `zen start` stop every managed subprocess immediately when any subprocess exits or crashes, so reruns do not hit stale-process port collisions.

**Architecture:** Keep the existing `devPlan` and `startPlan` structure, but harden supervision in two places. First, make `ManagedProcess.Run` own a full subprocess group so shell-launched descendants are torn down when the direct command exits or its context is canceled. Second, centralize multi-process supervision so the first exiting process cancels the rest and the CLI waits for every managed process to stop before returning.

**Tech Stack:** Go, `os/exec`, Unix process groups, existing `zencli` unit tests.

---

### Task 1: Reproduce the descendant-leak behavior in tests

**Files:**
- Modify: `zencli/internal/zencli/process_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestManagedProcessStopsDescendantsOnContextCancel(t *testing.T) {
	t.Helper()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd zencli && go test ./internal/zencli -run TestManagedProcessStopsDescendantsOnContextCancel`
Expected: FAIL or hang because the shell child survives cancellation and keeps the pipes open.

- [ ] **Step 3: Expand the test to record the spawned child PID and assert the process exits promptly after cancellation**

```go
pidPath := filepath.Join(t.TempDir(), "child.pid")
proc := ManagedProcess{
	Name: "test",
	Command: ProcessCommand{
		Name: "sh",
		Args: []string{"-c", "sleep 30 & echo $! > '" + pidPath + "'; wait"},
	},
}
```

- [ ] **Step 4: Run test to verify it still fails for the right reason**

Run: `cd zencli && go test ./internal/zencli -run TestManagedProcessStopsDescendantsOnContextCancel`
Expected: FAIL due to timeout or a still-running child process.

### Task 2: Reproduce the supervisor behavior when one subprocess exits

**Files:**
- Modify: `zencli/internal/zencli/process_test.go`

- [ ] **Step 1: Write the failing test for shared supervision**

```go
func TestRunManagedProcessesStopsSiblingsBeforeReturning(t *testing.T) {
	t.Helper()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd zencli && go test ./internal/zencli -run TestRunManagedProcessesStopsSiblingsBeforeReturning`
Expected: FAIL because no shared supervisor waits for the surviving subprocess tree to stop.

- [ ] **Step 3: Build the scenario with one short-lived process and one long-lived shell process**

```go
plan := []NamedProcessCommand{
	{Name: "fast", Command: ProcessCommand{Name: "sh", Args: []string{"-c", "exit 0"}}},
	{Name: "slow", Command: ProcessCommand{Name: "sh", Args: []string{"-c", "sleep 30 & echo $! > '" + pidPath + "'; wait"}}},
}
```

- [ ] **Step 4: Run test to verify it still fails for the right reason**

Run: `cd zencli && go test ./internal/zencli -run TestRunManagedProcessesStopsSiblingsBeforeReturning`
Expected: FAIL or not compile until the shared supervisor helper exists.

### Task 3: Implement process-group teardown and coordinated supervision

**Files:**
- Modify: `zencli/internal/zencli/process.go`
- Modify: `zencli/internal/zencli/dev.go`
- Modify: `zencli/internal/zencli/start.go`
- Create: `zencli/internal/zencli/process_unix.go`
- Create: `zencli/internal/zencli/process_other.go`

- [ ] **Step 1: Add platform helpers for command process-group setup and group termination**

```go
func configureManagedCommand(cmd *exec.Cmd)
func terminateManagedProcess(cmd *exec.Cmd) error
```

- [ ] **Step 2: Update `ManagedProcess.Run` to configure the command, wait for the direct process, then tear down any remaining descendants before draining log pipes**

```go
waitErr := cmd.Wait()
terminateErr := terminateManagedProcess(cmd)
```

- [ ] **Step 3: Add a shared supervisor that starts all managed processes, cancels siblings on first exit, waits for all goroutines to finish, and returns the triggering error**

```go
func runManagedProcesses(ctx context.Context, plan []NamedProcessCommand, stdout, stderr io.Writer, afterStart func(context.Context) error) error
```

- [ ] **Step 4: Switch `runDev` and `runStart` to the shared supervisor and keep their existing renderer health checks in the `afterStart` callback**

Run: `cd zencli && go test ./internal/zencli -run 'TestManagedProcessStopsDescendantsOnContextCancel|TestRunManagedProcessesStopsSiblingsBeforeReturning'`
Expected: PASS

### Task 4: Verify the full zencli package

**Files:**
- No code changes expected

- [ ] **Step 1: Run zencli tests**

Run: `cd zencli && go test ./...`
Expected: PASS

- [ ] **Step 2: Review local diff for scope**

Run: `git diff -- zencli/internal/zencli docs/superpowers/plans/2026-06-22-zencli-subprocess-shutdown.md`
Expected: Only the supervision fix, tests, and plan file are present.
