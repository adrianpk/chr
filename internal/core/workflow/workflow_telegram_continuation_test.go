package workflow

import (
	"context"
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramWorkContinuationGuard rejects revoked authority under the common
// acceptance lock, before a checkpoint, worker or other invocation change.
func TestTelegramWorkContinuationGuard(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request).Invocation
	calls := 0
	check := func(context.Context) error { calls++; return workflowstate.ErrConflict }

	_, err := ContinueWorkflow(t.Context(), store, run.ID, ContinuationRequest{ExpectedRevision: run.Revision, StepLimit: 1, AuthorizationCheck: check}, ContinuationPorts{})
	if !errors.Is(err, workflowstate.ErrConflict) || calls != 1 {
		t.Fatal("continuation authority was not checked", calls, err)
	}

	after, err := store.ReadWorkflow(t.Context(), run.ID)
	if err != nil || after.Revision != run.Revision || after.Continuation != nil || len(after.Attempts) != 0 {
		t.Fatal("revoked activation changed native work", after, err)
	}
}
