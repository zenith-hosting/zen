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
    process.stdout.write(`Zen production renderer listening on http://${args.host}:${args.port}\n`);
  });
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
