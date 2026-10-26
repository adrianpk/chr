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
	if outcome := m.updateGuidedOrHarnessKey(msg); outcome.Handled {
		return outcome.Command
	}

	if outcome := m.updateActiveModeKey(msg); outcome.Handled {
		return outcome.Command
	}

	if cmd, handled := m.handleGlobalKey(msg); handled {
		return cmd
	}

	if m.SelectedView() == ViewProject {
		return m.updateProjectKey(msg)
	}

	if m.SelectedView() == ViewCases {
		return m.updateCasesKey(msg)
	}

	if m.SelectedView() == ViewDumps {
		return m.updateDumpsKey(msg)
	}

	if m.SelectedView() == ViewQuizzes {
		return m.updateQuizzesKey(msg)
	}

	if m.SelectedView() == ViewSpecs {
		return m.updateSpecsKey(msg)
	}

	if m.SelectedView() == ViewPlans {
		return m.updatePlansKey(msg)
	}

	return m.updateNavigationKey(msg)
}

//nolint:cyclop // Active mode dispatch is centralized to preserve modal ordering.
func (m *Model) updateActiveModeKey(msg tea.KeyMsg) updateOutcome {
	switch {
	case m.artifactAction.Pending:
		cmd := m.updateArtifactActionKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.dumpAction.Pending:
		cmd := m.updateDumpActionKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.dumpEdit.Active:
		cmd := m.updateDumpEditKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.ticketAction.Pending:
		cmd := m.updateTicketActionKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.specAction.Pending:
		cmd := m.updateSpecActionKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.specBusy:
		return updateOutcome{Command: nil, Handled: true}
	case m.planAction.Pending:
		cmd := m.updatePlanActionKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.planBusy:
		m.statusMessage = "Plan action already running."

		return updateOutcome{Command: nil, Handled: true}
	case m.workBusy:
		m.statusMessage = "Work action already running."

		return updateOutcome{Command: nil, Handled: true}
	case m.workAction.Pending:
		cmd := m.updateWorkActionKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.commandPalette.Open:
		cmd := m.updateCommandPaletteKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.filterEntryMode:
		cmd := m.updateFilterKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	default:
		return m.updateActiveQuizModeKey(msg)
	}
}

func (m *Model) updateActiveQuizModeKey(msg tea.KeyMsg) updateOutcome {
	switch {
	case m.quizAction.Pending:
		cmd := m.updateQuizActionKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	case m.quizBusy:
		return updateOutcome{Command: nil, Handled: true}
	case m.quizAnswer.Active:
		cmd := m.updateQuizAnswerKey(msg)

		return updateOutcome{Command: cmd, Handled: true}
	default:
		return updateOutcome{Command: nil, Handled: false}
	}
}

func (m *Model) handleGlobalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "alt+m":
		m.toggleMaximizedPane()

		return nil, true
	case "alt+e":
		if m.SelectedView() != ViewDumps {
			return m.openSelectedSource(), true
		}

		m.planDumpAction(DumpActionExternalEdit)

		return m.confirmDumpAction(), true
	default:
		return nil, false
	}
}
