import test from "node:test";
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { once } from "node:events";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const here = dirname(fileURLToPath(import.meta.url));
const serverPath = join(here, "prod-renderer.mjs");
const okEntry = join(here, "..", "fixtures", "entry-server-ok.mjs");
const errorEntry = join(here, "..", "fixtures", "entry-server-error.mjs");

function startRenderer(entry, port) {
  return spawn(process.execPath, [
    serverPath,
    "--entry",
    entry,
    "--host",
    "127.0.0.1",
    "--port",
    String(port)
  ], {
    stdio: ["ignore", "pipe", "pipe"]
  });
}

async function waitForHealth(port) {
  const url = `http://127.0.0.1:${port}/__zen/health`;

  for (let i = 0; i < 50; i++) {
    try {
      const res = await fetch(url);
      if (res.ok) {
        return;
      }
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 25));
    }
  }

  throw new Error(`renderer did not become healthy on port ${port}`);
}

test("prod renderer health endpoint works", async () => {
  const port = 4771;
  const child = startRenderer(okEntry, port);

  try {
    await waitForHealth(port);

    const res = await fetch(`http://127.0.0.1:${port}/__zen/health`);
    const body = await res.json();

    assert.equal(res.status, 200);
    assert.deepEqual(body, {
      ok: true,
      mode: "production"
    });
  } finally {
    child.kill();
    await once(child, "exit");
  }
});

test("prod renderer renders page and island requests", async () => {
  const port = 4772;
  const child = startRenderer(okEntry, port);

  try {
    await waitForHealth(port);

    const res = await fetch(`http://127.0.0.1:${port}/__zen/render`, {
      method: "POST",
      headers: {
        "content-type": "application/json"
      },
      body: JSON.stringify({
        url: "/",
        page: "Home",
        props: {
          title: "Hello"
        }
      })
    });

    const body = await res.json();

    assert.equal(res.status, 200);
    assert.equal(body.html, `<main data-page="Home">Hello</main>`);

    const islandRes = await fetch(`http://127.0.0.1:${port}/__zen/render`, {
      method: "POST",
      headers: {
        "content-type": "application/json"
      },
      body: JSON.stringify({
        mode: "island",
        url: "/counter",
        island: "Counter",
        props: {
          count: 0
        }
      })
    });

    const islandBody = await islandRes.json();

    assert.equal(islandRes.status, 200);
    assert.equal(islandBody.html, `<button data-island="Counter">0</button>`);
  } finally {
    child.kill();
    await once(child, "exit");
  }
});

test("prod renderer returns structured render errors", async () => {
  const port = 4773;
  const child = startRenderer(errorEntry, port);

  try {
    await waitForHealth(port);

    const res = await fetch(`http://127.0.0.1:${port}/__zen/render`, {
      method: "POST",
      headers: {
        "content-type": "application/json"
      },
      body: JSON.stringify({
        url: "/",
        page: "Home",
        props: {}
      })
    });

    const body = await res.json();

    assert.equal(res.status, 500);
    assert.equal(body.error.message, "fixture render failed");
  } finally {
    child.kill();
    await once(child, "exit");
  }
});
