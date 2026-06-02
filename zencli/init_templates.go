package zencli

func starterFiles() map[string]string {
	return map[string]string{
		"zen.config.json": starterZenConfigJSON,
		".air.toml":       starterAirTOML,
		"go.mod":          starterGoMod,
		"main.go":         starterMainGo,
		"package.json":    starterRootPackageJSON,

		"frontend/package.json":   starterFrontendPackageJSON,
		"frontend/tsconfig.json":  starterTSConfigJSON,
		"frontend/vite.config.ts": starterViteConfigTS,
		"frontend/index.html":     starterIndexHTML,

		"frontend/src/app.css":          starterAppCSS,
		"frontend/src/pages.ts":         starterPagesTS,
		"frontend/src/entry-client.tsx": starterEntryClientTSX,
		"frontend/src/entry-server.tsx": starterEntryServerTSX,
		"frontend/src/pages/Home.tsx":   starterHomeTSX,
		"frontend/src/pages/User.tsx":   starterUserTSX,

		"frontend/.zen/renderers/renderer-shared.mjs": starterRendererSharedMJS,
		"frontend/.zen/renderers/dev-renderer.mjs":    starterDevRendererMJS,
		"frontend/.zen/renderers/prod-renderer.mjs":   starterProdRendererMJS,
	}
}

const starterZenConfigJSON = `{
  "appCommand": "go run .",
  "airCommand": "air -c .air.toml",
  "frontendDir": "frontend",
  "devRendererPort": 5173,
  "prodRendererPort": 4174,
  "binaryPath": "./bin/app"
}
`

const starterAirTOML = `root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/zen-app ."
entrypoint = "./tmp/zen-app"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["frontend/node_modules", "frontend/dist", "tmp", "bin"]
delay = 1000
stop_on_error = true
send_interrupt = true
kill_delay = "500ms"

[log]
time = false

[misc]
clean_on_exit = true
`

const starterGoMod = `module zen-app

go 1.23

require (
	github.com/gofiber/fiber/v3 v3.0.0-rc.3
	github.com/zenith/zen v0.0.0
)
`

const starterMainGo = `package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/zenith/zen/zen"
)

func main() {
	app := fiber.New()

	dev := os.Getenv("ZEN_ENV") != "prod" && os.Getenv("ZEN_ENV") != "production"

	cfg := zen.Config{
		Dev:           dev,
		ViteURL:       "http://localhost:5173",
		RenderURL:     "http://localhost:5173/__zen/render",
		ClientDist:    "./frontend/dist/client",
		Manifest:      "./frontend/dist/client/.vite/manifest.json",
		DefaultTitle:  "Zen App",
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
			"title": "Zen App",
			"body":  "Fiber route, Preact page, Vite renderer.",
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

		return c.Redirect().To("/")
	})

	log.Fatal(app.Listen(":3000"))
}
`

const starterRootPackageJSON = `{
  "name": "zen-app",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "zen dev",
    "build": "zen build",
    "start": "zen start",
    "check": "zen check"
  }
}
`

const starterFrontendPackageJSON = `{
  "name": "zen-app-frontend",
  "private": true,
  "type": "module",
  "scripts": {
    "build:client": "vite build --outDir dist/client --manifest",
    "build:server": "vite build --ssr src/entry-server.tsx --outDir dist/server",
    "build": "pnpm build:client && pnpm build:server"
  },
  "dependencies": {
    "@preact/preset-vite": "^2.0.0",
    "@tailwindcss/vite": "^4.0.0",
    "preact": "^10.0.0",
    "preact-render-to-string": "^6.0.0",
    "tailwindcss": "^4.0.0",
    "vite": "^7.0.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}
`

const starterTSConfigJSON = `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "jsx": "react-jsx",
    "jsxImportSource": "preact",
    "strict": true,
    "types": ["vite/client"],
    "skipLibCheck": true
  },
  "include": ["src"]
}
`

