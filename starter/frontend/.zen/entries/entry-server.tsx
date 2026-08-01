import type { ComponentType } from "react";
import { renderToString } from "react-dom/server";

type ComponentModule = {
  default: ComponentType<Record<string, unknown>>;
};

type RenderRequest = {
  mode?: "page" | "island";
  url: string;
  page?: string;
  island?: string;
  identifierPrefix?: string;
  props?: Record<string, unknown>;
};

const pageModules = import.meta.glob<ComponentModule>(
  [
    "../../src/pages/**/*.tsx",
    "!../../src/pages/**/*.test.tsx",
    "!../../src/pages/**/*.spec.tsx"
  ],
  { eager: true }
);

const pages = Object.fromEntries(
  Object.entries(pageModules).map(([path, mod]) => [
    path.replace("../../src/pages/", "").replace(/\.tsx$/, ""),
    mod.default
  ])
);

const islandModules = import.meta.glob<ComponentModule>(
  [
    "../../src/islands/**/*.tsx",
    "!../../src/islands/**/*.test.tsx",
    "!../../src/islands/**/*.spec.tsx"
  ],
  { eager: true }
);

const islands = Object.fromEntries(
  Object.entries(islandModules).map(([path, mod]) => [
    path.replace("../../src/islands/", "").replace(/\.tsx$/, ""),
    mod.default
  ])
);

export async function render(request: RenderRequest) {
  const mode = request.mode ?? "page";
  const props = request.props ?? {};

  if (mode === "island") {
    const Island = islands[request.island ?? ""];

    if (!Island) {
      throw new Error(`Unknown island: ${request.island}`);
    }

    return {
      html: renderToString(<Island {...props} />, {
        identifierPrefix: request.identifierPrefix
      }),
      head: ""
    };
  }

  if (mode !== "page") {
    throw new Error(`Unknown render mode: ${mode}`);
  }

  const Page = pages[request.page ?? ""];

  if (!Page) {
    throw new Error(`Unknown page: ${request.page}`);
  }

  const html = renderToString(<Page {...props} />, {
    identifierPrefix: request.identifierPrefix
  });

  return {
    html,
    head: ""
  };
}
