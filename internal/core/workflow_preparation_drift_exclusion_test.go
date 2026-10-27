package core

import (
	"context"
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeDriftExclusionGit struct {
	OSWorkflowPreparationGit
	store       *WorkflowStore
	key, target string
}

func (git nativeDriftExclusionGit) PrepareWorkflowIntegration(ctx context.Context, query WorkflowPreparationQuery) (WorkflowIntegrationProof, error) {
	if err := git.store.guardWorkflowSubmissionTarget(ctx, git.key, git.target); err != nil {
		return WorkflowIntegrationProof{}, err
	}

	return git.OSWorkflowPreparationGit.PrepareWorkflowIntegration(ctx, query)
}

// TestNativeDriftExclusion keeps uncertain transport excluded, then releases
// the target for another queue before preparing a changed-base replacement.
// The sealed failed submission remains consumed and readable after restart.
func TestNativeDriftExclusion(t *testing.T) {
	f, candidate := nativeReviewFixture(t, workflowstate.ReviewWorkflowModeTrustedAutoMerge)
	nativeReviewAuthorize(t, f)

	store, id := f.f.store, f.f.run.ID
	entry := nativeRoleTransition(t, store, nativeReviewEntry(t, store, id), "ready_to_publish")
	submit := nativeSubmissionFixture{store, entry, candidate, f.provider}
	submission := beginNativeSubmission(t, submit)
	nativeDriftMoveTarget(t, &f, "Externally advanced target")

	if _, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, f.provider, f.f.git); !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("pending native intent admitted replacement", err)
	}

	nativeDriftSealDispatch(t, submit, submission, f.f.base)

	store = NewWorkflowStore(store.layout, store.projectID)
	git := nativeDriftExclusionGit{f.f.git, store, contenthash.Sum([]byte("independent queue")), contenthash.Sum([]byte(candidate.Native.SourceRepository + "\n" + candidate.Native.TargetRef))}

	fresh, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, f.provider, git)
	if err != nil || fresh.Input.Git.BaseCommit != f.f.base {
		t.Fatal("target exclusion remained held during replacement", err)
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || budget.UnitCounters[0].Drift != 1 || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("replacement refunded dispatch or failed to charge drift", err)
	}

	if retained, err := store.ReadWorkflowSubmission(t.Context(), id); err != nil || retained.ID != submission.ID || retained.State != "unpublished" {
		t.Fatal("replacement lost the sealed submission", err)
	}
}

func nativeDriftSealDispatch(t *testing.T, submit nativeSubmissionFixture, submission WorkflowSubmission, target string) {
	t.Helper()

	nativeDriftSealInputs(t, submit, submission, submit.candidate.Input.Git.SourceCommit, target)
}

func nativeDriftSealInputs(t *testing.T, submit nativeSubmissionFixture, submission WorkflowSubmission, source, target string) {
	t.Helper()

	store, id := submit.store, submit.entry.InvocationID
	update := nativeSubmissionUpdate(t, submit, submission)
	identity, stop := nativeSubmissionDispatcher(t)
	stop()

	update.Facts.DispatchState, update.Facts.ApplicationState, update.Facts.DispatchIdentity = "quiescent", "not_applied", &identity
	proof := WorkflowNativeNonApplicationProof{SchemaVersion: 1, SubmissionID: submission.ID, CannotApply: true, InputChanged: true, ObservedSource: source, ObservedTarget: target}
	ref := retainNativeRecordEvidence(t, store, id, "target-drift-non-application", proof)

	update.Facts.NonApplicationRef = &ref
	if sealed, err := store.RetainWorkflowNativeSubmission(t.Context(), update); err != nil || sealed.State != "unpublished" {
		t.Fatal("changed input did not seal the failed dispatch", err)
	}
}

// TestNativeSealedSourceDrift requires an ownership/scope decision after an
// atomic rejection caused by a new source. Non-application frees the target,
// but cannot authorize the appended source content or refund submission usage.
func TestNativeSealedSourceDrift(t *testing.T) {
	f, candidate := nativeReviewFixture(t, workflowstate.ReviewWorkflowModeTrustedAutoMerge)
	nativeReviewAuthorize(t, f)

	store, id := f.f.store, f.f.run.ID
	entry := nativeRoleTransition(t, store, nativeReviewEntry(t, store, id), "ready_to_publish")
	submit := nativeSubmissionFixture{store, entry, candidate, f.provider}
	submission := beginNativeSubmission(t, submit)
	nativeDriftMoveSource(t, &f, "foreign scope", "unapproved.txt")
	nativeDriftSealInputs(t, submit, submission, f.provider.pr.HeadCommit, f.f.base)

	provider := nativeDriftProviderFixture{f.provider, f.f.remote}
	if _, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, provider, f.f.git); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("sealed source drift did not require a scope decision", err)
	}

	blocked := nativeReviewEntry(t, store, id)
	if blocked.State != "blocked" || blocked.Blocker.Reason != "source_revision_requires_reconciliation" || blocked.ClaimGeneration != entry.ClaimGeneration {
		t.Fatal("sealed source drift admitted replacement scope")
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || budget.UnitCounters[0].Submission != 1 || budget.UnitCounters[0].Drift != 0 {
		t.Fatal("source decision changed consumed actions", err)
	}
}
