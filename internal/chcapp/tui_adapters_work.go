package chcapp

import (
	"errors"

	"forge.adrianpk.com/chc/internal/chctui"
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
	request chctui.WorkActionRequest,
) (chctui.WorkActionResult, error) {
	switch request.Action {
	case chctui.WorkActionStartTicket:
		result, err := e.operations.StartSelectedTicket(request.TicketKey)
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	case chctui.WorkActionStartReadyLoop:
		result, err := e.operations.StartReadyTicketLoop()
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	case chctui.WorkActionRunAllReady:
		result, err := e.operations.StartAllReadyAgentWork()
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(workOperationResultFromAgentWork(result))
	case chctui.WorkActionRequestStop:
		result, err := e.operations.RequestDaemonStop()
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	default:
		return chctui.WorkActionResult{}, errors.New("unknown work action")
	}
}

func (e tuiWorkActionExecutor) workActionResult(
	result WorkOperationResult,
) (chctui.WorkActionResult, error) {
	observability, err := loadTUIObservability(e.operations.ctx, e.operations.env, e.operations.projectContext, e.operations.assetsFS)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	return chctui.WorkActionResult{
		Message:       result.Message,
		Tickets:       result.Tickets,
		Observability: observability,
	}, nil
}
