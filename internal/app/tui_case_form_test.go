package app

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// Real metadata validation keeps every draft field and writes no partial case;
// one corrected Create submits the complete form through the production adapter.
func TestTUICaseForm(t *testing.T) {
	operations := navigationTerminalFixture(t)
	model := tui.NewModel(tui.Options{ProjectDir: operations.projectContext.ProjectDir, CaseExecutor: newTUICaseMetadataExecutor(operations)})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	runGuidedTUIKey(t, model, "f1", false)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Description sentinel.")})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Context sentinel.")})

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))

	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if !strings.Contains(model.View(), "validation_failed") || !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("invalid form wrote a partial case", model.View())
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Complete case")})
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	saved, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadPath(authoringcore.CaseMetadataPath(operations.projectContext.ProjectDir, "complete-case"))
	if err != nil || saved.Description != "Description sentinel." || saved.Context != "Context sentinel." {
		t.Fatal(saved, err)
	}
}
