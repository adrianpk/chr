package app

import (
	"maps"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowIntegrationRole displays one actual owned serial agent without
// deriving a second lane from a pending integration unit or changing its claim.
func TestWorkflowIntegrationRole(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.AdmitWorkflowIntegration(t.Context(), core.WorkflowIntegrationAdmission{InvocationID: id, ExpectedRevision: run.Revision, Unit: workflowstate.Ref{Kind: "slice", ID: hierarchy.Tasks[0].Task.SliceID}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"}); err != nil {
		t.Fatal(err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	})

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.Integration == nil || len(status.Integration.Roles) != 1 {
		t.Fatal("status lost the actual owned serial agent", err)
	}

	checkWorkflowIntegrationRole(t, status, claim.AttemptID)

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("serial display changed the retained owner or budget")
	}
}

func checkWorkflowIntegrationRole(t *testing.T, status WorkflowWorkStatus, attempt string) {
	t.Helper()

	cli := strings.Join(workflowStatusLines(status), "\n")
	tui := strings.Join(tuiWorkflowView(status).Lines, "\n")

	for _, text := range []string{cli, tui} {
		for _, fact := range []string{"Agent capacity: serial; limit 1; available 0", "Active agents: 1", "Agent lane 1: executor; running; attempt " + attempt} {
			if !strings.Contains(text, fact) {
				t.Fatal("display lost actual serial occupancy", fact)
			}
		}

		if strings.Contains(text, "Agent lane 2") {
			t.Fatal("serial display invented a parallel agent")
		}
	}
}

// TestWorkflowIntegrationFacts distinguishes observed parallel roles, bounded
// grants, unanswered choices and uncertain publication without inferred success.
func TestWorkflowIntegrationFacts(t *testing.T) {
	limit, available := int64(1000), int64(700)

	cases := []struct {
		name   string
		status WorkflowIntegrationStatus
		facts  []string
	}{
		{"parallel roles", WorkflowIntegrationStatus{MaximumRoles: 3, AvailableRoles: 1, Roles: []WorkflowIntegrationRoleStatus{{Role: "executor", Slot: 1, State: "running"}, {Role: "merger", Slot: 3, State: "reconciliation_required"}}}, []string{"Agent capacity: parallel; limit 3; available 1", "Active agents: 2", "Agent lane 1: executor; running", "Agent lane 3: merger; reconciliation_required"}},
		{"unmetered", WorkflowIntegrationStatus{Budget: WorkflowIntegrationBudgetStatus{TokensSettled: 200, TokensReserved: 100}}, []string{"Token budget: unmetered; settled 200; reserved 100"}},
		{"finite tokens", WorkflowIntegrationStatus{Budget: WorkflowIntegrationBudgetStatus{AdmissionLimit: &limit, TokenLimit: &limit, TokensAvailable: &available, TokensSettled: 200, TokensReserved: 100}}, []string{"Admission token limit: 1000", "Token budget: limit 1000; settled 200; reserved 100; available 700"}},
		{"budget block", WorkflowIntegrationStatus{Phase: "blocked", Publication: "application_unknown", Blocker: &core.WorkflowIntegrationBlocker{Reason: "submission_budget_exhausted", RequiredAction: "grant"}, Budget: WorkflowIntegrationBudgetStatus{AdmissionLimits: core.WorkflowIntegrationLimits{Submission: 3}, Limits: core.WorkflowIntegrationLimits{Submission: 4}, Consumed: core.WorkflowIntegrationLimits{Submission: 3}, Remaining: core.WorkflowIntegrationLimits{Submission: 1}}}, []string{"Integration phase: blocked", "Publication: application_unknown", "Required action: grant", "Integration blocker: submission_budget_exhausted", "Submission budget: admitted 3; limit 4; consumed 3; remaining 1"}},
		{"decision wait", WorkflowIntegrationStatus{Decision: &WorkflowIntegrationDecisionStatus{ID: "choices", Revision: 2, Hash: "displayed-hash", State: "waiting", Requests: []core.WorkflowMergerDecision{{ID: "meaning", Issue: "Choose intended behavior", Authority: "operator", Options: []string{"preserve source", "preserve target"}}}}}, []string{"Decision binding: choices revision 2; waiting; hash displayed-hash", "Question meaning: Choose intended behavior; authority operator; options preserve source; preserve target"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := strings.Join(workflowIntegrationLines(&tc.status), "\n")
			for _, fact := range tc.facts {
				if !strings.Contains(text, fact) {
					t.Fatal("integration display lost a retained decision fact", fact)
				}
			}
		})
	}
}
