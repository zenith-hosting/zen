# HTTP Renderer Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the fragile stdin/stdout Node renderer bridge with a boring HTTP renderer service that supports concurrent render requests, timeouts, readable errors, health checks, and normal debugging.

**Architecture:** Fiber remains the public app server. Zen’s Go renderer calls a long-lived Node renderer over HTTP using keep-alive connections. The Node renderer exposes `POST /__zen/render` and `GET /__zen/health`; dev mode loads Preact SSR through Vite, while production mode loads the built SSR bundle directly.

**Tech Stack:** Go, Fiber, `net/http`, Vite, Preact, Node HTTP server, TypeScript/TSX SSR entry, Tailwind.

---

This follows the uploaded planning instructions: start with file structure, give exact paths, use bite-sized checklist steps, write failing tests first, include real implementation snippets, and end with a self-review. 

## Migration Summary

Current bad thing:

```text
Go Fiber route
  -> Zen renderer
  -> Node child process
  -> stdin/stdout JSON messages
  -> sadness every second request
```

Target good-enough thing:

```text
Go Fiber route
  -> Zen renderer
  -> HTTP keep-alive POST /__zen/render
  -> Node renderer service
  -> Vite dev SSR or built Preact SSR bundle
```

The migration keeps the internal render interface stable:

```go
type ssrClient interface {
	Render(ctx context.Context, req ssrRequest) (ssrResponse, error)
}
```

Then swaps the implementation from pipe slop to HTTP slop. Still slop, obviously, but now at least the kind that debuggers, curl, logs, timeouts, and concurrent requests understand.

---

## File Structure

```text
zen/
  config.go
  config_test.go

  render.go
  render_test.go

  ssr_client.go
  ssr_client_test.go

  ssr_http_client.go
  ssr_http_client_test.go

  manifest.go
  document.go
  escape.go
  static.go

js/
  renderer-shared.mjs
  renderer-shared.test.mjs
  dev-renderer.mjs
  dev-renderer.test.mjs
  prod-renderer.mjs
  prod-renderer.test.mjs

  fixtures/
    entry-server-ok.mjs
    entry-server-error.mjs

examples/
  basic/
    main.go
    package.json

    frontend/
      package.json
      vite.config.ts
      src/
        entry-server.tsx
        entry-client.tsx
        pages.ts
```

## Responsibility Map

| File                          | Responsibility                                                      |
| ----------------------------- | ------------------------------------------------------------------- |
| `zen/ssr_client.go`           | Shared request/response structs and `ssrClient` interface.          |
| `zen/ssr_http_client.go`      | HTTP implementation of `ssrClient`.                                 |
| `zen/config.go`               | Replace `SSRCommand` with `RenderURL` and timeout settings.         |
| `zen/render.go`               | Use HTTP client during `New()` and render through it.               |
| `js/renderer-shared.mjs`      | Shared JSON helpers for Node renderer servers.                      |
| `js/dev-renderer.mjs`         | Dev renderer server using Vite middleware and `vite.ssrLoadModule`. |
| `js/prod-renderer.mjs`        | Production renderer server importing built SSR bundle once.         |
| `examples/basic/main.go`      | Point Zen at `RenderURL`, no child process management.              |
| `examples/basic/package.json` | Run Go app and Node renderer as explicit processes.                 |

---

## Task 1: Freeze the Existing SSR Contract

**Files:**

* Modify: `zen/ssr_client.go`

* Modify: `zen/ssr_client_test.go`

* [ ] **Step 1: Update `zen/ssr_client.go` to contain only the shared contract**

```go
package zen

import "context"

type ssrClient interface {
	Render(ctx context.Context, req ssrRequest) (ssrResponse, error)
}

type ssrRequest struct {
	URL   string `json:"url"`
	Page  string `json:"page"`
	Props any    `json:"props"`
}

type ssrResponse struct {
	HTML string `json:"html"`
	Head string `json:"head"`
}
```

* [ ] **Step 2: Update `zen/ssr_client_test.go`**

```go
package zen

import (
	"context"
	"testing"
)

type fakeSSRClient struct {
	req ssrRequest
	res ssrResponse
	err error
}

func (f *fakeSSRClient) Render(ctx context.Context, req ssrRequest) (ssrResponse, error) {
	f.req = req
	return f.res, f.err
}

func TestSSRClientInterfaceCapturesRenderRequest(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{HTML: "<main>Hello</main>"},
	}

	res, err := client.Render(context.Background(), ssrRequest{
		URL:   "/",
		Page:  "Home",
		Props: map[string]string{"title": "Hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.req.Page != "Home" {
		t.Fatalf("expected page Home, got %q", client.req.Page)
	}

	if res.HTML != "<main>Hello</main>" {
		t.Fatalf("expected rendered html, got %q", res.HTML)
	}
}
```

* [ ] **Step 3: Run the contract test**

```bash
go test ./zen -run 'TestSSRClientInterfaceCapturesRenderRequest' -v
```

Expected: PASS.

* [ ] **Step 4: Commit**

```bash
git add zen/ssr_client.go zen/ssr_client_test.go
git commit -m "refactor: isolate ssr client contract"
```

---

## Task 2: Add HTTP Renderer Client Tests

**Files:**

* Create: `zen/ssr_http_client.go`

* Create: `zen/ssr_http_client_test.go`

* [ ] **Step 1: Write failing tests**

Create `zen/ssr_http_client_test.go`:

