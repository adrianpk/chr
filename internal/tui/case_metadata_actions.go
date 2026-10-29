package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

const (
	caseFieldName = iota
	caseFieldDescription
	caseFieldContext
	caseFieldCount
)

// CaseMetadataExecutor persists use-case metadata edits.
type CaseMetadataExecutor interface {
	SaveCaseMetadata(metadata authoringcore.CaseMetadata) (CaseMetadataResult, error)
}

// CaseMetadataResult returns refreshed use-case metadata after a save.
type CaseMetadataResult struct {
	Outcome operation.Result
	Message string
	Case    authoringcore.CaseMetadata
	Cases   []authoringcore.CaseMetadata
}

func (m *Model) updateCasesKey(msg tea.KeyMsg) tea.Cmd {
	if m.caseEdit.Active {
		cmd := m.updateCaseEditKey(msg)

		return cmd
	}

	if m.caseFilterEditing {
		m.updateCaseFilterKey(msg)

		return nil
	}

	if msg.String() == "w" || msg.String() == "g" {
		m.openWizardCreateForm()

		return nil
	}

	if msg.String() == "o" {
		command := m.openGuided()

		return command
	}

	if m.handleCaseSelectionKey(msg) {
		return nil
	}

	if m.handleCaseActionKey(msg) {
		return nil
	}

	return m.updateNavigationKey(msg)
}

func (m *Model) handleCaseSelectionKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "up", "k":
		switch m.selectedPane {
		case 0:
			m.selectPreviousCase()
		case 1:
			m.selectPreviousCaseField()
		default:
			return false
		}
	case "down", "j":
		switch m.selectedPane {
		case 0:
			m.selectNextCase()
		case 1:
			m.selectNextCaseField()
		default:
			return false
		}
	default:
		return false
	}

	return true
}

func (m *Model) handleCaseActionKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "enter", "e":
		return m.handleCaseEnterKey()
	case "n":
		m.newCase()
	case "ctrl+s", "s":
		m.saveSelectedCase()
	case "r":
		m.refreshWorkspace()
	case "/":
		if m.selectedPane != 0 {
			return false
		}

		m.caseFilterEditing = true
		m.statusMessage = ""
	case "esc":
		m.handleCaseEscapeKey()
	default:
		return false
	}

	return true
}

func (m *Model) handleCaseEnterKey() bool {
	switch m.selectedPane {
	case 0:
		m.captureCaseContext()

		m.selectedPane = 1
	case 1:
		m.startCaseFieldEdit()
	default:
		return false
	}

	return true
}

func (m *Model) handleCaseEscapeKey() {
	if m.selectedPane == 0 {
		m.caseFilter = ""

		return
	}

	m.selectPrimaryPane()
}

func (m *Model) updateCaseEditKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		if m.caseEdit.Input.Value() != m.caseEdit.Original {
			return m.confirmDraftChange("Discard the unsent case field?", discardCaseField)
		}

		return discardCaseField(m)

	case tea.KeyEnter:
		m.applyCaseFieldEdit()

		return nil
	case tea.KeySpace:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	}

	input, cmd := m.caseEdit.Input.Update(msg)
	m.caseEdit.Input = input

	return cmd
}

func (m *Model) updateCaseFilterKey(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEsc:
		m.caseFilterEditing = false
		m.statusMessage = ""

		return
	case tea.KeyEnter:
		m.caseFilterEditing = false
		m.statusMessage = ""

		return
	case tea.KeyBackspace:
		if m.caseFilter != "" {
			runes := []rune(m.caseFilter)
			m.caseFilter = string(runes[:len(runes)-1])
		}
	case tea.KeySpace:
		m.caseFilter += " "
	case tea.KeyRunes:
		m.caseFilter += string(msg.Runes)
	}

	m.clampSelectedCase()
}

func (m *Model) selectPreviousCase() {
	cases := m.filteredCases()
	if len(cases) == 0 {
		return
	}

	position := m.selectedCasePosition(cases)
	if position == 0 {
		m.selectedCase = m.caseIndexForSlug(cases[len(cases)-1].Slug)

		return
	}

	m.selectedCase = m.caseIndexForSlug(cases[position-1].Slug)
}

func (m *Model) selectNextCase() {
	cases := m.filteredCases()
	if len(cases) == 0 {
		return
	}

	position := m.selectedCasePosition(cases)
	if position >= len(cases)-1 {
		m.selectedCase = m.caseIndexForSlug(cases[0].Slug)

		return
	}

	m.selectedCase = m.caseIndexForSlug(cases[position+1].Slug)
}

func (m *Model) selectedCasePosition(cases []authoringcore.CaseMetadata) int {
	selected, ok := m.selectedCaseMetadata()
	if !ok {
		return 0
	}

	for index, current := range cases {
		if current.Slug == selected.Slug {
			return index
		}
	}

	return 0
}

