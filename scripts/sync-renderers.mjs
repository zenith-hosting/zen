import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), "..");

const files = [
  "dev-renderer.mjs",
  "prod-renderer.mjs",
  "renderer-shared.mjs"
];

const sourceDir = join(repoRoot, "js");
const targetDir = join(repoRoot, "zencli", "init_template", "frontend", ".zen", "renderers");

async function read(path) {
  return readFile(path, "utf8");
}

async function main() {
  const stale = [];

  await mkdir(targetDir, {
    recursive: true
  });

  for (const file of files) {
    const sourcePath = join(sourceDir, file);
    const targetPath = join(targetDir, file);
    const source = await read(sourcePath);
    let target = "";

    try {
      target = await read(targetPath);
    } catch (error) {
      if (error.code !== "ENOENT") {
        throw error;
      }
    }

    if (source === target) {
      continue;
    }

    stale.push(file);

    await writeFile(targetPath, source);
  }

  if (stale.length === 0) {
    return;
  }

  process.stdout.write(`Synced ${stale.length} renderer file(s).\n`);
}

main().catch((error) => {
  process.stderr.write((error && error.stack ? error.stack : String(error)) + "\n");
  process.exit(1);
});