```go
package zen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPSSRClientRendersPage(t *testing.T) {
	var received ssrRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/__zen/render" {
			t.Fatalf("expected /__zen/render, got %s", r.URL.Path)
		}

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(ssrResponse{
			HTML: "<main>Hello</main>",
			Head: "<title>Hello</title>",
		})
	}))
	defer server.Close()

	client := newHTTPSSRClient(httpSSRClientConfig{
		RenderURL: server.URL + "/__zen/render",
		Timeout:   time.Second,
	})

	res, err := client.Render(context.Background(), ssrRequest{
		URL:   "/",
		Page:  "Home",
		Props: map[string]string{"title": "Hello"},
	})
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}

	if received.Page != "Home" {
		t.Fatalf("expected page Home, got %q", received.Page)
	}

	if res.HTML != "<main>Hello</main>" {
		t.Fatalf("expected rendered html, got %q", res.HTML)
	}

	if res.Head != "<title>Hello</title>" {
		t.Fatalf("expected head html, got %q", res.Head)
	}
}

func TestHTTPSSRClientReturnsRendererError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		_ = json.NewEncoder(w).Encode(httpRendererErrorResponse{
			Error: httpRendererError{
				Message: "Unknown page: Admin",
				Stack:   "Error: Unknown page: Admin",
			},
		})
	}))
	defer server.Close()

	client := newHTTPSSRClient(httpSSRClientConfig{
		RenderURL: server.URL,
		Timeout:   time.Second,
	})

	_, err := client.Render(context.Background(), ssrRequest{
		URL:   "/admin",
		Page:  "Admin",
		Props: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected render error")
	}

	if !strings.Contains(err.Error(), "Unknown page: Admin") {
		t.Fatalf("expected renderer error message, got %v", err)
	}
}

func TestHTTPSSRClientHandlesParallelRequests(t *testing.T) {
	var count atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)

		var req ssrRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(ssrResponse{
			HTML: "<main>" + req.Page + "</main>",
		})
	}))
	defer server.Close()

	client := newHTTPSSRClient(httpSSRClientConfig{
		RenderURL: server.URL,
		Timeout:   time.Second,
	})

	errs := make(chan error, 25)

	for i := 0; i < 25; i++ {
		go func() {
			_, err := client.Render(context.Background(), ssrRequest{
				URL:   "/",
				Page:  "Home",
				Props: map[string]string{},
			})
			errs <- err
		}()
	}

	for i := 0; i < 25; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("parallel render failed: %v", err)
		}
	}

	if count.Load() != 25 {
		t.Fatalf("expected 25 requests, got %d", count.Load())
	}
}

func TestHTTPSSRClientRespectsContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newHTTPSSRClient(httpSSRClientConfig{
		RenderURL: server.URL,
		Timeout:   time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	_, err := client.Render(ctx, ssrRequest{
		URL:   "/",
		Page:  "Home",
		Props: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
```

* [ ] **Step 2: Run tests to verify failure**

```bash
go test ./zen -run 'TestHTTPSSRClient' -v
```

Expected: FAIL because `newHTTPSSRClient`, `httpSSRClientConfig`, and renderer error types are undefined.

* [ ] **Step 3: Implement HTTP renderer client**

Create `zen/ssr_http_client.go`:

```go
package zen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type httpSSRClientConfig struct {
	RenderURL string
	Timeout   time.Duration
}

type httpSSRClient struct {
	renderURL string
	client    *http.Client
}

type httpRendererErrorResponse struct {
	Error httpRendererError `json:"error"`
}

type httpRendererError struct {
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

func newHTTPSSRClient(config httpSSRClientConfig) *httpSSRClient {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	return &httpSSRClient{
		renderURL: config.RenderURL,
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

func (c *httpSSRClient) Render(ctx context.Context, req ssrRequest) (ssrResponse, error) {
	if c.renderURL == "" {
		return ssrResponse{}, errors.New("zen: renderer RenderURL is required")
	}

	raw, err := json.Marshal(req)
	if err != nil {
		return ssrResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.renderURL, bytes.NewReader(raw))
	if err != nil {
		return ssrResponse{}, err
	}

	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("accept", "application/json")

	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return ssrResponse{}, err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode < 200 || httpRes.StatusCode >= 300 {
		var errorBody httpRendererErrorResponse
		if err := json.NewDecoder(httpRes.Body).Decode(&errorBody); err != nil {
			return ssrResponse{}, fmt.Errorf("zen renderer returned status %d", httpRes.StatusCode)
		}

		if errorBody.Error.Message == "" {
			return ssrResponse{}, fmt.Errorf("zen renderer returned status %d", httpRes.StatusCode)
		}

		return ssrResponse{}, fmt.Errorf("zen renderer: %s", errorBody.Error.Message)
	}

	var out ssrResponse
	if err := json.NewDecoder(httpRes.Body).Decode(&out); err != nil {
		return ssrResponse{}, err
	}

	return out, nil
}
```

* [ ] **Step 4: Run tests**

```bash
go test ./zen -run 'TestHTTPSSRClient' -v
```

Expected: PASS.

* [ ] **Step 5: Commit**

```bash
git add zen/ssr_http_client.go zen/ssr_http_client_test.go
git commit -m "feat: add http ssr client"
```

---

## Task 3: Replace `SSRCommand` Config With `RenderURL`

**Files:**

* Modify: `zen/config.go`

* Modify: `zen/config_test.go`

* [ ] **Step 1: Replace config tests**

