package core

import (
	"context"
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeGitInspectorFixture struct {
	inspector WorkflowGitPublicationInspector
	after     func(context.Context)
}

func (p nativeGitInspectorFixture) InspectWorkflowGitPublication(ctx context.Context, request WorkflowGitPublication) (WorkflowGitPublicationAdvertisement, error) {
	advertisement, err := p.inspector.InspectWorkflowGitPublication(ctx, request)
	p.after(ctx)

	return advertisement, err
}

// TestNativeGitFence rejects an observation whose queue owner changed during IO.
// The common journal is unchanged; a fresh observer recovers without another send.
func TestNativeGitFence(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	publisher, remote := nativeDispatchPublisher(t, f)

	submission, _, err := f.store.PublishWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider, publisher)
	if err != nil {
		t.Fatal(err)
	}

	transport := NewWorkflowGitPublisher(f.candidate.Native.SourceRepository, remote)
	git := NewWorkflowIntegrationGit(f.candidate.Native.SourceRepository, remote)
	guard := nativeBudgetGuard(t, f.store, submission.InvocationID)

	inspector := nativeGitInspectorFixture{transport, func(ctx context.Context) {
		entry := nativeReviewEntry(t, f.store, submission.InvocationID)

		run, err := f.store.readWorkflowInvocation(submission.InvocationID)
		if err != nil {
			t.Fatal(err)
		}

		request := WorkflowIntegrationTransition{InvocationID: run.ID, EntryID: entry.ID, InvocationRevision: run.Revision, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, CandidateHash: entry.CandidateRef.Hash, State: "blocked"}

		request.Blocker = &WorkflowIntegrationBlocker{Reason: "observation_hold", RequiredAction: "observe current owner", ResumeState: "observing"}
		if _, err := f.store.TransitionWorkflowIntegration(ctx, request); err != nil {
			t.Fatal(err)
		}
	}}
	if _, err := f.store.ReconcileWorkflowNativeGitSubmission(t.Context(), guard, inspector, git); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale observation changed the common journal", err)
	}

	retained, err := f.store.ReadWorkflowSubmission(t.Context(), submission.InvocationID)
	if err != nil || !workflowNativeJSONEqual(retained, submission) {
		t.Fatal("stale callback rewrote retained evidence", err)
	}

	retained, err = f.store.ReconcileWorkflowNativeGitSubmission(t.Context(), nativeBudgetGuard(t, f.store, submission.InvocationID), transport, git)
	if err != nil || retained.Native.ApplicationState != "git_applied" || publisher.calls.Load() != 1 {
		t.Fatal("fresh observation failed to recover the same submission", err)
	}
}
