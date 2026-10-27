package app

import (
	"errors"
	"maps"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func serviceIntegrationControl(status WorkflowWorkStatus, action string) WorkflowControlRequest {
	current := status.Integration

	return WorkflowControlRequest{Action: action, ID: status.ID, ExpectedRevision: status.Revision, BindingHash: status.ApprovalHash, Integration: &WorkflowIntegrationControlGuard{EntryID: current.EntryID, EntryRevision: current.EntryRevision, BudgetRevision: current.Budget.Revision, CandidateHash: current.CandidateHash}}
}

// TestServiceIntegrationGrant rejects every stale displayed fence before the
// common owner can extend a ceiling, then preserves consumption on a fresh grant.
func TestServiceIntegrationGrant(t *testing.T) {
	operations, id := serviceIntegrationReady(t)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	request := serviceIntegrationControl(status, "grant")
	request.BudgetGrant = &core.WorkflowIntegrationGrant{ID: "more-submissions", Kind: "submission", Amount: 1, Actor: "operator", Reason: "Authorize one additional bounded submission"}
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	cases := []struct {
		name   string
		change func(*WorkflowControlRequest)
	}{
		{"invocation", func(r *WorkflowControlRequest) { r.ExpectedRevision++ }},
		{"entry", func(r *WorkflowControlRequest) { r.Integration.EntryRevision++ }},
		{"budget", func(r *WorkflowControlRequest) { r.Integration.BudgetRevision++ }},
		{"candidate", func(r *WorkflowControlRequest) { r.Integration.CandidateHash = "stale-candidate" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stale := request
			guard := *request.Integration
			stale.Integration = &guard
			tc.change(&stale)

			if _, err := operations.ControlWorkExtent(t.Context(), stale); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("stale native grant accepted", err)
			}

			if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
				t.Fatal("rejected grant changed authority")
			}
		})
	}

	checkServiceIntegrationGrant(t, operations, request, status)
}

func checkServiceIntegrationGrant(t *testing.T, operations service, request WorkflowControlRequest, before WorkflowWorkStatus) {
	t.Helper()

	after, err := operations.ControlWorkExtent(t.Context(), request)
	if err != nil || after.Integration == nil || after.Integration.Budget.Limits.Submission != before.Integration.Budget.Limits.Submission+1 || after.Integration.Budget.Consumed != before.Integration.Budget.Consumed || after.Integration.Budget.AdmissionLimits != before.Integration.Budget.AdmissionLimits {
		t.Fatal("explicit grant erased consumption or lost the bounded extension", after, err)
	}

	images := workflowCLIImage(t, operations.projectContext.ProjectDir)
	if _, err := operations.ControlWorkExtent(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("old surface revision granted again", err)
	}

	if !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("stale retry changed authority")
	}
}

// TestServiceIntegrationApproval binds manual review to the displayed native
// candidate while leaving publication, receipts and member finalization pending.
func TestServiceIntegrationApproval(t *testing.T) {
	operations, id := serviceIntegrationReady(t)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	request := serviceIntegrationControl(status, "approve")

	after, err := operations.ControlWorkExtent(t.Context(), request)
	checkServiceIntegrationApproval(t, status, after, err)

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	if _, err := operations.ControlWorkExtent(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale surface approval accepted", err)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("stale approval changed authority")
	}
}

func checkServiceIntegrationApproval(t *testing.T, before, after WorkflowWorkStatus, err error) {
	t.Helper()

	if err != nil || after.Integration == nil || after.ReviewState != "authorized" || after.Integration.Phase != before.Integration.Phase || after.Integration.Publication != "not_started" || after.IntegratedUnits != 0 || after.ConfirmedTickets != 0 {
		t.Fatal("native review approval failed or fabricated delivery", after, err)
	}
}
