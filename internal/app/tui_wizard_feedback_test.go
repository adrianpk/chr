package app

import (
	"errors"
	"maps"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// A real partially published preparation exposes its saved plan and blocker;
// observing it neither retries generation nor replaces the pending command.
func TestTUIWizardFeedback(t *testing.T) {
	operations, session, generation := guidedPreparationFixture(t, "default")
	generation.fail = true
	command := guidedQuizCommand(session, "retained-prepare", "prepare", "")
	executor := tuiGuidedExecutor{operations: operations}

	view, err := executor.ExecuteGuided(tui.GuidedRequest{Action: "input", Command: command})
	if err != nil {
		t.Fatal(err)
	}

	assertTUIWizardFeedback(t, view, []string{"Plan and ticket preparation has a pending operation", "Plan saved; ticket preparation is pending", "Generated plan retained", "Correct native contract generation"})

	if view.Retry == nil || view.Retry.Command != command || view.Session.Guided.PendingOperation == nil || generation.calls != 1 {
		t.Fatal("feedback replaced original preparation or retried it")
	}

	checkTUIWizardFeedbackReentry(t, operations)
	assertTUIWizardSuspended(t, operations, *view.Session)
}

// Lost acknowledgement is still an error, accompanied by the real retained
// observation. It cannot be presented as success or silently resubmitted.
func TestTUIWizardFeedbackFailure(t *testing.T) {
	operations, session, generation := guidedPreparationFixture(t, "default")
	lost := errors.New("preparation acknowledgement unavailable")
	operations.afterGuidedPreparation = func(string) error { return lost }

	view, err := (tuiGuidedExecutor{operations: operations}).ExecuteGuided(tui.GuidedRequest{Action: "input", Command: guidedQuizCommand(session, "retained-failure", "prepare", "")})
	if !errors.Is(err, lost) || view.Session == nil || view.Retry == nil || view.Session.Guided.PendingOperation.ID != "retained-failure" || len(generation.base.requests) != 1 {
		t.Fatal("failed acknowledgement lost pending observation", err)
	}

	assertTUIWizardFeedback(t, view, []string{"pending operation", "Generated plan retained"})
}

func assertTUIWizardFeedback(t *testing.T, view tui.GuidedView, markers []string) {
	t.Helper()

	text := strings.Join(append(append([]string{view.Prompt}, view.Progress...), view.Problems...), "\n")
	for _, marker := range markers {
		if !strings.Contains(text, marker) {
			t.Fatal("missing retained feedback", marker, text)
		}
	}

	if view.InputAction != "" {
		t.Fatal("pending preparation advertised a new input")
	}
}

// Status presentation follows the existing owner for each authoring stage; it
// does not substitute a terminal controller or fabricate automatic progression.
func TestTUIWizardStageFeedback(t *testing.T) {
	for _, stage := range []string{"quiz", "specification", "preparation"} {
		t.Run(stage, func(t *testing.T) {
			session := &authoringcore.Session{Guided: &authoringcore.GuidedUseCaseState{Stage: stage, ProgressionIntent: true, Status: "running", PendingOperation: &authoringcore.GuidedUseCaseOperation{ID: "original", Content: "SAVED_INPUT"}}}
			view := tuiGuidedView(GuidedUseCaseInspection{Session: session})
			assertTUIWizardFeedback(t, view, []string{guidedStageName(stage), "pending operation", "SAVED_INPUT"})
		})
	}
}

func checkTUIWizardFeedbackReentry(t *testing.T, operations service) {
	t.Helper()

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	_, model := guidedTUIModel(operations)
	openGuidedTUI(t, model)

	if !strings.Contains(model.View(), "Plan saved; ticket preparation is pending") {
		t.Fatal(model.View())
	}

	if !maps.Equal(workflowCLIImage(t, operations.projectContext.ProjectDir), before) {
		t.Fatal("re-entry dispatched preparation")
	}
}

func assertTUIWizardSuspended(t *testing.T, operations service, session authoringcore.Session) {
	t.Helper()

	paused, err := operations.ControlGuidedUseCase(t.Context(), guidedQuizCommand(session, "pause", "pause", ""))
	if err != nil {
		t.Fatal(err)
	}

	observed, err := operations.InspectGuidedUseCase(t.Context(), paused.ID, paused.Namespace)
	if err != nil {
		t.Fatal(err)
	}

	suspended := tuiGuidedView(observed)
	if !strings.Contains(suspended.Prompt, "suspended") || suspended.Retry != nil || strings.Join(suspended.Controls, ",") != "resume,cancel" {
		t.Fatal("suspended preparation advertised dispatch", suspended)
	}
}
