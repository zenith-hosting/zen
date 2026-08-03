# Zen

Zen is tiny glue for building server-rendered Go apps with any Go HTTP framework, Vite, React, Tailwind, and Air. The library itself uses only the Go standard library.

It does not replace those tools. It keeps your HTTP framework in charge of requests and responses, Vite in charge of frontend builds and dev behavior, React in charge of components and hydration, Tailwind in charge of styling, and Air in charge of Go hot reload.

> [!WARNING]
> Zen is very experimental. Breaking changes WILL happen. Proceed with caution.

## Install

Use the library from Go apps:

```bash
go get github.com/zenith-hosting/zen
```

For a new app, clone the starter:

```bash
git clone https://github.com/zenith-hosting/zen-starter my-app
cd my-app
go mod edit -module example.com/my-app
pnpm tidy
pnpm dev
```

The starter source also lives in this repository under `starter/`.

## Usage

Zen returns a framework-neutral response from ordinary Go handlers:

```go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zenith-hosting/zen"
)

func main() {
	dev := os.Getenv("ZEN_ENV") != "prod"
	renderer, err := zen.New(zen.Config{
		Dev:              dev,
		InlineStyles:     true,
		FrontendDir:      "frontend",
		DevRendererPort:  5173,
		ProdRendererPort: 4174,
		DefaultTitle:     "Zen App",
		RenderTimeout:    5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	if !dev {
		mux.Handle("/assets/", http.StripPrefix(
			"/assets/",
			http.FileServer(http.Dir(renderer.AssetsDir())),
		))
	}

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, request *http.Request) {
		response, err := renderer.RenderPage(
			request.Context(),
			request.URL.RequestURI(),
			"Home",
			map[string]any{},
			zen.WithTitle("Home"),
			zen.WithMeta(zen.Name("description"), zen.Content("Zen app home page")),
			zen.WithLink(zen.Rel("canonical"), zen.Href("https://example.com/")),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", response.ContentType)
		w.WriteHeader(response.Status)
		_, _ = w.Write(response.Body)
	})

	log.Fatal(http.ListenAndServe(":3000", mux))
}
```

Chi uses the same `net/http` handler shape. Gin and Fiber handlers pass their request context and original URL to Zen, then write `Response.Status`, `Response.ContentType`, and `Response.Body` with their native APIs.

The Go app calls the Node renderer over HTTP:

```text
POST /__zen/render
GET  /__zen/health
```

`FrontendDir`, `DevRendererPort`, and `ProdRendererPort` default to the values shown in the example. The package scripts use those defaults directly; if you change one in `zen.Config`, update the corresponding `cd` or renderer `--port` argument in `package.json` too.

If the app does not run from the project root, set `ProjectRoot` in `zen.Config` so Zen can resolve frontend build paths.

Zen builds the HTML document in Go. Custom head elements use escaped attributes:

```go
response, err := renderer.RenderPage(ctx, url, "Home", props,
	zen.WithBase(zen.Href("/")),
	zen.WithMeta(zen.Name("description"), zen.Content("Page description")),
	zen.WithLink(zen.Rel("canonical"), zen.Href("https://example.com/"), zen.Attr("data-source", "go")),
	zen.WithStyle(`body { color: red; }`),
	zen.WithScript(zen.Type("application/ld+json"), zen.Text(`{"name":"Home"}`)),
)
```

`RenderIsland` returns an SSR fragment that the client entry hydrates as an independent React root. Islands can be included in the initial page props or fetched and inserted later; a `MutationObserver` hydrates newly inserted fragments.

```go
counter, err := renderer.RenderIsland(ctx, url, "Counter", map[string]any{"count": 0})
```

In production, Zen reads client scripts and styles from Vite's manifest. Serve `renderer.AssetsDir()` at `/assets/`. Set `InlineStyles: true` in `zen.Config` when you want compiled CSS embedded in every rendered page instead of linked.

## Project workflow

```bash
pnpm tidy
pnpm dev
pnpm build
pnpm start
```

These commands are ordinary `package.json` scripts:

* `tidy` runs `go mod tidy` and installs frontend dependencies.
* `dev` starts the Vite/React renderer on port `5173` and Air; Air proxies `localhost:3000` to the Go app on `30001`.
* `build` creates the client bundle, SSR bundle, Vite manifest, and `./bin/app`.
* `start` checks those production artifacts, then starts the renderer on `4174` and the Go app on `3000`. It never builds.

## Repository

This repository contains the Zen library module:

```text
github.com/zenith-hosting/zen
```

Canonical frontend runtime sources are in `starter/frontend/.zen`; there are no separate CLI, examples, or JavaScript source copies elsewhere in this repository.

## Development

Run the main checks:

```bash
go test ./...
pnpm --dir starter/frontend exec tsc --noEmit
pnpm --dir starter/frontend build
```
