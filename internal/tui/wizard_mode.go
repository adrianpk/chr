package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) explainWizardMode() {
	m.statusMessage = "Read-only wizard stage. Press c to continue at the current stage; clear case context for manual work."
}

func (m *Model) wizardManualCommand(command Command) bool {
	return m.wizardBound() && (strings.HasPrefix(string(command.Action), "plan_") || command.Action == CommandActionOpenSelectedSource)
}

func (m *Model) wizardReadKey(key string) bool {
	return key == ":" || key == "tab" || key == "shift+tab" || key == "esc" || key == "alt+m"
}

func (m *Model) wizardCaseKey(key string) bool {
	if m.SelectedView() != ViewCases {
		return false
	}

	switch key {
	case "up", "down", "j", "k", "/", "o", "w", "g":
		return true
	case "enter":
		if metadata, ok := m.selectedCaseMetadata(); ok && metadata.Slug != m.caseContext.CaseSlug {
			return true
		}
	}

	return false
}

func (m *Model) wizardActionKey(key string) (tea.Cmd, bool) {
	if m.wizardReadKey(key) || m.wizardCaseKey(key) {
		return nil, false
	}

	if m.SelectedView() != m.wizardStageView() {
		m.explainWizardMode()

		return nil, true
	}

	return m.updateWizardActionKey(key)
}

func (m *Model) wizardRequestCurrent(request GuidedRequest) bool {
	if request.Action != "input" {
		return true
	}

	if !m.wizardBound() || request.Command.RunID != m.caseContext.RunID {
		m.statusMessage = "The wizard case or run changed. Reopen its current stage before submitting."

		return false
	}

	return true
}
