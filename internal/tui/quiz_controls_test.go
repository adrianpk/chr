package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Mode selection is visible before Start and local to the selected target;
// pending-answer markers distinguish an open question from an untouched scope.
func TestQuizModeSelection(t *testing.T) {
	model := NewModel(Options{Quizzes: []QuizView{{Title: "Ready", Status: "not_started"}, {Title: "Pending", Status: "active", ActiveQuestionSummary: "Current question?"}}})
	model.selectView(ViewQuizzes)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})

	quiz, _ := model.selectedQuizView()
	if quiz.Mode != "fast" || model.quizzes[1].Mode != "deep" || !strings.Contains(model.View(), "Mode: Fast") || !strings.Contains(model.View(), "[answer pending]") {
		t.Fatal("mode choice or pending target was hidden", model.View())
	}

	request := model.newQuizActionState(QuizActionStart).Request
	if request.Mode != "fast" {
		t.Fatal("Start did not capture the visible mode", request)
	}
}