Update `zen/config_test.go`:

```go
package zen

import (
	"testing"
	"time"
)

func TestConfigWithDefaultsDev(t *testing.T) {
	cfg := Config{
		Dev: true,
	}

	got := cfg.withDefaults()

	if got.ViteURL != "http://localhost:5173" {
		t.Fatalf("expected default ViteURL, got %q", got.ViteURL)
	}

	if got.RenderURL != "http://localhost:5173/__zen/render" {
		t.Fatalf("expected default RenderURL, got %q", got.RenderURL)
	}

	if got.AppElementID != "app" {
		t.Fatalf("expected app element id app, got %q", got.AppElementID)
	}

	if got.DataElementID != "__ZEN_DATA__" {
		t.Fatalf("expected data element id __ZEN_DATA__, got %q", got.DataElementID)
	}

	if got.RenderTimeout != 5*time.Second {
		t.Fatalf("expected render timeout 5s, got %s", got.RenderTimeout)
	}
}

func TestConfigWithDefaultsProduction(t *testing.T) {
	cfg := Config{
		Dev:       false,
		RenderURL: "http://127.0.0.1:4174/__zen/render",
	}

	got := cfg.withDefaults()

	if got.RenderURL != "http://127.0.0.1:4174/__zen/render" {
		t.Fatalf("expected configured RenderURL, got %q", got.RenderURL)
	}

	if got.RenderTimeout != 5*time.Second {
		t.Fatalf("expected render timeout 5s, got %s", got.RenderTimeout)
	}
}

func TestConfigValidateProductionRequiresPaths(t *testing.T) {
	cfg := Config{
		Dev:       false,
		RenderURL: "http://127.0.0.1:4174/__zen/render",
	}

	err := cfg.withDefaults().validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConfigValidateRequiresRenderURL(t *testing.T) {
	cfg := Config{
		Dev:       true,
		ViteURL:   "http://localhost:5173",
		RenderURL: "",
	}

	err := cfg.withDefaults().validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConfigValidateDevRequiresViteURL(t *testing.T) {
	cfg := Config{
		Dev:     true,
		ViteURL: "",
	}

	err := cfg.withDefaults().validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}
```

* [ ] **Step 2: Run tests to verify failure**

```bash
go test ./zen -run 'TestConfig' -v
```

Expected: FAIL because `RenderURL` and `RenderTimeout` do not exist.

* [ ] **Step 3: Update `zen/config.go`**

```go
package zen

import (
	"errors"
	"strings"
	"time"
)

type Config struct {
	Dev bool

	ViteURL   string
	RenderURL string

	ClientDist string
	Manifest   string

	RenderTimeout time.Duration

	AppElementID  string
	DataElementID string
	DefaultTitle  string
}

func (c Config) withDefaults() Config {
	if c.ViteURL == "" && c.Dev {
		c.ViteURL = "http://localhost:5173"
	}

	if c.RenderURL == "" && c.Dev && c.ViteURL != "" {
		c.RenderURL = strings.TrimRight(c.ViteURL, "/") + "/__zen/render"
	}

	if c.AppElementID == "" {
		c.AppElementID = "app"
	}

	if c.DataElementID == "" {
		c.DataElementID = "__ZEN_DATA__"
	}

	if c.DefaultTitle == "" {
		c.DefaultTitle = "Zen"
	}

	if c.RenderTimeout == 0 {
		c.RenderTimeout = 5 * time.Second
	}

	return c
}

func (c Config) validate() error {
	if c.AppElementID == "" {
		return errors.New("zen: AppElementID is required")
	}

	if c.DataElementID == "" {
		return errors.New("zen: DataElementID is required")
	}

	if strings.TrimSpace(c.RenderURL) == "" {
		return errors.New("zen: RenderURL is required")
	}

	if c.RenderTimeout <= 0 {
		return errors.New("zen: RenderTimeout must be greater than zero")
	}

	if c.Dev {
		if strings.TrimSpace(c.ViteURL) == "" {
			return errors.New("zen: ViteURL is required in dev mode")
		}
		return nil
	}

	if strings.TrimSpace(c.ClientDist) == "" {
		return errors.New("zen: ClientDist is required in production mode")
	}

	if strings.TrimSpace(c.Manifest) == "" {
		return errors.New("zen: Manifest is required in production mode")
	}

	return nil
}
```

* [ ] **Step 4: Run tests**

```bash
go test ./zen -run 'TestConfig' -v
```

Expected: PASS.

* [ ] **Step 5: Commit**

```bash
git add zen/config.go zen/config_test.go
git commit -m "refactor: configure renderer over http"
```

---

## Task 4: Wire HTTP Client Into Renderer Constructor

**Files:**

* Modify: `zen/render.go`

* Modify: `zen/render_test.go`

* [ ] **Step 1: Update constructor tests**

Update the constructor-related tests in `zen/render_test.go`:

```go
func TestNewRendererAppliesDefaults(t *testing.T) {
	r, err := New(Config{
		Dev: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.config.ViteURL != "http://localhost:5173" {
		t.Fatalf("expected default vite url, got %q", r.config.ViteURL)
	}

	if r.config.RenderURL != "http://localhost:5173/__zen/render" {
		t.Fatalf("expected default render url, got %q", r.config.RenderURL)
	}

	if r.ssr == nil {
		t.Fatal("expected renderer to create ssr client")
	}
}

func TestNewRendererRejectsInvalidProductionConfig(t *testing.T) {
	_, err := New(Config{
		Dev: false,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewRendererCreatesProductionHTTPSSRClient(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")

	err := os.WriteFile(manifestPath, []byte(`{
		"src/entry-client.tsx": {
			"file": "assets/entry-client.js"
		}
	}`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	r, err := New(Config{
		Dev:        false,
		RenderURL:  "http://127.0.0.1:4174/__zen/render",
		ClientDist: dir,
		Manifest:   manifestPath,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.ssr == nil {
		t.Fatal("expected production ssr client")
	}
}
```

