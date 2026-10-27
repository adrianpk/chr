package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

type harnessSelectionFixture struct {
	view    HarnessSelectionView
	err     error
	saveErr error
	request HarnessSelectionRequest
}

func (f *harnessSelectionFixture) InspectHarnessSelections() (HarnessSelectionView, error) {
	return f.view, f.err
}

func (f *harnessSelectionFixture) SaveHarnessSelections(request HarnessSelectionRequest) (HarnessSelectionView, error) {
	f.request = request

	return f.view, f.saveErr
}

// The project key opens an asynchronous compact picker. Late reads cannot
// reopen a cancelled interaction or replace a newer project's visible choices.
func TestTUIHarnessSelectionRead(t *testing.T) {
	executor := &harnessSelectionFixture{view: HarnessSelectionView{Snapshot: harnesscore.HarnessSelectionSnapshot{Revision: strings.Repeat("a", 64)}}}
	model := NewModel(Options{HarnessExecutor: executor})
	updated, first := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})

	model = requireUpdatedModel(t, updated)
	if first == nil || !model.harnessSelection.Active || !model.harnessSelection.Busy {
		t.Fatal("project control did not start a bounded read")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = requireUpdatedModel(t, updated)
	updated, second := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	model = requireUpdatedModel(t, updated)
	updated, _ = model.Update(first())

	model = requireUpdatedModel(t, updated)
	if !model.harnessSelection.Busy {
		t.Fatal("late read replaced the reopened picker")
	}

	updated, _ = model.Update(second())

	model = requireUpdatedModel(t, updated)
	if model.harnessSelection.Busy || model.harnessSelection.Worker != harnessKeepProfile {
		t.Fatal("current read did not initialize choices")
	}
}

// Read failures give a public setup diagnostic, retain existing preview input
// and never render the executor's private error. The modal owns mouse focus.
func TestTUIHarnessSelectionFailure(t *testing.T) {
	executor := &harnessSelectionFixture{err: errors.New("private credential path")}
	model := NewModel(Options{HarnessExecutor: executor})
	model.harnessSelection = harnessSelectionState{Active: true, Worker: "attempted"}
	command := model.reloadHarnessPicker()
	updated, _ := model.Update(command())

	model = requireUpdatedModel(t, updated)
	if model.harnessSelection.Worker != "attempted" || !strings.Contains(model.harnessSelection.Message, "private profile setup") || strings.Contains(model.View(), "credential path") {
		t.Fatal("failed read discarded input or exposed setup")
	}

	updated, _ = model.Update(tea.MouseMsg{X: 5, Y: 23, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if requireUpdatedModel(t, updated).SelectedView() != ViewProject {
		t.Fatal("mouse navigation escaped the modal")
	}
}
