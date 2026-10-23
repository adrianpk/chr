package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
)

const agentWorkRefreshInterval = 2 * time.Second

// AgentWorkRefreshExecutor reloads dashboard agent work state.
type AgentWorkRefreshExecutor interface {
	RefreshAgentWork(AgentWorkRefreshRequest) (AgentWorkRefreshResult, error)
}

// AgentWorkRefreshRequest identifies the displayed retained page and invocation.
type AgentWorkRefreshRequest struct {
	WorkflowID     string
	WorkflowOffset int
}

// AgentWorkRefreshResult contains refreshed agent work rows for the dashboard.
type AgentWorkRefreshResult struct {
	Workflow      *WorkflowView
	Tickets       []core.IndexedTicket
	Observability ObservabilityView
	Readiness     WorkReadinessView
}

type agentWorkRefreshTickMsg struct{}

type agentWorkRefreshFinishedMsg struct {
	result AgentWorkRefreshResult
	err    error
}

func agentWorkRefreshTickCmd() tea.Cmd {
	return tea.Tick(agentWorkRefreshInterval, func(time.Time) tea.Msg {
		return agentWorkRefreshTickMsg{}
	})
}

func (m Model) startAgentWorkRefresh() tea.Cmd {
	if m.agentWorkRefreshExecutor == nil {
		return nil
	}

	request := AgentWorkRefreshRequest{WorkflowID: m.workflow.ID, WorkflowOffset: m.workflowPage.Offset}
	if request.WorkflowID == "" && len(m.observability.WorkflowPage.Items) > 0 {
		request.WorkflowID = m.observability.WorkflowPage.Items[0].ID
	}

	return func() tea.Msg {
		result, err := m.agentWorkRefreshExecutor.RefreshAgentWork(request)

		return agentWorkRefreshFinishedMsg{result: result, err: err}
	}
}

func (m *Model) completeAgentWorkRefresh(msg agentWorkRefreshFinishedMsg) tea.Cmd {
	m.agentWorkRefreshInFlight = false

	if msg.err != nil {
		m.statusMessage = "Agent work refresh failed: " + msg.err.Error()

		return agentWorkRefreshTickCmd()
	}

	if msg.result.Tickets != nil {
		m.tickets = NewTicketViews(msg.result.Tickets)
		m.selectFirstFilteredTicket()
	}

	m.applyWorkflowPage(msg.result.Observability.WorkflowPage)

	if view := msg.result.Workflow; view != nil && view.ID == m.workflow.ID && view.Revision >= m.workflow.Revision {
		m.workflow = *view
	}

	m.observability = msg.result.Observability
	m.workReadiness = msg.result.Readiness

	return agentWorkRefreshTickCmd()
}

func (m Model) handleAgentWorkRefreshTick() (tea.Model, tea.Cmd) {
	if m.guided.Active {
		command := m.refreshGuided()

		return m, tea.Batch(agentWorkRefreshTickCmd(), command)
	}

	if m.SelectedView() != ViewDash || m.agentWorkRefreshExecutor == nil {
		return m, agentWorkRefreshTickCmd()
	}

	if m.workAction.Pending {
		return m, agentWorkRefreshTickCmd()
	}

	if m.agentWorkRefreshInFlight {
		return m, agentWorkRefreshTickCmd()
	}

	m.agentWorkRefreshInFlight = true

	return m, m.startAgentWorkRefresh()
}

func (m Model) updateAgentWorkRefreshMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch typedMsg := msg.(type) {
	case agentWorkRefreshTickMsg:
		updated, cmd := m.handleAgentWorkRefreshTick()

		return updated, cmd, true
	case agentWorkRefreshFinishedMsg:
		cmd := m.completeAgentWorkRefresh(typedMsg)

		return m, cmd, true
	default:
		return m, nil, false
	}
}
