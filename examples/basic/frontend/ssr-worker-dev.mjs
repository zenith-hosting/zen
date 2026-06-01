import fs from "node:fs";
import { createServer } from "vite";

const vite = await createServer({
  server: {
    middlewareMode: true
  },
  appType: "custom"
});

async function handleLine(line) {
  if (!line.trim()) {
    return;
  }

  let incoming;

  try {
    incoming = JSON.parse(line);
  } catch (error) {
    process.stdout.write(JSON.stringify({
      id: null,
      error: `invalid json: ${error.message}`
    }) + "\n");
    return;
  }

  try {
    const mod = await vite.ssrLoadModule("/src/entry-server.tsx");

    if (typeof mod.render !== "function") {
      throw new Error("SSR entry must export render(request)");
    }

    const result = await mod.render(incoming.request);

    process.stdout.write(JSON.stringify({
      id: incoming.id,
      result
    }) + "\n");
  } catch (error) {
    vite.ssrFixStacktrace(error);

    process.stdout.write(JSON.stringify({
      id: incoming.id,
      error: error && error.stack ? error.stack : String(error)
    }) + "\n");
  }
}

let buffer = "";
const chunk = Buffer.alloc(64 * 1024);

for (;;) {
  const bytesRead = fs.readSync(0, chunk, 0, chunk.length, null);
  if (bytesRead === 0) {
    break;
  }

  buffer += chunk.subarray(0, bytesRead).toString("utf8");

  for (;;) {
    const index = buffer.indexOf("\n");
    if (index === -1) {
      break;
    }

    const line = buffer.slice(0, index);
    buffer = buffer.slice(index + 1);
    await handleLine(line);
  }
}
