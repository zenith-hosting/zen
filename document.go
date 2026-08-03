package zen

import (
	"html"
	"strings"
)

type documentInput struct {
	Title            string
	HeadHTML         string
	BaseElements     []headElement
	MetaElements     []headElement
	LinkElements     []headElement
	HeadScripts      []headElement
	CustomCSS        string
	AppHTML          string
	HydrationJSON    string
	StylesheetURLs   []string
	CompiledCSS      string
	ModuleScriptURLs []string
}

func renderDocument(input documentInput) string {
	var b strings.Builder

	b.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>")
	b.WriteString(html.EscapeString(input.Title))
	b.WriteString("</title>\n")
	b.WriteString(input.HeadHTML)
	if input.HeadHTML != "" && !strings.HasSuffix(input.HeadHTML, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString(headElementTags(input.BaseElements))
	b.WriteString(headElementTags(input.MetaElements))
	b.WriteString(headElementTags(input.LinkElements))
	b.WriteString(styleTag(input.CustomCSS))
	b.WriteString(styleTag(input.CompiledCSS))
	b.WriteString(stylesheetTags(input.StylesheetURLs))
	b.WriteString(headElementTags(input.HeadScripts))
	b.WriteString("</head>\n<body>\n<div id=\"app\">")
	b.WriteString(input.AppHTML)
	b.WriteString("</div>\n")
	b.WriteString(hydrationDataScript(input.HydrationJSON))
	b.WriteByte('\n')
	b.WriteString(scriptTags(input.ModuleScriptURLs))
	b.WriteString("</body>\n</html>\n")

	return b.String()
}

func stylesheetTags(styles []string) string {
	var b strings.Builder

	for _, href := range styles {
		b.WriteString(`<link rel="stylesheet" href="`)
		b.WriteString(html.EscapeString(href))
		b.WriteString(`">` + "\n")
	}

	return b.String()
}

func scriptTags(scripts []string) string {
	var b strings.Builder

	for _, src := range scripts {
		b.WriteString(`<script type="module" src="`)
		b.WriteString(html.EscapeString(src))
		b.WriteString(`"></script>` + "\n")
	}

	return b.String()
}

func hydrationDataScript(json string) string {
	return `<script id="__ZEN_DATA__" type="application/json">` + json + "</script>"
}

func renderIslandFragment(island, renderedHTML, hydrationJSON string) string {
	var b strings.Builder

	b.WriteString(`<div data-zen-island-root>`)
	b.WriteString(`<div data-zen-island="`)
	b.WriteString(html.EscapeString(island))
	b.WriteString(`">`)
	b.WriteString(renderedHTML)
	b.WriteString("</div>")
	b.WriteString(`<script type="application/json" data-zen-island-props>`)
	b.WriteString(hydrationJSON)
	b.WriteString("</script>")
	b.WriteString("</div>\n")

	return b.String()
}
