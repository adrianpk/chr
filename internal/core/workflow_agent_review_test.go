package core

import (
	"errors"
	"strings"
	"testing"
)

// TestNativeAgentReview counts source workers and a separate reviewer together,
// while aggregate host validation remains unmetered and raw review cannot bypass
// the certified driver even when it shares an executor implementation.
func TestNativeAgentReview(t *testing.T) {
	store, _, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))
	request.Policy.AgentTokenBudget = 300
	request.Policy.MaxWorkerAttempts = 3

	workflowReviewConfig(t, store, ReviewWorkflowModeAgenticReview)

	run := workflowTestStart(t, store, request)

	harness := nativeAgentHarness(t, store, run.Invocation.ID, "success")
	nativeAgentReviewSource(t, store, run.Invocation.ID, harness)

	candidate, err := store.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	provider := nativeAgentReviewProvider(t, store, request, run.Invocation.ID, candidate)

	if _, err := store.ReviewWorkflowSlice(t.Context(), run.Invocation.ID, provider, workflowReviewWorker("pass", "", "")); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("raw reviewer bypassed invocation accounting", err)
	}

	review, err := store.ReviewWorkflowSliceWithHarness(t.Context(), run.Invocation.ID, provider, harness)
	if err != nil || review.State != "authorized" {
		t.Fatal("certified reviewer failed", err)
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.Invocation.ID)
	if err != nil || budget.TokensSettled != 120 || budget.TokensReserved != 0 || len(budget.Reservations) != 4 || budget.Reservations[3].Role != "reviewer" {
		t.Fatal("roles lost shared usage or duplicated attempt consumption", err)
	}
}

func nativeAgentReviewProvider(t *testing.T, store *WorkflowStore, request WorkflowStartRequest, id string, candidate WorkflowSliceCandidate) *workflowReviewProviderFixture {
	t.Helper()

	pr := WorkflowPullRequestSnapshot{URL: request.Selection.Target.Repository + "/pulls/7", Repository: request.Selection.Target.Repository, HeadRepository: request.Selection.Target.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: request.Selection.Target.Commit, State: "open"}

	provider := &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}, target: pr.BaseCommit}
	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), id, workflowPRContent(), provider); err != nil {
		t.Fatal(err)
	}

	return provider
}

func nativeAgentReviewSource(t *testing.T, store *WorkflowStore, id string, harness WorkflowAgentHarness) {
	t.Helper()

	for range 3 {
		claim := workflowTestClaim(t, store, id)
		if _, err := store.ExecuteWorkflowTaskWithHarness(t.Context(), claim, harness); err != nil {
			t.Fatal(err)
		}

		if err := claim.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
