# Zen CLI

`zencli` is the command-line tool for Zen apps.

It creates starter projects, starts the development renderer and Go app, builds production artifacts, and starts existing production artifacts.

The Zen library is a separate Go module:

```text
github.com/zenith-hosting/zen
```

## Install

```bash
go install github.com/zenith-hosting/zen/zencli/cmd/zen@latest
```

That command installs a `zen` binary because the main package path ends in `/zen`.

For a local `zen` binary, build with:

```bash
go build -o /tmp/zen ./cmd/zen
```

## Commands

Create a complete starter app:

```bash
zen init
go mod tidy
pnpm --dir frontend install
zen dev
```

Start development mode:

```bash
zen dev
```

Build production artifacts:

```bash
zen build
```

Start existing production artifacts:

```bash
zen start
```

`zen start` does not build. If artifacts are missing, run `zen build`.

## Starter Project

`zen init` writes a runnable Fiber, Vite, Preact, Tailwind, and Air app. It refuses to overwrite existing files.

After init, users are expected to run:

```bash
go mod tidy
pnpm --dir frontend install
```

The generated app imports:

```go
import "github.com/zenith-hosting/zen"
```

Renderer runtime files are embedded into generated projects under:

```text
frontend/.zen/entries/
frontend/.zen/renderers/
```

The renderer process runs from the frontend directory so Vite resolves from `frontend/node_modules`.

## Development

Run CLI tests:

```bash
go test ./...
```

Build a local binary:

```bash
go build -o /tmp/zen ./cmd/zen
```

This module lives in the main `github.com/zenith-hosting/zen` repository under `zencli/`.
