import test from "node:test";
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { once } from "node:events";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const scriptPath = join(here, "sync-renderers.mjs");

test("renderer template files are in sync with js sources", async () => {
  const child = spawn(process.execPath, [scriptPath], {
    stdio: ["ignore", "pipe", "pipe"]
  });

  const [exitCode] = await once(child, "exit");

  assert.equal(exitCode, 0);
});
