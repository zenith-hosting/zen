import fs from "node:fs";
import { pathToFileURL } from "node:url";

function parseArgs(argv) {
  const entryIndex = argv.indexOf("--entry");
  if (entryIndex === -1 || !argv[entryIndex + 1]) {
    throw new Error("missing required --entry argument");
  }

  return {
    entry: argv[entryIndex + 1]
  };
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const entryURL = pathToFileURL(args.entry).href;
  const mod = await import(entryURL);

  if (typeof mod.render !== "function") {
    throw new Error("SSR entry must export render(request)");
  }

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
      const result = await mod.render(incoming.request);
      process.stdout.write(JSON.stringify({
        id: incoming.id,
        result
      }) + "\n");
    } catch (error) {
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
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
