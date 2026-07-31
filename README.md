# Zen

Zen is tiny glue for building server-rendered Go apps with Fiber, Vite, Preact, Tailwind, and Air.

It does not replace those tools. It keeps Fiber in charge of routes, Vite in charge of frontend builds and dev behavior, Preact in charge of components and hydration, Tailwind in charge of styling, and Air in charge of Go hot reload.

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

Zen fits inside ordinary Fiber handlers:

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/zenith-hosting/zen"
)

func main() {
	app := fiber.New()
	dev := os.Getenv("ZEN_ENV") != "prod"

	cfg := zen.Config{
		Dev:           dev,
		DefaultTitle:  "Zen App",
		RenderTimeout: 5 * time.Second,
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
		return renderer.RenderPage(c, "Home", map[string]any{
			"title": "Zen App",
		},
			zen.WithTitle("Home"),
			zen.WithMeta(zen.Name("description"), zen.Content("Zen app home page")),
			zen.WithLink(zen.Rel("canonical"), zen.Href("https://example.com/")),
		)
	})

	log.Fatal(app.Listen(":3000"))
}
```

The Go app reads renderer ports and frontend paths from `zen.config.json`, then calls the Node renderer over HTTP:

```text
POST /__zen/render
GET  /__zen/health
```

If the app does not run from the project root, set `ProjectRoot` in `zen.Config` so Zen can find `zen.config.json`.

Custom head elements are built from escaped attributes and injected into the matching slots in `frontend/index.html`:

```go
renderer.RenderPage(c, "Home", props,
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

These commands are ordinary `package.json` scripts. Tidy updates Go modules, installs frontend dependencies, and approves pnpm builds. Development starts the Vite/Preact renderer and Air, with Tailwind handled by Vite. Build creates the frontend bundles and `./bin/app`. Start runs those existing production artifacts and never builds.

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
