package workflow

import (
	"context"
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowSubmissionMethods publishes actual merge/rebase/squash results
// through a server-side ref transaction that verifies source and base together.
// Intent precedes IO; acceptance waits for fetched canonical method proof.
func TestWorkflowSubmissionMethods(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) { testWorkflowSubmissionMethod(t, method) })
	}
}

func testWorkflowSubmissionMethod(t *testing.T, method string) {
	t.Helper()

	f, provider := newWorkflowSubmissionFixture(t, method)
	provider.before = func(ctx context.Context, submission workflowstate.Submission) {
		assertWorkflowSubmissionIntent(t, ctx, f, submission)
	}

	git := workflowSubmissionObserver(f)

	submission, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, git)
	if err != nil || submission.State != "awaiting_observation" || submission.Response != "accepted" || provider.calls != 1 {
		t.Fatalf("accepted response was not retained: %+v, %v", submission, err)
	}

	checkWorkflowPreparationTaskProofs(t, f.store, workflowTestRead(t, f.store, f.run.ID))

	confirmed, err := f.store.ReconcileWorkflowSubmission(t.Context(), f.run.ID, provider, git)
	if err != nil || confirmed.State != "confirmed" || confirmed.Proof == nil || confirmed.Proof.ResultCommit != submission.Guard.Git.ResultCommit {
		t.Fatalf("actual delivery: %+v, %v", confirmed, err)
	}

	assertWorkflowSubmissionRetry(t, f, provider, submission)
}

func assertWorkflowSubmissionIntent(t *testing.T, ctx context.Context, f workflowPreparationFixture, submission workflowstate.Submission) {
	t.Helper()

	retained, err := f.store.ReadWorkflowSubmission(ctx, f.run.ID)
	if err != nil || retained.ID != submission.ID || retained.State != "awaiting_observation" {
		t.Fatalf("provider IO preceded durable intent: %+v, %v", retained, err)
	}

	run, err := f.store.ReadWorkflow(ctx, f.run.ID)
	if err != nil || run.Submission == nil || run.Submission.ID != submission.ID || run.Status != "waiting_integration" || run.Outcome != "awaiting_observation" {
		t.Fatal("provider IO preceded invocation projection")
	}
}

func assertWorkflowSubmissionRetry(t *testing.T, f workflowPreparationFixture, provider *workflowSubmissionProviderFixture, submission workflowstate.Submission) {
	t.Helper()

	again, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f))
	if err != nil || again.ID != submission.ID || again.State != "confirmed" || provider.calls != 1 {
		t.Fatal("confirmed retry allocated another submission")
	}

	if workflowTestRead(t, f.store, f.run.ID).DeliveryRef != "" {
		t.Fatal("submission observation bypassed receipt finalization")
	}
}

// TestWorkflowSubmissionLostResponse observes actual Git delivery after a lost
// provider response and restart without allocating another attempt or worker.
func TestWorkflowSubmissionLostResponse(t *testing.T) {
	f, provider := newWorkflowSubmissionFixture(t, "rebase")
	provider.lost = true
	git := workflowSubmissionObserver(f)

	first, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, git)
	if !errors.Is(err, context.DeadlineExceeded) || first.State != "awaiting_observation" || first.Response != "error_unknown" {
		t.Fatalf("lost response: %+v, %v", first, err)
	}

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	confirmed, err := restarted.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, git)
	if err != nil || confirmed.ID != first.ID || confirmed.State != "confirmed" || confirmed.Attempt != 1 || provider.calls != 1 {
		t.Fatalf("restart repeated publication: %+v, %v", confirmed, err)
	}

	checkWorkflowPreparationTaskProofs(t, restarted, workflowTestRead(t, restarted, f.run.ID))
}
