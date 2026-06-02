import renderToString from "preact-render-to-string";
import type { ComponentType } from "preact";

type PageModule = {
  default: ComponentType<Record<string, unknown>>;
};

type RenderRequest = {
  url: string;
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

export async function render(request: RenderRequest) {
  const Page = pages[request.page];

  if (!Page) {
    throw new Error(`Unknown page: ${request.page}`);
  }

  const html = renderToString(<Page {...request.props} />);

  return {
    html,
    head: ""
  };
}
