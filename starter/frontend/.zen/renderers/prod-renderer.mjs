import http from "node:http";
import { pathToFileURL } from "node:url";
import { parseArgs } from "node:util";
import {
  readJSON,
  writeJSON,
  writeRendererError
} from "./renderer-shared.mjs";

async function main() {
  const { values: args } = parseArgs({
    options: {
      entry: { type: "string", default: "./dist/server/entry-server.js" },
      host: { type: "string", default: "127.0.0.1" },
      port: { type: "string", default: "4174" }
    }
  });
  args.port = Number(args.port);

  if (!Number.isInteger(args.port) || args.port <= 0) {
    throw new Error("port must be a positive integer");
  }
  const entryURL = pathToFileURL(args.entry).href;
  const mod = await import(entryURL);

  if (typeof mod.render !== "function") {
    throw new Error("SSR entry must export render(request)");
  }

  const server = http.createServer(async (req, res) => {
    if (req.method === "GET" && req.url === "/__zen/health") {
      writeJSON(res, 200, { ok: true, mode: "production" });
      return;
    }

    if (req.method === "POST" && req.url === "/__zen/render") {
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
    process.stdout.write(`Zen production renderer listening on http://${args.host}:${args.port}\n`);
  });
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
