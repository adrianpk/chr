package workflow

import (
	"reflect"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeGrantEarly extends an exhausted invocation before queue admission
// without refunding unknown usage or replacing its original execution policy.
func TestNativeGrantEarly(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Policy.AgentTokenBudget = 100
	request.Policy.MaxWorkerAttempts = 1
	run := workflowTestStart(t, store, request)
	harness := nativeAgentHarness(t, store, run.Invocation.ID, "missing")

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	if _, err := ExecuteTaskWithHarness(t.Context(), store, claim, harness); err == nil {
		t.Fatal("missing usage accepted")
	}

	if err := claim.Close(); err != nil {
		t.Fatal(err)
	}

	grant := IntegrationGrantRequest{IntegrationBudgetGuard: nativeOperatorBudgetGuard(t, store, run.Invocation.ID), Grant: workflowstate.IntegrationGrant{ID: "token-extension", Kind: "tokens", Amount: 50, Actor: "operator", Reason: "Authorize one bounded retry after unknown usage"}}

	budget, err := store.GrantWorkflowIntegration(t.Context(), grant)
	if err != nil || budget.TokenLimit == nil || *budget.TokenLimit != 150 || budget.TokensReserved != 100 || *budget.AdmissionTokenLimit != 100 {
		t.Fatal("pre-queue grant failed or erased consumption", err)
	}

	again, err := store.GrantWorkflowIntegration(t.Context(), grant)
	if err != nil || !reflect.DeepEqual(budget, again) {
		t.Fatal("grant retry extended twice", err)
	}

	nativeOperatorWorkerRetry(t, store, run.Invocation.ID, grant, harness)
}

func nativeOperatorWorkerRetry(t *testing.T, store *Store, id string, grant IntegrationGrantRequest, harness AgentHarness) {
	t.Helper()

	grant.IntegrationBudgetGuard = nativeOperatorBudgetGuard(t, store, id)

	grant.Grant.ID, grant.Grant.Kind, grant.Grant.Amount = "worker-extension", "executor", 1
	if _, err := store.GrantWorkflowIntegration(t.Context(), grant); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ResumeWorkflow(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	second := workflowTestClaim(t, store, id)
	defer func() { _ = second.Close() }() //nolint:errcheck

	harness.InputTokens, harness.OutputTokens = 30, 20
	if _, err := store.prepareWorkflowTaskHarness(t.Context(), second, harness.Command, "automatic", &harness); err != nil {
		t.Fatal("explicit worker extension did not permit retry", err)
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)

	retained := workflowTestRead(t, store, id)
	if err != nil || budget.TokensReserved != 150 || len(budget.Grants) != 2 || retained.Policy.MaxWorkerAttempts != 1 || budget.Usage.WorkerCount != 2 {
		t.Fatal("retry reset history or admission policy", err)
	}
}

func nativeOperatorBudgetGuard(t *testing.T, store *Store, id string) IntegrationBudgetGuard {
	t.Helper()

	run := workflowTestRead(t, store, id)

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	return IntegrationBudgetGuard{InvocationID: id, InvocationRevision: run.Revision, BudgetRevision: budget.Revision}
}

// TestNativeGrantUnmetered creates the existing invocation ledger on its first
// explicit extension and migrates already consumed unmetered worker attempts.
func TestNativeGrantUnmetered(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "fail"), 30)
	request.Policy.MaxWorkerAttempts = 1
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	if _, err := ExecuteTask(t.Context(), store, claim, []string{"false"}); err == nil {
		t.Fatal("fixture worker unexpectedly passed")
	}

	if err := claim.Close(); err != nil {
		t.Fatal(err)
	}

	retained := workflowTestRead(t, store, run.Invocation.ID)
	grant := IntegrationGrantRequest{IntegrationBudgetGuard: IntegrationBudgetGuard{InvocationID: retained.ID, InvocationRevision: retained.Revision}, Grant: workflowstate.IntegrationGrant{ID: "worker-extension", Kind: "executor", Amount: 1, Actor: "operator", Reason: "Authorize one bounded worker retry"}}

	budget, err := store.GrantWorkflowIntegration(t.Context(), grant)
	if err != nil || budget.TokenLimit != nil || budget.Usage.WorkerCount != 1 || budget.AdmissionLimits.Worker != 1 || budget.LimitsSnapshot.Worker != 2 {
		t.Fatal("unmetered grant lost consumed history", err)
	}

	if _, err := store.ResumeWorkflow(t.Context(), run.Invocation.ID); err != nil {
		t.Fatal(err)
	}

	retry := workflowTestClaim(t, store, run.Invocation.ID)
	if err := retry.Close(); err != nil {
		t.Fatal(err)
	}
}
