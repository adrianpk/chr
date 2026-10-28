package workflow

import (
	"context"
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowPublicationRevision preserves sealed attempt consumption across a
// real repaired task and new aggregate candidate. Native retries consume the
// next attempt; external delivery never confirms the old candidate's command.
func TestWorkflowPublicationRevision(t *testing.T) {
	for _, external := range []bool{false, true} {
		name := "guarded retry"
		if external {
			name = "external delivery"
		}

		t.Run(name, func(t *testing.T) { testWorkflowPublicationRevision(t, external) })
	}
}

func testWorkflowPublicationRevision(t *testing.T, external bool) {
	t.Helper()

	f, p := newWorkflowSubmissionFixture(t, "rebase")

	first := workflowPublicationSealedAttempt(t, f, p)
	if _, err := f.store.RepairWorkflowSlice(t.Context(), f.run.ID, f.run.Members[2].Admission.Key); err != nil {
		t.Fatal(err)
	}

	candidate, err := f.store.ExecuteWorkflowSlice(t.Context(), f.run.ID, workflowPRWorker())
	if err != nil {
		t.Fatal(err)
	}

	p.pr.HeadCommit = candidate.Result.Commit
	workflowIntegrationCommand(t, candidate.Lane.Worktree, "push", f.remote, candidate.Result.Commit+":"+candidate.Lane.Branch)

	if _, err := f.store.EnsureWorkflowSlicePullRequest(t.Context(), f.run.ID, workflowPRContent(), p); err != nil {
		t.Fatal(err)
	}

	if external {
		workflowPublicationExternalResult(t, f, p, "rebase")
	}

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))
	if err != nil {
		t.Fatal("sealed candidate history prevented reconciliation", err)
	}

	checkWorkflowPublicationRevision(t, f, p, first, progress, external)
}

func workflowPublicationSealedAttempt(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture) workflowstate.Submission {
	t.Helper()

	p.unknown = true

	submission, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, p, workflowSubmissionObserver(f))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}

	p.seal(submission.ID)

	p.unknown = false
	if _, err := f.store.ReconcileWorkflowSubmission(t.Context(), f.run.ID, p, workflowSubmissionObserver(f)); err != nil {
		t.Fatal(err)
	}

	return submission
}

func checkWorkflowPublicationRevision(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture, first workflowstate.Submission, progress IntegrationProgress, external bool) {
	t.Helper()

	if external {
		checkWorkflowPublicationRevisedExternal(t, f, p, progress)

		return
	}

	if progress.Receipt != nil || progress.Invocation.Submission.Attempt != 2 || progress.Invocation.Submission.ID == first.ID || p.calls != 2 {
		t.Fatal("repaired candidate reset or reused a consumed attempt")
	}

	final, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))
	if err != nil || final.Receipt == nil || final.Receipt.Receipt.SubmissionID != progress.Invocation.Submission.ID {
		t.Fatal("repaired guarded result did not finalize its new intent", err)
	}
}

func checkWorkflowPublicationRevisedExternal(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture, progress IntegrationProgress) {
	t.Helper()

	if progress.Receipt == nil || p.calls != 1 || progress.Invocation.Submission.State != workflowstate.SubmissionUnpublished || progress.Receipt.Receipt.SubmissionID != progress.Invocation.PullRequest.ID {
		t.Fatal("external revised delivery was attributed to an old intent")
	}

	checkWorkflowPublicationFinalized(t, f, progress)
}
