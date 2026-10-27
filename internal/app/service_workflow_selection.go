package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowBoundPreview freezes the displayed native selection for confirmation.
type WorkflowBoundPreview struct {
	ReadyTicket *core.WorkflowTicketAdmission `json:"ready_ticket,omitempty"`
	Request     core.WorkflowPreviewRequest   `json:"request"`
	Selected    *core.WorkflowSelectionGuard  `json:"selected,omitempty"`
	View        WorkflowWorkPreview           `json:"preview"`
}

// PreviewSelectedWork resolves a selected ticket's native hierarchy identities.
// An explicit use-case identity is supplied by the case view, without a ticket.
func (s service) PreviewSelectedWork(ctx context.Context, request core.WorkflowPreviewRequest, key string) (result WorkflowBoundPreview, err error) {
	result.Request = request
	result.Request.Limit = core.WorkflowPageLimit

	if key != "" && !validWorkflowControlIdentity(key) {
		return result, fmt.Errorf("invalid selected workflow ticket identity")
	}

	if key != "" {
		selector, guard, err := s.selectedWorkflowIdentity(ctx, request.Selector.Extent, key)
		if err != nil {
			return result, err
		}

		result.Request.Selector, result.Selected = selector, &guard
	}

	result.View, err = s.PreviewWorkExtent(ctx, result.Request)
	result.Request.SnapshotHash = result.View.Preview.SnapshotHash

	result.Request.Target = result.View.Target
	if err == nil && result.Request.Selector.Extent == "ticket" && len(result.View.Preview.Members) == 1 {
		store, storeErr := s.workflowStore()
		if storeErr != nil {
			return result, storeErr
		}

		result.ReadyTicket, err = store.PreviewWorkflowTicketAdmission(ctx, result.View.Preview.Members[0])
	}

	return result, err
}

func (s service) selectedWorkflowIdentity(ctx context.Context, extent, key string) (core.WorkflowSelector, core.WorkflowSelectionGuard, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowSelector{}, core.WorkflowSelectionGuard{}, err
	}

	snapshot, err := store.InspectWorkflowRecords(ctx)
	if err != nil {
		return core.WorkflowSelector{}, core.WorkflowSelectionGuard{}, err
	}

	for _, record := range snapshot.Records {
		if record.Task != nil && record.Task.Key == key {
			selector := core.WorkflowSelector{Extent: extent, Identity: workflowSelectedExtentID(extent, record.Task)}

			return selector, core.WorkflowSelectionGuard{Key: key, Revision: record.Revision}, nil
		}
	}

	return core.WorkflowSelector{}, core.WorkflowSelectionGuard{}, fmt.Errorf("selected ticket has no native workflow identity: %s", key)
}

func workflowSelectedExtentID(extent string, task *core.WorkflowTask) string {
	switch extent {
	case "ticket":
		return task.Key
	case "slice":
		return task.SliceID
	case "slice_set":
		return task.SetID
	case "use_case":
		return task.UseCase
	default:
		return ""
	}
}

// StartBoundWorkExtent requires the displayed snapshot and selected revision.
// Domain admission compares both under exclusion before any attempt can start.
func (s service) StartBoundWorkExtent(ctx context.Context, request core.WorkflowPreviewRequest, selected *core.WorkflowSelectionGuard, steps int, admission ...WorkflowBoundAdmission) (AgentWorkResult, error) {
	s.ctx = ctx

	if len(admission) != 0 {
		return s.admitBoundWorkExtent(ctx, request, selected, steps, admission)
	}

	if len(request.SnapshotHash) != 64 || request.Offset != 0 || steps < 0 || steps > core.WorkflowPageLimit {
		return AgentWorkResult{}, fmt.Errorf("confirmed workflow admission requires a snapshot, offset zero and bounded steps")
	}

	config, err := s.workflowExtentConfig()
	if err != nil {
		return AgentWorkResult{}, err
	}

	if _, err := s.PreviewWorkExtent(ctx, request); err != nil {
		return AgentWorkResult{}, err
	}

	return s.startNativeAgentWork(config, request, steps, "", selected)
}

// WorkflowPage is a bounded read-only dashboard projection of retained work.
type WorkflowPage struct {
	Offset     int                  `json:"offset"`
	Items      []WorkflowWorkStatus `json:"items"`
	NextOffset int                  `json:"next_offset"`
}

// WorkflowStatusPage observes retained invocations without claiming or resuming.
func (s service) WorkflowStatusPage(ctx context.Context, offset, limit int) (result WorkflowPage, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result.Offset = offset

	store, err := s.workflowStore()
	if err != nil {
		return result, err
	}

	runs, next, err := store.InspectWorkflowPage(ctx, offset, limit)
	if err != nil {
		return result, err
	}

	result.Items, result.NextOffset = []WorkflowWorkStatus{}, next

	for _, run := range runs {
		inspection, err := store.InspectWorkflow(ctx, run.ID)
		if err != nil {
			return result, err
		}

		result.Items = append(result.Items, workflowWorkStatus(inspection, WorkflowStatusRequest{Limit: 8}))
	}

	return result, nil
}

// PreviewWorkflow exposes one adapter-neutral, read-only selection operation.
func (s service) PreviewWorkflow(ctx context.Context, request core.WorkflowPreviewRequest) (core.WorkflowPreview, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowPreview{}, err
	}

	return store.Preview(ctx, request)
}

var errWorkflowConfirmationBound = errors.New("complete workflow confirmation exceeds source bounds")

func (s service) workflowConfirmationRecords(ctx context.Context, preview WorkflowBoundPreview) ([]core.WorkflowRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	selection := preview.View.Preview
	if selection.Total != len(selection.Members) || len(selection.Members) > core.WorkflowPageLimit {
		return nil, errWorkflowConfirmationBound
	}

	store, err := s.workflowStore()
	if err != nil {
		return nil, err
	}

	var records []core.WorkflowRecord

	size := 0

	for _, member := range selection.Members {
		record, err := store.Read(ctx, member.Path)
		if err != nil {
			return nil, err
		}

		if !s.workflowConfirmationRecordMatches(record, member) {
			return nil, workflowstate.ErrConflict
		}

		size += len(record.Body) + len(record.Domain["execution_contract"])
		if size > core.WorkflowFileLimit {
			return nil, errWorkflowConfirmationBound
		}

		records = append(records, record)
	}

	return records, nil
}

func (s service) workflowConfirmationRecordMatches(record core.WorkflowRecord, member core.WorkflowPreviewMember) bool {
	return record.Task != nil && record.ProjectID == s.projectContext.ProjectID && record.Namespace == s.workflowNamespace && record.Task.Key == member.Key && record.Revision == member.Revision && record.SourceHash == member.SourceHash && record.Task.ContractHash == member.ContractHash
}
