package tui

import (
	"crypto/rand"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateGuidedOrHarnessKey(message tea.KeyMsg) updateOutcome {
	if m.guided.Active && !m.textForm.Active {
		command := m.updateGuidedKey(message)

		return updateOutcome{Command: command, Handled: true}
	}

	if m.harnessSelection.Active {
		command := m.updateHarnessKey(message)

		return updateOutcome{Command: command, Handled: true}
	}

	return updateOutcome{Command: nil, Handled: false}
}

func (m *Model) updateGuidedKey(message tea.KeyMsg) tea.Cmd {
	if message.String() == "ctrl+c" {
		return tea.Quit
	}

	if m.guided.Editing {
		return m.updateGuidedInputKey(message)
	}

	if message.String() == "esc" {
		m.guided.Active = false

		return nil
	}

	if m.guided.Busy {
		return nil
	}

	if m.moveGuidedSelection(message.String()) {
		return nil
	}

	return m.guidedActionKey(message.String())
}

func (m *Model) guidedActionKey(key string) tea.Cmd {
	switch key {
	case "n":
		m.openWizardCreateForm()

		return nil
	case "r":
		command := m.startGuidedRequest(m.guidedRequest("inspect"))

		return command
	case "b":
		request := m.guidedRequest("inspect")
		request.Command.RunID = ""
		m.guided.Editing, m.guided.Scroll = false, 0

		command := m.startGuidedRequest(request)

		return command
	case "y":
		if m.guided.Pending != nil {
			command := m.startGuidedRequest(*m.guided.Pending)

			return command
		}
	case "enter", "e":
		command := m.openGuidedInput()

		return command
	case "f":
		command := m.submitGuidedStep("finish")

		return command
	default:
		command := m.guidedExtraKey(key)

		return command
	}

	return nil
}

func (m *Model) moveGuidedSelection(key string) bool {
	delta := map[string]int{"up": -1, "k": -1, "down": 1, "j": 1, "pgup": -10, "pgdown": 10}[key]
	if delta == 0 {
		return false
	}

	if m.guided.View.Session == nil {
		m.guided.Selected = max(0, min(m.guided.Selected+delta, len(m.guided.View.Runs)-1))
	} else {
		m.guided.Scroll = max(0, m.guided.Scroll+delta)
	}

	return true
}

func (m *Model) openGuidedInput() tea.Cmd {
	if m.guided.View.Session == nil {
		if len(m.guided.View.Runs) == 0 {
			return nil
		}

		request := m.guidedRequest("inspect")
		request.Command.RunID = m.guided.View.Runs[min(m.guided.Selected, len(m.guided.View.Runs)-1)].ID
		m.guided.Scroll = 0

		return m.startGuidedRequest(request)
	}

	return m.editGuidedInput(m.guided.View.InputAction)
}

func (m *Model) editGuidedInput(action string) tea.Cmd {
	if action == "" {
		return nil
	}

	if m.guided.Bound.Command.RunID == m.guided.View.Session.ID && m.guided.Bound.Command.Action == action && m.guided.Input.Value() != "" {
		m.guided.Editing = true

		return m.guided.Input.Focus()
	}

	m.guided.Bound = m.guidedRequest("input")
	m.guided.Bound.Command.Action = action
	m.guided.Bound.Command.OperationID = "tui-" + strings.ToLower(rand.Text())
	m.guided.Input = textarea.New()
	m.guided.Input.CharLimit, m.guided.Input.ShowLineNumbers = 65536, false
	m.guided.Input.SetWidth(max(m.width-8, 40))
	m.guided.Input.SetHeight(5)

	m.guided.Editing = true

	return m.guided.Input.Focus()
}

func (m *Model) updateGuidedInputKey(message tea.KeyMsg) tea.Cmd {
	switch message.String() {
	case "esc":
		m.guided.Editing = false

		return nil
	case "enter":
		request := m.guided.Bound
		request.Command.Content = m.guided.Input.Value()

		command := m.startGuidedRequest(request)

		return command
	case "ctrl+j", "shift+enter":
		message = tea.KeyMsg{Type: tea.KeyEnter}
	}

	var command tea.Cmd

	m.guided.Input, command = m.guided.Input.Update(message)

	return command
}

func (m *Model) submitGuidedStep(action string) tea.Cmd {
	if m.guided.View.Session == nil || !slices.Contains(m.guided.View.Controls, action) {
		return nil
	}

	request := m.guidedRequest("input")
	request.Command.Action, request.Command.OperationID = action, "tui-"+strings.ToLower(rand.Text())

	return m.startGuidedRequest(request)
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
