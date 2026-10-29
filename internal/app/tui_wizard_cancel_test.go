package app

import (
	"maps"
	"slices"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// Pause follows the existing Web execution policy; stopped, input and blocked
// stages cannot advertise an active-execution action that has no useful effect.
func TestTUIWizardPausePolicy(t *testing.T) {
	for _, test := range []struct {
		name, stage, status string
		want                bool
	}{{"capture", "dump", "", false}, {"quiz", "quiz", "", false}, {"preparation", "preparation", "", false}, {"running", "execution", "running", true}, {"review", "execution", "waiting_review", true}, {"blocked", "execution", "blocked", false}, {"finished", "results", "completed", false}} {
		t.Run(test.name, func(t *testing.T) {
			observed := GuidedUseCaseInspection{Session: &authoringcore.Session{Guided: &authoringcore.GuidedUseCaseState{Stage: test.stage, Status: "running", ProgressionIntent: true}}}
			if test.status != "" {
				observed.Work = &WorkflowWorkStatus{WorkflowWorkOutcome: WorkflowWorkOutcome{Status: test.status}}
			}

			if got := slices.Contains(tuiGuidedControlChoices(observed), "pause"); got != test.want {
				t.Fatal("Pause presentation violates current stage policy", got)
			}
		})
	}
}

// An owned production terminal confirms cancellation, preserves exact committed
// inputs and the completed worker's candidate, and retains consumed work limits.
func TestTUIWizardCancelTerminal(t *testing.T) {
	operations, session, _, _ := guidedResultFixture(t)
	before := assertGuidedNativeAdmission(t, operations, session)
	originals := tuiWizardOriginalSnapshot(t, operations, session)
	candidate := serviceWorkflowDependencyGit(t, operations.projectContext.ProjectDir, "show", before.PullRequest.SourceCommit+":result.txt")
	terminal := startGuidedTerminal(t, operations)
	terminal.catalog(t)
	terminal.see(t, "Native review requires approval")
	terminal.keys(t, "p")
	terminal.see(t, "Dispatch stopped: pause")

	paused := onlyGuidedTUIRun(t, operations)
	image := workflowCLIImage(t, operations.projectContext.ProjectDir)
	terminal.keys(t, "x")
	terminal.see(t, "Saved documents and completed changes will be kept.")
	terminal.keys(t, "\x1b")
	terminal.see(t, "u Resume")
	assertTUIWizardCancelUnchanged(t, operations, paused, image)
	terminal.keys(t, "x")
	terminal.see(t, "Saved documents and completed changes will be kept.")
	terminal.keys(t, "\r")
	terminal.see(t, "Dispatch stopped: cancel")
	terminal.stop(t)

	after := onlyGuidedTUIRun(t, operations)

	native := assertGuidedNativeAdmission(t, operations, after)
	if after.Guided.Status != "cancelled" || native.ID != before.ID || native.Policy != before.Policy || serviceWorkTaskAttempts(native) != serviceWorkTaskAttempts(before) {
		t.Fatal("cancel replaced ownership or refilled consumed limits")
	}

	assertTUIWizardCancelledInputs(t, operations, after, originals)

	if native.PullRequest == nil || native.PullRequest.SourceCommit != before.PullRequest.SourceCommit || serviceWorkflowDependencyGit(t, operations.projectContext.ProjectDir, "show", native.PullRequest.SourceCommit+":result.txt") != candidate {
		t.Fatal("cancel replaced completed candidate")
	}
}

func assertTUIWizardCancelUnchanged(t *testing.T, operations service, paused authoringcore.Session, image map[string][32]byte) {
	t.Helper()

	current := onlyGuidedTUIRun(t, operations)
	if current.Guided.Status != "paused" || current.Revision != paused.Revision || !maps.Equal(image, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("dismissed cancellation changed retained state")
	}
}

func tuiWizardOriginalSnapshot(t *testing.T, operations service, session authoringcore.Session) map[string]GuidedUseCaseArtifact {
	t.Helper()

	originals := map[string]GuidedUseCaseArtifact{}

	for _, kind := range []string{"dump", "quiz", "specification", "plan", "tracker"} {
		original, err := operations.ReadGuidedUseCaseOriginal(t.Context(), session.ID, session.Namespace, kind)
		if err != nil {
			t.Fatal(err)
		}

		originals[kind] = original
	}

	return originals
}

func assertTUIWizardCancelledInputs(t *testing.T, operations service, session authoringcore.Session, before map[string]GuidedUseCaseArtifact) {
	t.Helper()

	after := tuiWizardOriginalSnapshot(t, operations, session)
	if !maps.Equal(before, after) || len(session.Guided.Checkpoints) != 4 {
		t.Fatal("cancel changed original bytes, references, commits or checkpoints")
	}
}
