import renderToString from "preact-render-to-string";
//@ts-ignore
import { pages, type PageName } from "../../src/pages";

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

  //@ts-ignore
  const html = renderToString(<Page {...request.props} />);

  return {
    html,
    head: ""
  };
}