* [ ] **Step 2: Run tests to verify failure**

```bash
go test ./zen -run 'TestNewRenderer' -v
```

Expected: FAIL where constructor still expects process-based renderer slop.

* [ ] **Step 3: Update `New` in `zen/render.go`**

Replace the `New` function:

```go
func New(config Config) (*Renderer, error) {
	cfg := config.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	r := &Renderer{
		config: cfg,
		ssr: newHTTPSSRClient(httpSSRClientConfig{
			RenderURL: cfg.RenderURL,
			Timeout:   cfg.RenderTimeout,
		}),
	}

	if !cfg.Dev {
		manifest, err := readManifest(cfg.Manifest)
		if err != nil {
			return nil, err
		}
		r.manifest = manifest
	}

	return r, nil
}
```

* [ ] **Step 4: Remove process close expectations**

If `Renderer.Close()` only existed to kill the Node child process, replace it with a no-op:

```go
func (r *Renderer) Close() error {
	return nil
}
```

The renderer service is now explicit process orchestration, not a child process hidden under Go’s floorboards. Good. One less haunted basement.

* [ ] **Step 5: Run renderer tests**

```bash
go test ./zen -run 'TestNewRenderer|TestRender' -v
```

Expected: PASS.

* [ ] **Step 6: Commit**

```bash
git add zen/render.go zen/render_test.go
git commit -m "feat: use http ssr client in renderer"
```

---

## Task 5: Add Shared Node Renderer Helpers

**Files:**

* Create: `js/renderer-shared.mjs`

* Create: `js/renderer-shared.test.mjs`

* [ ] **Step 1: Write failing tests**

Create `js/renderer-shared.test.mjs`:

```js
import test from "node:test";
import assert from "node:assert/strict";
import { Readable } from "node:stream";
import {
  readJSON,
  writeJSON,
  writeRendererError,
  createHealthResponse
} from "./renderer-shared.mjs";

function mockResponse() {
  return {
    statusCode: 0,
    headers: {},
    body: "",
    setHeader(name, value) {
      this.headers[name.toLowerCase()] = value;
    },
    end(value) {
      this.body = value;
    }
  };
}

test("readJSON parses request body", async () => {
  const req = Readable.from([
    JSON.stringify({
      page: "Home",
      props: {
        title: "Hello"
      }
    })
  ]);

  const got = await readJSON(req);

  assert.equal(got.page, "Home");
  assert.equal(got.props.title, "Hello");
});

test("writeJSON writes JSON response", () => {
  const res = mockResponse();

  writeJSON(res, 201, {
    ok: true
  });

  assert.equal(res.statusCode, 201);
  assert.equal(res.headers["content-type"], "application/json");
  assert.equal(res.body, '{"ok":true}');
});

test("writeRendererError writes structured error", () => {
  const res = mockResponse();
  const error = new Error("render failed");

  writeRendererError(res, 500, error, {
    includeStack: true
  });

  const body = JSON.parse(res.body);

  assert.equal(res.statusCode, 500);
  assert.equal(body.error.message, "render failed");
  assert.match(body.error.stack, /render failed/);
});

test("createHealthResponse includes mode", () => {
  const got = createHealthResponse("dev");

  assert.deepEqual(got, {
    ok: true,
    mode: "dev"
  });
});
```

* [ ] **Step 2: Run tests to verify failure**

```bash
node --test js/renderer-shared.test.mjs
```

Expected: FAIL because `js/renderer-shared.mjs` does not exist.

* [ ] **Step 3: Implement helpers**

Create `js/renderer-shared.mjs`:

```js
export async function readJSON(req) {
  let body = "";

  for await (const chunk of req) {
    body += chunk.toString("utf8");
  }

  if (!body.trim()) {
    return {};
  }

  return JSON.parse(body);
}

export function writeJSON(res, status, value) {
  res.statusCode = status;
  res.setHeader("content-type", "application/json");
  res.end(JSON.stringify(value));
}

export function writeRendererError(res, status, error, options = {}) {
  const includeStack = Boolean(options.includeStack);

  writeJSON(res, status, {
    error: {
      message: error && error.message ? error.message : String(error),
      stack: includeStack && error && error.stack ? error.stack : ""
    }
  });
}

export function createHealthResponse(mode) {
  return {
    ok: true,
    mode
  };
}

export function isRenderRequest(req) {
  return req.method === "POST" && req.url === "/__zen/render";
}

export function isHealthRequest(req) {
  return req.method === "GET" && req.url === "/__zen/health";
}
```

* [ ] **Step 4: Run tests**

```bash
node --test js/renderer-shared.test.mjs
```

Expected: PASS.

* [ ] **Step 5: Commit**

```bash
git add js/renderer-shared.mjs js/renderer-shared.test.mjs
git commit -m "feat: add shared node renderer helpers"
```

---

## Task 6: Add Production Node Renderer Server

**Files:**

* Create: `js/prod-renderer.mjs`

