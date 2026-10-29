package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// A pending read cannot consume typing or Enter. A write supersedes that read's
// presentation ownership, and its late reply cannot clear the captured command.
func TestWizardPollingInput(t *testing.T) {
	model := guidedTerminalFixture()
	requests := []GuidedRequest{}
	model.guidedExecutor = guidedTestExecutor(func(request GuidedRequest) (GuidedView, error) {
		requests = append(requests, request)

		return model.guided.View, nil
	})
	model.editGuidedInput("answer")
	model.quizAnswer.Input.SetValue("BEFORE")

	poll := model.refreshGuided()
	if poll == nil {
		t.Fatal("passive read unavailable")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" AFTER")})

	if model.quizAnswer.Input.Value() != "BEFORE AFTER" {
		t.Fatal("poll consumed typed text")
	}

	_, submit := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if submit == nil {
		t.Fatal("poll consumed the captured submission")
	}

	model.Update(poll())

	if !model.guided.Busy || model.guided.Pending == nil || model.guided.Pending.Command.Content != "BEFORE AFTER" {
		t.Fatal("late read replaced the write")
	}

	model.Update(submit())

	if len(requests) != 2 || requests[0].Action != "inspect" || requests[1].Command.Content != "BEFORE AFTER" {
		t.Fatal("poll dispatched or rebound input", requests)
	}
}

// Passive observations retain the selected case, tab and focus. Confirmation
// waits keep their captured context, and newer stages never redirect a draft.
func TestWizardPollingContext(t *testing.T) {
	model := guidedTerminalFixture()
	observed := model.guided.View
	observed.Cases = append([]authoringcore.CaseMetadata{{Slug: "other", Name: "Other"}}, model.cases...)
	model.cases = observed.Cases
	model.selectView(ViewCases)

	model.selectedCase = 0
	model.guidedExecutor = guidedTestExecutor(func(GuidedRequest) (GuidedView, error) { return observed, nil })
	poll := model.refreshGuided()
	model.Update(poll())

	selected, _ := model.selectedCaseMetadata()
	if selected.Slug != "other" || model.SelectedView() != ViewCases || model.selectedPane != 0 {
		t.Fatal("poll changed selected case, tab or focus")
	}

	model.guided.View.Controls = []string{"cancel"}
	model.submitGuidedStep("cancel")

	if model.refreshGuided() != nil {
		t.Fatal("poll changed a captured confirmation")
	}
}

// Stage changes from another owner preserve the old prompt, complete draft and
// chosen Quizzes tab until the user explicitly resolves that captured input.
func TestWizardPollingStage(t *testing.T) {
	for _, stage := range []string{"specification", "preparation", "execution"} {
		t.Run(stage, func(t *testing.T) {
			model := guidedTerminalFixture()
			newer := wizardNewerQuiz(model.guided.View)
			newer.Session.Guided.Stage = stage
			model.guidedExecutor = guidedTestExecutor(func(GuidedRequest) (GuidedView, error) { return newer, nil })
			model.editGuidedInput("answer")
			model.quizAnswer.Input.SetValue("COMPLETE_UNSENT_INPUT")

			poll := model.refreshGuided()
			model.Update(poll())

			if model.SelectedView() != ViewQuizzes || model.quizAnswer.Input.Value() != "COMPLETE_UNSENT_INPUT" || !strings.Contains(model.View(), "What should remain bounded?") || strings.Contains(model.View(), "NEWER_QUESTION") {
				t.Fatal("stage observation redirected or replaced draft", model.View())
			}
		})
	}
}

// An in-flight guided read cannot roll back the native revision installed by a
// completed approval/control. The next passive read can observe current state.
func TestWizardPollingNativeRevision(t *testing.T) {
	model := guidedTerminalFixture()
	view := model.guided.View
	view.Native = &WorkflowView{ID: "native", Revision: 3}
	model.guidedExecutor = guidedTestExecutor(func(GuidedRequest) (GuidedView, error) { return view, nil })
	poll := model.refreshGuided()
	model.workflow = WorkflowView{ID: "native", Revision: 4}
	model.Update(poll())

	if model.workflow.Revision != 4 || model.guided.Problem != "" || model.refreshGuided() == nil {
		t.Fatal("old read rolled back native control or stopped polling")
	}
}
