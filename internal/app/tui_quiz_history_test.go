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

// Real retained entries feed both adapters. Cancellation, resume, Skip and Done
// keep saved history and session ownership across several question transitions.
func TestTUIQuizHistory(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	started, err := operations.StartQuiz(QuizStartRequest{Mode: quizcore.ModeFast})
	if err != nil {
		t.Fatal(err)
	}

	state := started.State
	for _, answer := range []string{"FIRST_SAVED_ANSWER", "SECOND_SAVED_ANSWER"} {
		result, err := operations.AnswerQuiz(QuizAnswerRequest{Binding: quizStateBinding(state, answer), Answer: answer})
		if err != nil {
			t.Fatal(err)
		}

		state = result.State
	}

	views, err := loadTUIQuizzes(operations)
	if err != nil || len(views[0].History) != 2 || views[0].History[0].Answer != "SECOND_SAVED_ANSWER" {
		t.Fatal(views, err)
	}

	web := webQuizDetail(t, provider, views[0].CuratedPath)
	if web.Quiz.History[0].Answer != views[0].History[0].Answer {
		t.Fatal("history order differs between adapters")
	}

	model := tui.NewModel(tui.Options{Quizzes: views, QuizExecutor: newTUIQuizActionExecutor(operations)})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	runGuidedTUIKey(t, model, "f3", false)
	runGuidedTUIKey(t, model, "enter", false)

	for _, want := range []string{"FIRST_SAVED_ANSWER", "SECOND_SAVED_ANSWER", state.ActiveQuestion.Prompt} {
		if !strings.Contains(model.View(), want) {
			t.Fatal("question or history is hidden", want, model.View())
		}
	}

	controlTUIQuizHistory(t, operations, model)
}

func controlTUIQuizHistory(t *testing.T, operations service, model *tui.Model) {
	t.Helper()

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x"), Alt: true})
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("cancelled confirmation changed retained sources")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x"), Alt: true})
	runGuidedTUIKey(t, model, "enter", true)
	runGuidedTUIKey(t, model, "u", true)
	runGuidedTUIKey(t, model, "alt+s", true)
	runGuidedTUIKey(t, model, "alt+d", true)

	observed, err := operations.quizRepository().Observe(t.Context(), operations.quizScope(quizcore.Scope{}))
	if err != nil || observed.State.AnswersRecorded != 2 || observed.State.SkipsRecorded != 1 || observed.State.Status != quizcore.RuntimeStatusDone {
		t.Fatal(observed.State, err)
	}

	if !strings.Contains(model.View(), "SECOND_SAVED_ANSWER") {
		t.Fatal("controls lost retained history", model.View())
	}
}