* Create: `js/prod-renderer.test.mjs`

* Ensure exists: `js/fixtures/entry-server-ok.mjs`

* Ensure exists: `js/fixtures/entry-server-error.mjs`

* [ ] **Step 1: Ensure fixtures exist**

Create `js/fixtures/entry-server-ok.mjs`:

```js
export async function render(request) {
  return {
    html: `<main data-page="${request.page}">${request.props.title}</main>`,
    head: ""
  };
}
```

Create `js/fixtures/entry-server-error.mjs`:

```js
export async function render() {
  throw new Error("fixture render failed");
}
```

* [ ] **Step 2: Write failing production renderer tests**

Create `js/prod-renderer.test.mjs`:

```js
import test from "node:test";
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { once } from "node:events";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const here = dirname(fileURLToPath(import.meta.url));
const serverPath = join(here, "prod-renderer.mjs");
const okEntry = join(here, "fixtures", "entry-server-ok.mjs");
const errorEntry = join(here, "fixtures", "entry-server-error.mjs");

function startRenderer(entry, port) {
  return spawn(process.execPath, [
    serverPath,
    "--entry",
    entry,
    "--host",
    "127.0.0.1",
    "--port",
    String(port)
  ], {
    stdio: ["ignore", "pipe", "pipe"]
  });
}

async function waitForHealth(port) {
  const url = `http://127.0.0.1:${port}/__zen/health`;

  for (let i = 0; i < 50; i++) {
    try {
      const res = await fetch(url);
      if (res.ok) {
        return;
      }
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 25));
    }
  }

  throw new Error(`renderer did not become healthy on port ${port}`);
}

test("prod renderer health endpoint works", async () => {
  const port = 4771;
  const child = startRenderer(okEntry, port);

  try {
    await waitForHealth(port);

    const res = await fetch(`http://127.0.0.1:${port}/__zen/health`);
    const body = await res.json();

    assert.equal(res.status, 200);
    assert.deepEqual(body, {
      ok: true,
      mode: "production"
    });
  } finally {
    child.kill();
    await once(child, "exit");
  }
});

test("prod renderer renders request", async () => {
  const port = 4772;
  const child = startRenderer(okEntry, port);

  try {
    await waitForHealth(port);

    const res = await fetch(`http://127.0.0.1:${port}/__zen/render`, {
      method: "POST",
      headers: {
        "content-type": "application/json"
      },
      body: JSON.stringify({
        url: "/",
        page: "Home",
        props: {
          title: "Hello"
        }
      })
    });

    const body = await res.json();

    assert.equal(res.status, 200);
    assert.equal(body.html, `<main data-page="Home">Hello</main>`);
  } finally {
    child.kill();
    await once(child, "exit");
  }
});

test("prod renderer returns structured render errors", async () => {
  const port = 4773;
  const child = startRenderer(errorEntry, port);

  try {
    await waitForHealth(port);

    const res = await fetch(`http://127.0.0.1:${port}/__zen/render`, {
      method: "POST",
      headers: {
        "content-type": "application/json"
      },
      body: JSON.stringify({
        url: "/",
        page: "Home",
        props: {}
      })
    });

    const body = await res.json();

    assert.equal(res.status, 500);
    assert.equal(body.error.message, "fixture render failed");
  } finally {
    child.kill();
    await once(child, "exit");
  }
});
```

* [ ] **Step 3: Run tests to verify failure**

```bash
node --test js/prod-renderer.test.mjs
```

Expected: FAIL because `js/prod-renderer.mjs` does not exist.

* [ ] **Step 4: Implement production renderer**

Create `js/prod-renderer.mjs`:

```js
import http from "node:http";
import { pathToFileURL } from "node:url";
import {
  createHealthResponse,
  isHealthRequest,
  isRenderRequest,
  readJSON,
  writeJSON,
  writeRendererError
} from "./renderer-shared.mjs";

