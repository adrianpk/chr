package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
)

// The unsupported contextual-authoring entry explains availability before any
// confirmation or executor call; supported scoped runtime quizzes remain separate.
func TestQuizContextUnavailable(t *testing.T) {
	executor := &fakeArtifactActionExecutor{}
	model := NewModel(Options{ActionExecutor: executor})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("context-aware")})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if model.artifactAction.Pending || executor.action != "" || model.feedback.Result.Status != operation.Blocked || !strings.Contains(model.View(), "chr author quiz new") {
		t.Fatal("unsupported entry invited a known failing operation", model.View())
	}
}
