package core

import (
	"context"
	"maps"
	"reflect"
	"slices"
)

// WorkflowLooseQueue retains admission order independently of live pending order.
// Excluded members remain admitted facts; they are never counted as delivered.
type WorkflowLooseQueue struct {
	Admitted []string          `json:"admitted"`
	Cursor   int               `json:"cursor"`
	Eligible int               `json:"eligible"`
	Excluded map[string]string `json:"excluded"`
}

func newWorkflowLooseQueue(members []WorkflowMember) *WorkflowLooseQueue {
	queue := &WorkflowLooseQueue{Eligible: len(members), Excluded: map[string]string{}}
	for _, member := range members {
		queue.Admitted = append(queue.Admitted, member.Admission.Key)
	}

	return queue
}

func validateWorkflowLooseQueue(run workflowInvocationFile) error {
	queue := run.LooseQueue
	if queue == nil {
		return nil
	}

	if !validWorkflowLooseQueueBounds(run) {
		return ErrWorkflowConflict
	}

	keys := map[string]bool{}
	for _, key := range queue.Admitted {
		if key == "" || keys[key] {
			return ErrWorkflowConflict
		}

		keys[key] = true
	}

	for _, member := range run.Members {
		if !keys[member.Admission.Key] || member.Admission.SetID != "" || member.Admission.SliceID != "" {
			return ErrWorkflowConflict
		}
	}

	return validateWorkflowLooseExclusions(queue, keys)
}

func validWorkflowLooseQueueBounds(run workflowInvocationFile) bool {
	queue := run.LooseQueue

	return run.Selection.Selector.Extent == "loose" && len(queue.Admitted) == len(run.Members) && queue.Cursor >= 0 && queue.Cursor <= run.Cursor && queue.Eligible >= 0 && queue.Eligible <= len(run.Members)-queue.Cursor && len(queue.Excluded) <= len(run.Members)
}

func validateWorkflowLooseExclusions(queue *WorkflowLooseQueue, keys map[string]bool) error {
	for key, reason := range queue.Excluded {
		if !keys[key] || !slices.Contains([]string{"status_mismatch", "filter_mismatch", "below_minimum", "integrated"}, reason) {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func (s *WorkflowStore) reconcileWorkflowLooseQueue(ctx context.Context, run workflowInvocationFile) (workflowInvocationFile, error) {
	if workflowInvocationTerminal(run.Status) || run.StopIntent != "" || slices.Contains([]string{"paused", "blocked"}, run.Status) {
		return run, nil
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return run, err
	}

	if run.Lane != nil {
		return run, nil
	}

	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return run, err
	}

	next, err := s.selectWorkflowLooseQueue(ctx, run, selection)
	if err != nil || reflect.DeepEqual(next.WorkflowInvocation, run.WorkflowInvocation) {
		return run, err
	}

	if err := s.commitWorkflowState(&next); err != nil {
		return run, err
	}

	return next, nil
}

func (s *WorkflowStore) selectWorkflowLooseQueue(ctx context.Context, run workflowInvocationFile, selection workflowSelection) (workflowInvocationFile, error) {
	preview, err := selection.preview(run.ExecutionSelection())
	if err != nil {
		return run, err
	}

	next := run
	next.Members = slices.Clone(run.Members)
	next.Excluded = maps.Clone(preview.Excluded)

	next.LooseQueue = newWorkflowLooseQueue(run.Members)
	if run.LooseQueue != nil {
		next.LooseQueue.Admitted = slices.Clone(run.LooseQueue.Admitted)
	}

	if err := s.admitWorkflowLooseLater(ctx, &next, selection); err != nil {
		return run, err
	}

	selectWorkflowLoosePending(&next, selection)

	if next.LooseQueue.Eligible == 0 {
		next.Status, next.Outcome = "completed", "loose_exhausted"
	}

	return next, nil
}

func selectWorkflowLoosePending(run *workflowInvocationFile, selection workflowSelection) {
	tickets := map[string]IndexedTicket{}
	for _, ticket := range selection.tickets {
		tickets[ticket.Key] = ticket
	}

	eligible, excluded := []WorkflowPreviewMember{}, []WorkflowMember{}
	known := map[string]WorkflowMember{}

	for _, member := range run.Members[run.Cursor:] {
		key := member.Admission.Key

		reason := workflowPreviewExclusion(run.Selection, tickets[key], selection.records[key], true)
		if reason == "" {
			eligible = append(eligible, member.Admission)
			known[key] = member
		} else {
			excluded = append(excluded, member)
			run.LooseQueue.Excluded[key] = reason
		}
	}

	selection.orderLooseMembers(eligible, run.Selection.Order)

	run.Members = slices.Clone(run.Members[:run.Cursor])
	for _, member := range eligible {
		run.Members = append(run.Members, known[member.Key])
	}

	run.Members = append(run.Members, excluded...)
	run.LooseQueue.Cursor = run.Cursor
	run.LooseQueue.Eligible = len(eligible)
}
