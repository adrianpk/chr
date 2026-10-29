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

// Newline edits are new submission intent just like typed text; reusing an
// earlier request identity could reconcile a different captured answer body.
func TestQuizInputIdentity(t *testing.T) {
	model := NewModel(Options{Quizzes: []QuizView{{Status: "active", ActiveQuestionSummary: "Question?", Binding: QuizActionBinding{ExpectedRevision: "revision"}}}})
	model.selectView(ViewQuizzes)
	model.startQuizAnswer()

	before := model.quizAnswer.Binding.OperationID
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if model.quizAnswer.Binding.OperationID == before || !strings.Contains(model.quizAnswer.Input.Value(), "\n") {
		t.Fatal("newline reused an older input identity")
	}
}

// A read-only refresh may observe a newer question while editing. Its prompt
// cannot be paired with an answer still bound to the earlier question.
func TestQuizDraftRefresh(t *testing.T) {
	model := NewModel(Options{Quizzes: []QuizView{{Status: "active", ActiveQuestionSummary: "CAPTURED_PROMPT", Binding: QuizActionBinding{QuestionID: "first", ExpectedRevision: "before"}}}})
	model.selectView(ViewQuizzes)
	model.startQuizAnswer()
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("CAPTURED_INPUT")})
	model.applyQuizActionResult(QuizActionResult{Quizzes: []QuizView{{Status: "active", ActiveQuestionSummary: "NEWER_PROMPT", Binding: QuizActionBinding{QuestionID: "second", ExpectedRevision: "after"}}}})

	view := model.View()
	if !strings.Contains(view, "CAPTURED_PROMPT") || strings.Contains(view, "NEWER_PROMPT") || model.quizAnswer.Input.Value() != "CAPTURED_INPUT" || model.quizAnswer.Binding.QuestionID != "first" {
		t.Fatal("refresh paired the draft with another question", view)
	}
}
