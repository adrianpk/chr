package ui

import (
	"fmt"
	"html/template"
	"strings"

	"forge.adrianpk.com/chc/internal/webui/htmx"
)

type Variant string

const (
	VariantDefault Variant = ""
	VariantPrimary Variant = "primary"
	VariantMuted   Variant = "muted"
	VariantDanger  Variant = "danger"
)

type Button struct {
	text    string
	variant Variant
	type_   string
	attrs   *htmx.Attrs
}

func NewButton(text string) *Button {
	return &Button{text: text, type_: "button"}
}

func (b *Button) Primary() *Button {
	b.variant = VariantPrimary

	return b
}

func (b *Button) Danger() *Button {
	b.variant = VariantDanger

	return b
}

func (b *Button) Type(value string) *Button {
	b.type_ = value

	return b
}

func (b *Button) WithAttrs(attrs *htmx.Attrs) *Button {
	b.attrs = attrs

	return b
}

func (b *Button) Render() template.HTML {
	classes := []string{"button"}
	if b.variant != "" {
		classes = append(classes, "button--"+string(b.variant))
	}

	attrs := []string{
		fmt.Sprintf(`class="%s"`, template.HTMLEscapeString(strings.Join(classes, " "))),
		fmt.Sprintf(`type="%s"`, template.HTMLEscapeString(b.type_)),
	}
	if b.attrs != nil && b.attrs.String() != "" {
		attrs = append(attrs, b.attrs.String())
	}

	return template.HTML(fmt.Sprintf(
		`<button %s>%s</button>`,
		strings.Join(attrs, " "),
		template.HTMLEscapeString(b.text),
	))
}

type Chip struct {
	text    string
	variant Variant
}

func NewChip(text string) *Chip {
	return &Chip{text: text}
}

func (c *Chip) Muted() *Chip {
	c.variant = VariantMuted

	return c
}

func (c *Chip) Primary() *Chip {
	c.variant = VariantPrimary

	return c
}

func (c *Chip) Render() template.HTML {
	classes := []string{"chip"}
	if c.variant != "" {
		classes = append(classes, "chip--"+string(c.variant))
	}

	return template.HTML(fmt.Sprintf(
		`<span class="%s">%s</span>`,
		template.HTMLEscapeString(strings.Join(classes, " ")),
		template.HTMLEscapeString(c.text),
	))
}

type NavLink struct {
	Label  string
	Href   string
	Active bool
}

type Nav struct {
	links []NavLink
}

func NewNav(links ...NavLink) *Nav {
	return &Nav{links: links}
}

func (n *Nav) Render() template.HTML {
	var html strings.Builder
	html.WriteString(`<nav class="app-nav" aria-label="Primary">`)
	for _, link := range n.links {
		class := "app-nav__link"
		if link.Active {
			class += " app-nav__link--active"
		}

		fmt.Fprintf(
			&html,
			`<a class="%s" href="%s">%s</a>`,
			template.HTMLEscapeString(class),
			template.HTMLEscapeString(link.Href),
			template.HTMLEscapeString(link.Label),
		)
	}
	html.WriteString(`</nav>`)

	return template.HTML(html.String())
}

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"button": NewButton,
		"chip":   NewChip,
		"nav": func(links []NavLink) *Nav {
			return NewNav(links...)
		},
	}
}
