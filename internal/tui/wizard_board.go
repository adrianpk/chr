package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) wizardBoardCurrent() bool {
	if !m.wizardBound() || m.SelectedView() != ViewBoard {
		return false
	}

	stage := m.guided.View.Session.Guided.Stage

	return (stage == "execution" || stage == "results") && m.guided.View.Document == nil
}

func (m *Model) observeGuidedBoard(view GuidedView) {
	first := view.Native != nil && view.Native.ID != m.workflow.ID
	if view.Native != nil {
		m.workflow = *view.Native
	}

	if view.Tickets == nil {
		return
	}

	m.replaceTickets(view.Tickets)

	if first {
		m.boardColumn = defaultBoardColumn(m.caseTickets())
		m.selectFirstFilteredTicket()
	}
}

func (m *Model) updateWizardComponentKey(message tea.KeyMsg) (tea.Cmd, bool) {
	if command, handled := m.updateWizardQuizKey(message); handled {
		return command, true
	}

	return m.updateWizardBoardKey(message)
}

func (m *Model) updateWizardBoardKey(message tea.KeyMsg) (tea.Cmd, bool) {
	if !m.wizardBoardCurrent() {
		return nil, false
	}

	key := message.String()
	switch key {
	case "up", "down", "j", "k", "left", "right", "h", "l", "tab", "shift+tab", "pgup", "pgdown", "alt+a", "alt+m":
		return m.updateNavigationKey(message), true
	case "/":
		m.filterEntryMode = true

		return nil, true
	case "R", "g", "d":
		action := map[string]WorkActionID{"R": WorkActionWorkflowReconcile, "g": WorkActionWorkflowGrant, "d": WorkActionWorkflowAnswer}[key]
		if !m.wizardNativeActionAllowed(action) {
			m.statusMessage = "This native control is unavailable in the retained state."

			return nil, true
		}

		return m.handleWorkflowControlKey(key)
	default:
		return nil, false
	}
}

func (m *Model) wizardNativeActionAllowed(action WorkActionID) bool {
	if m.guided.View.Native == nil {
		return false
	}

	for _, option := range m.guided.View.Native.Actions {
		if option.Action == action {
			return option.Reason == ""
		}
	}

	return false
}

func (m *Model) wizardBoardNativeHint() string {
	if !m.wizardBoardCurrent() {
		return ""
	}

	hint := ""

	for _, item := range []struct {
		action WorkActionID
		label  string
	}{{WorkActionWorkflowReconcile, "R Reconcile"}, {WorkActionWorkflowGrant, "g Allowance"}, {WorkActionWorkflowAnswer, "d Decision"}} {
		if m.wizardNativeActionAllowed(item.action) {
			hint += item.label + " · "
		}
	}

	return hint
}

func (m *Model) wizardInteractionActive() bool {
	for _, active := range []bool{m.textForm.Active, m.dumpEdit.Active, m.quizAnswer.Active, m.quizAction.Pending, m.workAction.Pending, m.filterEntryMode, m.commandPalette.Open} {
		if active {
			return true
		}
	}

	return false
}
