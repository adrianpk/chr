package chcapp

import (
	"forge.adrianpk.com/chc/internal/chctui"
)

type tuiAgentWorkRefreshExecutor struct {
	operations service
}

func newTUIAgentWorkRefreshExecutor(operations service) tuiAgentWorkRefreshExecutor {
	return tuiAgentWorkRefreshExecutor{operations: operations}
}

func (e tuiAgentWorkRefreshExecutor) RefreshAgentWork() (chctui.AgentWorkRefreshResult, error) {
	tickets, err := e.operations.listProjectTickets()
	if err != nil {
		return chctui.AgentWorkRefreshResult{}, err
	}

	observability, err := loadTUIObservability(e.operations.ctx, e.operations.env, e.operations.projectContext, e.operations.assetsFS)
	if err != nil {
		return chctui.AgentWorkRefreshResult{}, err
	}

	return chctui.AgentWorkRefreshResult{
		Tickets:       tickets,
		Observability: observability,
	}, nil
}
