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
  -> Zen injects SSR output into frontend/index.html
  -> browser hydrates Preact page
```

Development mode:

```text
pnpm dev
  -> starts dev renderer
  -> starts Go app through Air
```

Production mode:

```text
pnpm build
  -> builds frontend client assets
  -> builds frontend SSR bundle
  -> builds Go binary

pnpm start
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

Run the starter app in development:

```bash
cd starter
pnpm dev
```

Build the starter app:

```bash
cd starter
pnpm build
```

Start the starter app from production artifacts:

```bash
cd starter
pnpm start
```

Create a new project:

```bash
git clone https://github.com/zenith-hosting/zen-starter my-app
cd my-app
pnpm tidy
pnpm dev
```

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

starter/
  package.json
  main.go
  go.mod
  frontend/
    .zen/
      entries/
        entry-client.tsx
        entry-server.tsx
      renderers/
        renderer-shared.mjs
        dev-renderer.mjs
        prod-renderer.mjs
    src/

```

The root Go module is:

```text
github.com/zenith-hosting/zen
```

The starter defines the workflow directly in `package.json` scripts invoked through `pnpm tidy`, `pnpm dev`, `pnpm build`, and `pnpm start`. Zen does not need a general-purpose CLI or process manager.

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

The `pnpm dev` package script should be the only workflow that starts Air.

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

Renderer and entry sources live in:

```text
starter/frontend/.zen/
```

---

## HTML Document Template Rules

Full-page responses are assembled from the developer-owned frontend document template:

```text
frontend/index.html
```

`Config.DocumentPath` defaults to:

```text
<ProjectRoot>/<frontendDir>/index.html
```

where `frontendDir` comes from `zen.config.json` and defaults to `frontend`.

The Go renderer reads this template for `RenderPage`, replaces Zen slots, and sends the final HTML response. Keep this as simple string slot replacement. Do not introduce an HTML parser, metadata DSL, layout system, Vite `transformIndexHtml`, or a Next-style document API unless explicitly requested and carefully justified.

Required template slots:

```html
<!--zen:title-->
<!--zen:head-->
<!--zen:base-->
<!--zen:meta-->
<!--zen:link-->
<!--zen:style-->
<!--zen:styles-->
<!--zen:script-->
<!--zen:app-->
<!--zen:data-->
<!--zen:scripts-->
```

Slot meanings:

* `<!--zen:title-->`: escaped page title from `WithTitle` or `Config.DefaultTitle`
* `<!--zen:head-->`: raw `head` returned by the Node renderer
* `<!--zen:base-->`: Go-generated `<base>` tags from `WithBase` / `Base`
* `<!--zen:meta-->`: Go-generated `<meta>` tags from `WithMeta` / `Meta`
* `<!--zen:link-->`: Go-generated `<link>` tags from `WithLink` / `Link`
* `<!--zen:style-->`: Go-generated `<style>` tag from `WithStyle` / `Style`
* `<!--zen:styles-->`: production CSS links from the Vite manifest
* `<!--zen:script-->`: Go-generated `<script>` tags from `WithScript` / `Script`
* `<!--zen:app-->`: raw Preact SSR HTML
* `<!--zen:data-->`: Go-generated safe hydration JSON script using `Config.DataElementID`
* `<!--zen:scripts-->`: Vite dev scripts or production client entry scripts

Go-owned head elements should use the structured render options:

```go
renderer.RenderPage(c, "Home", props,
	zen.WithTitle("Home"),
	zen.WithBase(zen.Href("/")),
	zen.WithMeta(zen.Name("description"), zen.Content("Page description")),
	zen.WithLink(zen.Rel("canonical"), zen.Href("https://example.com/")),
	zen.WithStyle(`body { color: red; }`),
	zen.WithScript(zen.Type("application/ld+json"), zen.Text(`{"name":"Home"}`)),
)
```

Attribute values are escaped by Zen. Use `Attr(name, value)` for less-common attributes; unsafe attribute names are skipped. `Text` is raw element text for scripts, and `WithStyle` takes raw CSS directly as a string.

The starter should keep this shape:

```html
<div id="app"><!--zen:app--></div>
<!--zen:data-->
<!--zen:scripts-->
```

The current frontend hydration entry looks for:

```text
document.getElementById("app")
document.getElementById("__ZEN_DATA__")
```

Do not change the template IDs without also changing the hydration entry and tests.

`RenderIsland` is intentionally separate. It returns a hydratable fragment from `renderIslandFragment`; it should not read or inject `frontend/index.html`.

When changing document template behavior, update:

```text
document.go
document_test.go
head.go
render.go
render_test.go
starter/frontend/index.html
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

The HTTP protocol stops here. The Node renderer returns fragments and optional head HTML; Go owns template slot injection and the final full document response.

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

## Project Script Behavior

The starter contains app-owned `package.json` scripts:

```text
pnpm tidy
pnpm dev
pnpm build
pnpm start
```

There is no init command. New projects come from the starter repository.

`pnpm tidy` runs `go mod tidy` and `pnpm --dir frontend install` in order.

`pnpm dev` starts the development renderer from `frontend/` and the Go app through Air with `ZEN_ENV=dev`. Tailwind is handled by the existing Vite plugin; do not add a separate Tailwind watcher.

`pnpm build` runs `pnpm --dir frontend build`, creates `bin/`, and runs `go build -o ./bin/app .`.

`pnpm start` verifies the production binary, SSR entry, and Vite manifest, then starts the production renderer from `frontend/` and `ZEN_ENV=prod ./bin/app`. It never builds.

Keep the scripts project-local and conventional. Do not add JSON parsing, log prefixing, health polling, a template generator, or a compatibility binary unless a demonstrated need justifies it.

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
  pnpm build
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
```

When changing the starter, validate:

```bash
tmpdir="$(mktemp -d)"
cp -R starter/. "$tmpdir"
cd "$tmpdir"
pnpm tidy
pnpm dev
```

When changing production behavior, validate:

```bash
pnpm build
pnpm start
```

Then request:

```bash
curl -i http://127.0.0.1:3000/
```

In dev output, the HTML should include Vite client scripts.

The response should preserve the surrounding `frontend/index.html` template shell and include SSR HTML in `<!--zen:app-->`.

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
