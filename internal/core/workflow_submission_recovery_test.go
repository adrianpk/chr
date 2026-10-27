package core

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// TestWorkflowSubmissionCrashProcess is the owned writer that exits during
// common intent, invocation projection, provider publication or response writes.
func TestWorkflowSubmissionCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_SUBMISSION_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	id, stage := os.Getenv("CHR_SUBMISSION_CRASH_ID"), os.Getenv("CHR_SUBMISSION_CRASH_STAGE")

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	observation, err := store.readWorkflowPullRequestObservation(*run.PullRequest)
	if err != nil {
		t.Fatal(err)
	}

	pr := observation.Snapshots[0]
	remote := os.Getenv("CHR_SUBMISSION_CRASH_REMOTE")
	base := &workflowPreparationProviderFixture{workflowReviewProviderFixture: &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}}, guarded: true, revision: "guard-v1"}
	provider := &workflowSubmissionProviderFixture{workflowPreparationProviderFixture: base, remote: remote, sealed: map[string]bool{}}
	provider.after = func(WorkflowSubmission) {
		if stage == "provider" {
			os.Exit(17)
		}
	}
	store.afterReplace = func(index int) error {
		if workflowSubmissionCrashReady(t, store, run, stage, index) {
			os.Exit(17)
		}

		return nil
	}

	_, err = store.SubmitWorkflowSliceIntegration(t.Context(), id, provider, NewWorkflowIntegrationGit(pr.Repository, remote))
	t.Fatalf("owned submission writer did not terminate: %v", err)
}

func workflowSubmissionCrashReady(t *testing.T, store *WorkflowStore, run workflowInvocationFile, stage string, index int) bool {
	t.Helper()

	if stage == "projection" {
		return index == 0
	}

	if index != -1 && (stage != "confirmation_projection" || index != 0) {
		return false
	}

	if stage == "intent" {
		return true
	}

	journal, err := store.readWorkflowSubmissionJournal(t.Context(), run)
	if err != nil {
		t.Fatal(err)
	}

	last := journal.Attempts[len(journal.Attempts)-1]

	switch stage {
	case "confirmed", "confirmation_projection":
		return last.State == "confirmed"
	case "response":
		return last.Response == "accepted"
	default:
		return false
	}
}

// TestWorkflowSubmissionJournal recovers actual host death around publication.
// Unstarted intent still excludes the target until sealed; a delivered target is
// fetched and verified after restart without another provider mutation.
func TestWorkflowSubmissionJournal(t *testing.T) {
	for _, stage := range []string{"intent", "projection", "provider", "response", "confirmed", "confirmation_projection"} {
		t.Run(stage, func(t *testing.T) { testWorkflowSubmissionJournal(t, stage) })
	}
}

func testWorkflowSubmissionJournal(t *testing.T, stage string) {
	t.Helper()

	f, provider := newWorkflowSubmissionFixture(t, "rebase")
	if stage == "confirmed" || stage == "confirmation_projection" {
		if _, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f)); err != nil {
			t.Fatal(err)
		}
	}

	baseline := provider.calls
	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowSubmissionCrashProcess$")

	cmd.Env = append(os.Environ(), "CHR_SUBMISSION_CRASH_DIR="+f.store.layout.ProjectDir(), "CHR_SUBMISSION_CRASH_ID="+f.run.ID, "CHR_SUBMISSION_CRASH_STAGE="+stage, "CHR_SUBMISSION_CRASH_REMOTE="+f.remote)
	output, err := cmd.CombinedOutput()

	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 17 {
		t.Fatalf("owned writer crash: %v: %s", err, output)
	}

	assertWorkflowSubmissionRecovery(t, f, provider, stage, baseline)
}

func assertWorkflowSubmissionRecovery(t *testing.T, f workflowPreparationFixture, provider *workflowSubmissionProviderFixture, stage string, baseline int) {
	t.Helper()

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	first, err := restarted.ReadWorkflowSubmission(t.Context(), f.run.ID)
	if err != nil || first.Attempt != 1 || first.State != "awaiting_observation" && first.State != "confirmed" {
		t.Fatalf("durable crash intent: %+v, %v", first, err)
	}

	expected := sealWorkflowSubmissionCrash(t, restarted, f, provider, first, stage)
	last, err := restarted.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f))

	if err != nil || last.ID != first.ID || last.State != expected || provider.calls != baseline {
		t.Fatalf("crash recovery repeated publication: %+v, %v", last, err)
	}

	checkWorkflowPreparationTaskProofs(t, restarted, workflowTestRead(t, restarted, f.run.ID))
}

func sealWorkflowSubmissionCrash(t *testing.T, store *WorkflowStore, f workflowPreparationFixture, provider *workflowSubmissionProviderFixture, first WorkflowSubmission, stage string) string {
	t.Helper()

	if stage != "intent" && stage != "projection" {
		return "confirmed"
	}

	if _, err := store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f)); !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("an unobserved crash licensed a retry", err)
	}

	provider.seal(first.ID)

	return "unpublished"
}
