package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
)

func (m *Model) workflowActionAvailable(action WorkActionID) bool {
	for _, option := range m.workflow.Actions {
		if option.Action == action && option.Reason != "" {
			m.statusMessage = option.Reason
			m.recordOperation(operation.Result{Status: operation.Blocked, Summary: option.Reason})

			return false
		}
	}

	return true
}

func (m *Model) handleWorkflowControlKey(key string) (tea.Cmd, bool) {
	if key == "b" {
		m.openWorkPreviewForm()

		return nil, true
	}

	actions := map[string]WorkActionID{"c": WorkActionWorkflowContinue, "p": WorkActionWorkflowPause, "x": WorkActionWorkflowCancel, "u": WorkActionWorkflowResume, "a": WorkActionWorkflowApprove, "R": WorkActionWorkflowReconcile, "g": WorkActionWorkflowGrant, "d": WorkActionWorkflowAnswer}
	if action, ok := actions[key]; ok {
		m.planWorkAction(action)

		if m.workAction.Pending {
			return m.confirmWorkAction(), true
		}

		return nil, true
	}

	if key == "t" || key == "C" {
		m.openWorkflowLink(key == "C")

		return nil, true
	}

	return nil, false
}

func (m *Model) workflowLinkedTicket() string {
	if m.workflow.TicketKey != "" {
		return m.workflow.TicketKey
	}

	if len(m.workflow.Members) != 0 {
		return m.workflow.Members[0].Key
	}

	return ""
}

func (m *Model) openWorkflowLink(openCase bool) {
	if openCase && m.workflow.Extent == "use_case" {
		m.bindCaseContext(m.workflow.Identity, "")
		m.selectView(ViewCases)

		return
	}

	key := m.workflowLinkedTicket()
	if !openCase && key != "" {
		m.readTicketForm(TicketActionRequest{Action: TicketActionSave, Key: key})

		return
	}

	if key != "" && m.ticketExecutor != nil {
		result, err := m.ticketExecutor.ExecuteTicketAction(TicketActionRequest{Action: TicketActionRead, Key: key})
		if err == nil && result.Editor != nil && result.Editor.CaseSlug != "" {
			m.bindCaseContext(result.Editor.CaseSlug, "")
			m.selectView(ViewCases)

			return
		}
	}

	m.statusMessage = "No linked case or ticket is available on this member page."
}