function parseArgs(argv) {
  const args = {
    host: "127.0.0.1",
    port: 4174,
    entry: ""
  };

  for (let i = 0; i < argv.length; i++) {
    const item = argv[i];

    if (item === "--entry") {
      args.entry = argv[++i] ?? "";
      continue;
    }

    if (item === "--host") {
      args.host = argv[++i] ?? "127.0.0.1";
      continue;
    }

    if (item === "--port") {
      args.port = Number(argv[++i] ?? "4174");
      continue;
    }
  }

  if (!args.entry) {
    throw new Error("missing required --entry argument");
  }

  if (!Number.isInteger(args.port) || args.port <= 0) {
    throw new Error("port must be a positive integer");
  }

  return args;
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const entryURL = pathToFileURL(args.entry).href;
  const mod = await import(entryURL);

  if (typeof mod.render !== "function") {
    throw new Error("SSR entry must export render(request)");
  }

  const server = http.createServer(async (req, res) => {
    if (isHealthRequest(req)) {
      writeJSON(res, 200, createHealthResponse("production"));
      return;
    }

    if (isRenderRequest(req)) {
      try {
        const body = await readJSON(req);
        const result = await mod.render(body);
        writeJSON(res, 200, result);
      } catch (error) {
        writeRendererError(res, 500, error, {
          includeStack: process.env.NODE_ENV !== "production"
        });
      }

      return;
    }

    writeJSON(res, 404, {
      error: {
        message: "not found"
      }
    });
  });

  server.listen(args.port, args.host, () => {
    process.stdout.write(`Zen production renderer listening on http://${args.host}:${args.port}\n`);
  });
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
```

* [ ] **Step 5: Run tests**

```bash
node --test js/prod-renderer.test.mjs
```

Expected: PASS.

* [ ] **Step 6: Commit**

```bash
git add js/prod-renderer.mjs js/prod-renderer.test.mjs js/fixtures/entry-server-ok.mjs js/fixtures/entry-server-error.mjs
git commit -m "feat: add production http renderer"
```

---

## Task 7: Add Dev Node Renderer Server Using Vite

**Files:**

* Create: `js/dev-renderer.mjs`

* Create: `js/dev-renderer.test.mjs`

* [ ] **Step 1: Write failing dev renderer smoke test**

Create `js/dev-renderer.test.mjs`:

```js
import test from "node:test";
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { once } from "node:events";
import { mkdtemp, mkdir, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { dirname } from "node:path";

const here = dirname(fileURLToPath(import.meta.url));
const serverPath = join(here, "dev-renderer.mjs");

async function createViteFixture() {
  const root = await mkdtemp(join(tmpdir(), "zen-vite-fixture-"));
  const src = join(root, "src");

  await mkdir(src, {
    recursive: true
  });

  await writeFile(join(root, "package.json"), JSON.stringify({
    type: "module",
    dependencies: {
      vite: "^7.0.0"
    }
  }));

  await writeFile(join(src, "entry-server.js"), `
    export async function render(request) {
      return {
        html: '<main data-page="' + request.page + '">' + request.props.title + '</main>',
        head: ''
      };
    }
  `);

  return root;
}

async function waitForHealth(port) {
  const url = `http://127.0.0.1:${port}/__zen/health`;

  for (let i = 0; i < 50; i++) {
    try {
      const res = await fetch(url);
      if (res.ok) {
        return;
      }
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 25));
    }
  }

  throw new Error(`renderer did not become healthy on port ${port}`);
}

test("dev renderer renders through vite", async () => {
  const root = await createViteFixture();
  const port = 4781;

  const child = spawn(process.execPath, [
    serverPath,
    "--root",
    root,
    "--entry",
    "/src/entry-server.js",
    "--host",
    "127.0.0.1",
    "--port",
    String(port)
  ], {
    stdio: ["ignore", "pipe", "pipe"]
  });

  try {
    await waitForHealth(port);

    const res = await fetch(`http://127.0.0.1:${port}/__zen/render`, {
      method: "POST",
      headers: {
        "content-type": "application/json"
      },
      body: JSON.stringify({
        url: "/",
        page: "Home",
        props: {
          title: "Hello"
        }
      })
    });

    const body = await res.json();

    assert.equal(res.status, 200);
    assert.equal(body.html, `<main data-page="Home">Hello</main>`);
  } finally {
    child.kill();
    await once(child, "exit");
  }
});
```

* [ ] **Step 2: Run test to verify failure**

```bash
node --test js/dev-renderer.test.mjs
```

Expected: FAIL because `js/dev-renderer.mjs` does not exist.

* [ ] **Step 3: Implement dev renderer**

Create `js/dev-renderer.mjs`:

```js
import http from "node:http";
import { createServer as createViteServer } from "vite";
import {
  createHealthResponse,
  isHealthRequest,
  isRenderRequest,
  readJSON,
  writeJSON,
  writeRendererError
} from "./renderer-shared.mjs";

function parseArgs(argv) {
  const args = {
    root: process.cwd(),
    entry: "/src/entry-server.tsx",
    host: "127.0.0.1",
    port: 5173
  };

  for (let i = 0; i < argv.length; i++) {
    const item = argv[i];

    if (item === "--root") {
      args.root = argv[++i] ?? process.cwd();
      continue;
    }

    if (item === "--entry") {
      args.entry = argv[++i] ?? "/src/entry-server.tsx";
      continue;
    }

    if (item === "--host") {
      args.host = argv[++i] ?? "127.0.0.1";
      continue;
    }

    if (item === "--port") {
      args.port = Number(argv[++i] ?? "5173");
      continue;
    }
  }

  if (!Number.isInteger(args.port) || args.port <= 0) {
    throw new Error("port must be a positive integer");
  }

  return args;
}

