import test from "node:test";
import assert from "node:assert/strict";
import { Readable } from "node:stream";
import {
  readJSON,
  writeJSON,
  writeRendererError,
  createHealthResponse
} from "./renderer-shared.mjs";

function mockResponse() {
  return {
    statusCode: 0,
    headers: {},
    body: "",
    setHeader(name, value) {
      this.headers[name.toLowerCase()] = value;
    },
    end(value) {
      this.body = value;
    }
  };
}

test("readJSON parses request body", async () => {
  const req = Readable.from([
    JSON.stringify({
      page: "Home",
      props: {
        title: "Hello"
      }
    })
  ]);

  const got = await readJSON(req);

  assert.equal(got.page, "Home");
  assert.equal(got.props.title, "Hello");
});

test("writeJSON writes JSON response", () => {
  const res = mockResponse();

  writeJSON(res, 201, {
    ok: true
  });

  assert.equal(res.statusCode, 201);
  assert.equal(res.headers["content-type"], "application/json");
  assert.equal(res.body, '{"ok":true}');
});

test("writeRendererError writes structured error", () => {
  const res = mockResponse();
  const error = new Error("render failed");

  writeRendererError(res, 500, error, {
    includeStack: true
  });

  const body = JSON.parse(res.body);

  assert.equal(res.statusCode, 500);
  assert.equal(body.error.message, "render failed");
  assert.match(body.error.stack, /render failed/);
});

test("createHealthResponse includes mode", () => {
  const got = createHealthResponse("dev");

  assert.deepEqual(got, {
    ok: true,
    mode: "dev"
  });
});
