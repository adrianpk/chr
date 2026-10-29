package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (m *Model) changeDashQuery() {
	m.queueSequence++
	m.queueRequest.Priority, m.queueRequest.Order = m.agentWorkPriorityFilter, m.agentWorkOrder
	m.queueRequest.Offset = 0
	m.queuePending = m.workObservation != nil
}

func (m *Model) handleDashQueryKey(key string) (tea.Cmd, bool) {
	if m.SelectedView() != ViewDash {
		return nil, false
	}

	switch key {
	case "p":
		m.cycleAgentWorkPriorityFilter()
	case "o":
		m.cycleAgentWorkOrder()
	case "[":
		m.setDashQueueOffset(max(0, m.queueRequest.Offset-workflowstate.PageLimit))
	case "]":
		if m.workObservation == nil || m.workObservation.QueueNext < 0 || m.queuePending {
			return nil, true
		}

		m.setDashQueueOffset(m.workObservation.QueueNext)
	default:
		return nil, false
	}

	m.queuePending = m.workObservation != nil

	return m.requestDashRefresh(), true
}

func (m *Model) requestDashRefresh() tea.Cmd {
	if m.agentWorkRefreshExecutor == nil || m.agentWorkRefreshInFlight {
		return nil
	}

	m.agentWorkRefreshInFlight = true

	return m.startAgentWorkRefresh()
}

func (m *Model) setDashQueueOffset(offset int) {
	m.queueSequence++
	m.queueRequest.Offset = offset
}
