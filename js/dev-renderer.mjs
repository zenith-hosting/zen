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
    process.stdout.write(`Zen dev renderer listening on http://${args.host}:${args.port}\n`);
  });
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
