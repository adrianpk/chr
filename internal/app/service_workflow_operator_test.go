package app

import (
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestServiceWorkflowGrant uses the common owner before queue admission and
// retains the frozen policy while rejecting a stale second operator request.
func TestServiceWorkflowGrant(t *testing.T) {
	operations, hierarchy, target := serviceWorkflowTaskFixture(t)
	request := core.WorkflowStartRequest{RequestKey: "budget-grant", Selection: core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: core.WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}}, Policy: core.WorkflowExecutionPolicy{AgentTokenBudget: 100}}

	started, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), started.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	grant := core.WorkflowIntegrationGrantRequest{WorkflowIntegrationBudgetGuard: core.WorkflowIntegrationBudgetGuard{InvocationID: started.Invocation.ID, InvocationRevision: started.Invocation.Revision, BudgetRevision: budget.Revision}, Grant: workflowstate.WorkflowIntegrationGrant{ID: "operator-grant", Kind: "tokens", Amount: 25, Actor: "operator", Reason: "Authorize one bounded extension"}}

	budget, err = operations.GrantWorkflowIntegration(t.Context(), grant)
	if err != nil || *budget.TokenLimit != 125 || *budget.AdmissionTokenLimit != 100 {
		t.Fatal("service grant lost frozen policy", err)
	}

	grant.Grant.ID = "stale-grant"
	if _, err := operations.GrantWorkflowIntegration(t.Context(), grant); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("service bypassed revision guard", err)
	}

	if _, err := operations.ReadWorkflowIntegrationDecisions(t.Context(), started.Invocation.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("service invented an approval", err)
	}
}
