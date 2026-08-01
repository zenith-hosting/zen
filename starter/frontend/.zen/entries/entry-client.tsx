import type { ComponentType } from "react";
import { hydrateRoot } from "react-dom/client";
import "../../src/globals.css";

type PageModule = {
  default: ComponentType<Record<string, unknown>>;
};

type HydrationData = {
  page?: string;
  island?: string;
  identifierPrefix?: string;
  props: Record<string, unknown>;
};

const pageModules = import.meta.glob<PageModule>([
  "../../src/pages/**/*.tsx",
  "!../../src/pages/**/*.test.tsx",
  "!../../src/pages/**/*.spec.tsx"
]);

const islandModules = import.meta.glob<PageModule>([
  "../../src/islands/**/*.tsx",
  "!../../src/islands/**/*.test.tsx",
  "!../../src/islands/**/*.spec.tsx"
]);

async function hydratePage() {
  const element = document.getElementById("__ZEN_DATA__");

  if (!element) {
    return;
  }

  if (!element.textContent) {
    throw new Error("Missing Zen hydration data");
  }

  const data = JSON.parse(element.textContent) as HydrationData;
  const load = pageModules[`../../src/pages/${data.page}.tsx`];

  if (!load) {
    throw new Error(`Unknown page: ${data.page}`);
  }

  const app = document.getElementById("app");

  if (!app) {
    throw new Error("Missing app element");
  }

  const { default: Page } = await load();

  hydrateRoot(app, <Page {...data.props} />, {
    identifierPrefix: data.identifierPrefix
  });
}

const hydratedIslands = new WeakSet<HTMLElement>();

async function hydrateIsland(root: HTMLElement) {
  if (hydratedIslands.has(root)) {
    return;
  }

  const mount = root.querySelector<HTMLElement>("[data-zen-island]");
  const element = root.querySelector<HTMLScriptElement>("[data-zen-island-props]");

  if (!mount || !element || !element.textContent) {
    throw new Error("Missing Zen island hydration data");
  }

  const data = JSON.parse(element.textContent) as HydrationData;
  const load = islandModules[`../../src/islands/${data.island}.tsx`];

  if (!load) {
    throw new Error(`Unknown island: ${data.island}`);
  }

  hydratedIslands.add(root);
  const { default: Island } = await load();

  hydrateRoot(mount, <Island {...data.props} />, {
    identifierPrefix: data.identifierPrefix
  });
}

function hydrateIslands(root: ParentNode = document) {
  if (root instanceof HTMLElement && root.matches("[data-zen-island-root]")) {
    void hydrateIsland(root);
  }

  for (const island of root.querySelectorAll<HTMLElement>("[data-zen-island-root]")) {
    void hydrateIsland(island);
  }
}

void hydratePage();
hydrateIslands();

new MutationObserver((records) => {
  for (const record of records) {
    for (const node of record.addedNodes) {
      if (node instanceof HTMLElement) {
        hydrateIslands(node);
      }
    }
  }
}).observe(document.body, { childList: true, subtree: true });
