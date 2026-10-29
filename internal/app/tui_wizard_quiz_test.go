package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// The real retained clarification must render the same mode, history and
// answer component as manual quizzes, without invoking a manual quiz owner.
func TestTUIWizardQuiz(t *testing.T) {
	operations, session, _ := guidedQuizFixture(t, "default")

	current, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	_, model := guidedTUIModel(operations)
	openGuidedTUI(t, model)

	for _, marker := range []string{"Mode: Deep", "Answer history", "m mode"} {
		if !strings.Contains(model.View(), marker) {
			t.Fatal("shared quiz component missing", marker, model.View())
		}
	}

	runGuidedTUIKey(t, model, "m", true)

	after, err := authoringcore.ReadGuidedUseCase(t.Context(), operations.projectContext.ProjectDir, guidedQuizCommand(current, "", "", "").GuidedUseCaseScope)
	if err != nil || observeGuidedQuiz(t, operations, after).State.Mode != "fast" {
		t.Fatal("shared mode control did not reach guided owner", err)
	}
}

// Answers, history, skip and Done use the same editor and native clarification
// contract; Cancel reaches the guided stop owner instead of cancelling a manual quiz.
func TestTUIWizardQuizControls(t *testing.T) {
	operations, session, generation := guidedQuizFixture(t, "default")
	now := serviceQuizClock()()
	more := serviceQuizModelResponse("more", "question-2", now)
	more.Question.Prompt = "Which existing boundary remains?"
	later := serviceQuizModelResponse("later", "question-3", now)
	later.Question.Prompt = "Which delivery evidence is required?"
	generation.responses = append(generation.responses[:1], more, later, more, generation.responses[1])

	_, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	_, model := guidedTUIModel(operations)
	openGuidedTUI(t, model)
	runGuidedTUIKey(t, model, "e", false)
	runGuidedTUIKey(t, model, "Keep the agreed boundary.", false)
	runGuidedTUIKey(t, model, "enter", true)

	for _, marker := range []string{"Keep the agreed boundary.", "Which existing boundary remains?", "Answer history"} {
		if !strings.Contains(model.View(), marker) {
			t.Fatal("shared history lost retained content", marker, model.View())
		}
	}

	runGuidedTUIKey(t, model, "e", false)

	_, skip := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}, Alt: true})
	if skip == nil {
		t.Fatal("shared Skip did not submit")
	}

	model.Update(skip())

	after := onlyGuidedTUIRun(t, operations)
	if observeGuidedQuiz(t, operations, after).State.SkipsRecorded != 1 {
		t.Fatal("Skip did not persist")
	}

	runGuidedTUIKey(t, model, "e", false)

	_, done := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}, Alt: true})
	if done == nil {
		t.Fatal("shared Done did not submit", model.View(), after.Guided)
	}

	model.Update(done())

	if after = onlyGuidedTUIRun(t, operations); after.Guided.Stage != "quiz" {
		t.Fatal("Done bypassed clarification sufficiency")
	}

	runGuidedTUIKey(t, model, "x", false)
	runGuidedTUIKey(t, model, "enter", true)

	if after = onlyGuidedTUIRun(t, operations); after.Guided.Status != "cancelled" {
		t.Fatal("shared Cancel missed guided owner", model.View())
	}
}
