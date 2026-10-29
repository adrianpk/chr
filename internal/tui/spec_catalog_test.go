package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Refresh reconciles the selected concrete identity, even when another document
// in its case is inserted before it. Removing it chooses a current visible row.
func TestSpecCatalogSelection(t *testing.T) {
	first := SpecView{Title: "First", RelativePath: "first", Scope: QuizScope{CaseSlug: "shared"}}
	second := SpecView{Title: "Second", RelativePath: "second", Scope: first.Scope}
	model := NewModel(Options{Specs: []SpecView{first, second}})
	model.selectView(ViewSpecs)

	model.selectedSpec = 1
	model.applySpecActionResult(SpecActionResult{Specs: []SpecView{first, {Title: "Inserted", RelativePath: "inserted", Scope: first.Scope}, second}, Scope: first.Scope})

	selected, ok := model.selectedSpecView()
	if !ok || selected.RelativePath != "second" {
		t.Fatal("refresh changed concrete selection", selected)
	}

	model.applySpecActionResult(SpecActionResult{Specs: []SpecView{first}, Scope: first.Scope})

	selected, ok = model.selectedSpecView()
	if !ok || selected.RelativePath != "first" {
		t.Fatal("removed selection retained a stale target", selected)
	}
}

// Real palette keys reach the same form owner as direct creation, without a
// separate confirmation path or loss of its text-entry mode.
func TestSpecFormPalette(t *testing.T) {
	model := NewModel(Options{})
	model.selectView(ViewSpecs)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("create spec")})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !model.textForm.Active || model.textForm.SubmitLabel != "Create" || model.specAction.Pending {
		t.Fatal("palette forked the form owner")
	}
}
