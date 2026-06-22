package zen

import (
	"html"
	"strings"
)

type Attribute struct {
	name  string
	value string
	text  string
}

type headElement struct {
	tag        string
	attrs      []Attribute
	textChunks []string
}

func Attr(name string, value string) Attribute {
	return Attribute{name: name, value: value}
}

func Name(value string) Attribute {
	return Attribute{name: "name", value: value}
}

func Content(value string) Attribute {
	return Attribute{name: "content", value: value}
}

func Rel(value string) Attribute {
	return Attribute{name: "rel", value: value}
}

func Href(value string) Attribute {
	return Attribute{name: "href", value: value}
}

func Type(value string) Attribute {
	return Attribute{name: "type", value: value}
}

func Src(value string) Attribute {
	return Attribute{name: "src", value: value}
}

func Property(value string) Attribute {
	return Attribute{name: "property", value: value}
}

func Text(value string) Attribute {
	return Attribute{text: value}
}

func WithBase(attrs ...Attribute) RenderOption {
	return func(opts *renderOptions) {
		opts.Base = append(opts.Base, newHeadElement("base", attrs...))
	}
}

func Base(attrs ...Attribute) RenderOption {
	return WithBase(attrs...)
}

func WithMeta(attrs ...Attribute) RenderOption {
	return func(opts *renderOptions) {
		opts.Meta = append(opts.Meta, newHeadElement("meta", attrs...))
	}
}

func Meta(attrs ...Attribute) RenderOption {
	return WithMeta(attrs...)
}

func WithLink(attrs ...Attribute) RenderOption {
	return func(opts *renderOptions) {
		opts.Link = append(opts.Link, newHeadElement("link", attrs...))
	}
}

func Link(attrs ...Attribute) RenderOption {
	return WithLink(attrs...)
}

func WithScript(attrs ...Attribute) RenderOption {
	return func(opts *renderOptions) {
		opts.Script = append(opts.Script, newHeadElement("script", attrs...))
	}
}

func Script(attrs ...Attribute) RenderOption {
	return WithScript(attrs...)
}

func WithStyle(style string) RenderOption {
	return func(opts *renderOptions) {
		opts.Style = style
	}
}

func Style(style string) RenderOption {
	return WithStyle(style)
}

func newHeadElement(tag string, attrs ...Attribute) headElement {
	element := headElement{tag: tag}
	for _, attr := range attrs {
		if attr.text != "" {
			element.textChunks = append(element.textChunks, attr.text)
			continue
		}
		element.attrs = append(element.attrs, attr)
	}
	return element
}

func headElementTags(elements []headElement) string {
	var b strings.Builder

	for _, element := range elements {
		if element.tag == "" {
			continue
		}

		b.WriteString("<")
		b.WriteString(element.tag)
		for _, attr := range element.attrs {
			if !isSafeAttributeName(attr.name) {
				continue
			}
			b.WriteString(" ")
			b.WriteString(attr.name)
			b.WriteString(`="`)
			b.WriteString(html.EscapeString(attr.value))
			b.WriteString(`"`)
		}

		if element.tag == "script" || element.tag == "style" {
			b.WriteString(">")
			b.WriteString(strings.Join(element.textChunks, ""))
			b.WriteString("</")
			b.WriteString(element.tag)
			b.WriteString(">\n")
			continue
		}

		b.WriteString(">\n")
	}

	return b.String()
}

func styleTag(style string) string {
	if style == "" {
		return ""
	}
	return "<style>" + style + "</style>\n"
}

func isSafeAttributeName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		switch r {
		case '-', '_', ':', '.':
			continue
		}
		return false
	}
	return true
}
