package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// History paging and resize keep the current prompt and unsent textarea intact,
// and the complete final saved answer remains reachable in a narrow terminal.
func TestQuizHistoryPaging(t *testing.T) {
	history := []QuizAnswerRecord{}
	for index := range 20 {
		history = append(history, QuizAnswerRecord{Question: "Saved question?", Answer: fmt.Sprintf("SAVED_ANSWER_%02d", index)})
	}

	model := NewModel(Options{Quizzes: []QuizView{{Status: "active", ActiveQuestionSummary: "CAPTURED_QUESTION", History: history}}})
	model.selectView(ViewQuizzes)
	model.startQuizAnswer()
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("UNSENT_ANSWER")})
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for range 100 {
		model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	}

	view := strings.Join(model.renderSelectedQuizDetailLines(76), "\n")
	if !strings.Contains(view, "SAVED_ANSWER_19") || !strings.Contains(view, "CAPTURED_QUESTION") || model.quizAnswer.Input.Value() != "UNSENT_ANSWER" {
		t.Fatal("paging lost history or the current input", view)
	}
}
