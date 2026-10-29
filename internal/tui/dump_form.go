package tui

import (
	"crypto/rand"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openDumpCreateForm() {
	scope := m.defaultDumpScope()
	if scope.CaseSlug != "" && !caseSliceContainsSlug(m.cases, scope.CaseSlug) {
		m.statusMessage = "The selected case is unavailable. Refresh or clear case context before creating a dump."

		return
	}

	field := newTextFormField("Scope", false, 0)

	field.Choices = []textFormChoice{{Label: "Project-wide"}}
	for _, metadata := range m.cases {
		field.Choices = append(field.Choices, textFormChoice{Label: metadata.Name, Value: metadata.Slug, Revision: metadata.SourceHash})
		if metadata.Slug == scope.CaseSlug {
			field.Choice = len(field.Choices) - 1
		}
	}

	m.openTextForm(textFormState{
		Title: "Create Dump", SubmitLabel: "Create", Submit: (*Model).submitDumpForm,
		DumpRequest: DumpActionRequest{Action: DumpActionCreate, OperationID: rand.Text()},
		Fields:      []textFormField{field, newTextFormField("Title", false, 512), newTextFormField("Body", true, 65536)},
	})
}

func (m *Model) submitDumpForm() tea.Cmd {
	fields := m.textForm.Fields

	request := m.textForm.DumpRequest
	if request.Action == DumpActionAppend {
		request.Content = fields[0].Input.Value()

		return m.executeDumpForm(request)
	}

	choice := fields[0].Choices[fields[0].Choice]
	request.Scope, request.Revision = DumpScope{CaseSlug: choice.Value}, choice.Revision
	request.Title, request.Content = fields[1].Input.Value(), fields[2].Input.Value()

	return m.executeDumpForm(request)
}

func (m *Model) executeDumpForm(request DumpActionRequest) tea.Cmd {
	if m.dumpExecutor == nil {
		m.statusMessage = "Dump action executor is unavailable."

		return nil
	}

	result, err := m.dumpExecutor.ExecuteDumpAction(request)
	m.recordOperation(result.Outcome)

	if err != nil {
		m.statusMessage = "Dump submission failed: " + err.Error()
		m.bindDumpFormRetry(request)

		return nil
	}

	m.textForm = textFormState{}
	m.selectView(ViewDumps)
	m.bindCaseContext(request.Scope.CaseSlug, "")
	m.applyDumpActionResult(result)

	return nil
}

func (m *Model) bindDumpFormRetry(request DumpActionRequest) {
	if !m.feedback.Result.Retry {
		return
	}

	context := m.currentInteraction()
	m.feedback.Context = context
	m.feedback.Retry = func(m *Model) tea.Cmd {
		if context != m.currentInteraction() {
			return nil
		}

		m.feedback.Open = false

		return m.executeDumpForm(request)
	}
}

func (m *Model) openDumpAppendForm() {
	dump, ok := m.selectedDumpView()
	if !ok {
		m.statusMessage = "No dump selected."

		return
	}

	request := m.newSelectedDumpActionState(DumpActionAppend, "Append dump").Request
	request.Scope = dump.Scope
	m.openTextForm(textFormState{
		Title: "Append to " + dump.Title, SubmitLabel: "Append", Submit: (*Model).submitDumpForm,
		DumpRequest: request, Fields: []textFormField{newTextFormField("Text", true, 1<<20)},
	})
}
