package app

import (
	"context"
	"fmt"
	"maps"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowStatusRequest bounds a retained invocation's member page.
type WorkflowStatusRequest struct {
	Offset           int
	Limit            int
	ExpectedRevision uint64
}

// WorkflowWorkEvidence exposes accepted checkpoints without host credentials.
type WorkflowWorkEvidence struct {
	ID         string `json:"id"`
	Purpose    string `json:"purpose"`
	TicketKey  string `json:"ticket_key"`
	Outcome    string `json:"outcome"`
	Slot       int    `json:"capacity_slot,omitempty"`
	ResultRef  string `json:"result_ref,omitempty"`
	ResultHash string `json:"result_hash,omitempty"`
}

// WorkflowWorkStatus distinguishes host task validation from confirmed delivery.
type WorkflowWorkStatus struct {
	WorkTitle string `json:"work_title,omitempty"`
	WorkflowWorkBindings
	Integration *WorkflowIntegrationStatus `json:"integration,omitempty"`
	WorkflowWorkLifecycle
	workflowstate.Identity
	WorkflowWorkSelection
	WorkflowWorkOutcome
	WorkflowWorkScope
	WorkflowWorkCounts
	WorkflowWorkControl
	WorkflowWorkDelivery
	WorkflowWorkPage
}

// WorkflowWorkBindings groups the retained bindings fields without changing their wire order.
type WorkflowWorkBindings struct {
	HarnessBinding *harnesscore.Binding         `json:"harness_binding,omitempty"`
	ControlReceipt *workflowstate.ControlResult `json:"control_receipt,omitempty"`
}

// WorkflowWorkLifecycle groups the retained lifecycle fields without changing their wire order.
type WorkflowWorkLifecycle struct {
	FrozenPolicy         workflowstate.ExecutionPolicy        `json:"execution_policy"`
	Continuation         *workflowcore.ContinuationCheckpoint `json:"continuation,omitempty"`
	NextActionOwner      string                               `json:"next_action_owner"`
	CapacityResumeStatus string                               `json:"capacity_resume_status,omitempty"`
	SelectedTicket       string                               `json:"selected_ticket,omitempty"`
	SelectedSlice        string                               `json:"selected_slice,omitempty"`
}

// WorkflowWorkSelection groups the retained selection fields without changing their wire order.
type WorkflowWorkSelection struct {
	Selector  workflowstate.Selector `json:"selector"`
	Authority string                 `json:"authority"`
}

// WorkflowWorkOutcome groups the retained outcome fields without changing their wire order.
type WorkflowWorkOutcome struct {
	Status  string `json:"status"`
	Outcome string `json:"outcome"`
}

// WorkflowWorkScope groups the retained scope fields without changing their wire order.
type WorkflowWorkScope struct {
	Target       workflowstate.PreviewTarget `json:"target"`
	Policy       workflowstate.ReviewPolicy  `json:"publication_policy"`
	SetOrderRule string                      `json:"set_order_rule"`
	Excluded     map[string]int              `json:"excluded"`
}

// WorkflowWorkCounts groups the retained counts fields without changing their wire order.
type WorkflowWorkCounts struct {
	Total            int `json:"total_members"`
	ValidatedTasks   int `json:"validated_tasks"`
	ConfirmedTickets int `json:"confirmed_tickets"`
	IntegratedUnits  int `json:"integrated_units"`
	Remaining        int `json:"remaining_tickets"`
}

// WorkflowWorkControl groups the retained control fields without changing their wire order.
type WorkflowWorkControl struct {
	StopIntent string                `json:"stop_intent"`
	Blocker    *workflowcore.Blocker `json:"blocker,omitempty"`
}

// WorkflowWorkDelivery groups the retained delivery fields without changing their wire order.
type WorkflowWorkDelivery struct {
	CandidateRef  string                           `json:"candidate_ref,omitempty"`
	CandidateHash string                           `json:"candidate_hash,omitempty"`
	ApprovalHash  string                           `json:"approval_hash,omitempty"`
	ReviewState   string                           `json:"review_state,omitempty"`
	PullRequest   *workflowcore.PullRequestBinding `json:"pull_request,omitempty"`
	DeliveryRef   string                           `json:"delivery_ref,omitempty"`
	DeliveryHash  string                           `json:"delivery_hash,omitempty"`
}

// WorkflowWorkPage groups the retained page fields without changing their wire order.
type WorkflowWorkPage struct {
	Members       []workflowcore.PreviewMember `json:"members"`
	Evidence      []WorkflowWorkEvidence       `json:"recent_evidence"`
	TotalAttempts int                          `json:"total_attempts"`
	NextOffset    int                          `json:"next_offset"`
}

// WorkflowStatus is read-only, including when a transition needs reconciliation.
func (s service) WorkflowStatus(ctx context.Context, id string, request WorkflowStatusRequest) (WorkflowWorkStatus, error) {
	if request.Offset < 0 || request.Offset > workflowstate.RecordLimit || request.Limit < 1 || request.Limit > workflowstate.PageLimit {
		return WorkflowWorkStatus{}, fmt.Errorf("workflow status page bounds exceeded")
	}

	store, err := s.workflowStore()
	if err != nil {
		return WorkflowWorkStatus{}, err
	}

	inspection, err := store.InspectWorkflow(ctx, id)
	if err != nil {
		return WorkflowWorkStatus{}, err
	}

	if request.ExpectedRevision != 0 && request.ExpectedRevision != inspection.Invocation.Revision {
		return WorkflowWorkStatus{}, fmt.Errorf("%w: restart status pagination", workflowstate.ErrConflict)
	}

	return workflowWorkStatus(inspection, request), nil
}

func workflowWorkStatus(inspection workflowcore.Inspection, request WorkflowStatusRequest) WorkflowWorkStatus {
	run := inspection.Invocation

	result := WorkflowWorkStatus{Identity: run.Identity, WorkflowWorkSelection: WorkflowWorkSelection{Selector: run.Selection.Selector, Authority: run.Authority}, WorkflowWorkOutcome: WorkflowWorkOutcome{Status: string(run.Status), Outcome: run.Outcome}, WorkflowWorkScope: WorkflowWorkScope{Target: run.Selection.Target, Policy: inspection.Policy, SetOrderRule: run.SetOrderRule, Excluded: maps.Clone(run.Excluded)}, WorkflowWorkCounts: WorkflowWorkCounts{Total: len(run.Members), ValidatedTasks: run.Cursor, ConfirmedTickets: run.ConfirmedCursor, IntegratedUnits: len(run.Deliveries), Remaining: len(run.Members) - run.ConfirmedCursor}, WorkflowWorkControl: WorkflowWorkControl{StopIntent: run.StopIntent, Blocker: run.Blocker}, WorkflowWorkDelivery: WorkflowWorkDelivery{CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, PullRequest: run.PullRequest, DeliveryRef: run.DeliveryRef, DeliveryHash: run.DeliveryHash}, WorkflowWorkPage: WorkflowWorkPage{Members: []workflowcore.PreviewMember{}, Evidence: []WorkflowWorkEvidence{}, TotalAttempts: len(run.Attempts), NextOffset: -1}}

	for _, record := range inspection.Tasks {
		if record.Task != nil && run.Selection.Selector.Extent == "ticket" && record.Task.Key == run.Selection.Selector.Identity {
			result.WorkTitle = nativeRecordView(record, RecordTicket).Title
		}
	}

	result.HarnessBinding = run.HarnessBinding
	if run.Review != nil {
		result.ApprovalHash, result.ReviewState = run.Review.InputHash, string(run.Review.State)
	}

	workflowWorkProgress(&result, run)

	result.Integration = workflowIntegrationStatus(inspection.Integration)

	workflowWorkMembers(&result, inspection, request)

	for _, attempt := range run.Attempts[max(0, len(run.Attempts)-workflowstate.PageLimit):] {
		purpose := attempt.Purpose
		if purpose == "" {
			purpose = "task"
		}

		result.Evidence = append(result.Evidence, WorkflowWorkEvidence{ID: attempt.ID, Purpose: purpose, TicketKey: attempt.TaskKey, Outcome: attempt.Outcome, Slot: attempt.CapacitySlot, ResultRef: attempt.ResultRef, ResultHash: attempt.ResultHash})
	}

	return result
}

func workflowWorkMembers(result *WorkflowWorkStatus, inspection workflowcore.Inspection, request WorkflowStatusRequest) {
	records := map[string]workflowcore.Record{}

	for _, record := range inspection.Tasks {
		if record.Task != nil {
			records[record.Task.Key] = record
		}
	}

	end := min(result.Total, request.Offset+request.Limit)
	for index := request.Offset; index < end; index++ {
		member := inspection.Invocation.Members[index].Admission
		if record, found := records[member.Key]; found {
			member.Revision, member.Status, member.Phase = record.Revision, record.Task.Status, record.Task.Phase
			if member.Phase != "pending" {
				member.Readiness, member.Blocker = member.Phase, nil
			}
		}

		result.Members = append(result.Members, member)
	}

	if end < result.Total {
		result.NextOffset = end
	}
}

func workflowWorkProgress(result *WorkflowWorkStatus, run workflowcore.Invocation) {
	if run.ExecutionTarget != nil {
		result.Target = *run.ExecutionTarget
	}

	result.FrozenPolicy, result.Continuation = run.Policy, run.Continuation
	result.CapacityResumeStatus = string(run.CapacityResumeStatus)

	result.NextActionOwner = "host"
	if run.Cursor < len(run.Members) {
		result.SelectedTicket = run.Members[run.Cursor].Admission.Key
		result.SelectedSlice = run.Members[run.Cursor].Admission.SliceID
	}

	if result.SelectedSlice == "" && run.Lane != nil && run.Lane.Unit.Kind == "slice" {
		result.SelectedSlice = run.Lane.Unit.ID
	}

	if run.Blocker != nil {
		result.NextActionOwner = run.Blocker.NextActionOwner

		return
	}

	switch run.Status {
	case workflowstate.InvocationWaitingCapacity, workflowstate.InvocationRunning:
		result.NextActionOwner = "current_owner"
	case workflowstate.InvocationWaitingReview, workflowstate.InvocationPaused, workflowstate.InvocationBlocked:
		result.NextActionOwner = "operator"
	case workflowstate.InvocationCompleted, workflowstate.InvocationCancelled:
		result.NextActionOwner = "none"
	}
}
