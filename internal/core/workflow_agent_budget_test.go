package core

import (
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeAgentAdmission refuses unmetered execution before any worker birth
// and freezes the invocation ceiling before a queue entry exists.
func TestNativeAgentAdmission(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Policy.AgentTokenBudget = 100
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	defer func() { _ = claim.Close() }() //nolint:errcheck

	_, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker"))
	if !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("finite invocation launched an uncertified worker", err)
	}

	retained := workflowTestRead(t, store, run.Invocation.ID)
	if len(retained.Attempts[0].Processes) != 0 {
		t.Fatal("uncertified worker obtained a launch permit")
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.Invocation.ID)
	if err != nil || budget.TokenLimit == nil || *budget.TokenLimit != 100 || len(budget.Reservations) != 0 {
		t.Fatal("invocation ceiling absent or charged before admission", err)
	}
}

// TestNativeAgentCeiling refuses a grant larger than the remaining ceiling;
// neither admission retries nor a different role can create an uncharged call.
func TestNativeAgentCeiling(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Policy.AgentTokenBudget = 90
	run := workflowTestStart(t, store, request)
	harness := nativeAgentHarness(t, store, run.Invocation.ID, "success")

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	defer func() { _ = claim.Close() }() //nolint:errcheck

	if _, err := store.ExecuteWorkflowTaskWithHarness(t.Context(), claim, harness); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("call exceeded remaining invocation ceiling", err)
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.Invocation.ID)
	if err != nil || budget.TokensReserved != 0 || len(budget.Reservations) != 0 {
		t.Fatal("rejected grant mutated consumption", err)
	}

	if _, err := store.StartWorkflow(t.Context(), WorkflowStartRequest{RequestKey: request.RequestKey, Selection: request.Selection, Policy: workflowstate.WorkflowExecutionPolicy{AgentTokenBudget: 91}}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("entry retry replaced admission ceiling", err)
	}
}

// TestNativeAgentLateCeiling refuses to introduce a finite ceiling after an
// unmetered call whose usage was never reserved or certified.
func TestNativeAgentLateCeiling(t *testing.T) {
	store, record, run := workflowTicketValidated(t)
	limit := int64(100)

	_, err := store.AdmitWorkflowIntegration(t.Context(), WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: workflowstate.Ref{Kind: "ticket", ID: record.Task.Key}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev", TokenLimit: &limit})
	if !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("finite admission forgot earlier unmetered usage", err)
	}
}

// TestNativeAgentPermitBudget fences a raw input prepared before a ceiling
// appears. Even an injected host-state change cannot grant its launch permit.
func TestNativeAgentPermitBudget(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	defer func() { _ = claim.Close() }() //nolint:errcheck

	prepared, err := store.prepareWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker"))
	if err != nil {
		t.Fatal(err)
	}

	nativeAgentInjectCeiling(t, store, run.Invocation.ID)

	if _, err := workflowRunAgentCommand(t.Context(), claim, prepared, prepared.input.Worker); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("raw input obtained a launch permit after ceiling changed", err)
	}

	retained := workflowTestRead(t, store, run.Invocation.ID)
	if len(retained.Attempts[0].Processes) != 0 || workflowSelectionGit(t, store, "rev-parse", prepared.input.Lane.Branch) != request.Selection.Target.Commit {
		t.Fatal("uncertified worker executed")
	}
}

func nativeAgentInjectCeiling(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	if err := store.withLock(t.Context(), func() error {
		run, err := store.readWorkflowInvocation(id)
		if err != nil {
			return err
		}

		limit := int64(100)
		limits := workflowIntegrationAdmissionLimits(run.Policy)
		budget := WorkflowIntegrationBudget{SchemaVersion: 1, InvocationID: run.ID, TokenLimit: &limit, AdmissionTokenLimit: &limit, AdmissionLimits: limits, LimitsSnapshot: limits}

		return store.commitWorkflowIntegrationBudget(t.Context(), &run, &budget)
	}); err != nil {
		t.Fatal(err)
	}
}
