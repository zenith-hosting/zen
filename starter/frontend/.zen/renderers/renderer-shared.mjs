export async function readJSON(req) {
  let body = "";

  for await (const chunk of req) {
    body += chunk.toString("utf8");
  }

  if (!body.trim()) {
    return {};
  }

  return JSON.parse(body);
}

export function writeJSON(res, status, value) {
  res.statusCode = status;
  res.setHeader("content-type", "application/json");
  res.end(JSON.stringify(value));
}

export function writeRendererError(res, status, error, options = {}) {
  const includeStack = Boolean(options.includeStack);

  writeJSON(res, status, {
    error: {
      message: error && error.message ? error.message : String(error),
      stack: includeStack && error && error.stack ? error.stack : ""
    }
  });
}
