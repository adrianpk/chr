package app

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/tui"
)

// A queued terminal answer cannot follow a concurrent writer to another
// question. Confirmed reload discards only the old draft; duplicate execution
// of a successful captured command reconciles once without another answer.
func TestTUIQuizDraft(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	started, err := operations.StartQuiz(QuizStartRequest{Mode: quizcore.ModeFast})
	if err != nil {
		t.Fatal(err)
	}

	views, err := loadTUIQuizzes(operations)
	if err != nil {
		t.Fatal(err)
	}

	model := tui.NewModel(tui.Options{Quizzes: views, QuizExecutor: newTUIQuizActionExecutor(operations)})
	runGuidedTUIKey(t, model, "f3", false)
	runGuidedTUIKey(t, model, "enter", false)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("RETAINED_STALE_ANSWER")})

	_, command := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if command == nil {
		t.Fatal("answer did not capture a command")
	}

	if _, err := operations.AnswerQuiz(QuizAnswerRequest{Binding: quizStateBinding(started.State, "concurrent"), Answer: "CONCURRENT_ANSWER"}); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))

	model.Update(command())

	if !strings.Contains(model.View(), "Operation: stale") || !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("queued answer changed a newer question", model.View())
	}

	reloadTUIQuizDraft(t, model, started.State.ActiveQuestion.Prompt)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("FRESH_ANSWER")})

	_, command = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model.Update(command())

	after := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))

	model.Update(command())

	if !maps.Equal(after, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("duplicate command answered another question")
	}
}

func reloadTUIQuizDraft(t *testing.T, model *tui.Model, prompt string) {
	t.Helper()
	runGuidedTUIKey(t, model, "r", false)
	runGuidedTUIKey(t, model, "esc", false)
	runGuidedTUIKey(t, model, "esc", false)

	if !strings.Contains(model.View(), "RETAINED_STALE_ANSWER") || !strings.Contains(model.View(), prompt) {
		t.Fatal("cancelled reload changed input or its captured question", model.View())
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o"), Alt: true})
	runGuidedTUIKey(t, model, "r", false)
	runGuidedTUIKey(t, model, "enter", true)

	if strings.Contains(model.View(), "RETAINED_STALE_ANSWER") || strings.Contains(model.View(), prompt) && !strings.Contains(model.View(), "CONCURRENT_ANSWER") {
		t.Fatal("confirmed reload did not retain current history", model.View())
	}
}
