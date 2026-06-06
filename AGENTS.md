# AGENTS.md

## Project Overview

Zen is a tiny Go-first SSR framework that glues together:

* Fiber for backend routing and request handling
* Vite for frontend dev/build behavior
* Preact for server-rendered and hydrated frontend pages
* Tailwind for styling
* Air for Go hot reload during development

Zen is not trying to replace these tools. It should make them work together with as little magic as possible.

The core philosophy is:

> Developers should feel like they are using Fiber, Vite, Preact, Tailwind, and Air directly. Zen should only remove the repetitive bridge slop between them.

Do not add abstractions that make developers learn “the Zen way” for things Fiber, Vite, Preact, Tailwind, or Air already do well.

---

## Current Architecture

A Zen app works like this:

```text
Browser
  -> Fiber app
  -> normal Fiber route handler
  -> renderer.Render(c, "Page", props)
  -> HTTP call to Node renderer
  -> Vite/Preact SSR
  -> Zen document assembly
  -> browser hydrates Preact page
```

Development mode:

```text
zen dev
  -> starts dev renderer
  -> starts Go app through Air
```

Production mode:

```text
zen build
  -> builds frontend client assets
  -> builds frontend SSR bundle
  -> builds Go binary

zen start
  -> starts production renderer
  -> starts compiled Go binary with ZEN_ENV=prod
```

The renderer bridge uses HTTP. Do not reintroduce stdin/stdout process communication.

---

## Important Commands

Run all Go tests:

```bash
go test ./...
```

Run Node renderer tests:

```bash
node --test js/renderers/*.test.mjs
node --test scripts/*.test.mjs
```

Build the Zen CLI:

```bash
go build -o ./bin/zen ./zencli/cmd/zen
```

Run the example app in development:

```bash
cd examples/basic
../../bin/zen dev
```

Build the example app:

```bash
cd examples/basic
../../bin/zen build
```

Start the example app from production artifacts:

```bash
cd examples/basic
../../bin/zen start
```

Initialize a new project:

```bash
zen init
zen dev
```

`zen init` runs `go mod tidy`, `pnpm --dir frontend install`, and `pnpm --dir frontend approve-builds --all` after writing files.

---

## Repository Layout

Expected high-level structure:

```text
.
  config.go
  render.go
  document.go
  escape.go
  manifest.go
  static.go
  ssr_client.go
  ssr_http_client.go

zencli/
  go.mod
  cmd/
    zen/
      main.go
  internal/
    zencli/

js/
  entries/
    entry-client.tsx
    entry-server.tsx
  renderers/
    renderer-shared.mjs
    dev-renderer.mjs
    prod-renderer.mjs

examples/
  basic/
    main.go
    zen.config.json
    .air.toml
    frontend/
      package.json
      vite.config.ts
      src/
```

The root Go module is:

```text
github.com/zenith-hosting/zen
```

The CLI is an in-repo Go module:

```text
github.com/zenith-hosting/zen/zencli
```

Do not merge CLI implementation code into the root Zen library package. Keeping `zencli` as a separate subdirectory module lets the root module stay usable as a library dependency without CLI code.

---

## Boundaries

### Fiber owns

* Routes
* Middleware
* Request parsing
* Forms
* Redirects
* Cookies
* Sessions
* Errors
* Static asset route registration

Zen should not wrap Fiber into a second routing framework.

Good:

```go
app.Get("/", func(c fiber.Ctx) error {
	return renderer.Render(c, "Home", props)
})
```

Bad:

```go
zen.Page("/", HomePage)
zen.Loader(...)
zen.Action(...)
```

Do not add this kind of API unless explicitly requested and carefully justified.

---

### Vite owns

* TypeScript/TSX transforms
* Preact plugin behavior
* Tailwind integration
* Frontend dev server behavior
* Frontend production builds
* SSR bundle generation
* HMR

Zen should not reimplement Vite behavior in Go.

---

### Preact owns

* Page components
* SSR rendering
* Hydration
* Frontend component model

Zen should not invent component wrappers, custom page definitions, or a Zen-specific frontend API.

---

### Tailwind owns

* Styling
* CSS pipeline
* Utility classes

Zen should not add its own styling abstraction.

---

### Air owns

* Go hot reload
* Rebuilding/restarting the Go app during development

Zen should only start Air as part of `zen dev`.

---

## Renderer Runtime Rules

The renderer files must be available to initialized projects.

Renderer runtime files should live under the frontend package:

```text
frontend/.zen/renderers/renderer-shared.mjs
frontend/.zen/renderers/dev-renderer.mjs
frontend/.zen/renderers/prod-renderer.mjs
```

The renderer process must run with:

```text
Dir: cfg.FrontendDir
```

This matters because Vite is installed in:

```text
frontend/node_modules
```

If the renderer is started from the project root, imports like this will fail:

```js
import { createServer as createViteServer } from "vite";
```

Do not move renderer runtime files back to root `.zen/renderers` unless dependency resolution is redesigned.

Canonical renderer and entry sources live in:

```text
js/entries/
js/renderers/
```

The starter-template copies live in the `zencli` module:

```text
zencli/internal/zencli/init_template/frontend/.zen/
```

After changing renderer or entry source files, run:

```bash
node scripts/sync-renderers.mjs
```

---

## HTTP Renderer Protocol

The Go app calls the renderer over HTTP.

Render endpoint:

```text
POST /__zen/render
```

Health endpoint:

