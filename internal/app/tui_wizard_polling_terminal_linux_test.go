package app

import (
	"maps"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// Real managed polling observes a peer's retained idea while the production
// Quiz editor holds unsent text. Typing, cancelled discard and earlier-tab
// navigation keep their state; the reads cannot dispatch or save an answer.
func TestTUIWizardPollingTerminal(t *testing.T) {
	operations, session, _ := guidedQuizFixture(t, "default")

	current, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	terminal := startGuidedTerminal(t, operations)
	terminal.catalog(t)
	terminal.see(t, "What should the workflow preserve")
	terminal.keys(t, "e")
	terminal.see(t, "alt+x cancels quiz")
	terminal.keys(t, "UNSENT_BEFORE_POLL ")
	terminal.see(t, "UNSENT_BEFORE_POLL")

	observed, err := operations.ChangeGuidedUseCase(t.Context(), guidedQuizCommand(current, "peer-idea-one", "idea", "POLL_OBSERVED_ONE"))
	if err != nil {
		t.Fatal(err)
	}

	image := workflowCLIImage(t, operations.projectContext.ProjectDir)
	terminal.see(t, "POLL_OBSERVED_ONE")
	terminal.keys(t, "_AFTER_POLL ")
	terminal.see(t, "UNSENT_BEFORE_POLL _AFTER_POLL")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Discard the unsent answer?")
	terminal.keys(t, "\x1b")
	terminal.see(t, "UNSENT_BEFORE_POLL _AFTER_POLL")
	assertTUIWizardPollingOwner(t, operations, observed, image)
	finishTUIWizardPollingNavigation(t, terminal, operations, observed)
}

func assertTUIWizardPollingOwner(t *testing.T, operations service, expected authoringcore.Session, image map[string][32]byte) {
	t.Helper()

	current := onlyGuidedTUIRun(t, operations)
	if current.Revision != expected.Revision || current.Guided.Stage != "quiz" || current.Guided.PendingOperation != nil || !maps.Equal(image, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("passive terminal reads dispatched or saved input")
	}

	if observeGuidedQuiz(t, operations, current).State.AnswersRecorded != 0 {
		t.Fatal("unsent text was submitted by polling")
	}
}

func finishTUIWizardPollingNavigation(t *testing.T, terminal *guidedTerminal, operations service, current authoringcore.Session) {
	t.Helper()
	terminal.keys(t, "\x1b")
	terminal.see(t, "Discard the unsent answer?")
	terminal.keys(t, "\r")
	terminal.see(t, "Draft discarded.")
	terminal.keys(t, "\x1bOQ")
	terminal.see(t, "Read-only stage.")

	observed, err := operations.ChangeGuidedUseCase(t.Context(), guidedQuizCommand(current, "peer-idea-two", "idea", "POLL_OBSERVED_TWO"))
	if err != nil {
		t.Fatal(err)
	}

	image := workflowCLIImage(t, operations.projectContext.ProjectDir)
	terminal.see(t, "POLL_OBSERVED_TWO")
	terminal.see(t, "Committed original (read-only)")
	terminal.stop(t)
	assertTUIWizardPollingOwner(t, operations, observed, image)
}
