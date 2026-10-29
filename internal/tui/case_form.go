package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func (m *Model) submitCaseForm() tea.Cmd {
	if m.caseExecutor == nil {
		m.statusMessage = "Case metadata executor is unavailable."

		return nil
	}

	fields := m.textForm.Fields
	metadata := authoringcore.CaseMetadata{
		SchemaVersion: authoringcore.CaseMetadataSchemaVersion,
		Name:          fields[0].Input.Value(), Description: fields[1].Input.Value(), Context: fields[2].Input.Value(),
	}
	result, err := m.caseExecutor.SaveCaseMetadata(metadata)
	m.recordOperation(result.Outcome)

	if err != nil {
		m.statusMessage = "Case creation failed: " + err.Error()

		return nil
	}

	m.textForm = textFormState{}
	m.caseFilter = ""
	m.applyCaseMetadataResult(result)
	m.bindCaseContext(result.Case.Slug, "")

	m.selectedPane = 1
	if m.refreshWorkspaceGraphAfterWrite() {
		m.statusMessage = result.Message
	}

	return nil
}
