package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Unsubmitted fields survive navigation attempts, resize and a cancelled
// discard; confirming discard creates no catalog row or metadata write.
func TestCaseFormCancel(t *testing.T) {
	executor := &fakeCaseMetadataExecutor{}
	model := NewModel(Options{CaseExecutor: executor})
	model.newCase()
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("UNSENT_CASE")})
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model.Update(tea.KeyMsg{Type: tea.KeyF2})

	if model.SelectedView() != ViewCases {
		t.Fatal("form navigation escaped its owner")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !model.textForm.Active || model.textForm.Fields[0].Input.Value() != "UNSENT_CASE" {
		t.Fatal("cancelled discard lost input")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if model.textForm.Active || len(model.cases) != 0 || executor.saved.Name != "" {
		t.Fatal("discard created a blank case")
	}
}
