package core

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeGrantGuards rejects nonpositive, overflowing, stale and unbounded
// extensions before changing the ledger, including before a queue entry exists.
func TestNativeGrantGuards(t *testing.T) {
	store, _, start := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	start.Policy.AgentTokenBudget = 100
	run := workflowTestStart(t, store, start)

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	base := WorkflowIntegrationGrantRequest{WorkflowIntegrationBudgetGuard: nativeOperatorBudgetGuard(t, store, run.Invocation.ID), Grant: WorkflowIntegrationGrant{ID: "extension", Kind: "tokens", Amount: 1, Actor: "operator", Reason: "Authorize a bounded retry"}}

	cases := []struct {
		name   string
		change func(*WorkflowIntegrationGrantRequest)
	}{
		{"zero", func(r *WorkflowIntegrationGrantRequest) { r.Grant.Amount = 0 }},
		{"negative", func(r *WorkflowIntegrationGrantRequest) { r.Grant.Amount = -1 }},
		{"overflow", func(r *WorkflowIntegrationGrantRequest) { r.Grant.Amount = math.MaxInt64 }},
		{"role bound", func(r *WorkflowIntegrationGrantRequest) {
			r.Grant.Kind = "executor"
			r.Grant.Amount = WorkflowRecordLimit
		}},
		{"invocation revision", func(r *WorkflowIntegrationGrantRequest) { r.InvocationRevision++ }},
		{"budget revision", func(r *WorkflowIntegrationGrantRequest) { r.BudgetRevision++ }},
		{"foreign entry", func(r *WorkflowIntegrationGrantRequest) { r.EntryID = "foreign" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := base
			tc.change(&request)

			if _, err := store.GrantWorkflowIntegration(t.Context(), request); err == nil {
				t.Fatal("invalid extension accepted")
			}

			again, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.Invocation.ID)
			if err != nil || !reflect.DeepEqual(budget, again) {
				t.Fatal("rejected extension mutated history", err)
			}
		})
	}
}

// TestNativeGrantConcurrent lets exactly one fresh pre-queue grant use a shared
// revision. A losing operator cannot silently add another extension.
func TestNativeGrantConcurrent(t *testing.T) {
	store, _, start := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	start.Policy.AgentTokenBudget = 100
	run := workflowTestStart(t, store, start)
	guard := nativeOperatorBudgetGuard(t, store, run.Invocation.ID)
	results := make(chan error, 2)

	var group sync.WaitGroup
	for _, id := range []string{"first-grant", "second-grant"} {
		group.Go(func() {
			_, err := store.GrantWorkflowIntegration(t.Context(), WorkflowIntegrationGrantRequest{WorkflowIntegrationBudgetGuard: guard, Grant: WorkflowIntegrationGrant{ID: id, Kind: "tokens", Amount: 50, Actor: "operator", Reason: "Authorize one bounded extension"}})
			results <- err
		})
	}

	group.Wait()
	close(results)

	successes := 0

	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal(err)
		}
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.Invocation.ID)
	if err != nil || successes != 1 || len(budget.Grants) != 1 || *budget.TokenLimit != 150 {
		t.Fatal("shared revision granted twice", err)
	}
}
