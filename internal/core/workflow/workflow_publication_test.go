package workflow

import (
	"context"
	"errors"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowPublicationPorts(f workflowPreparationFixture, p PullRequestProvider) PublicationPorts {
	return PublicationPorts{Provider: p, Preparer: f.git, Observer: workflowSubmissionObserver(f), Reviewer: workflowReviewWorker("pass", "", "")}
}

// TestWorkflowPublicationModes drives each policy through its own stop and one
// actual guarded publication. Lost responses finalize from canonical Git on the
// next call; a completed checkpoint reuses the receipt without new IO or work.
func TestWorkflowPublicationModes(t *testing.T) {
	for _, mode := range []workflowstate.ReviewMode{workflowstate.ReviewHuman, workflowstate.ReviewAgentic, workflowstate.ReviewTrustedAutoMerge} {
		t.Run(string(mode), func(t *testing.T) { testWorkflowPublicationMode(t, mode) })
	}
}

func testWorkflowPublicationMode(t *testing.T, mode workflowstate.ReviewMode) {
	t.Helper()

	f := newWorkflowPreparationFixture(t, "rebase", workflowstate.Validation{WaiverReason: "Publication fixture observes actual Git"})
	p := &workflowSubmissionProviderFixture{workflowPreparationProviderFixture: f.provider, remote: f.remote, sealed: map[string]bool{}, lost: true}
	workflowReviewConfig(t, f.store, mode)

	ports := workflowPublicationPorts(f, p)

	progress, err := workflowPublicationModeStart(t, f, p, ports, mode)

	if !errors.Is(err, context.DeadlineExceeded) || progress.Receipt != nil || progress.Invocation.Submission.State != "awaiting_observation" || p.calls != 1 {
		t.Fatal("response replaced observation or repeated publication", err)
	}

	checkWorkflowPublicationConfirmation(t, f, p, ports, progress.Invocation)
}

func checkWorkflowPublicationFinalized(t *testing.T, f workflowPreparationFixture, progress IntegrationProgress) {
	t.Helper()

	for _, member := range progress.Invocation.Members {
		record, err := f.store.Read(t.Context(), member.Admission.Path)
		if err != nil || record.Task.Status != string(ticketcore.StatusDone) || record.Task.DeliveryRef != progress.Receipt.Path {
			t.Fatal("member task lacks the shared verified receipt", err)
		}
	}

	if progress.Invocation.Cursor != f.run.Cursor || f.provider.creates != 1 {
		t.Fatal("delivery advanced the frontier or created another PR")
	}
}

// TestWorkflowPublicationMovedBase revokes a previously approved combined tree.
// The shared operation preserves task proofs, repeats the combined gate/review,
// and waits for approval of that new binding before any provider mutation.
func TestWorkflowPublicationMovedBase(t *testing.T) {
	f, p := newWorkflowSubmissionFixture(t, "rebase")
	workflowReviewConfig(t, f.store, workflowstate.ReviewHuman)

	ports := workflowPublicationPorts(f, p)

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if err != nil {
		t.Fatal(err)
	}

	old := progress.Invocation
	if _, err := f.store.ApproveWorkflowSlice(t.Context(), f.run.ID, old.Review.InputHash, p); err != nil {
		t.Fatal(err)
	}

	moved := workflowIntegrationCommit(t, f.actor, "movement.txt", "New canonical target content")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	progress, err = f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if err != nil || progress.Invocation.Review.State != "waiting_approval" || progress.Invocation.Review.InputHash == old.Review.InputHash || progress.Invocation.Preparation.Round != old.Preparation.Round+1 || p.calls != 0 {
		t.Fatal("moved base reused old validation or approval", err)
	}

	assertWorkflowPublicationMoved(t, f, p, progress, moved, old.Review.InputHash)
}

func assertWorkflowPublicationMoved(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture, progress IntegrationProgress, moved, oldHash string) {
	t.Helper()

	if _, err := f.store.ApproveWorkflowSlice(t.Context(), f.run.ID, oldHash, p); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("old manual approval survived target movement", err)
	}

	guarded, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	prepared, err := f.store.readWorkflowPreparation(guarded)
	if err != nil || prepared.Input.Git.BaseCommit != moved {
		t.Fatal("fresh host validation did not bind the actual base", err)
	}

	checkWorkflowPreparationTaskProofs(t, f.store, progress.Invocation)

	if _, err := f.store.ApproveWorkflowSlice(t.Context(), f.run.ID, progress.Invocation.Review.InputHash, p); err != nil {
		t.Fatal(err)
	}

	ports := workflowPublicationPorts(f, p)
	if _, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports); err != nil {
		t.Fatal(err)
	}

	final, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if err != nil || final.Receipt == nil || p.calls != 1 {
		t.Fatal("new exact approval did not produce one verified delivery", err)
	}
}

func workflowPublicationModeStart(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture, ports PublicationPorts, mode workflowstate.ReviewMode) (IntegrationProgress, error) {
	t.Helper()

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if mode == workflowstate.ReviewHuman {
		if err != nil || progress.Invocation.Review.State != "waiting_approval" || p.calls != 0 {
			t.Fatal("human policy skipped exact approval", err)
		}

		if _, err := f.store.ApproveWorkflowSlice(t.Context(), f.run.ID, progress.Invocation.Review.InputHash, p); err != nil {
			t.Fatal(err)
		}

		progress, err = f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	}

	return progress, err
}

func checkWorkflowPublicationConfirmation(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture, ports PublicationPorts, before Invocation) {
	t.Helper()

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	progress, err := restarted.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if err != nil || progress.Receipt == nil || progress.Invocation.Status != "completed" || len(progress.Invocation.Attempts) != len(before.Attempts) || p.calls != 1 {
		t.Fatal("verified delivery did not finalize without more work", err)
	}

	if progress.Receipt.Receipt.SubmissionID != before.Submission.ID || progress.Invocation.Submission.ExternalDelivery {
		t.Fatal("receipt lost its exact publication intent")
	}

	checkWorkflowPublicationFinalized(t, f, progress)

	checkWorkflowPublicationReuse(t, restarted, f.run.ID, ports, progress)
}

func checkWorkflowPublicationReuse(t *testing.T, restarted *Store, id string, ports PublicationPorts, progress IntegrationProgress) {
	t.Helper()

	again, err := restarted.AdvanceWorkflowSliceIntegration(t.Context(), id, ports)
	if err != nil || again.Invocation.Revision != progress.Invocation.Revision || again.Receipt.SourceHash != progress.Receipt.SourceHash {
		t.Fatal("completed slice repeated finalization", err)
	}
}
