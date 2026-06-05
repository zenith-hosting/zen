# Split Zen And Zencli Modules Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `github.com/zenith-hosting/zen` a standalone library dependency and `github.com/zenith-hosting/zencli` a separate CLI project.

**Architecture:** The repository root is the library module with module path `github.com/zenith-hosting/zen`. The CLI package remains in `zencli/`, but that directory becomes a separate Go module with module path `github.com/zenith-hosting/zencli`, and its `cmd/zen` binary entrypoint moves inside that module. A root `go.work` ties the local modules and examples together.

**Tech Stack:** Go modules/workspaces, Fiber v3, existing Node renderer tests.

---

## Chunk 1: Module Split

### Task 1: Lock Desired Import Paths

**Files:**
- Modify: `zencli/init_templates_test.go`
- Modify: `examples/basic/main.go`
- Modify: `examples/todo/main.go`

- [ ] **Step 1: Write failing tests/import expectations**

Update starter-template tests to expect `github.com/zenith-hosting/zen`, not the old nested import path.

- [ ] **Step 2: Run focused test to verify failure**

Run: `go test ./zencli -run TestStarterFilesMainGoImportsZen -v`
Expected: FAIL because the starter template still imports `github.com/zenith/zen/zen`.

- [ ] **Step 3: Update source imports**

Change application and starter imports from `github.com/zenith/zen/zen` to `github.com/zenith-hosting/zen`.

- [ ] **Step 4: Run focused test to verify pass**

Run: `go test ./zencli -run TestStarterFilesMainGoImportsZen -v`
Expected: PASS.

### Task 2: Create Separate Go Modules

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `zencli/go.mod`
- Create: `go.work`
- Move: `cmd/zen/main.go` to `zencli/main.go`

- [ ] **Step 1: Write failing workspace/module verification**

Run: `go test ./...`
Expected: FAIL once root module no longer owns all packages, until `go.work` and per-module files are in place.

- [ ] **Step 2: Add module files**

Keep root `go.mod` as `module github.com/zenith-hosting/zen`. Create `zencli/go.mod` with `module github.com/zenith-hosting/zencli`. Create root `go.work` covering `.`, `./zencli`, and examples.

- [ ] **Step 3: Move CLI entrypoint**

Move `cmd/zen/main.go` into `zencli/main.go` and update its import to `github.com/zenith-hosting/zencli/internal/zencli`.

- [ ] **Step 4: Keep library internal test helper under root module**

Keep `internal/testutil` under the root library module and update library tests to import `github.com/zenith-hosting/zen/internal/testutil`.

- [ ] **Step 5: Verify modules**

Run:
- `go test ./...`
- `go test ./zencli/...`
- `go test ./examples/basic/...`
- `go test ./examples/todo/...`

Expected: all PASS.

### Task 3: Update Examples And Templates

**Files:**
- Modify: `examples/basic/go.mod`
- Modify: `examples/todo/go.mod`
- Modify: `zencli/init_template/go.mod.template`
- Modify: `zencli/init_templates_test.go`
- Modify: `examples/basic/zen`
- Modify: `examples/todo/zen`

- [ ] **Step 1: Update module dependencies**

Examples and starter templates should require `github.com/zenith-hosting/zen`. Examples use local `replace github.com/zenith-hosting/zen => ../..`.

- [ ] **Step 2: Update example CLI wrappers**

Example `zen` scripts should run the CLI entrypoint from `../../zencli`.

- [ ] **Step 3: Verify init template tests**

Run: `go test ./zencli -run 'TestStarterFiles' -v`
Expected: PASS.

### Task 4: Final Verification

**Files:**
- All changed files.

- [ ] **Step 1: Run Go module tests**

Run:
- `go test ./...`
- `go test ./zencli/...`
- `go test ./examples/basic/...`
- `go test ./examples/todo/...`

Expected: all PASS.

- [ ] **Step 2: Run Node renderer tests**

Run: `node --test js/renderers/*.test.mjs`
Expected: PASS.

- [ ] **Step 3: Build CLI**

Run: `go build -o ./bin/zen ./zencli`
Expected: build exits 0.
