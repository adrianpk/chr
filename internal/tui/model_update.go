package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateSystemMessage(msg tea.Msg) updateOutcome {
	if finished, ok := msg.(guidedFinishedMsg); ok {
		m.completeGuided(finished)

		return updateOutcome{Command: nil, Handled: true}
	}

	if outcome := m.updateAgentWorkRefreshMessage(msg); outcome.Handled {
		return outcome
	}

	if outcome := m.updateActionFinishedMessage(msg); outcome.Handled {
		return outcome
	}

	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typedMsg.Width
		m.height = typedMsg.Height

		return updateOutcome{Command: nil, Handled: true}
	case tea.MouseMsg:
		cmd := m.updateMouse(typedMsg)

		return updateOutcome{Command: cmd, Handled: true}
	case dumpDraftEditorFinishedMsg:
		m.completeDumpDraftEditor(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	case dumpEditorFinishedMsg:
		m.completeDumpEdit(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	case sourceEditorFinishedMsg:
		m.completeSourceEdit(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	default:
		return updateOutcome{Command: nil, Handled: false}
	}
}

func (m *Model) updateActionFinishedMessage(msg tea.Msg) updateOutcome {
	switch typedMsg := msg.(type) {
	case harnessSelectionSavedMsg:
		m.completeHarnessSave(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	case harnessSelectionLoadedMsg:
		m.completeHarnessRead(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	case quizActionFinishedMsg:
		m.completeQuizAction(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	case specActionFinishedMsg:
		m.completeSpecAction(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	case planActionFinishedMsg:
		m.completePlanAction(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	case workActionFinishedMsg:
		m.completeWorkAction(typedMsg)

		return updateOutcome{Command: nil, Handled: true}
	default:
		return updateOutcome{Command: nil, Handled: false}
	}
}

func (m *Model) updateKey(msg tea.KeyMsg) tea.Cmd {
	if msg.Type == tea.KeyCtrlC {
		return tea.Quit
	}

	if command, handled := m.updateFeedbackKey(msg); handled {
		return command
	}

	if command, handled := m.handleShellShortcut(msg); handled {
		return command
	}

	if outcome := m.updateGuidedOrHarnessKey(msg); outcome.Handled {
		return outcome.Command
	}

	if outcome := m.updateActiveModeKey(msg); outcome.Handled {
		return outcome.Command
	}

	if cmd, handled := m.handleGlobalKey(msg); handled {
		return cmd
	}

	if handler, ok := viewKeyHandlers[m.SelectedView()]; ok {
		return handler(m, msg)
	}

	return m.updateNavigationKey(msg)
}

func (m *Model) updateActiveModeKey(msg tea.KeyMsg) updateOutcome {
	modes := []keyMode{
		{Active: m.textForm.Active, Handle: (*Model).updateTextFormKey},
		{Active: m.artifactAction.Pending, Handle: (*Model).updateArtifactActionKey},
		{Active: m.dumpAction.Pending, Handle: (*Model).updateDumpActionKey},
		{Active: m.dumpEdit.Active, Handle: (*Model).updateDumpEditKey},
		{Active: m.ticketAction.Pending, Handle: (*Model).updateTicketActionKey},
		{Active: m.specAction.Pending, Handle: (*Model).updateSpecActionKey},
		{Active: m.specBusy},
		{Active: m.planAction.Pending, Handle: (*Model).updatePlanActionKey},
		{Active: m.planBusy, Message: "Plan action already running."},
		{Active: m.workBusy, Message: "Work action already running."},
		{Active: m.workAction.Pending, Handle: (*Model).updateWorkActionKey},
		{Active: m.commandPalette.Open, Handle: (*Model).updateCommandPaletteKey},
		{Active: m.filterEntryMode, Handle: (*Model).updateFilterKey},
	}
	if outcome := m.updateModeKey(msg, modes); outcome.Handled {
		return outcome
	}

	return m.updateActiveQuizModeKey(msg)
}

func (m *Model) updateActiveQuizModeKey(msg tea.KeyMsg) updateOutcome {
	return m.updateModeKey(msg, []keyMode{
		{Active: m.quizAction.Pending, Handle: (*Model).updateQuizActionKey},
		{Active: m.quizBusy},
		{Active: m.quizAnswer.Active, Handle: (*Model).updateQuizAnswerKey},
	})
}

func (m *Model) handleGlobalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "alt+m":
		m.toggleMaximizedPane()

		return nil, true
	case "alt+e":
		if m.wizardBound() {
			m.explainWizardMode()

			return nil, true
		}

		if m.SelectedView() != ViewDumps {
			return m.openSelectedSource(), true
		}

		m.planDumpAction(DumpActionExternalEdit)

		return m.confirmDumpAction(), true
	default:
		return nil, false
	}
}
