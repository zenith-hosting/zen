package zen

import (
	"html"
	"strings"
)

type documentInput struct {
	Title         string
	AppElementID  string
	DataElementID string
	HTML          string
	HydrationJSON string
	Styles        []string
	Scripts       []string
	DevScripts    []string
}

type islandFragmentInput struct {
	Island        string
	HTML          string
	HydrationJSON string
	Styles        []string
	Scripts       []string
	DevScripts    []string
}

func renderDocument(input documentInput) string {
	var b strings.Builder

	b.WriteString("<!doctype html>\n")
	b.WriteString("<html lang=\"en\">\n")
	b.WriteString("<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>")
	b.WriteString(html.EscapeString(input.Title))
	b.WriteString("</title>\n")

	for _, href := range input.Styles {
		b.WriteString(`<link rel="stylesheet" href="`)
		b.WriteString(html.EscapeString(href))
		b.WriteString(`">` + "\n")
	}

	for _, src := range input.DevScripts {
		b.WriteString(`<script type="module" src="`)
		b.WriteString(html.EscapeString(src))
		b.WriteString(`"></script>` + "\n")
	}

	b.WriteString("</head>\n")
	b.WriteString("<body>\n")
	b.WriteString(`<div id="`)
	b.WriteString(html.EscapeString(input.AppElementID))
	b.WriteString(`">`)
	b.WriteString(input.HTML)
	b.WriteString("</div>\n")

	b.WriteString(`<script id="`)
	b.WriteString(html.EscapeString(input.DataElementID))
	b.WriteString(`" type="application/json">`)
	b.WriteString(input.HydrationJSON)
	b.WriteString("</script>\n")

	for _, src := range input.Scripts {
		b.WriteString(`<script type="module" src="`)
		b.WriteString(html.EscapeString(src))
		b.WriteString(`"></script>` + "\n")
	}

	b.WriteString("</body>\n")
	b.WriteString("</html>\n")

	return b.String()
}

func renderIslandFragment(input islandFragmentInput) string {
	var b strings.Builder

	for _, href := range input.Styles {
		b.WriteString(`<link rel="stylesheet" href="`)
		b.WriteString(html.EscapeString(href))
		b.WriteString(`">` + "\n")
	}

	b.WriteString(`<div data-zen-island-root>`)
	b.WriteString(`<div data-zen-island="`)
	b.WriteString(html.EscapeString(input.Island))
	b.WriteString(`">`)
	b.WriteString(input.HTML)
	b.WriteString("</div>")
	b.WriteString(`<script type="application/json" data-zen-island-props>`)
	b.WriteString(input.HydrationJSON)
	b.WriteString("</script>")
	b.WriteString("</div>\n")

	for _, src := range input.DevScripts {
		b.WriteString(`<script type="module" src="`)
		b.WriteString(html.EscapeString(src))
		b.WriteString(`"></script>` + "\n")
	}

	for _, src := range input.Scripts {
		b.WriteString(`<script type="module" src="`)
		b.WriteString(html.EscapeString(src))
		b.WriteString(`"></script>` + "\n")
	}

	return b.String()
}
