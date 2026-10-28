package workflow

import (
	"context"
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowPublicationLateResponse finalizes verified external delivery while
// the original publisher is still live. The late guarded attempt cannot replace
// its moved base or overwrite the shared receipt and confirmed external fact.
func TestWorkflowPublicationLateResponse(t *testing.T) {
	f, p := newWorkflowSubmissionFixture(t, "rebase")
	started, release := make(chan struct{}), make(chan struct{})
	p.before = func(_ context.Context, _ Submission) { close(started); <-release }
	finished := make(chan error, 1)

	go func() {
		_, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, p, workflowSubmissionObserver(f))
		finished <- err
	}()

	<-started
	workflowPublicationExternalResult(t, f, p, "rebase")

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))

	close(release)

	lateErr := <-finished
	if err != nil || progress.Receipt == nil || lateErr == nil || !progress.Invocation.Submission.ExternalDelivery {
		t.Fatal("live publisher prevented or replaced canonical delivery", err, lateErr)
	}

	checkWorkflowPublicationReuse(t, f.store, f.run.ID, workflowPublicationPorts(f, p), progress)
}

// TestWorkflowPublicationDuringIO rejects target drift during Git observation.
// The next read observes the newly reachable target without another worker or
// submission; the stale observation never becomes an immutable delivery proof.
func TestWorkflowPublicationDuringIO(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "squash", workflowstate.Validation{WaiverReason: "Publication observation fixture"})
	p := &workflowSubmissionProviderFixture{workflowPreparationProviderFixture: f.provider, remote: f.remote}
	workflowPublicationExternalResult(t, f, p, "squash")

	ports := workflowPublicationPorts(f, p)
	moved := false
	ports.Observer = workflowIntegrationHook{workflowSubmissionObserver(f), func() {
		if !moved {
			moved = true

			workflowIntegrationCommit(t, f.actor, "observed-later.txt", "Target changed during observation")
			workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")
		}
	}}

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if !errors.Is(err, workflowstate.ErrConflict) || progress.Receipt != nil || progress.Invocation.DeliveryProofRef != "" {
		t.Fatal("target changed during IO but stale delivery proof was accepted", err)
	}

	progress, err = f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))
	if err != nil || progress.Receipt == nil || p.calls != 0 {
		t.Fatal("fresh canonical observation failed or submitted work", err)
	}
}
