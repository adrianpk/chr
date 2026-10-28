package workflow

import (
	"context"
	"errors"
	"fmt"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowSubmissionBudget preserves the consumed per-unit attempts through
// restart and new combined candidate revisions. Only a sealed result licenses
// another attempt; the default three submissions cannot be silently reset.
func TestWorkflowSubmissionBudget(t *testing.T) {
	f, provider := newWorkflowSubmissionFixture(t, "rebase")
	provider.unknown = true
	git := workflowSubmissionObserver(f)

	var previousHash string

	for attempt := 1; attempt <= 3; attempt++ {
		store := NewWorkflowStore(f.store.layout, f.store.projectID)
		if attempt > 1 {
			workflowSubmissionMoveBase(t, f, provider, store, attempt)
		}

		submission, err := SubmitIntegration(t.Context(), store, f.run.ID, provider, git)
		if !errors.Is(err, context.DeadlineExceeded) || submission.Attempt != attempt || submission.Guard.PreparationHash == previousHash {
			t.Fatalf("attempt %d: %+v, %v", attempt, submission, err)
		}

		previousHash = submission.Guard.PreparationHash
		provider.seal(submission.ID)

		if _, err := ReconcileSubmission(t.Context(), store, f.run.ID, provider, git); err != nil {
			t.Fatal(err)
		}
	}

	assertWorkflowSubmissionBudget(t, f, provider)
}

func assertWorkflowSubmissionBudget(t *testing.T, f workflowPreparationFixture, provider *workflowSubmissionProviderFixture) {
	t.Helper()

	if _, err := SubmitIntegration(t.Context(), f.store, f.run.ID, provider, workflowSubmissionObserver(f)); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatalf("exhausted budget: %v", err)
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Status != workflowstate.InvocationBlocked || run.Blocker.Reason != "integration_submission_budget_exhausted" || provider.calls != 3 || run.PreparationCount != 3 {
		t.Fatal("candidate replacement reset budget or launched extra work")
	}

	checkWorkflowPreparationTaskProofs(t, f.store, run)
}

func workflowSubmissionMoveBase(t *testing.T, f workflowPreparationFixture, provider *workflowSubmissionProviderFixture, store *Store, number int) {
	t.Helper()

	workflowIntegrationCommit(t, f.actor, fmt.Sprintf("target-%d.txt", number), "More target content")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	if _, err := PrepareIntegration(t.Context(), store, f.run.ID, provider, f.git); err != nil {
		t.Fatal(err)
	}

	if _, err := ReviewSlice(t.Context(), store, f.run.ID, provider, nil); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowSubmissionPolicy admits explicit finite limits while preserving
// legacy request identities when the newly delivered policy field is absent.
func TestWorkflowSubmissionPolicy(t *testing.T) {
	for _, limit := range []int{-1, 0, 1, 10, 11} {
		t.Run(fmt.Sprintf("limit %d", limit), func(t *testing.T) {
			store, _, request := workflowClaimFixture(t)
			request.Policy.MaxIntegrationAttempts = limit
			result, err := store.StartWorkflow(t.Context(), request)

			valid := limit >= 0 && limit <= 10
			if (err == nil) != valid {
				t.Fatalf("policy admission: %v", err)
			}

			if valid && result.Invocation.Policy.MaxIntegrationAttempts != limit {
				t.Fatal("policy filling rewrote an existing request identity")
			}
		})
	}
}
