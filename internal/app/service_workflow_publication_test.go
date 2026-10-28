package app

import (
	"context"
	"errors"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// TestServiceWorkflowPublication shares the combined gate, human approval and
// durable submission protocol. Retrying uncertainty never starts another role.
func TestServiceWorkflowPublication(t *testing.T) {
	operations, id, guarded := serviceWorkflowPreparationReady(t)
	provider := &serviceWorkflowSubmissionFixture{serviceWorkflowPreparationFixture: guarded}
	operations.workflowPullRequests = provider

	progress, err := operations.AdvanceWorkflowSliceIntegration(t.Context(), id, nil)
	if err != nil || progress.Invocation.Review.State != "waiting_approval" || len(provider.submissions) != 0 {
		t.Fatal("central operation skipped manual approval", err)
	}

	if _, err := operations.ApproveWorkflowSlice(t.Context(), id, progress.Invocation.Review.InputHash); err != nil {
		t.Fatal(err)
	}

	progress, err = operations.AdvanceWorkflowSliceIntegration(t.Context(), id, nil)
	if !errors.Is(err, context.DeadlineExceeded) || progress.Invocation.Submission.State != "awaiting_observation" {
		t.Fatal("central operation lost publication intent", err)
	}

	before := progress.Invocation

	progress, err = operations.AdvanceWorkflowSliceIntegration(t.Context(), id, nil)
	if !errors.Is(err, workflowcore.ErrWorkflowObservationRequired) || len(provider.submissions) != 1 || len(progress.Invocation.Attempts) != len(before.Attempts) {
		t.Fatal("central operation retried uncertain execution", err)
	}
}

// TestServiceWorkflowPublicationExternal uses a read-only provider and actual
// remote Git. The central operation finalizes shared delivery after cancellation
// without constructing another candidate, review role or publication command.
func TestServiceWorkflowPublicationExternal(t *testing.T) {
	operations, id, provider := serviceWorkflowDeliveryFixture(t)
	if _, err := operations.StopWorkflow(t.Context(), id, "cancel"); err != nil {
		t.Fatal(err)
	}

	progress, err := operations.AdvanceWorkflowSliceIntegration(t.Context(), id, nil)
	if err != nil || progress.Receipt == nil || progress.Invocation.Status != "cancelled" || progress.Invocation.StopIntent != "cancel" || provider.creates != 1 {
		t.Fatal("central service did not preserve external delivery and cancellation", err)
	}

	again, err := operations.AdvanceWorkflowSliceIntegration(t.Context(), id, nil)
	if err != nil || again.Invocation.Revision != progress.Invocation.Revision || again.Receipt.SourceHash != progress.Receipt.SourceHash {
		t.Fatal("central service repeated receipt finalization", err)
	}
}
