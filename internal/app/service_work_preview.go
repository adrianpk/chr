package app

import (
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s service) previewNativeAgentWork(config core.WorkspaceConfig, opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	if err := validateAgentWorkSelectionOptions(opts); err != nil {
		return AgentWorkResult{}, err
	}

	if opts.MaxIterations > workflowstate.PageLimit {
		return AgentWorkResult{}, fmt.Errorf("ticket authority max-iterations must contain 0..128 checkpoint steps")
	}

	return s.previewNativeAgentWorkSelection(config, nativeAgentWorkSelection(opts))
}

func (s service) previewNativeAgentWorkSelection(config core.WorkspaceConfig, selection workflowstate.PreviewRequest) (AgentWorkResult, error) {
	selection.Limit = workflowstate.PageLimit

	target, err := s.nativeAgentWorkTarget(config)
	if err != nil {
		return AgentWorkResult{}, err
	}

	selection.Target, err = s.previewWorkTarget(config, target)
	if err != nil {
		return AgentWorkResult{}, err
	}

	preview, err := s.PreviewWorkflow(s.ctx, selection)
	if err != nil {
		return AgentWorkResult{}, err
	}

	store, err := s.workflowStore()
	if err != nil {
		return AgentWorkResult{}, err
	}

	tickets, err := store.WorkflowTickets(s.ctx)
	queue := []ticketcore.Indexed{}

	if len(preview.Members) > 0 {
		for _, ticket := range tickets {
			if ticket.Key == preview.Members[0].Key {
				queue = append(queue, ticket)
			}
		}
	}

	return AgentWorkResult{Scope: selection.Selector.Extent, Authority: core.WorkStateAuthorityTickets, Excluded: preview.Excluded, EligibleCount: preview.Total, Tickets: tickets, SelectedQueue: queue, Message: fmt.Sprintf("Authority tickets; %s selection contains %d eligible tickets; %d slice-owned tickets excluded.", selection.Selector.Extent, preview.Total, preview.Excluded["slice_owned"])}, err
}
