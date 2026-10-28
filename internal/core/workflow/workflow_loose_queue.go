package workflow

import (
	"context"
	"maps"
	"reflect"
	"slices"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func newWorkflowLooseQueue(members []workflowstate.Member) *workflowstate.LooseQueue {
	queue := &workflowstate.LooseQueue{Eligible: len(members), Excluded: map[string]string{}}
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
		return workflowstate.ErrConflict
	}

	keys := map[string]bool{}
	for _, key := range queue.Admitted {
		if key == "" || keys[key] {
			return workflowstate.ErrConflict
		}

		keys[key] = true
	}

	for _, member := range run.Members {
		if !keys[member.Admission.Key] || member.Admission.SetID != "" || member.Admission.SliceID != "" {
			return workflowstate.ErrConflict
		}
	}

	return validateWorkflowLooseExclusions(queue, keys)
}

func validWorkflowLooseQueueBounds(run workflowInvocationFile) bool {
	queue := run.LooseQueue

	return run.Selection.Selector.Extent == "loose" && len(queue.Admitted) == len(run.Members) && queue.Cursor >= 0 && queue.Cursor <= run.Cursor && queue.Eligible >= 0 && queue.Eligible <= len(run.Members)-queue.Cursor && len(queue.Excluded) <= len(run.Members)
}

func validateWorkflowLooseExclusions(queue *workflowstate.LooseQueue, keys map[string]bool) error {
	for key, reason := range queue.Excluded {
		if !keys[key] || !slices.Contains([]string{"status_mismatch", "filter_mismatch", "below_minimum", "integrated"}, reason) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func (s *Store) reconcileWorkflowLooseQueue(ctx context.Context, run workflowInvocationFile) (workflowInvocationFile, error) {
	if workflowstate.InvocationTerminal(run.Status) || run.StopIntent != "" || slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationPaused, workflowstate.InvocationBlocked}, run.Status) {
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

	next, err := (workflowWorkObserver{store: s}).selectLoose(ctx, run, selection)
	if err != nil || reflect.DeepEqual(next.Invocation, run.Invocation) {
		return run, err
	}

	if err := s.commitWorkflowState(&next); err != nil {
		return run, err
	}

	return next, nil
}

func (host workflowWorkObserver) selectLoose(ctx context.Context, run workflowInvocationFile, selection workflowSelection) (workflowInvocationFile, error) {
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

	if err := host.store.admitWorkflowLooseLater(ctx, &next, selection); err != nil {
		return run, err
	}

	selectWorkflowLoosePending(&next, selection)

	if next.LooseQueue.Eligible == 0 {
		next.Status, next.Outcome = workflowstate.InvocationCompleted, "loose_exhausted"
	}

	return next, nil
}

func selectWorkflowLoosePending(run *workflowInvocationFile, selection workflowSelection) {
	tickets := map[string]ticketcore.Indexed{}
	for _, ticket := range selection.tickets {
		tickets[ticket.Key] = ticket
	}

	eligible, excluded := []workflowstate.PreviewMember{}, []workflowstate.Member{}
	known := map[string]workflowstate.Member{}

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