async function main() {
  const args = parseArgs(process.argv.slice(2));

  const vite = await createViteServer({
    root: args.root,
    server: {
      middlewareMode: true
    },
    appType: "custom"
  });

  const server = http.createServer(async (req, res) => {
    if (isHealthRequest(req)) {
      writeJSON(res, 200, createHealthResponse("dev"));
      return;
    }

    if (isRenderRequest(req)) {
      try {
        const body = await readJSON(req);
        const mod = await vite.ssrLoadModule(args.entry);

        if (typeof mod.render !== "function") {
          throw new Error("SSR entry must export render(request)");
        }

        const result = await mod.render(body);
        writeJSON(res, 200, result);
      } catch (error) {
        vite.ssrFixStacktrace(error);

        writeRendererError(res, 500, error, {
          includeStack: true
        });
      }

      return;
    }

    vite.middlewares(req, res);
  });

  server.listen(args.port, args.host, () => {
    process.stdout.write(`Zen dev renderer listening on http://${args.host}:${args.port}\n`);
  });
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
```

* [ ] **Step 4: Run test**

```bash
node --test js/dev-renderer.test.mjs
```

Expected: PASS.

* [ ] **Step 5: Commit**

```bash
git add js/dev-renderer.mjs js/dev-renderer.test.mjs
git commit -m "feat: add vite dev http renderer"
```

---

## Task 8: Update Example App to Use HTTP Renderer

**Files:**

* Modify: `examples/basic/main.go`

* Modify: `examples/basic/package.json`

* [ ] **Step 1: Update `examples/basic/main.go`**

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/zenith/zen/zen"
)

func main() {
	app := fiber.New()

	dev := os.Getenv("ZEN_ENV") != "production"

	cfg := zen.Config{
		Dev:           dev,
		ViteURL:       "http://localhost:5173",
		RenderURL:     "http://localhost:5173/__zen/render",
		ClientDist:    "./frontend/dist/client",
		Manifest:      "./frontend/dist/client/.vite/manifest.json",
		DefaultTitle:  "Zen Basic Example",
		RenderTimeout: 5 * time.Second,
	}

	if !dev {
		cfg.RenderURL = "http://127.0.0.1:4174/__zen/render"
	}

	renderer, err := zen.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Close()

	if !dev {
		app.Get("/assets/*", renderer.Static())
	}

	app.Get("/", func(c fiber.Ctx) error {
		return renderer.Render(c, "Home", map[string]any{
			"title": "Zen Basic Example",
			"body":  "Fiber route, Preact page, Vite renderer. No pipe slop.",
		}, zen.WithTitle("Home"))
	})

	app.Get("/users/:id", func(c fiber.Ctx) error {
		id := c.Params("id")

		return renderer.Render(c, "User", map[string]any{
			"id": id,
		}, zen.WithTitle("User "+id))
	})

	app.Post("/contact", func(c fiber.Ctx) error {
		name := c.FormValue("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).SendString("name is required")
		}

		return c.Redirect("/")
	})

	log.Fatal(app.Listen(":3000"))
}
```

* [ ] **Step 2: Update `examples/basic/package.json`**

```json
{
  "name": "zen-basic-example",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "pnpm dev:renderer & go run .",
    "dev:renderer": "node ../../js/dev-renderer.mjs --root ./frontend --entry /src/entry-server.tsx --host 127.0.0.1 --port 5173",
    "build": "pnpm --dir frontend build && go build -o ./bin/basic .",
    "start": "ZEN_ENV=production ./bin/basic",
    "start:renderer": "node ../../js/prod-renderer.mjs --entry ./frontend/dist/server/entry-server.js --host 127.0.0.1 --port 4174"
  }
}
```

* [ ] **Step 3: Run example frontend build**

```bash
cd examples/basic
pnpm --dir frontend build
```

Expected: PASS.

* [ ] **Step 4: Run example Go build**

```bash
cd examples/basic
go build .
```

Expected: PASS.

* [ ] **Step 5: Commit**

```bash
git add examples/basic/main.go examples/basic/package.json
git commit -m "example: use http renderer service"
```

---

## Task 9: Remove Stdio Worker Slop

**Files:**

* Delete: `js/ssr-worker.mjs`

* Delete or rewrite: old process-client tests in `zen/ssr_client_test.go`

* Delete old process client implementation if it still exists in `zen/ssr_client.go`

* [ ] **Step 1: Delete old worker**

```bash
rm -f js/ssr-worker.mjs
```

* [ ] **Step 2: Search for stale `SSRCommand` references**

```bash
grep -R "SSRCommand\|ssr-worker\|newProcessSSRClient\|processSSRClient" -n .
```

Expected: matches only in old commits, not in working tree. If the command prints matches in tracked files, remove those references.

* [ ] **Step 3: Remove old process client test blocks**

Ensure `zen/ssr_client_test.go` only contains:

```go
package zen

import (
	"context"
	"testing"
)

type fakeSSRClient struct {
	req ssrRequest
	res ssrResponse
	err error
}

func (f *fakeSSRClient) Render(ctx context.Context, req ssrRequest) (ssrResponse, error) {
	f.req = req
	return f.res, f.err
}

func TestSSRClientInterfaceCapturesRenderRequest(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{HTML: "<main>Hello</main>"},
	}

	res, err := client.Render(context.Background(), ssrRequest{
		URL:   "/",
		Page:  "Home",
		Props: map[string]string{"title": "Hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.req.Page != "Home" {
		t.Fatalf("expected page Home, got %q", client.req.Page)
	}

	if res.HTML != "<main>Hello</main>" {
		t.Fatalf("expected rendered html, got %q", res.HTML)
	}
}
```

* [ ] **Step 4: Run stale-reference check again**

```bash
grep -R "SSRCommand\|ssr-worker\|newProcessSSRClient\|processSSRClient" -n .
```

Expected: no output.

* [ ] **Step 5: Run tests**

```bash
go test ./...
node --test js/*.test.mjs
```

Expected: PASS.

* [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor: remove stdio renderer bridge"
```

---

## Task 10: Add Render Failure Integration Test Through Fiber

**Files:**

* Modify: `zen/render_test.go`

* [ ] **Step 1: Add test**

Append to `zen/render_test.go`:

```go
func TestRenderReturnsRendererHTTPErrorThroughFiber(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		_ = json.NewEncoder(w).Encode(httpRendererErrorResponse{
			Error: httpRendererError{
				Message: "renderer exploded",
			},
		})
	}))
	defer server.Close()

	r := &Renderer{
		config: Config{
			Dev:           true,
			ViteURL:       "http://localhost:5173",
			RenderURL:     server.URL,
			AppElementID:  "app",
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: newHTTPSSRClient(httpSSRClientConfig{
			RenderURL: server.URL,
			Timeout:   time.Second,
		}),
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.Render(c, "Home", map[string]string{})
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")

	if res.StatusCode == fiber.StatusOK {
		t.Fatal("expected non-200 response")
	}
}
```

Update imports in `zen/render_test.go` to include:

```go
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/zenith/zen/internal/testutil"
)
```

* [ ] **Step 2: Run test**

```bash
go test ./zen -run 'TestRenderReturnsRendererHTTPErrorThroughFiber' -v
```

Expected: PASS.

* [ ] **Step 3: Commit**

```bash
git add zen/render_test.go
git commit -m "test: cover renderer http failure through fiber"
```

---

## Task 11: Add Dev Smoke Validation Script

**Files:**

* Modify: `package.json`

* [ ] **Step 1: Update root `package.json`**

```json
{
  "name": "@zenith/zen",
  "private": true,
  "type": "module",
  "scripts": {
    "test": "go test ./... && node --test js/*.test.mjs",
    "test:go": "go test ./...",
    "test:js": "node --test js/*.test.mjs",
    "test:example-build": "pnpm --dir examples/basic/frontend build && cd examples/basic && go build ."
  },
  "devDependencies": {
    "@types/node": "^22.0.0",
    "vite": "^7.0.0",
    "preact": "^10.0.0",
    "preact-render-to-string": "^6.0.0"
  }
}
```

* [ ] **Step 2: Run root tests**

```bash
pnpm test
```

Expected: PASS.

* [ ] **Step 3: Run example build test**

```bash
pnpm test:example-build
```

Expected: PASS.

* [ ] **Step 4: Commit**

```bash
git add package.json
git commit -m "test: validate http renderer migration"
```

---

## Task 12: Manual Smoke Test

**Files:**

* No file changes.

* [ ] **Step 1: Start dev renderer**

```bash
cd examples/basic
pnpm dev:renderer
```

Expected output contains:

```text
Zen dev renderer listening on http://127.0.0.1:5173
```

* [ ] **Step 2: Start Fiber app in another terminal**

```bash
cd examples/basic
go run .
```

Expected output contains Fiber startup logs for port `3000`.

* [ ] **Step 3: Check health endpoint**

```bash
curl -s http://127.0.0.1:5173/__zen/health
```

Expected:

```json
{"ok":true,"mode":"dev"}
```

* [ ] **Step 4: Check render endpoint directly**

```bash
curl -s http://127.0.0.1:5173/__zen/render \
  -H 'content-type: application/json' \
  -d '{"url":"/","page":"Home","props":{"title":"Hello","body":"Direct renderer call"}}'
