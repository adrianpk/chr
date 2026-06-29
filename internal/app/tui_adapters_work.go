package app

import (
	"errors"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/tui"
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
		return e.startSelectedTicketResult(request)
	case tui.WorkActionStartReadyLoop:
		return e.startReadyLoopResult()
	case tui.WorkActionRunAllReady:
		return e.runAllReadyResult(request)
	case tui.WorkActionRepairReadiness:
		return e.repairReadinessResult()
	case tui.WorkActionRequestStop:
		return e.requestStopResult()
	default:
		return tui.WorkActionResult{}, errors.New("unknown work action")
	}
}

func (e tuiWorkActionExecutor) startSelectedTicketResult(request tui.WorkActionRequest) (tui.WorkActionResult, error) {
	result, err := e.operations.StartSelectedTicket(request.TicketKey)
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	return e.workActionResult(result)
}

func (e tuiWorkActionExecutor) startReadyLoopResult() (tui.WorkActionResult, error) {
	result, err := e.operations.StartReadyTicketLoop()
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	return e.workActionResult(result)
}

func (e tuiWorkActionExecutor) runAllReadyResult(request tui.WorkActionRequest) (tui.WorkActionResult, error) {
	result, err := e.operations.StartAgentWorkLoop(agentWorkSelectionOptionsFromTUIRequest(request))
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	return e.workActionResult(workOperationResultFromAgentWork(result))
}

func (e tuiWorkActionExecutor) requestStopResult() (tui.WorkActionResult, error) {
	result, err := e.operations.RequestDaemonStop()
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	return e.workActionResult(result)
}

func agentWorkSelectionOptionsFromTUIRequest(request tui.WorkActionRequest) AgentWorkSelectionOptions {
	opts := AgentWorkSelectionOptions{
		Statuses: []string{core.TicketStatusReadyForDev},
		Order:    request.Order,
	}
	if request.PriorityFilter != "" {
		opts.Filters.Priorities = []string{request.PriorityFilter}
	}

	return opts
}

func (e tuiWorkActionExecutor) repairReadinessResult() (tui.WorkActionResult, error) {
	workConfig, err := workSelectionWorkConfig(e.operations.projectContext.ProjectDir)
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	result, err := core.RepairWorkReadiness(e.operations.ctx, core.WorkReadinessRepairOptions{
		ProjectDir: e.operations.projectContext.ProjectDir,
		Policy:     workConfig.BranchPolicy,
		Workspace:  workConfig.Workspace,
	})
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	workResult, err := e.operations.reloadWorkOperationState()
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	message := "No safe readiness repair was applied."
	if len(result.Applied) > 0 {
		message = "Safe readiness repair applied. Retry the work action."
	}

	workResult.Message = message

	return e.workActionResultWithReadiness(workResult, tui.NewWorkReadinessRepairView(result))
}

func (e tuiWorkActionExecutor) workActionResult(
	result WorkOperationResult,
) (tui.WorkActionResult, error) {
	readiness, err := loadTUIWorkReadiness(e.operations.ctx, e.operations.projectContext)
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	return e.workActionResultWithReadiness(result, readiness)
}

func (e tuiWorkActionExecutor) workActionResultWithReadiness(result WorkOperationResult, readiness tui.WorkReadinessView) (tui.WorkActionResult, error) {
	observability, err := loadTUIObservability(e.operations.ctx, e.operations.env, e.operations.projectContext, e.operations.assetsFS)
	if err != nil {
		return tui.WorkActionResult{}, err
	}

	return tui.WorkActionResult{
		Message:       result.Message,
		Tickets:       result.Tickets,
		Observability: observability,
		Readiness:     readiness,
	}, nil
}
