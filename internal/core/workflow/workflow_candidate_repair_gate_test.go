package workflow

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeCandidateRepair composes actual semantic failure/review, a broker
// merger, host Git acceptance, renewed gate and fresh candidate-bound review.
func TestNativeCandidateRepair(t *testing.T) {
	for _, method := range []string{"merge", "squash", "rebase"} {
		for _, reason := range []string{"combined_gate_regression", "actionable_review_finding"} {
			t.Run(method+"/"+reason, func(t *testing.T) { nativeCandidateRepairGate(t, method, reason) })
		}
	}
}

func nativeCandidateRepairGate(t *testing.T, method, reason string) {
	t.Helper()

	f, request, harness := nativeCandidateRepairFixture(t, method, reason)
	store, id := f.f.store, f.f.run.ID
	before := workflowTestRead(t, store, id)

	attempt, err := ExecuteMergerRepair(t.Context(), store, request, harness, MergerRepairPorts{Provider: f.provider, Git: f.f.git})
	if err != nil || attempt.Acceptance == nil || attempt.Acceptance.Outcome != "accepted_pending_validation" || attempt.Acceptance.Next.Result == nil {
		t.Fatal("semantic repair was not accepted", attempt.Failure, err)
	}

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, f.provider); err == nil {
		t.Fatal("merger acceptance granted publication without new gate/review")
	}

	store = NewWorkflowStore(store.layout, store.projectID)

	candidate, err := PrepareIntegration(t.Context(), store, id, f.provider, f.f.git)
	if err != nil {
		t.Fatal("accepted result did not receive the unchanged host gate", err)
	}

	checkNativeCandidateRevalidated(t, f, store, before, candidate, attempt)
}

func checkNativeCandidateRevalidated(t *testing.T, f nativePreparationFixture, store *Store, before workflowstate.Invocation, candidate workflowstate.IntegrationCandidate, attempt MergerAttempt) {
	t.Helper()

	id := f.f.run.ID

	if candidate.Input.Git.ResultCommit != attempt.Acceptance.Next.Result.ResultCommit || candidate.Round <= before.Preparation.Round || len(candidate.Native.MergerAttemptRefs) != 1 {
		t.Fatal("new gate lost accepted P, history or merger evidence")
	}

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, f.provider); err == nil {
		t.Fatal("repaired gate inherited old review authority")
	}

	f.f.store = store
	binding := nativeReviewAuthorize(t, f)

	guard, err := store.CheckWorkflowPublicationGuard(t.Context(), id, f.provider)
	if err != nil || guard.Git.ResultCommit != candidate.Input.Git.ResultCommit || guard.ReviewHash != binding.InputHash {
		t.Fatal("fresh review did not authorize the repaired exact result", err)
	}

	assertWorkflowPreparationSource(t, f.f)
	checkNativeMergerBudget(t, f.f)
	checkNativeCandidateReplay(t, candidate, attempt)
}

func checkNativeCandidateReplay(t *testing.T, candidate workflowstate.IntegrationCandidate, attempt MergerAttempt) {
	t.Helper()

	proof := candidate.Input.Git
	if proof.Method != "rebase" {
		return
	}

	if len(proof.History) != len(candidate.Input.Source.Slice.Tasks) || proof.History[len(proof.History)-1].Source != proof.SourceCommit || len(proof.Provenance.Repairs) != 1 || proof.Provenance.Repairs[0].Input.InputHash != attempt.Input.InputHash {
		t.Fatal("semantic repair lost original replay attribution")
	}

	checkNativeReplayMessages(t, candidate.Lane.Worktree, proof)
}
