package app

import (
	"errors"

	"forge.adrianpk.com/chc/internal/tui"
	"forge.adrianpk.com/chc/internal/workspace"
)

type tuiWorkActionExecutor struct {
	operations service
}

func newTUIWorkActionExecutor(
	operations service,
) tuiWorkActionExecutor {
	return tuiWorkActionExecutor{
		operations: operations,
	}
}

func (e tuiWorkActionExecutor) ExecuteWorkAction(
	request tui.WorkActionRequest,
) (tui.WorkActionResult, error) {
	switch request.Action {
	case tui.WorkActionStartTicket:
		result, err := e.operations.StartSelectedTicket(request.TicketKey)
		if err != nil {
			return tui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	case tui.WorkActionStartReadyLoop:
		result, err := e.operations.StartReadyTicketLoop()
		if err != nil {
			return tui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	case tui.WorkActionRunAllReady:
		result, err := e.operations.StartAgentWorkLoop(agentWorkSelectionOptionsFromTUIRequest(request))
		if err != nil {
			return tui.WorkActionResult{}, err
		}

		return e.workActionResult(workOperationResultFromAgentWork(result))
	case tui.WorkActionRequestStop:
		result, err := e.operations.RequestDaemonStop()
		if err != nil {
			return tui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	default:
		return tui.WorkActionResult{}, errors.New("unknown work action")
	}
}

func agentWorkSelectionOptionsFromTUIRequest(request tui.WorkActionRequest) AgentWorkSelectionOptions {
	opts := AgentWorkSelectionOptions{
		Statuses: []string{workspace.TicketStatusReadyForDev},
		Order:    request.Order,
	}
	if request.PriorityFilter != "" {
		opts.Filters.Priorities = []string{request.PriorityFilter}
	}

	return opts
}

func (e tuiWorkActionExecutor) workActionResult(
	result WorkOperationResult,
) (tui.WorkActionResult, error) {
	observability, err := loadTUIObservability(e.operations.ctx, e.operations.env, e.operations.projectContext, e.operations.assetsFS)
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	return tui.WorkActionResult{
		Message:       result.Message,
		Tickets:       result.Tickets,
		Observability: observability,
	}, nil
}
