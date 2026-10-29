package tui

import (
	"crypto/rand"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func (m *Model) openWizardCreateForm() {
	if m.caseDirty {
		m.statusMessage = "Save the existing case changes before starting a wizard."

		return
	}

	request := GuidedRequest{Action: "start", Command: authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{ProjectID: m.projectID, RunID: "tui-" + rand.Text()}}}
	m.guided.Active = false
	m.openTextForm(textFormState{Title: "Wizard", SubmitLabel: "Create", Submit: (*Model).submitWizardCreate, GuidedRequest: &request, Fields: []textFormField{newTextFormField("Name", false, 512)}})
}

func (m *Model) submitWizardCreate() tea.Cmd {
	if m.guided.Busy {
		return nil
	}

	request := *m.textForm.GuidedRequest
	request.Command.Content = m.textForm.Fields[0].value()
	m.guided.Active = true

	return m.startGuidedRequest(request)
}
