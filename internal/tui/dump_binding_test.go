package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// An unchanged failed Save keeps its operation and observed revision. Editing
// the retained text creates a new intent while preserving that source guard.
func TestDumpEditBinding(t *testing.T) {
	executor := &fakeDumpActionExecutor{err: errors.New("lost acknowledgement")}
	model := NewModel(Options{DumpExecutor: executor})
	revision := strings.Repeat("a", 64)
	model.startDumpEdit("ops/default/dump/context/artifact.md", "Original material.", revision)

	operationID := model.dumpEdit.OperationID
	for range 2 {
		model.saveDumpEdit()

		if !model.dumpEdit.Active || executor.request.OperationID != operationID || executor.request.Revision != revision {
			t.Fatal("failed Save lost its intent or source binding", executor.request)
		}
	}

	updated, _ := model.updateDumpEditKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	edited, ok := updated.(*Model)
	if !ok {
		t.Fatalf("unexpected model %T", updated)
	}

	if edited.dumpEdit.OperationID == operationID || edited.dumpEdit.Revision != revision {
		t.Fatal("changed text reused the old intent or discarded its source guard")
	}
}
