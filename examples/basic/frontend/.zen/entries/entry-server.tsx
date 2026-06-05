import renderToString from "preact-render-to-string";
import type { ComponentType } from "preact";

type ComponentModule = {
  default: ComponentType<Record<string, unknown>>;
};

type RenderRequest = {
  mode?: "page" | "island";
  url: string;
  page?: string;
  island?: string;
  props?: Record<string, unknown>;
};

const pageModules = import.meta.glob<ComponentModule>("../../src/pages/**/*.tsx", {
  eager: true
});

const pages = Object.fromEntries(
  Object.entries(pageModules).map(([path, mod]) => [
    path.replace("../../src/pages/", "").replace(/\.tsx$/, ""),
    mod.default
  ])
);

const islandModules = import.meta.glob<ComponentModule>("../../src/islands/**/*.tsx", {
  eager: true
});

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
      html: renderToString(<Island {...props} />),
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

  const html = renderToString(<Page {...props} />);

  return {
    html,
    head: ""
  };
}
