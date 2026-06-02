import renderToString from "preact-render-to-string";
import { pages, type PageName } from "./pages";

type RenderRequest = {
  url: string;
  page: PageName;
  props: Record<string, unknown>;
};

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
