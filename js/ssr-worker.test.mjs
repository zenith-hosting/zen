import test from "node:test";
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { once } from "node:events";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const here = dirname(fileURLToPath(import.meta.url));
const workerPath = join(here, "ssr-worker.mjs");
const okEntry = join(here, "fixtures", "entry-server-ok.mjs");
const errorEntry = join(here, "fixtures", "entry-server-error.mjs");

function startWorker(entry) {
  return spawn(process.execPath, [workerPath, "--entry", entry], {
    stdio: ["pipe", "pipe", "pipe"]
  });
}

async function readLine(stream) {
  let buffer = "";
  for await (const chunk of stream) {
    buffer += chunk.toString("utf8");
    const index = buffer.indexOf("\n");
    if (index !== -1) {
      return buffer.slice(0, index);
    }
  }
  throw new Error("stream ended before line");
}

test("worker renders one request", async () => {
  const child = startWorker(okEntry);

  child.stdin.write(JSON.stringify({
    id: "1",
    request: {
      url: "/",
      page: "Home",
      props: {
        title: "Hello"
      }
    }
  }) + "\n");

  const line = await readLine(child.stdout);
  const message = JSON.parse(line);

  assert.equal(message.id, "1");
  assert.equal(message.result.html, `<main data-page="Home">Hello</main>`);

  child.kill();
  await once(child, "exit");
});

test("worker reports render errors", async () => {
  const child = startWorker(errorEntry);

  child.stdin.write(JSON.stringify({
    id: "2",
    request: {
      url: "/",
      page: "Home",
      props: {}
    }
  }) + "\n");

  const line = await readLine(child.stdout);
  const message = JSON.parse(line);

  assert.equal(message.id, "2");
  assert.match(message.error, /fixture render failed/);

  child.kill();
  await once(child, "exit");
});