const starterViteConfigTS = `import { defineConfig } from "vite";
import preact from "@preact/preset-vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [preact(), tailwindcss()],
  build: {
    rollupOptions: {
      input: {
        client: "src/entry-client.tsx"
      }
    }
  }
});
`

const starterIndexHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Zen App</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/entry-client.tsx"></script>
  </body>
</html>
`

const starterAppCSS = `@import "tailwindcss";

body {
  margin: 0;
}
`

const starterPagesTS = `import Home from "./pages/Home";
import User from "./pages/User";

export const pages = {
  Home,
  User
};

export type PageName = keyof typeof pages;
`

const starterEntryServerTSX = `import { h } from "preact";
import renderToString from "preact-render-to-string";
import { pages, type PageName } from "./pages";

type RenderRequest = {
  url: string;
  page: PageName;
  props: Record<string, unknown>;
};

export async function render(request: RenderRequest) {
  const Page = pages[request.page];

  if (!Page) {
    throw new Error(` + "`Unknown page: ${request.page}`" + `);
  }

  const html = renderToString(<Page {...request.props} />);

  return {
    html,
    head: ""
  };
}
`

const starterEntryClientTSX = `import { h, hydrate } from "preact";
import { pages, type PageName } from "./pages";
import "./app.css";

type HydrationData = {
  page: PageName;
  props: Record<string, unknown>;
};

const element = document.getElementById("__ZEN_DATA__");

if (!element || !element.textContent) {
  throw new Error("Missing Zen hydration data");
}

const data = JSON.parse(element.textContent) as HydrationData;
const Page = pages[data.page];

if (!Page) {
  throw new Error(` + "`Unknown page: ${data.page}`" + `);
}

const app = document.getElementById("app");

if (!app) {
  throw new Error("Missing app element");
}

hydrate(<Page {...data.props} />, app);
`

const starterHomeTSX = `type HomeProps = {
  title?: string;
  body?: string;
};

export default function Home(props: HomeProps) {
  return (
    <main class="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section class="mx-auto max-w-3xl rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <p class="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
          Zen
        </p>

        <h1 class="text-4xl font-bold tracking-tight">
          {props.title ?? "Zen App"}
        </h1>

        <p class="mt-4 text-lg leading-8 text-zinc-300">
          {props.body ?? "Tiny Fiber + Vite + Preact SSR glue."}
        </p>

        <form class="mt-8 flex gap-3" method="post" action="/contact">
          <input
            class="min-w-0 flex-1 rounded-xl border border-zinc-700 bg-zinc-950 px-4 py-3 text-zinc-100"
            name="name"
            placeholder="Normal Fiber form field"
          />
          <button
            class="rounded-xl bg-white px-5 py-3 font-semibold text-zinc-950"
            type="submit"
          >
            Submit
          </button>
        </form>

        <a class="mt-8 inline-block text-zinc-300 underline" href="/users/42">
          Visit dynamic Fiber route
        </a>
      </section>
    </main>
  );
}
`

const starterUserTSX = `type UserProps = {
  id?: string;
};

export default function User(props: UserProps) {
  return (
    <main class="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section class="mx-auto max-w-3xl rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <p class="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
          Fiber param
        </p>

        <h1 class="text-4xl font-bold tracking-tight">
          User {props.id ?? "unknown"}
        </h1>

        <p class="mt-4 text-lg leading-8 text-zinc-300">
          This page was selected by a normal Fiber route and rendered by Preact.
        </p>

        <a class="mt-8 inline-block text-zinc-300 underline" href="/">
          Back home
        </a>
      </section>
    </main>
  );
}
`

const starterRendererSharedMJS = `
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
`

const starterDevRendererMJS = `
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
      hmr: false,
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
    process.stdout.write(` + "`" + `Zen dev renderer listening on http://${args.host}:${args.port}\n` + "`" + `);
  });
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
`
const starterProdRendererMJS = `
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
    process.stdout.write(` + "`" + `Zen production renderer listening on http://${args.host}:${args.port}\n` + "`" + `);
  });
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
`
