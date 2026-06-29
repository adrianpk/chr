// Package ui provides small server-rendered UI primitives.
package ui

import (
	"fmt"
	"html/template"
	"strings"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

// Variant identifies a visual component variant.
type Variant string

const (
	// VariantDefault uses the default visual style.
	VariantDefault Variant = ""
	// VariantPrimary uses the primary visual style.
	VariantPrimary Variant = "primary"
	// VariantMuted uses the muted visual style.
	VariantMuted Variant = "muted"
	// VariantDanger uses the danger visual style.
	VariantDanger Variant = "danger"
)

// Button renders a simple HTML button.
type Button struct {
	text    string
	variant Variant
	type_   string
	attrs   *htmx.Attrs
}

// NewButton creates a button primitive.
func NewButton(text string) *Button {
	return &Button{text: text, type_: "button"}
}

// Primary applies the primary button variant.
func (b *Button) Primary() *Button {
	b.variant = VariantPrimary

	return b
}

// Danger applies the danger button variant.
func (b *Button) Danger() *Button {
	b.variant = VariantDanger

	return b
}

// Type sets the button type attribute.
func (b *Button) Type(value string) *Button {
	b.type_ = value

	return b
}

// WithAttrs attaches htmx attributes to the button.
func (b *Button) WithAttrs(attrs *htmx.Attrs) *Button {
	b.attrs = attrs

	return b
}

// Render renders the button as escaped HTML.
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

// Chip renders a compact label.
type Chip struct {
	text    string
	variant Variant
}

// NewChip creates a chip primitive.
func NewChip(text any) *Chip {
	return &Chip{text: fmt.Sprint(text)}
}

// Muted applies the muted chip variant.
func (c *Chip) Muted() *Chip {
	c.variant = VariantMuted

	return c
}

// Primary applies the primary chip variant.
func (c *Chip) Primary() *Chip {
	c.variant = VariantPrimary

	return c
}

// Render renders the chip as escaped HTML.
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

// NavLink describes one navigation link.
type NavLink struct {
	Label  string
	Href   string
	Active bool
}

// Nav renders a primary navigation group.
type Nav struct {
	links []NavLink
}

// NewNav creates a navigation primitive.
func NewNav(links ...NavLink) *Nav {
	return &Nav{links: links}
}

// Render renders the navigation as escaped HTML.
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

// FuncMap returns UI helpers for templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"button": NewButton,
		"chip":   NewChip,
		"nav": func(links []NavLink) *Nav {
			return NewNav(links...)
		},
	}
}
