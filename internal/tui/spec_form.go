package tui

import (
	"crypto/rand"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openSpecCreateForm() {
	if !m.caseContextAvailable() {
		m.explainCaseContext()

		return
	}

	m.openSpecForm(SpecActionRequest{Action: SpecActionCreate, OperationID: rand.Text()}, SpecView{Scope: QuizScope{CaseSlug: m.caseContext.CaseSlug}, Status: "draft"})
}

func (m *Model) openSpecEditForm() {
	spec, ok := m.selectedSpecView()
	if !ok || m.specExecutor == nil {
		m.statusMessage = "No editable spec selected."

		return
	}

	result, err := m.specExecutor.ExecuteSpecAction(SpecActionRequest{Action: SpecActionEdit, Path: spec.RelativePath})
	m.recordOperation(result.Outcome)

	if err != nil || result.Document == nil {
		m.statusMessage = "The spec could not be opened. Refresh its catalog and try again."

		return
	}

	m.openSpecForm(SpecActionRequest{Action: SpecActionSave, OperationID: rand.Text(), Path: result.Document.RelativePath, Revision: result.Document.Revision, Scope: result.Document.Scope}, *result.Document)
}

func (m *Model) openSpecForm(request SpecActionRequest, document SpecView) {
	title, submit := "Create Spec", "Create"
	if request.Action == SpecActionSave {
		title, submit = "Edit Spec", "Save"
	}

	fields := []textFormField{m.specScopeField(request.Action, document.Scope), newTextFormField("Title", false, 512), newTextFormField("Body", true, 1<<20), newTextFormField("State", false, 0)}
	fields[1].Input.SetValue(document.Title)
	fields[2].Input.SetValue(document.Content)

	fields[3].Choices = []textFormChoice{{Label: "Draft", Value: "draft"}, {Label: "Approved", Value: "approved"}}
	if document.Status == "approved" {
		fields[3].Choice = 1
	}

	m.openTextForm(textFormState{Title: title, SubmitLabel: submit, Submit: (*Model).submitSpecForm, SpecRequest: request, Fields: fields})
}

func (m *Model) specScopeField(action SpecActionID, scope QuizScope) textFormField {
	field := newTextFormField("Scope", false, 0)

	field.Choices = []textFormChoice{{Label: "Project-wide"}}
	for _, metadata := range m.cases {
		field.Choices = append(field.Choices, textFormChoice{Label: metadata.Name, Value: metadata.Slug})
		if metadata.Slug == scope.CaseSlug {
			field.Choice = len(field.Choices) - 1
		}
	}

	if action == SpecActionSave {
		choice := textFormChoice{Label: quizUseCaseLabel(scope), Value: scope.CaseSlug}
		if field.Choices[field.Choice].Value == scope.CaseSlug {
			choice = field.Choices[field.Choice]
		}

		field.Label, field.Choices, field.Choice = "Scope (retained)", []textFormChoice{choice}, 0
	}

	return field
}

func (m *Model) submitSpecForm() tea.Cmd {
	fields := m.textForm.Fields

	request := m.textForm.SpecRequest
	if request.Action == SpecActionCreate {
		request.Scope.CaseSlug = fields[0].value()
	}

	request.Title, request.Content, request.Status = fields[1].value(), fields[2].value(), fields[3].value()
	m.executeSpecForm(request)

	return nil
}

func (m *Model) executeSpecForm(request SpecActionRequest) {
	if m.specExecutor == nil {
		m.statusMessage = "Spec editor is unavailable."

		return
	}

	result, err := m.specExecutor.ExecuteSpecAction(request)
	m.recordOperation(result.Outcome)

	if err != nil {
		m.statusMessage = "Spec submission failed: " + err.Error()
		m.bindSpecFormRecovery(request)

		return
	}

	m.textForm = textFormState{}
	m.bindCaseContext(request.Scope.CaseSlug, "")
	m.applySpecActionResult(result)
}
