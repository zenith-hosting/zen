export async function render(request) {
  return {
    html: `<main data-page="${request.page}">${request.props.title}</main>`,
    head: ""
  };
}
