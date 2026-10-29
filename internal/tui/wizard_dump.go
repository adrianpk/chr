package tui

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openWizardDump() {
	if m.dumpEdit.Guided != nil {
		m.dumpEdit.Active = true

		return
	}

	session := m.guided.View.Session
	m.startDumpEdit("", session.Guided.Capture, strconv.FormatInt(session.Revision, 10))

	m.dumpEdit.Input.CharLimit = 65536
	request := m.guidedRequest("input")
	request.Command.Action = "finish"
	m.dumpEdit.Guided = &request
	m.statusMessage = ""
}

func (m *Model) submitWizardDump() tea.Cmd {
	request := *m.dumpEdit.Guided
	request.Command.OperationID = m.dumpEdit.OperationID
	request.Command.Content = m.dumpEdit.Input.Value()
	m.dumpEdit.Guided = &request

	return m.startGuidedRequest(request)
}
