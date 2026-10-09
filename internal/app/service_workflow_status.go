package app

import (
	"context"
	"fmt"
	"maps"

	"forge.adrianpk.com/chr/internal/core"
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
	FrozenPolicy         core.WorkflowExecutionPolicy         `json:"execution_policy"`
	Continuation         *core.WorkflowContinuationCheckpoint `json:"continuation,omitempty"`
	NextActionOwner      string                               `json:"next_action_owner"`
	CapacityResumeStatus string                               `json:"capacity_resume_status,omitempty"`
	SelectedTicket       string                               `json:"selected_ticket,omitempty"`
	SelectedSlice        string                               `json:"selected_slice,omitempty"`
	core.WorkflowIdentity
	Selector         core.WorkflowSelector            `json:"selector"`
	Authority        string                           `json:"authority"`
	Status           string                           `json:"status"`
	Outcome          string                           `json:"outcome"`
	Target           core.WorkflowPreviewTarget       `json:"target"`
	Policy           core.WorkflowReviewPolicy        `json:"publication_policy"`
	SetOrderRule     string                           `json:"set_order_rule"`
	Excluded         map[string]int                   `json:"excluded"`
	Total            int                              `json:"total_members"`
	ValidatedTasks   int                              `json:"validated_tasks"`
	ConfirmedTickets int                              `json:"confirmed_tickets"`
	IntegratedUnits  int                              `json:"integrated_units"`
	Remaining        int                              `json:"remaining_tickets"`
	StopIntent       string                           `json:"stop_intent"`
	Blocker          *core.WorkflowBlocker            `json:"blocker,omitempty"`
	CandidateRef     string                           `json:"candidate_ref,omitempty"`
	CandidateHash    string                           `json:"candidate_hash,omitempty"`
	ApprovalHash     string                           `json:"approval_hash,omitempty"`
	ReviewState      string                           `json:"review_state,omitempty"`
	PullRequest      *core.WorkflowPullRequestBinding `json:"pull_request,omitempty"`
	DeliveryRef      string                           `json:"delivery_ref,omitempty"`
	DeliveryHash     string                           `json:"delivery_hash,omitempty"`
	Members          []core.WorkflowPreviewMember     `json:"members"`
	Evidence         []WorkflowWorkEvidence           `json:"recent_evidence"`
	TotalAttempts    int                              `json:"total_attempts"`
	NextOffset       int                              `json:"next_offset"`
}

// WorkflowStatus is read-only, including when a transition needs reconciliation.
func (s service) WorkflowStatus(ctx context.Context, id string, request WorkflowStatusRequest) (WorkflowWorkStatus, error) {
	if request.Offset < 0 || request.Offset > core.WorkflowRecordLimit || request.Limit < 1 || request.Limit > core.WorkflowPageLimit {
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
		return WorkflowWorkStatus{}, fmt.Errorf("%w: restart status pagination", core.ErrWorkflowConflict)
	}

	return workflowWorkStatus(inspection, request), nil
}

func workflowWorkStatus(inspection core.WorkflowInspection, request WorkflowStatusRequest) WorkflowWorkStatus {
	run := inspection.Invocation

	result := WorkflowWorkStatus{WorkflowIdentity: run.WorkflowIdentity, Selector: run.Selection.Selector, Authority: run.Authority, Status: run.Status, Outcome: run.Outcome, Target: run.Selection.Target, Policy: inspection.Policy, SetOrderRule: run.SetOrderRule, Excluded: maps.Clone(run.Excluded), Total: len(run.Members), ValidatedTasks: run.Cursor, ConfirmedTickets: run.ConfirmedCursor, IntegratedUnits: len(run.Deliveries), Remaining: len(run.Members) - run.ConfirmedCursor, StopIntent: run.StopIntent, Blocker: run.Blocker, CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, PullRequest: run.PullRequest, DeliveryRef: run.DeliveryRef, DeliveryHash: run.DeliveryHash, Members: []core.WorkflowPreviewMember{}, Evidence: []WorkflowWorkEvidence{}, TotalAttempts: len(run.Attempts), NextOffset: -1}
	if run.Review != nil {
		result.ApprovalHash, result.ReviewState = run.Review.InputHash, run.Review.State
	}

	workflowWorkProgress(&result, run)

	workflowWorkMembers(&result, inspection, request)

	for _, attempt := range run.Attempts[max(0, len(run.Attempts)-core.WorkflowPageLimit):] {
		purpose := attempt.Purpose
		if purpose == "" {
			purpose = "task"
		}

		result.Evidence = append(result.Evidence, WorkflowWorkEvidence{ID: attempt.ID, Purpose: purpose, TicketKey: attempt.TaskKey, Outcome: attempt.Outcome, Slot: attempt.CapacitySlot, ResultRef: attempt.ResultRef, ResultHash: attempt.ResultHash})
	}

	return result
}

func workflowWorkMembers(result *WorkflowWorkStatus, inspection core.WorkflowInspection, request WorkflowStatusRequest) {
	records := map[string]core.WorkflowRecord{}

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

func workflowWorkProgress(result *WorkflowWorkStatus, run core.WorkflowInvocation) {
	if run.ExecutionTarget != nil {
		result.Target = *run.ExecutionTarget
	}

	result.FrozenPolicy, result.Continuation = run.Policy, run.Continuation
	result.CapacityResumeStatus = run.CapacityResumeStatus

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
	case "waiting_capacity", "running":
		result.NextActionOwner = "current_owner"
	case "waiting_review", "paused", "blocked":
		result.NextActionOwner = "operator"
	case "completed", "cancelled":
		result.NextActionOwner = "none"
	}
}
