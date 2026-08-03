import http from "node:http";
import { parseArgs } from "node:util";
import { createServer as createViteServer } from "vite";
import {
  readJSON,
  writeJSON,
  writeRendererError
} from "./renderer-shared.mjs";

async function main() {
  const { values: args } = parseArgs({
    options: {
      root: { type: "string", default: process.cwd() },
      entry: { type: "string", default: "/.zen/entries/entry-server.tsx" },
      host: { type: "string", default: "127.0.0.1" },
      port: { type: "string", default: "5173" }
    }
  });
  args.port = Number(args.port);

  if (!Number.isInteger(args.port) || args.port <= 0) {
    throw new Error("port must be a positive integer");
  }

  let vite;

  const server = http.createServer(async (req, res) => {
    if (req.method === "GET" && req.url === "/__zen/health") {
      writeJSON(res, 200, { ok: true, mode: "dev" });
      return;
    }

    if (req.method === "POST" && req.url === "/__zen/render") {
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

  vite = await createViteServer({
    root: args.root,
    server: {
      host: args.host,
      port: args.port,
      hmr: {
        server
      },
      middlewareMode: true
    },
    appType: "custom"
  });

  server.listen(args.port, args.host, () => {
    process.stdout.write(`Zen dev renderer listening on http://${args.host}:${args.port}\n`);
  });
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
