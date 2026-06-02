import { hydrate } from "preact";
import { pages, type PageName } from "../../src/pages";
import "./app.css";

type HydrationData = {
  page: PageName;
  props: Record<string, unknown>;
};

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
