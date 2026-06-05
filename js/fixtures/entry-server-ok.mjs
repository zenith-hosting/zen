export async function render(request) {
  if (request.mode === "island") {
    return {
      html: `<button data-island="${request.island}">${request.props.count}</button>`,
      head: ""
    };
  }

  return {
    html: `<main data-page="${request.page}">${request.props.title}</main>`,
    head: ""
  };
}
