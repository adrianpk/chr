package tui

import (
	"crypto/rand"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateGuidedOrHarnessKey(message tea.KeyMsg) updateOutcome {
	if m.wizardBound() && !m.wizardInteractionActive() {
		if command, handled := m.updateWizardKey(message); handled {
			return updateOutcome{Command: command, Handled: true}
		}
	}

	if m.harnessSelection.Active {
		return updateOutcome{Command: m.updateHarnessKey(message), Handled: true}
	}

	return updateOutcome{}
}

func (m *Model) updateWizardKey(message tea.KeyMsg) (tea.Cmd, bool) {
	if m.wizardWriteBusy() {
		return nil, true
	}

	key := message.String()
	if command, handled := m.updateWizardComponentKey(message); handled {
		return command, true
	}

	if m.wizardCaseKey(key) {
		return nil, false
	}

	if delta := map[string]int{"up": -1, "k": -1, "down": 1, "j": 1, "pgup": -10, "pgdown": 10}[key]; delta != 0 {
		m.scrollDetail(delta)

		return nil, true
	}

	switch key {
	case "r":
		return m.startGuidedRequest(m.guidedRequest("inspect")), true
	case "y":
		return m.retryWizardRequest(), true
	case "c":
		if m.SelectedView() != m.wizardStageView() {
			m.selectView(m.wizardStageView())

			return nil, true
		}

		return m.guidedExtraKey(key), true
	default:
		return m.wizardActionKey(key)
	}
}

func (m *Model) editGuidedInput(action string) tea.Cmd {
	if action == "" || m.guided.View.Session == nil {
		return nil
	}

	if action == "answer" {
		m.startQuizAnswer()

		return nil
	}

	if action == "append" {
		m.openWizardDump()

		return nil
	}

	if m.textForm.GuidedRequest != nil && m.textForm.GuidedRequest.Command.RunID == m.guided.View.Session.ID && m.textForm.GuidedRequest.Command.Action == action {
		m.textForm.Active = true

		return nil
	}

	request := m.guidedRequest("input")
	request.Command.Action, request.Command.OperationID = action, "tui-"+rand.Text()

	label := "Answer"
	if action == "idea" {
		label = "Later idea"
	}

	if action == "append" {
		label = "Capture"
	}

	m.openTextForm(textFormState{Title: label, SubmitLabel: "Submit", Submit: (*Model).submitWizardText, GuidedRequest: &request, Details: []string{m.guided.View.Prompt}, Fields: []textFormField{newTextFormField(label, true, 65536)}})

	return nil
}

func (m *Model) submitWizardText() tea.Cmd {
	request := *m.textForm.GuidedRequest
	request.Command.Content = m.textForm.Fields[0].value()
	m.textForm.GuidedRequest = &request

	return m.startGuidedRequest(request)
}

func (m *Model) submitGuidedStep(action string) tea.Cmd {
	if m.guided.View.Session == nil || !slices.Contains(m.guided.View.Controls, action) {
		return nil
	}

	request := m.guidedRequest("input")
	request.Command.Action, request.Command.OperationID = action, "tui-"+rand.Text()

	return m.confirmGuidedMutation(request)
}

func (m *Model) guidedDocumentKey(key string) tea.Cmd {
	kind := map[string]string{"1": "dump", "2": "quiz", "3": "specification", "4": "plan", "5": "tracker", "v": "results"}[key]
	if kind == "" || m.guided.View.Session == nil {
		return nil
	}

	request := m.guidedRequest("document")
	request.Document = kind

	return m.startGuidedRequest(request)
}

func (m *Model) updateWizardActionKey(key string) (tea.Cmd, bool) {
	switch key {
	case "enter", "e":
		return m.editGuidedInput(m.guided.View.InputAction), true
	case "f":
		if m.guided.View.Session.Guided.Stage == "dump" {
			m.openWizardDump()

			return nil, true
		}

		m.explainWizardMode()

		return nil, true
	case "p", "u", "x", "i", "a", "v", "1", "2", "3", "4", "5":
		return m.guidedExtraKey(key), true
	default:
		m.explainWizardMode()

		return nil, true
	}
}

func (m *Model) retryWizardRequest() tea.Cmd {
	if m.guided.Pending == nil {
		return nil
	}

	return m.startGuidedRequest(*m.guided.Pending)
}
