package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (m *Model) handleDashKey(key string) (tea.Cmd, bool) {
	if m.SelectedView() != ViewDash {
		return nil, false
	}

	if m.dashWorkflowDetails {
		if m.observability.Authority != workspaceconfig.WorkStateAuthorityTickets {
			m.dashWorkflowDetails = false
			m.statusMessage = "Native Preview Work and Workflow details require tickets authority."

			return nil, true
		}

		return m.handleWorkflowBrowseKey(key)
	}

	if command, handled := m.handleDashQueryKey(key); handled {
		return command, true
	}

	switch key {
	case "b":
		m.openWorkPreviewForm()

		return nil, true
	case "v":
		return m.inspectDashWork(WorkActionWorkflowTicket), true
	case "w":
		if m.observability.Authority != workspaceconfig.WorkStateAuthorityTickets {
			return m.inspectDashWork(WorkActionWorkflowStatus), true
		}

		m.dashWorkflowDetails = true
		m.selectedPane, m.detailScroll = 0, 0

		return m.readWorkflowPage(m.workflowPage.Offset), true
	default:
		return nil, m.handleDashActionKey(key)
	}
}

func (m *Model) inspectDashWork(action WorkActionID) tea.Cmd {
	if m.observability.Authority != workspaceconfig.WorkStateAuthorityTickets {
		m.statusMessage = "Native Preview Work and Workflow details require tickets authority."

		return nil
	}

	if m.queuePending {
		m.statusMessage = "Wait for the current queue observation."

		return nil
	}

	m.planWorkAction(action)

	if !m.workAction.Pending {
		return nil
	}
	// The first request is a read-only preview/status. Only its returned exact
	// snapshot and policy can create the separate executable confirmation.
	return m.confirmWorkAction()
}

func (m *Model) dashCurrentLines(limit int) []string {
	lines := m.renderDashReadinessLines(limit)

	lines = append(lines, "")
	if m.dashWorkflowDetails {
		return append(append(lines, m.nativeWorkflowLines()...), m.renderAgentWorkCurrentLines()...)
	}

	return append(append(lines, m.renderAgentWorkCurrentLines()...), m.nativeWorkflowLines()...)
}

func (m *Model) dashQueueActionLines() []string {
	lines := []string{"a Run all ready tickets", "enter Start selected ticket"}
	if m.workObservation == nil {
		return lines
	}

	if m.queuePending {
		return []string{"Wait for queue observation."}
	}

	if m.workObservation.Eligible == 0 {
		lines[0] = "a No eligible work"
	}

	ticket, _ := m.selectedDashTicketView()
	if item, ok := m.observedQueueItem(ticket.Key); !ok || item.Readiness != "ready" {
		lines[1] = "enter Selected work unavailable"
	}

	return lines
}
