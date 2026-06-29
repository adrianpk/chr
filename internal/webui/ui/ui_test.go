package ui

import (
	"strings"
	"testing"

	"forge.adrianpk.com/chc/internal/webui/htmx"
)

func TestButtonRenderEscapesTextAndIncludesHTMX(t *testing.T) {
	button := NewButton(`<Run>`).
		Primary().
		WithAttrs(htmx.HX().Post("/work").TargetID("workspace"))

	got := string(button.Render())
	for _, want := range []string{
		`class="button button--primary"`,
		`hx-post="/work"`,
		`hx-target="#workspace"`,
		`&lt;Run&gt;`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Button.Render() = %q, want %q", got, want)
		}
	}
}

func TestChipRenderEscapesText(t *testing.T) {
	got := string(NewChip(`<Ready>`).Muted().Render())
	for _, want := range []string{`chip chip--muted`, `&lt;Ready&gt;`} {
		if !strings.Contains(got, want) {
			t.Fatalf("Chip.Render() = %q, want %q", got, want)
		}
	}
}

func TestNavRenderMarksActiveLink(t *testing.T) {
	got := string(NewNav(
		NavLink{Label: "Project", Href: "/", Active: true},
		NavLink{Label: "Board", Href: "/board"},
	).Render())

	for _, want := range []string{
		`aria-label="Primary"`,
		`app-nav__link--active`,
		`href="/board"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Nav.Render() = %q, want %q", got, want)
		}
	}
}
