import test from "node:test";
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { once } from "node:events";
import { mkdtemp, mkdir, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { dirname } from "node:path";

const here = dirname(fileURLToPath(import.meta.url));
const serverPath = join(here, "dev-renderer.mjs");

async function createViteFixture() {
  const root = await mkdtemp(join(tmpdir(), "zen-vite-fixture-"));
  const src = join(root, "src");

  await mkdir(src, {
    recursive: true
  });

  await writeFile(join(root, "package.json"), JSON.stringify({
    type: "module",
    dependencies: {
      vite: "^7.0.0"
    }
  }));

  await writeFile(join(src, "entry-server.js"), `
    export async function render(request) {
      if (request.mode === "island") {
        return {
          html: '<button data-island="' + request.island + '">' + request.props.count + '</button>',
          head: ''
        };
      }

      return {
        html: '<main data-page="' + request.page + '">' + request.props.title + '</main>',
        head: ''
      };
    }
  `);

  return root;
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

async function waitForViteHMR(port) {
  const ws = new WebSocket(`ws://127.0.0.1:${port}/`, "vite-hmr");

  try {
    await new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error("timed out waiting for Vite HMR connection"));
      }, 1000);

      ws.addEventListener("message", (event) => {
        const payload = JSON.parse(String(event.data));

        if (payload.type === "connected") {
          clearTimeout(timeout);
          resolve();
        }
      });

      ws.addEventListener("error", () => {
        clearTimeout(timeout);
        reject(new Error("Vite HMR websocket failed"));
      });
    });
  } finally {
    ws.close();
  }
}

test("dev renderer renders page and island requests through vite", async () => {
  const root = await createViteFixture();
  const port = 4781;

  const child = spawn(process.execPath, [
    serverPath,
    "--root",
    root,
    "--entry",
    "/src/entry-server.js",
    "--host",
    "127.0.0.1",
    "--port",
    String(port)
  ], {
    stdio: ["ignore", "pipe", "pipe"]
  });

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

test("dev renderer exposes Vite HMR websocket", async () => {
  const root = await createViteFixture();
  const port = 4782;

  const child = spawn(process.execPath, [
    serverPath,
    "--root",
    root,
    "--entry",
    "/src/entry-server.js",
    "--host",
    "127.0.0.1",
    "--port",
    String(port)
  ], {
    stdio: ["ignore", "pipe", "pipe"]
  });

  try {
    await waitForHealth(port);
    await waitForViteHMR(port);
  } finally {
    child.kill();
    await once(child, "exit");
  }
});