func (m *Model) caseIndexForSlug(slug string) int {
	for index, current := range m.cases {
		if current.Slug == slug {
			return index
		}
	}

	return 0
}

func (m *Model) clampSelectedCase() {
	if !m.caseContextAvailable() {
		m.selectedCase = -1
		m.explainCaseContext()

		return
	}

	cases := m.filteredCases()
	if len(cases) == 0 {
		m.selectedCase = 0

		return
	}

	selected, ok := m.selectedCaseMetadata()
	if !ok || !caseSliceContainsSlug(cases, selected.Slug) {
		m.selectedCase = m.caseIndexForSlug(cases[0].Slug)
	}
}

func caseSliceContainsSlug(cases []authoringcore.CaseMetadata, slug string) bool {
	for _, current := range cases {
		if current.Slug == slug {
			return true
		}
	}

	return false
}

func (m *Model) selectPreviousCaseField() {
	if m.selectedCaseField == 0 {
		m.selectedCaseField = caseFieldCount - 1

		return
	}

	m.selectedCaseField--
}

func (m *Model) selectNextCaseField() {
	m.selectedCaseField++
	if m.selectedCaseField >= caseFieldCount {
		m.selectedCaseField = 0
	}
}

func (m *Model) startCaseFieldEdit() {
	m.draftSequence++

	metadata, ok := m.selectedCaseMetadata()
	if !ok {
		return
	}

	switch m.selectedCaseField {
	case caseFieldName:
		m.caseEdit = newProjectEditState(caseFieldName, metadata.Name)
	case caseFieldDescription:
		m.caseEdit = newProjectEditState(caseFieldDescription, metadata.Description)
	case caseFieldContext:
		m.caseEdit = newProjectEditState(caseFieldContext, metadata.Context)
	default:
		return
	}

	m.statusMessage = ""
}

func (m *Model) applyCaseFieldEdit() {
	if len(m.cases) == 0 || m.selectedCase < 0 || m.selectedCase >= len(m.cases) {
		m.caseEdit = ProjectEditState{}

		return
	}

	value := strings.TrimSpace(m.caseEdit.Input.Value())

	switch m.caseEdit.Field {
	case caseFieldName:
		m.cases[m.selectedCase].Name = value
	case caseFieldDescription:
		m.cases[m.selectedCase].Description = value
	case caseFieldContext:
		m.cases[m.selectedCase].Context = value
	}

	m.caseEdit = ProjectEditState{}
	m.caseDirty = true
	m.statusMessage = ""
}

func (m *Model) newCase() {
	if m.caseDirty {
		m.statusMessage = "Save the current case changes before creating another case."

		return
	}

	m.openTextForm(textFormState{
		Title: "Create Case", SubmitLabel: "Create", Submit: (*Model).submitCaseForm,
		Fields: []textFormField{newTextFormField("Name", false, 1<<20), newTextFormField("Description", true, 1<<20), newTextFormField("Context", true, 1<<20)},
	})
}

func (m *Model) saveSelectedCase() {
	if m.caseExecutor == nil {
		m.statusMessage = "Case metadata executor is unavailable."

		return
	}

	metadata, ok := m.selectedCaseMetadata()
	if !ok {
		m.statusMessage = "No case selected."

		return
	}

	result, err := m.caseExecutor.SaveCaseMetadata(metadata)
	m.recordOperation(result.Outcome)
	m.bindCaseRecovery(metadata)

	if err != nil {
		m.statusMessage = "Case metadata save failed: " + err.Error()

		return
	}

	m.applyCaseMetadataResult(result)

	m.caseDirty = false

	if !m.refreshWorkspaceGraphAfterWrite() {
		return
	}

	m.statusMessage = ""
}

func (m *Model) applyCaseMetadataResult(result CaseMetadataResult) {
	selectedSlug := result.Case.Slug
	if selectedSlug == "" {
		selected, ok := m.selectedCaseMetadata()
		if ok {
			selectedSlug = selected.Slug
		}
	}

	if len(result.Cases) > 0 {
		m.cases = append([]authoringcore.CaseMetadata(nil), result.Cases...)
	} else if result.Case.Slug != "" {
		m.upsertCase(result.Case)
	}

	if selectedSlug != "" {
		m.selectedCase = m.caseIndexForSlug(selectedSlug)
	}

	m.clampSelectedCase()
}

func (m *Model) upsertCase(metadata authoringcore.CaseMetadata) {
	for index, current := range m.cases {
		if current.Slug == metadata.Slug {
			m.cases[index] = metadata

			return
		}
	}

	m.cases = append(m.cases, metadata)
	m.selectedCase = len(m.cases) - 1
}
