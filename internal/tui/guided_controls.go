package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) guidedExtraKey(key string) tea.Cmd {
	if action := map[string]string{"p": "pause", "u": "resume", "x": "cancel"}[key]; action != "" {
		return m.submitGuidedStep(action)
	}

	switch key {
	case "c":
		if m.guided.View.Retry != nil {
			return m.startGuidedRequest(*m.guided.View.Retry)
		}
	case "i":
		if m.guided.View.CanRetainIdea {
			return m.editGuidedInput("idea")
		}
	case "d":
		m.textForm = textFormState{}
	case "a":
		m.inspectGuidedApproval()
	case "v":
		return m.toggleGuidedResult()
	default:
		return m.guidedDocumentKey(key)
	}

	return nil
}

func (m *Model) inspectGuidedApproval() {
	if !m.guided.View.CanApprove || m.guided.View.Native == nil {
		return
	}

	m.workflow = *m.guided.View.Native
	m.planWorkAction(WorkActionWorkflowApprove)
}

func (m *Model) guidedControlHint() string {
	text := ""

	labels := map[string]string{"finish": "e Continue to quiz", "pause": "p Pause", "resume": "u Resume", "cancel": "x Cancel"}
	for _, action := range m.guided.View.Controls {
		text += labels[action] + " · "
	}

	if m.guided.View.Retry != nil {
		text += "c Continue retained operation · "
	}

	if m.guided.View.CanApprove {
		text += "a Inspect native approval · "
	}

	return text + m.wizardBoardNativeHint()
}

func (m *Model) toggleGuidedResult() tea.Cmd {
	if m.guided.View.Document != nil && m.guided.View.Document.Kind == "results export" {
		m.guided.View.Document = nil

		return nil
	}

	return m.guidedDocumentKey("v")
}