```text
GET /__zen/health
```

Render request:

```json
{
  "url": "/users/42",
  "page": "User",
  "props": {
    "id": "42"
  }
}
```

Successful response:

```json
{
  "html": "<main>...</main>",
  "head": ""
}
```

Error response:

```json
{
  "error": {
    "message": "Unknown page: User",
    "stack": "..."
  }
}
```

Keep the protocol boring. Do not introduce WebSocket, JSON-RPC, gRPC, or a custom multiplexed protocol for the render path unless there is a benchmark-backed reason.

---

## CLI Behavior

### `zen init`

Should create a complete runnable starter project, not just config files.

It should write:

```text
zen.config.json
.air.toml
go.mod
main.go
package.json
frontend/package.json
frontend/tsconfig.json
frontend/vite.config.ts
frontend/index.html
frontend/src/app.css
frontend/src/pages.ts
frontend/src/pages/Home.tsx
frontend/src/pages/User.tsx
frontend/.zen/entries/entry-client.tsx
frontend/.zen/entries/entry-server.tsx
frontend/.zen/renderers/renderer-shared.mjs
frontend/.zen/renderers/dev-renderer.mjs
frontend/.zen/renderers/prod-renderer.mjs
```

Then it should install dependencies, leaving the user with a working project:

```sh
go mod tidy
pnpm --dir frontend install
pnpm --dir frontend approve-builds --all
```

`zen init` should refuse to overwrite existing files.

### `zen dev`

Should only create a development environment.

It should start:

```text
renderer: node .zen/renderers/dev-renderer.mjs
app:      go tool air -c .air.toml
```

It should not build production artifacts.

It should not inspect `ZEN_ENV=prod`.

It should not run production mode.

### `zen build`

Should build production artifacts.

It should run:

```text
pnpm --dir frontend build
go build -o ./bin/app .
```

### `zen start`

Should start production artifacts.

It should run:

```text
renderer: node .zen/renderers/prod-renderer.mjs
app:      ZEN_ENV=prod ./bin/app
```

`zen start` should not build. If artifacts are missing, it should clearly tell the user to run:

```bash
zen build
```

---

## Error Message Guidelines

Prefer actionable errors.

Bad:

```text
Cannot find package vite
```

Good:

```text
zen: missing frontend dependencies.

Run:
  pnpm --dir frontend install
```

Bad:

```text
open frontend/dist/server/entry-server.js: no such file or directory
```

Good:

```text
zen: missing production artifact frontend/dist/server/entry-server.js.

Run:
  zen build
```

Bad:

```text
renderer failed
```

Good:

```text
zen renderer: Unknown page: Admin
```

---

## Testing Rules

Use test-first development for behavior changes.

For Go:

```bash
go test ./...
```

For Node renderer behavior:

```bash
node --test js/renderers/*.test.mjs
node --test scripts/*.test.mjs
```

For CLI behavior, prefer small tests around pure planning functions:

* `devPlan`
* `startPlan`
* `buildPlan`
* `starterFiles`
* `ensureFrontendDependencies`
* `ensureProductionArtifacts`

CLI tests live in the `zencli` module. Run them from that module or through the workspace:

```bash
cd zencli
go test ./...
```

Avoid tests that require starting long-running processes unless the behavior cannot be tested another way.

---

## Style Rules

Keep files focused.

Prefer small units with boring names.

Do not add clever abstractions around simple tool invocations.

Do not introduce a plugin system.

Do not introduce a route system.

Do not introduce a form system.

Do not introduce generated frontend types.

Do not turn Zen into Next.js with a Go accent.

---

## Dependency Rules

Adding a dependency is allowed when it removes meaningful maintenance burden.

Adding a dependency is not allowed when it mostly creates a new abstraction developers have to understand.

Acceptable dependencies so far:

* Fiber
* Air
* Vite
* Preact
* Tailwind

Be cautious with protocol libraries, RPC frameworks, process managers, and anything that makes debugging less obvious.

---

## Development Philosophy

When in doubt, choose:

```text
boring over clever
explicit over magical
tool-native over Zen-specific
debuggable over elegant
small over complete
```

Zen should make this workflow pleasant:

```go
app.Get("/", func(c fiber.Ctx) error {
	return renderer.Render(c, "Home", props)
})
```

Not replace it with a shiny proprietary lifecycle.

If a feature requires users to learn a new Zen concept before they can understand what Fiber, Vite, or Preact is doing, the feature is probably too big or pointed in the wrong direction.

---

## Before Committing

Run:

```bash
go test ./...
node --test js/renderers/*.test.mjs
node --test scripts/*.test.mjs
```

When changing the CLI, also run:

```bash
go build -o ./bin/zen ./zencli/cmd/zen
```

When changing starter templates, validate:

```bash
tmpdir="$(mktemp -d)"
cd "$tmpdir"
/path/to/zen/bin/zen init
/path/to/zen/bin/zen dev
```

When changing production behavior, validate:

```bash
zen build
zen start
```

Then request:

```bash
curl -i http://127.0.0.1:3000/
```

In dev output, the HTML should include Vite client scripts.

In production output, the HTML should not include:

```text
/@vite/client
```

---

## Non-Goals

Zen is not:

* a full-stack application framework
* a frontend router
* a backend router
* an auth framework
* a database framework
* a form framework
* a deployment platform
* a replacement for Vite
* a replacement for Fiber
* a replacement for Air

Zen is glue.

Keep it that way.
