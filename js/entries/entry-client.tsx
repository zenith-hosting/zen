import { hydrate } from "preact";
import type { ComponentType } from "preact";
import "../../src/app.css";

type PageModule = {
  default: ComponentType<Record<string, unknown>>;
};

type HydrationData = {
  page: string;
  props: Record<string, unknown>;
};

const modules = import.meta.glob<PageModule>("../../src/pages/**/*.tsx", {
  eager: true
});

const pages = Object.fromEntries(
  Object.entries(modules).map(([path, mod]) => [
    path.replace("../../src/pages/", "").replace(/\.tsx$/, ""),
    mod.default
  ])
);

const element = document.getElementById("__ZEN_DATA__");

if (!element || !element.textContent) {
  throw new Error("Missing Zen hydration data");
}

const data = JSON.parse(element.textContent) as HydrationData;
const Page = pages[data.page];

if (!Page) {
  throw new Error(`Unknown page: ${data.page}`);
}

const app = document.getElementById("app");

if (!app) {
  throw new Error("Missing app element");
}

hydrate(<Page {...data.props} />, app);
