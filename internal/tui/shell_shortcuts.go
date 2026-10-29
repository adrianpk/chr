package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) handleShellShortcut(message tea.KeyMsg) (tea.Cmd, bool) {
	if m.shellShortcutBlocked() {
		return nil, false
	}

	view, navigation := functionKeyView(message)
	switch message.String() {
	case "?":
		view, navigation = ViewHelp, true
	case "alt+p":
		view, navigation = ViewProject, true
	}

	if !navigation {
		return nil, false
	}

	m.selectView(view)

	return nil, true
}

func (m *Model) shellShortcutBlocked() bool {
	for _, blocked := range []bool{
		m.feedback.Open, m.draftConfirmation.Active, m.textForm.Active,
		m.projectEdit.Active, m.caseEdit.Active, m.caseFilterEditing,
		m.dumpEdit.Active, m.quizAnswer.Active, m.filterEntryMode,
		m.commandPalette.Open, m.harnessSelection.Active,
		m.wizardWriteBusy(), m.specBusy, m.planBusy, m.quizBusy, m.workBusy,
		m.artifactAction.Pending, m.dumpAction.Pending, m.ticketAction.Pending,
		m.specAction.Pending, m.planAction.Pending, m.quizAction.Pending, m.workAction.Pending,
	} {
		if blocked {
			return true
		}
	}

	return false
}
