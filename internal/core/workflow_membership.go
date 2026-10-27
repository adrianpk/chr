package core

import (
	"context"
	"fmt"
	"maps"
	"slices"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReconcileWorkflowMembership reports later ready work under the frozen selector.
// Only the policy accepted at StartWorkflow can admit it. Reconciliation cannot
// reopen a terminal invocation or a completed set, or change an active unit.
// In loose scope it also reselects inactive admissions using current metadata.
func (s *WorkflowStore) ReconcileWorkflowMembership(ctx context.Context, id string) (result WorkflowInvocation, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		result = run.WorkflowInvocation
		run, readErr = owner.reconcileWorkflowMembership(ctx, run)
		result = run.WorkflowInvocation

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) reconcileWorkflowMembership(ctx context.Context, run workflowInvocationFile) (workflowInvocationFile, error) {
	if run.Selection.Selector.Extent == "loose" {
		return s.reconcileWorkflowLooseQueue(ctx, run)
	}

	if workflowMembershipDormant(run) {
		return run, nil
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return run, err
	}

	if run.Lane != nil || run.ClosingSet != "" {
		return run, nil
	}

	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return run, err
	}

	if selection.setOrderRule(run.Selection.Selector) != run.SetOrderRule {
		return run, fmt.Errorf("%w: frozen set ordering rule changed", workflowstate.ErrConflict)
	}

	return s.retainWorkflowLaterMembers(ctx, run, selection)
}

func workflowMembershipDormant(run workflowInvocationFile) bool {
	return workflowInvocationTerminal(run.Status) || run.StopIntent != "" || run.Status == "paused" || slices.Contains([]string{"ticket", "loose"}, run.Selection.Selector.Extent)
}

func (s *WorkflowStore) retainWorkflowLaterMembers(ctx context.Context, run workflowInvocationFile, selection workflowSelection) (workflowInvocationFile, error) {
	later := workflowLaterMembers(selection, run)
	if len(later) == 0 && run.Excluded["later"] == 0 {
		return run, nil
	}

	if !run.Policy.IncludeLaterTickets {
		return s.retainWorkflowLaterCount(run, len(later))
	}

	if len(later) == 0 {
		return s.retainWorkflowLaterCount(run, 0)
	}

	next := run
	next.Members = append([]WorkflowMember(nil), run.Members...)
	next.SetImages = append([]WorkflowInputImage(nil), run.SetImages...)

	next.Excluded = maps.Clone(run.Excluded)
	if err := s.admitWorkflowLaterMembers(ctx, &next, selection, later); err != nil {
		return run, err
	}

	if err := s.commitWorkflowState(&next); err != nil {
		return run, err
	}

	return next, nil
}

func workflowLaterMembers(selection workflowSelection, run workflowInvocationFile) []WorkflowPreviewMember {
	known := map[string]bool{}
	for _, member := range run.Members {
		known[member.Admission.Key] = true
	}

	later := []WorkflowPreviewMember{}

	for _, ticket := range selection.tickets {
		record := selection.records[ticket.Key]
		if known[ticket.Key] || record.Task == nil || record.Task.SetID == "" || !selection.matches(run.Selection.Selector, ticket, record) {
			continue
		}

		if record.Task.Status != string(ticketcore.StatusReadyForDev) || record.Task.Phase != "pending" || selection.sets[record.Task.SetID].Set.Status == "completed" {
			continue
		}

		later = append(later, workflowPreviewMember(record))
	}

	return later
}

func (s *WorkflowStore) retainWorkflowLaterCount(run workflowInvocationFile, count int) (workflowInvocationFile, error) {
	if run.Excluded["later"] == count {
		return run, nil
	}

	if run.Excluded == nil {
		run.Excluded = map[string]int{}
	}

	delete(run.Excluded, "later")

	if count > 0 {
		run.Excluded["later"] = count
	}

	err := s.commitWorkflowState(&run)

	return run, err
}
