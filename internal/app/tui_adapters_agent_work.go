package app

import (
	"forge.adrianpk.com/chc/internal/tui"
)

type tuiAgentWorkRefreshExecutor struct {
	operations service
}

func newTUIAgentWorkRefreshExecutor(operations service) tuiAgentWorkRefreshExecutor {
	return tuiAgentWorkRefreshExecutor{operations: operations}
}

func (e tuiAgentWorkRefreshExecutor) RefreshAgentWork() (tui.AgentWorkRefreshResult, error) {
	tickets, err := e.operations.listProjectTickets()
	if err != nil {
		return tui.AgentWorkRefreshResult{}, err
	}

	observability, err := loadTUIObservability(e.operations.ctx, e.operations.env, e.operations.projectContext, e.operations.assetsFS)
	if err != nil {
		return tui.AgentWorkRefreshResult{}, err
	}

	return tui.AgentWorkRefreshResult{
		Tickets:       tickets,
		Observability: observability,
	}, nil
}
