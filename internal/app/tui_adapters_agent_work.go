package app

import (
	"forge.adrianpk.com/chr/internal/tui"
)

type tuiAgentWorkRefreshExecutor struct {
	operations service
}

func newTUIAgentWorkRefreshExecutor(operations service) tuiAgentWorkRefreshExecutor {
	return tuiAgentWorkRefreshExecutor{operations: operations}
}

func (e tuiAgentWorkRefreshExecutor) RefreshAgentWork(request tui.AgentWorkRefreshRequest) (tui.AgentWorkRefreshResult, error) {
	tickets, err := loadTUITicketSnapshot(e.operations)
	if err != nil {
		return tui.AgentWorkRefreshResult{}, err
	}

	observability, err := loadTUIObservability(e.operations.ctx, e.operations.env, e.operations.projectContext, e.operations.assetsFS)
	if err != nil {
		return tui.AgentWorkRefreshResult{}, err
	}

	observation, err := e.operations.ObserveWorkPage(e.operations.ctx, request.Queue)
	if err != nil {
		return tui.AgentWorkRefreshResult{}, err
	}

	workflow, err := e.refreshNativeWorkflow(request, &observability)
	if err != nil {
		return tui.AgentWorkRefreshResult{}, err
	}

	return tui.AgentWorkRefreshResult{
		Workflow:      workflow,
		Tickets:       tickets,
		Observability: observability,
		Readiness:     tui.NewWorkReadinessView(observation.Readiness),
		Observation:   &observation,
	}, nil
}

func (e tuiAgentWorkRefreshExecutor) refreshNativeWorkflow(request tui.AgentWorkRefreshRequest, observability *tui.ObservabilityView) (*tui.WorkflowView, error) {
	if request.WorkflowOffset != 0 {
		page, err := e.operations.WorkflowStatusPage(e.operations.ctx, request.WorkflowOffset, 16)
		if err != nil {
			return nil, err
		}

		observability.WorkflowPage = tuiWorkflowPage(page)
	}

	if request.WorkflowID == "" {
		return nil, nil
	}

	status, err := e.operations.WorkflowStatus(e.operations.ctx, request.WorkflowID, WorkflowStatusRequest{Limit: 8, Offset: request.MemberOffset, ExpectedRevision: request.MemberRevision})
	if err != nil {
		return nil, err
	}

	view := tuiWorkflowView(status)
	view.Offset = request.MemberOffset

	return &view, nil
}
