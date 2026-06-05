import { hydrate } from "preact";
import type { ComponentType } from "preact";
import "../../src/app.css";

type PageModule = {
  default: ComponentType<Record<string, unknown>>;
};

type HydrationData = {
  page?: string;
  island?: string;
  props: Record<string, unknown>;
};

const pageModules = import.meta.glob<PageModule>("../../src/pages/**/*.tsx", {
  eager: true
});

const pages = Object.fromEntries(
  Object.entries(pageModules).map(([path, mod]) => [
    path.replace("../../src/pages/", "").replace(/\.tsx$/, ""),
    mod.default
  ])
);

const islandModules = import.meta.glob<PageModule>("../../src/islands/**/*.tsx", {
  eager: true
});

const islands = Object.fromEntries(
  Object.entries(islandModules).map(([path, mod]) => [
    path.replace("../../src/islands/", "").replace(/\.tsx$/, ""),
    mod.default
  ])
);

function hydratePage() {
  const element = document.getElementById("__ZEN_DATA__");

  if (!element) {
    return;
  }

  if (!element.textContent) {
    throw new Error("Missing Zen hydration data");
  }

  const data = JSON.parse(element.textContent) as HydrationData;
  const Page = pages[data.page ?? ""];

  if (!Page) {
    throw new Error(`Unknown page: ${data.page}`);
  }

  const app = document.getElementById("app");

  if (!app) {
    throw new Error("Missing app element");
  }

  hydrate(<Page {...data.props} />, app);
}

function hydrateIslands() {
  for (const root of document.querySelectorAll<HTMLElement>("[data-zen-island-root]")) {
    const mount = root.querySelector<HTMLElement>("[data-zen-island]");
    const element = root.querySelector<HTMLScriptElement>("[data-zen-island-props]");

    if (!mount || !element || !element.textContent) {
      throw new Error("Missing Zen island hydration data");
    }

    const data = JSON.parse(element.textContent) as HydrationData;
    const Island = islands[data.island ?? ""];

    if (!Island) {
      throw new Error(`Unknown island: ${data.island}`);
    }

    hydrate(<Island {...data.props} />, mount);
  }
}

hydratePage();
hydrateIslands();
