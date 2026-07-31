# Zen

Zen is tiny glue for building server-rendered Go apps with any Go HTTP framework, Vite, React, Tailwind, and Air.

It does not replace those tools. It keeps your HTTP framework in charge of requests and responses, Vite in charge of frontend builds and dev behavior, React in charge of components and hydration, Tailwind in charge of styling, and Air in charge of Go hot reload.

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
		Dev:           dev,
		DefaultTitle:  "Zen App",
		RenderTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Close()

	if !dev {
		http.Handle("/assets/", http.StripPrefix(
			"/assets/",
			http.FileServer(http.Dir(renderer.AssetsDir())),
		))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, request *http.Request) {
		response, err := renderer.RenderPage(
			request.Context(),
			request.URL.RequestURI(),
			"Home",
			map[string]any{
			"title": "Zen App",
			},
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

	log.Fatal(http.ListenAndServe(":3000", nil))
}
```

Chi uses the same `net/http` handler shape. Gin and Fiber handlers pass their request context and original URL to Zen, then write `Response.Status`, `Response.ContentType`, and `Response.Body` with their native APIs.

The Go app reads renderer ports and frontend paths from `zen.config.json`, then calls the Node renderer over HTTP:

```text
POST /__zen/render
GET  /__zen/health
```

If the app does not run from the project root, set `ProjectRoot` in `zen.Config` so Zen can find `zen.config.json`.

Custom head elements are built from escaped attributes and injected into the matching slots in `frontend/index.html`:

```go
response, err := renderer.RenderPage(ctx, url, "Home", props,
	zen.WithBase(zen.Href("/")),
	zen.WithMeta(zen.Name("description"), zen.Content("Page description")),
	zen.WithLink(zen.Rel("canonical"), zen.Href("https://example.com/"), zen.Attr("data-source", "go")),
	zen.WithStyle(`body { color: red; }`),
	zen.WithScript(zen.Type("application/ld+json"), zen.Text(`{"name":"Home"}`)),
)
```

Short aliases such as `zen.Base(...)`, `zen.Meta(...)`, and `zen.Link(...)` are also available.

## Project workflow

```bash
pnpm tidy
pnpm dev
pnpm build
pnpm start
```

These commands are ordinary `package.json` scripts. Tidy updates Go modules and installs frontend dependencies. Development starts the Vite/React renderer and Air, with Tailwind handled by Vite. Build creates the frontend bundles and `./bin/app`. Start runs those existing production artifacts and never builds.

## Repository

This repository contains the Zen library module:

```text
github.com/zenith-hosting/zen
```

Canonical frontend runtime sources are in `starter/frontend/.zen`.

## Development

Run the main checks:

```bash
go test ./...
```
