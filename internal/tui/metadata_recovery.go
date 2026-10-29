package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

type projectMetadataReader interface {
	ReadProjectMetadata() (ProjectMetadataResult, error)
}
type caseMetadataReader interface {
	ReadCaseMetadata(string) (CaseMetadataResult, error)
}

func (m *Model) bindProjectRecovery() {
	reader, ok := m.projectExecutor.(projectMetadataReader)
	if !ok || !m.feedback.Result.Reload {
		return
	}

	context := m.currentInteraction()
	m.feedback.Context = context
	m.feedback.Reload = func(m *Model) tea.Cmd {
		if context != m.currentInteraction() {
			return nil
		}

		return m.confirmDraftChange("Reload project metadata and discard local changes?", func(m *Model) tea.Cmd {
			result, err := reader.ReadProjectMetadata()
			m.recordOperation(result.Outcome)

			if err != nil {
				m.statusMessage = "Reload failed: " + err.Error()

				return nil
			}

			m.projectMetadata, m.projectDirty, m.projectEdit = result.Metadata, false, ProjectEditState{}
			m.statusMessage = result.Message

			return nil
		})
	}
}

func (m *Model) bindCaseRecovery(metadata authoringcore.CaseMetadata) {
	reader, ok := m.caseExecutor.(caseMetadataReader)
	if !ok || !m.feedback.Result.Reload {
		return
	}

	context := m.currentInteraction()
	m.feedback.Context = context
	m.feedback.Reload = func(m *Model) tea.Cmd {
		if context != m.currentInteraction() {
			return nil
		}

		return m.confirmDraftChange("Reload this case and discard its local changes?", func(m *Model) tea.Cmd {
			result, err := reader.ReadCaseMetadata(metadata.Slug)
			m.recordOperation(result.Outcome)

			if err != nil {
				m.statusMessage = "Reload failed: " + err.Error()

				return nil
			}

			m.upsertCase(result.Case)

			m.caseDirty, m.caseEdit = false, ProjectEditState{}
			m.statusMessage = result.Message

			return nil
		})
	}
}
