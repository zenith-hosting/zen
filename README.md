# Zen

Zen is tiny glue for building server-rendered Go apps with Fiber, Vite, Preact, Tailwind, and Air.

It does not replace those tools. It keeps Fiber in charge of routes, Vite in charge of frontend builds and dev behavior, Preact in charge of components and hydration, Tailwind in charge of styling, and Air in charge of Go hot reload.

## Install

Use the library from Go apps:

```bash
go get github.com/zenith-hosting/zen
```

The CLI lives in this repo under `zencli/`:

```bash
go install github.com/zenith-hosting/zen/zencli/cmd/zen@latest
```

That command installs a `zen` binary. Build a local `zen` binary from this repo with:

```bash
go build -o ./bin/zen ./zencli/cmd/zen
```

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
		}, zen.WithTitle("Home"))
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

## CLI Workflow

Create a starter project:

```bash
zen init
zen dev
```

`zen init` writes the starter project, runs `go mod tidy`, installs frontend dependencies, and approves required pnpm builds.

Build and start production artifacts:

```bash
zen build
zen start
```

`zen dev` starts the Vite/Preact renderer and the Go app through Air. `zen build` builds the frontend and Go binary. `zen start` starts existing production artifacts and does not build.

## Repository

This repository is the Zen library module:

```text
github.com/zenith-hosting/zen
```

The CLI lives in this repository as its own Go module:

```text
github.com/zenith-hosting/zen/zencli
```

Canonical frontend runtime sources are in `js/entries` and `js/renderers`. The starter-template copies live under `zencli/internal/zencli/init_template`. Sync them with:

```bash
node scripts/sync-renderers.mjs
```

## Development

Run the main checks:

```bash
go test ./...
node --test js/renderers/*.test.mjs
node --test scripts/*.test.mjs
```

Run CLI checks from the `zencli` module:

```bash
cd zencli
go test ./...
go build -o /tmp/zen ./cmd/zen
```
