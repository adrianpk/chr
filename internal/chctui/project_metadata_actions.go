package chctui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chc/internal/chc"
)

const (
	projectFieldName = iota
	projectFieldDescription
	projectFieldContext
	projectFieldCount
)

type ProjectMetadataExecutor interface {
	SaveProjectMetadata(metadata chc.ProjectMetadata) (ProjectMetadataResult, error)
}

type ProjectMetadataResult struct {
	Message  string
	Metadata chc.ProjectMetadata
}

type ProjectEditState struct {
	Active bool
	Field  int
	Input  textinput.Model
}

func (m Model) updateProjectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.projectEdit.Active {
		cmd := m.updateProjectEditKey(msg)

		return m, cmd
	}

	switch msg.String() {
	case "up", "k":
		m.selectPreviousProjectField()
	case "down", "j":
		m.selectNextProjectField()
	case "enter", "e":
		m.startProjectFieldEdit()
	case "ctrl+s", "s":
		m.saveProjectMetadata()
	case "r":
		m.refreshWorkspace()
	default:
		return m.updateNavigationKey(msg)
	}

	return m, nil
}

func (m *Model) updateProjectEditKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.projectEdit = ProjectEditState{}
		m.statusMessage = ""

		return nil
	case tea.KeyEnter:
		m.applyProjectFieldEdit()

		return nil
	case tea.KeySpace:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	}

	input, cmd := m.projectEdit.Input.Update(msg)
	m.projectEdit.Input = input

	return cmd
}

func (m *Model) selectPreviousProjectField() {
	if m.selectedProjectField == 0 {
		m.selectedProjectField = projectFieldCount - 1

		return
	}

	m.selectedProjectField--
}

func (m *Model) selectNextProjectField() {
	m.selectedProjectField++
	if m.selectedProjectField >= projectFieldCount {
		m.selectedProjectField = 0
	}
}

func (m *Model) startProjectFieldEdit() {
	switch m.selectedProjectField {
	case projectFieldName:
		m.projectEdit = newProjectEditState(projectFieldName, m.projectMetadata.Name)
	case projectFieldDescription:
		m.projectEdit = newProjectEditState(projectFieldDescription, m.projectMetadata.Description)
	case projectFieldContext:
		m.projectEdit = newProjectEditState(projectFieldContext, m.projectMetadata.Context)
	default:
		return
	}

	m.statusMessage = ""
}

func newProjectEditState(field int, value string) ProjectEditState {
	input := textinput.New()
	input.Prompt = ""
	input.SetValue(value)
	input.CursorStart()
	input.Focus()

	return ProjectEditState{
		Active: true,
		Field:  field,
		Input:  input,
	}
}

func (m *Model) applyProjectFieldEdit() {
	value := strings.TrimSpace(m.projectEdit.Input.Value())

	switch m.projectEdit.Field {
	case projectFieldName:
		m.projectMetadata.Name = value
	case projectFieldDescription:
		m.projectMetadata.Description = value
	case projectFieldContext:
		m.projectMetadata.Context = value
	}

	m.projectEdit = ProjectEditState{}
	m.projectDirty = true
	m.statusMessage = ""
}

func (m *Model) saveProjectMetadata() {
	if m.projectExecutor == nil {
		m.statusMessage = "Project metadata executor is unavailable."

		return
	}

	result, err := m.projectExecutor.SaveProjectMetadata(m.projectMetadata)
	if err != nil {
		m.statusMessage = "Project metadata save failed: " + err.Error()

		return
	}

	m.projectMetadata = result.Metadata
	m.projectDirty = false

	if !m.refreshWorkspaceGraphAfterWrite() {
		return
	}

	m.statusMessage = ""
}
