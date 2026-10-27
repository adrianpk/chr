package core

import (
	"context"
	"fmt"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) admitWorkflowLooseLater(ctx context.Context, run *workflowInvocationFile, selection workflowSelection) error {
	later := workflowLooseLaterMembers(*run, selection)
	delete(run.Excluded, "later")

	if !run.Policy.IncludeLaterTickets {
		if len(later) != 0 {
			run.Excluded["later"] = len(later)
		}

		return nil
	}

	if len(run.Members)+len(later) > workflowstate.RecordLimit {
		return fmt.Errorf("workflow loose admission member bound exceeded")
	}

	if len(later) == 0 {
		return nil
	}

	selection.orderLooseMembers(later, run.Selection.Order)

	preview := WorkflowPreview{Members: later, admitted: workflowAdmittedRefs(append(workflowLooseAdmissions(run.Members), later...), run.Selection.Selector)}
	if err := s.observePrerequisites(ctx, selection, run.ExecutionSelection(), &preview); err != nil {
		return err
	}

	if err := s.checkWorkflowLaterOwnership(ctx, *run, preview.Members); err != nil {
		return err
	}

	run.Members = append(run.Members, workflowNewMembers(preview.Members)...)
	for _, member := range preview.Members {
		run.LooseQueue.Admitted = append(run.LooseQueue.Admitted, member.Key)
	}

	return nil
}

func workflowLooseLaterMembers(run workflowInvocationFile, selection workflowSelection) []WorkflowPreviewMember {
	known := map[string]bool{}
	for _, member := range run.Members {
		known[member.Admission.Key] = true
	}

	later := []WorkflowPreviewMember{}

	for _, ticket := range selection.tickets {
		record, prepared := selection.records[ticket.Key]
		if known[ticket.Key] || workflowPreviewExclusion(run.Selection, ticket, record, prepared) != "" {
			continue
		}

		if record.Task.Status == ticketcore.StatusReadyForDev && record.Task.Phase == "pending" {
			later = append(later, workflowPreviewMember(record))
		}
	}

	return later
}

func workflowLooseAdmissions(members []WorkflowMember) []WorkflowPreviewMember {
	result := make([]WorkflowPreviewMember, 0, len(members))
	for _, member := range members {
		result = append(result, member.Admission)
	}

	return result
}