```

Expected includes:

```json
{"html":
```

and:

```html
Direct renderer call
```

* [ ] **Step 5: Check Fiber-rendered page**

```bash
curl -i http://127.0.0.1:3000/users/42
```

Expected response includes:

```html
User 42
```

and:

```html
http://localhost:5173/@vite/client
```

* [ ] **Step 6: Check parallel requests**

```bash
seq 1 50 | xargs -n1 -P10 -I{} curl -s http://127.0.0.1:3000/users/{} >/tmp/zen-parallel.out
```

Expected: command exits successfully. No every-second-request clown behavior. No pipe desync. No dead renderer.

* [ ] **Step 7: Commit validation note**

```bash
git commit --allow-empty -m "test: smoke test http renderer migration"
```

---

# Self-Review

## Spec Coverage

| Requirement                         | Covered By                       |
| ----------------------------------- | -------------------------------- |
| Replace fragile stdin/stdout bridge | Tasks 2, 4, 9                    |
| Use HTTP renderer service           | Tasks 2, 6, 7                    |
| Keep Fiber as public route handler  | Task 8                           |
| Keep Vite/Preact SSR                | Tasks 6, 7, 8                    |
| Support parallel requests           | Task 2, Task 12                  |
| Add timeouts                        | Tasks 2, 3                       |
| Add renderer health endpoint        | Tasks 5, 6, 7, 12                |
| Make renderer debuggable with curl  | Tasks 6, 7, 12                   |
| Avoid WebSocket/RPC detour          | Entire plan uses plain HTTP JSON |
| Remove old pipe slop                | Task 9                           |
| Test-first migration                | Tasks 1–11                       |
| Frequent commits                    | Every task                       |

## Placeholder Scan

No task contains “TBD,” “fill in later,” “add appropriate handling,” or “similar to previous.” Every changed file has concrete contents or exact replacement snippets.

## Type Consistency

The shared Go renderer contract remains:

```go
type ssrClient interface {
	Render(ctx context.Context, req ssrRequest) (ssrResponse, error)
}
```

The HTTP request body matches `ssrRequest`:

```json
{
  "url": "/users/42",
  "page": "User",
  "props": {
    "id": "42"
  }
}
```

The HTTP success response matches `ssrResponse`:

```json
{
  "html": "<main>...</main>",
  "head": ""
}
```

The HTTP error response matches `httpRendererErrorResponse`:

```json
{
  "error": {
    "message": "renderer exploded",
    "stack": ""
  }
}
```

The final result is still Zen’s intended shape:

```text
Fiber handles the request.
Zen calls the renderer.
Vite/Preact render the page.
Zen assembles the document.
The browser hydrates.
```

Only now, the renderer bridge is normal HTTP instead of fragile pipe slop that falls apart the moment two requests show up and ask it to walk and chew gum.
