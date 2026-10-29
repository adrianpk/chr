package app

import (
	"context"
	"errors"
	"maps"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/tui"
)

type tuiWizardAuthoringRecorder struct {
	owner    tuiGuidedExecutor
	requests []tui.GuidedRequest
}

func (e *tuiWizardAuthoringRecorder) ExecuteGuided(request tui.GuidedRequest) (tui.GuidedView, error) {
	if request.Action == "input" {
		e.requests = append(e.requests, request)
	}

	return e.owner.ExecuteGuided(request)
}

// Restart and explicit Continue reconcile the same real preparation through
// contract failure, deadline and lost plan/materialization acknowledgements.
// Generation, native ticket identity and consumed operation intent stay retained.
func TestTUIWizardPreparationRetry(t *testing.T) {
	for _, phase := range []string{"contracts", "timeout", "plan", "materialization"} {
		t.Run(phase, func(t *testing.T) { testTUIWizardPreparationRetry(t, phase) })
	}
}

func testTUIWizardPreparationRetry(t *testing.T, phase string) {
	t.Helper()

	operations, _, generation := guidedPreparationFixture(t, "default")
	generation.fail = phase == "contracts"
	tuiWizardPreparationFault(&operations, phase)

	executor := &tuiWizardAuthoringRecorder{owner: tuiGuidedExecutor{operations: operations}}
	model := tui.NewModel(tui.Options{ProjectID: operations.projectContext.ProjectID, GuidedExecutor: executor})
	openGuidedTUI(t, model)
	runGuidedTUIKey(t, model, "c", true)

	if len(executor.requests) != 1 {
		t.Fatal("preparation did not start once", model.View())
	}

	original := executor.requests[0]
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	reopened := tui.NewModel(tui.Options{ProjectID: operations.projectContext.ProjectID, GuidedExecutor: executor})
	openGuidedTUI(t, reopened)

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) || len(executor.requests) != 1 {
		t.Fatal("re-entry wrote or dispatched work")
	}

	generation.fail = false

	runGuidedTUIKey(t, reopened, "c", true)

	if len(executor.requests) != 2 || executor.requests[1] != original {
		t.Fatal("restart replaced the original retained command", executor.requests)
	}

	assertTUIWizardRetryResult(t, operations, generation, phase)
}

// A delayed real answer cannot apply after another writer changes the retained
// run. The shared editor keeps its original input until confirmed read-only reload.
func TestTUIWizardQuizStale(t *testing.T) {
	operations, session, _ := guidedQuizFixture(t, "default")

	current, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	_, model := guidedTUIModel(operations)
	openGuidedTUI(t, model)
	runGuidedTUIKey(t, model, "e", false)
	runGuidedTUIKey(t, model, "ORIGINAL_QUEUED_ANSWER", false)

	_, queued := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if queued == nil {
		t.Fatal("answer was not captured")
	}

	newer, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(current, "other-writer", "quiz-mode", "fast"))
	if err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	model.Update(queued())

	if !strings.Contains(model.View(), "ORIGINAL_QUEUED_ANSWER") || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("stale answer changed the owner or lost input", model.View())
	}

	checkTUIWizardQuizReload(t, model)

	after := onlyGuidedTUIRun(t, operations)
	if after.Revision != newer.Revision || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("reload dispatched or replaced the canonical run")
	}
}

func tuiWizardPreparationFault(operations *service, phase string) {
	faulted := false
	operations.afterGuidedPreparation = func(current string) error {
		matches := current == phase || phase == "timeout" && current == "plan"
		if !matches || faulted {
			return nil
		}

		faulted = true

		if phase == "timeout" {
			return context.DeadlineExceeded
		}

		return errors.New("lost " + phase + " acknowledgement")
	}
}

func assertTUIWizardRetryResult(t *testing.T, operations service, generation *planDeliveryTestModel, phase string) {
	t.Helper()

	after := onlyGuidedTUIRun(t, operations)
	assertGuidedPreparation(t, operations, after)
	assertOneGuidedMaterialization(t, operations)

	expected := 1
	if phase == "contracts" {
		expected = 2
	}

	if len(generation.base.requests) != 1 || generation.calls != expected || after.Guided.Work.Invocation != nil {
		t.Fatal("retry duplicated generation or bypassed native admission")
	}
}

func checkTUIWizardQuizReload(t *testing.T, model *tui.Model) {
	t.Helper()
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}, Alt: true})
	runGuidedTUIKey(t, model, "esc", false)

	if !strings.Contains(model.View(), "ORIGINAL_QUEUED_ANSWER") {
		t.Fatal("cancelled reload lost input")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}, Alt: true})
	runGuidedTUIKey(t, model, "enter", true)

	if strings.Contains(model.View(), "ORIGINAL_QUEUED_ANSWER") || !strings.Contains(model.View(), "Mode: Fast") {
		t.Fatal("confirmed reload did not read current binding", model.View())
	}
}
