package core

import (
	"testing"
)

// TestWorkflowAssuranceLost loses every guarded publication response in a
// dependent loose queue. Restart observes actual Git, finalizes once and starts
// the successor without resetting attempts or creating another source PR.
func TestWorkflowAssuranceLost(t *testing.T) {
	store, request, ports, provider := workflowAssuranceLooseFixture(t)
	provider.lostPublisher = nativeLostResponsePublisher(t, request.Selection.Target.Repository, provider.remote)
	run := workflowTestStart(t, store, request).Invocation
	completed, losses := workflowAssuranceLosses(t, store, run.ID, ports)
	checkWorkflowAssuranceDelivery(t, store, completed, provider, []string{"PRJ-90", "PRJ-91", "PRJ-92"}, 3)

	if losses != 3 {
		t.Fatal("the fixture did not lose exactly one response per publication", losses)
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)
	before := workflowTestRead(t, restarted, run.ID)

	progress, err := restarted.ContinueWorkflow(t.Context(), run.ID, WorkflowContinuationRequest{Worker: workflowTaskCommand(t, "worker"), PullRequest: workflowPRContent()}, ports)
	if err != nil || progress.Steps != 0 || progress.Invocation.Revision != before.Revision || provider.calls != 3 {
		t.Fatal("terminal retry repeated work, publication or finalization", err)
	}
}

func workflowAssuranceLosses(t *testing.T, store *WorkflowStore, id string, ports WorkflowContinuationPorts) (WorkflowInvocation, int) {
	t.Helper()

	losses := 0

	for range 80 {
		restarted := NewWorkflowStore(store.layout, store.projectID)
		request := WorkflowContinuationRequest{Worker: workflowTaskCommand(t, "worker"), PullRequest: workflowPRContent(), StepLimit: 1}

		progress, err := restarted.ContinueWorkflow(t.Context(), id, request, ports)
		losses += workflowAssuranceLostResponse(t, store, progress, err)

		if progress.Invocation.Status == "completed" {
			return progress.Invocation, losses
		}
	}

	t.Fatal("lost-response recovery exceeded eighty bounded calls")

	return WorkflowInvocation{}, losses
}

func workflowAssuranceLostResponse(t *testing.T, store *WorkflowStore, progress WorkflowContinuationProgress, failure error) int {
	t.Helper()

	if failure == nil {
		return 0
	}

	if progress.Invocation.Submission == nil || progress.Invocation.Submission.State != "awaiting_observation" {
		t.Fatal("a lost response did not retain its uncertain submission", failure)
	}

	submission, err := store.ReadWorkflowSubmission(t.Context(), progress.Invocation.ID)
	if err != nil || submission.Native == nil || submission.Native.TransportResponseRef == nil {
		t.Fatal("lost native response has no retained transport evidence", err)
	}

	checkWorkflowAssuranceLostProof(t, store, submission)

	return 1
}

func checkWorkflowAssuranceLostProof(t *testing.T, store *WorkflowStore, submission WorkflowSubmission) {
	t.Helper()

	var response WorkflowGitPublicationResult
	if err := store.readWorkflowNativeEvidence(submission.InvocationID, *submission.Native.TransportResponseRef, &response); err != nil || response.State != "unknown" || !response.CommandsSent || response.PublishedSourceCommit != "" {
		t.Fatal("lost native report was replaced with an applied response", err)
	}
}
