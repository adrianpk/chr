package app

import (
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestServiceWorkflowContinuation runs one actual task through the shared bounded
// operation and preserves that checkpoint on equivalent invocation attachment.
func TestServiceWorkflowContinuation(t *testing.T) {
	operations, hierarchy, target := serviceWorkflowSliceFixture(t)
	request := workflowcore.StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "bounded-service",
		Policy:     workflowstate.ExecutionPolicy{CheckpointStepLimit: 1},
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice_set", Identity: hierarchy.Set.ID}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}},
	}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	operations.workflowPullRequests = &serviceWorkflowPRFixture{}
	worker := []string{"sh", "-c", "printf 'Task\\n' >> result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task'"}

	progress, err := operations.ContinueWorkflow(t.Context(), run.Invocation.ID, workflowcore.ContinuationRequest{Worker: worker}, nil)
	if err != nil {
		t.Fatal(err)
	}

	checkServiceWorkflowContinuation(t, progress)

	attached, err := operations.StartWorkflow(t.Context(), request)
	if err != nil || !attached.Attached || attached.Invocation.ID != run.Invocation.ID || attached.Invocation.Continuation.EvidenceHash != progress.Invocation.Continuation.EvidenceHash {
		t.Fatal("service attachment lost accepted task continuation", err)
	}
}

func checkServiceWorkflowContinuation(t *testing.T, progress workflowcore.ContinuationProgress) {
	t.Helper()

	if progress.Steps != 1 || progress.Reason != "step_limit" || progress.Invocation.Cursor != 1 || len(progress.Invocation.Attempts) != 1 || progress.Invocation.Continuation.Sequence != 1 {
		t.Fatalf("service failed to retain one bounded task: %+v", progress)
	}
}
