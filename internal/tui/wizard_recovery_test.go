package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// A newer observation cannot pair an old draft with the new question. Read-only
// inspection preserves a failed request; reload requires explicit confirmation.
func TestWizardQuizRecovery(t *testing.T) {
	model := guidedTerminalFixture()
	before := model.guided.View
	newer := wizardNewerQuiz(before)
	requests := []GuidedRequest{}
	model.guidedExecutor = guidedTestExecutor(func(request GuidedRequest) (GuidedView, error) {
		requests = append(requests, request)
		if request.Action == "input" {
			return newer, workflowstate.ErrConflict
		}

		return newer, nil
	})
	model.editGuidedInput("answer")
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ORIGINAL_UNSENT_ANSWER")})
	model.completeGuided(guidedFinishedMsg{Sequence: model.guided.Sequence, Request: GuidedRequest{Action: "inspect"}, View: newer})

	_, submit := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if submit == nil {
		t.Fatal("captured answer did not submit")
	}

	model.Update(submit())

	original := *model.guided.Pending
	assertWizardQuizCapture(t, model, before)

	inspect := model.startGuidedRequest(model.guidedRequest("inspect"))
	model.Update(inspect())

	if model.guided.Pending == nil || *model.guided.Pending != original {
		t.Fatal("read-only inspection lost retry identity")
	}

	assertWizardQuizReload(t, model)

	if len(requests) != 3 || requests[2].Action != "inspect" || model.quizAnswer.Binding.QuestionID != "second" || model.quizAnswer.Input.Value() != "" {
		t.Fatal("confirmed reload did not open current question", requests)
	}
}

func wizardNewerQuiz(before GuidedView) GuidedView {
	after := before
	session := *before.Session
	guided := *session.Guided
	quiz := *guided.Quiz
	quiz.QuestionID = "second"
	guided.Quiz = &quiz
	session.Guided, session.Revision = &guided, session.Revision+1
	after.Session, after.Prompt = &session, "NEWER_QUESTION"
	projected := *before.Quiz
	projected.ActiveQuestionSummary, projected.Binding.QuestionID = "NEWER_QUESTION", "second"
	projected.Binding.ExpectedRevision = "newer-native"
	after.Quiz = &projected

	return after
}

func assertWizardQuizReload(t *testing.T, model *Model) {
	t.Helper()

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}, Alt: true}
	model.Update(key)

	if !model.draftConfirmation.Active {
		t.Fatal("reload skipped confirmation")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.quizAnswer.Input.Value() != "ORIGINAL_UNSENT_ANSWER" {
		t.Fatal("cancelled reload discarded the draft")
	}

	model.Update(key)

	_, reload := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if reload == nil {
		t.Fatal("confirmed guided reload was swallowed")
	}

	model.Update(reload())
}

// A lost acknowledgement retries the exact captured answer even after refresh;
// obsolete or regressing replies cannot clear a draft or replace newer state.
func TestWizardQuizReplyOrder(t *testing.T) {
	model := guidedTerminalFixture()
	calls := []GuidedRequest{}
	model.guidedExecutor = guidedTestExecutor(func(request GuidedRequest) (GuidedView, error) {
		calls = append(calls, request)

		return model.guided.View, errors.New("lost acknowledgement")
	})
	model.editGuidedInput("answer")
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("RETAINED_ANSWER")})

	_, command := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model.Update(command())

	original := *model.guided.Pending

	_, retry := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}, Alt: true})
	if retry == nil {
		t.Fatal("retry unavailable inside shared editor")
	}

	model.Update(retry())

	if len(calls) != 2 || calls[0] != original || calls[1] != original {
		t.Fatal("answer retry changed identity or content")
	}

	newer := wizardNewerQuiz(model.guided.View)
	model.guided.View = newer
	old := *newer.Session
	old.Revision--
	stale := newer
	stale.Session = &old
	model.completeGuided(guidedFinishedMsg{Sequence: model.guided.Sequence, Request: GuidedRequest{Action: "inspect"}, View: stale})

	if model.guided.View.Session.Revision != newer.Session.Revision || model.quizAnswer.Input.Value() != "RETAINED_ANSWER" {
		t.Fatal("regressing reply replaced newer state")
	}

	model.completeGuided(guidedFinishedMsg{Sequence: model.guided.Sequence - 1, View: stale})

	if model.guided.Pending == nil || *model.guided.Pending != original {
		t.Fatal("obsolete reply cleared retained retry")
	}
}

func assertWizardQuizCapture(t *testing.T, model *Model, before GuidedView) {
	t.Helper()

	original := *model.guided.Pending
	if original.Command.ExpectedRevision != before.Session.Revision || original.Command.QuestionID != "first" || strings.Contains(model.View(), "NEWER_QUESTION") || !strings.Contains(model.View(), "ORIGINAL_UNSENT_ANSWER") {
		t.Fatal("new state replaced the captured prompt or input", model.View())
	}
}

// Shared quiz controls keep their discard confirmation in wizard mode; Esc
// preserves unsent text, and Enter submits the originally captured control.
func TestWizardQuizControlDraft(t *testing.T) {
	model := guidedTerminalFixture()
	requests := []GuidedRequest{}
	model.guidedExecutor = guidedTestExecutor(func(request GuidedRequest) (GuidedView, error) {
		requests = append(requests, request)

		return model.guided.View, nil
	})
	model.editGuidedInput("answer")
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("UNSENT_DRAFT")})

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true}
	model.Update(key)

	if !model.draftConfirmation.Active || len(requests) != 0 {
		t.Fatal("shared cancel skipped confirmation")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.quizAnswer.Input.Value() != "UNSENT_DRAFT" {
		t.Fatal("cancelled control discarded input")
	}

	model.Update(key)

	_, command := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if command == nil {
		t.Fatal("confirmed shared control was swallowed")
	}

	model.Update(command())

	if len(requests) != 1 || requests[0].Command.Action != "cancel" || requests[0].Command.ExpectedRevision != 2 {
		t.Fatal("control lost its captured binding", requests)
	}
}
