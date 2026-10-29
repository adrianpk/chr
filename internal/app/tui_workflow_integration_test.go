package app

import (
	"errors"
	"maps"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

func tuiIntegrationRequest(shown WorkflowWorkStatus, action tui.WorkActionID) tui.WorkActionRequest {
	view := tuiWorkflowView(shown)

	return tui.WorkActionRequest{Action: action, Workflow: &tui.WorkflowActionRequest{Preview: true, ID: view.ID, ExpectedRevision: view.Revision, BindingHash: view.ApprovalHash, Integration: view.Integration, DecisionRevision: view.DecisionRevision, DecisionHash: view.DecisionHash, RequestKey: "surface-input", Fields: map[string]string{}}}
}

// Native grants preview without mutation, reject all displayed stale fences,
// then extend one ceiling without resetting admission or consumed accounting.
func TestTUIWorkflowGrant(t *testing.T) {
	operations, id := serviceIntegrationReady(t)

	shown, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	request := tuiIntegrationRequest(shown, tui.WorkActionWorkflowGrant)
	request.Workflow.Fields = map[string]string{"kind": "submission", "amount": "1", "actor": "operator", "reason": "Authorize another bounded submission"}
	executor := newTUIWorkActionExecutor(operations)
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	preview, err := executor.ExecuteWorkAction(request)
	if err != nil || preview.Confirmation == nil || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("grant preview mutated authority", err)
	}

	checkTUIIntegrationStale(t, operations, preview.Confirmation.Request)

	after, err := executor.ExecuteWorkAction(preview.Confirmation.Request)
	if err != nil || after.Workflow == nil {
		t.Fatal("grant failed", err)
	}

	checkTUIGrantBudget(t, operations, id, shown)
	checkTUIIntegrationStale(t, operations, preview.Confirmation.Request)
}

func checkTUIIntegrationStale(t *testing.T, operations service, request tui.WorkActionRequest) {
	t.Helper()

	cases := []struct {
		name   string
		change func(*tui.WorkflowActionRequest)
	}{
		{"invocation", func(r *tui.WorkflowActionRequest) { r.ExpectedRevision++ }},
		{"entry", func(r *tui.WorkflowActionRequest) { r.Integration.EntryRevision++ }},
		{"budget", func(r *tui.WorkflowActionRequest) { r.Integration.BudgetRevision++ }},
		{"candidate", func(r *tui.WorkflowActionRequest) { r.Integration.CandidateHash = strings.Repeat("d", 64) }},
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stale := *request.Workflow
			guard := *stale.Integration
			stale.Integration = &guard
			tc.change(&stale)

			candidate := request
			candidate.Workflow = &stale

			result, err := newTUIWorkActionExecutor(operations).ExecuteWorkAction(candidate)
			if !errors.Is(err, workflowstate.ErrConflict) || result.Workflow == nil || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
				t.Fatal("stale control changed authority or hid retained state", err)
			}
		})
	}
}

// Retained sealed questions use the shared answer owner. Resolve/defer never
// starts an agent, approves a candidate or fabricates an integrated unit.
func TestTUIWorkflowDecision(t *testing.T) {
	for _, action := range []string{"resolve", "defer"} {
		t.Run(action, func(t *testing.T) { checkTUIDecisionInput(t, action) })
	}
}

func checkTUIDecisionInput(t *testing.T, action string) {
	t.Helper()

	operations, id := serviceIntegrationDecisionReady(t)

	shown, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	request := tuiIntegrationRequest(shown, tui.WorkActionWorkflowAnswer)
	question := shown.Integration.Decision.Requests[0]
	request.Workflow.Fields = map[string]string{"decision_id": question.ID, "choice": question.Options[0], "answer_action": action, "actor": "operator", "authority": question.Authority, "reason": "Preserve both requirements"}
	executor := newTUIWorkActionExecutor(operations)
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	preview, err := executor.ExecuteWorkAction(request)
	if err != nil || preview.Confirmation == nil || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("decision preview changed authority", err)
	}

	stale := *preview.Confirmation.Request.Workflow
	stale.DecisionRevision++
	candidate := preview.Confirmation.Request

	candidate.Workflow = &stale
	if _, err := executor.ExecuteWorkAction(candidate); !errors.Is(err, workflowstate.ErrConflict) || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("obsolete question accepted", err)
	}

	if _, err := executor.ExecuteWorkAction(preview.Confirmation.Request); err != nil {
		t.Fatal(err)
	}

	after, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})

	state := "resolved"
	if action == "defer" {
		state = "waiting"
	}

	checkWorkflowIntegrationAnswer(t, shown, after, state, err)
}

func checkTUIGrantBudget(t *testing.T, operations service, id string, shown WorkflowWorkStatus) {
	t.Helper()

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.Integration.Budget.Limits.Submission != shown.Integration.Budget.Limits.Submission+1 || status.Integration.Budget.Consumed != shown.Integration.Budget.Consumed || status.Integration.Budget.AdmissionLimits != shown.Integration.Budget.AdmissionLimits {
		t.Fatal("grant changed consumption or admission", err)
	}
}

// Native approval captures the integration candidate as well as the review hash;
// confirming it authorizes review without publishing or marking tickets done.
func TestTUIWorkflowNativeApproval(t *testing.T) {
	operations, id := serviceIntegrationReady(t)

	shown, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	request := tuiIntegrationRequest(shown, tui.WorkActionWorkflowApprove)
	executor := newTUIWorkActionExecutor(operations)
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	preview, err := executor.ExecuteWorkAction(request)
	if err != nil || preview.Confirmation == nil || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("native approval preview changed authority", err)
	}

	checkTUIIntegrationStale(t, operations, preview.Confirmation.Request)

	if _, err := executor.ExecuteWorkAction(preview.Confirmation.Request); err != nil {
		t.Fatal(err)
	}

	after, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	checkServiceIntegrationApproval(t, shown, after, err)
}
