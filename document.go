package zen

import (
	"fmt"
	"html"
	"strings"
)

const (
	documentSlotTitle   = "<!--zen:title-->"
	documentSlotHead    = "<!--zen:head-->"
	documentSlotBase    = "<!--zen:base-->"
	documentSlotMeta    = "<!--zen:meta-->"
	documentSlotLink    = "<!--zen:link-->"
	documentSlotStyle   = "<!--zen:style-->"
	documentSlotStyles  = "<!--zen:styles-->"
	documentSlotScript  = "<!--zen:script-->"
	documentSlotApp     = "<!--zen:app-->"
	documentSlotData    = "<!--zen:data-->"
	documentSlotScripts = "<!--zen:scripts-->"
)

const defaultDocumentTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title><!--zen:title--></title>
<!--zen:head-->
<!--zen:base-->
<!--zen:meta-->
<!--zen:link-->
<!--zen:style-->
<!--zen:styles-->
<!--zen:script-->
</head>
<body>
<div id="app"><!--zen:app--></div>
<!--zen:data-->
<!--zen:scripts-->
</body>
</html>
`

type documentInput struct {
	Title         string
	AppElementID  string
	DataElementID string
	Head          string
	Base          []headElement
	Meta          []headElement
	Link          []headElement
	Script        []headElement
	Style         string
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

func requiredDocumentSlots() []string {
	return []string{
		documentSlotTitle,
		documentSlotHead,
		documentSlotBase,
		documentSlotMeta,
		documentSlotLink,
		documentSlotStyle,
		documentSlotStyles,
		documentSlotScript,
		documentSlotApp,
		documentSlotData,
		documentSlotScripts,
	}
}

func renderDocumentTemplate(template string, input documentInput) (string, error) {
	for _, slot := range requiredDocumentSlots() {
		if !strings.Contains(template, slot) {
			return "", fmt.Errorf("zen: missing required document slot %s", slot)
		}
	}

	replacements := map[string]string{
		documentSlotTitle:   html.EscapeString(input.Title),
		documentSlotHead:    input.Head,
		documentSlotBase:    headElementTags(input.Base),
		documentSlotMeta:    headElementTags(input.Meta),
		documentSlotLink:    headElementTags(input.Link),
		documentSlotStyle:   styleTag(input.Style),
		documentSlotStyles:  stylesheetTags(input.Styles),
		documentSlotScript:  headElementTags(input.Script),
		documentSlotApp:     input.HTML,
		documentSlotData:    hydrationDataScript(input.DataElementID, input.HydrationJSON),
		documentSlotScripts: scriptTags(append(append([]string{}, input.DevScripts...), input.Scripts...)),
	}

	out := template
	for slot, value := range replacements {
		out = strings.ReplaceAll(out, slot, value)
	}

	return out, nil
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

func hydrationDataScript(id string, json string) string {
	var b strings.Builder

	b.WriteString(`<script id="`)
	b.WriteString(html.EscapeString(id))
	b.WriteString(`" type="application/json">`)
	b.WriteString(json)
	b.WriteString("</script>")

	return b.String()
}

func renderIslandFragment(input islandFragmentInput) string {
	var b strings.Builder

	b.WriteString(stylesheetTags(input.Styles))

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

	b.WriteString(scriptTags(input.DevScripts))
	b.WriteString(scriptTags(input.Scripts))

	return b.String()
}
