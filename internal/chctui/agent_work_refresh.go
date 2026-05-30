package chctui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chc/internal/chc"
)

const agentWorkRefreshInterval = 2 * time.Second

type AgentWorkRefreshExecutor interface {
	RefreshAgentWork() (AgentWorkRefreshResult, error)
}

type AgentWorkRefreshResult struct {
	Tickets       []chc.IndexedTicket
	Observability ObservabilityView
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

	return func() tea.Msg {
		result, err := m.agentWorkRefreshExecutor.RefreshAgentWork()

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

	m.observability = msg.result.Observability

	return agentWorkRefreshTickCmd()
}

func (m Model) handleAgentWorkRefreshTick() (tea.Model, tea.Cmd) {
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
