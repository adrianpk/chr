package tui

import (
	"crypto/rand"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// DumpActionID identifies a dump action awaiting TUI confirmation.
type DumpActionID string

// Dump action identifiers define supported dump commands.
const (
	DumpActionCreate       DumpActionID = "create_dump"
	DumpActionAppend       DumpActionID = "append_dump"
	DumpActionEdit         DumpActionID = "edit_dump"
	DumpActionSave         DumpActionID = "save_dump"
	DumpActionExternalEdit DumpActionID = "external_edit_dump"
	DumpActionRefresh      DumpActionID = "refresh_dumps"
)

// DumpActionExecutor runs confirmed dump actions.
type DumpActionExecutor interface {
	ExecuteDumpAction(request DumpActionRequest) (DumpActionResult, error)
}

// DumpActionRequest carries selected dump action input.
type DumpActionRequest struct {
	OperationID, Revision string
	Action                DumpActionID
	Title                 string
	Content               string
	Scope                 DumpScope
	Path                  string
}

// DumpActionResult contains refreshed dump state after an action.
type DumpActionResult struct {
	Committed, ProjectionStale bool
	Revision                   string
	Message                    string
	Dumps                      []DumpView
	Content                    string
	SelectedPath               string
	EditorCommand              *exec.Cmd
	EditorCommandDescription   string
}

// DumpActionState stores pending confirmation state for a dump action.
type DumpActionState struct {
	Pending bool
	Request DumpActionRequest
	Title   string
	Details []string
}

// DumpEditState stores in-terminal dump editing state.
type DumpEditState struct {
	OperationID, Revision string
	Active                bool
	Path                  string
	Input                 textarea.Model
}

func (m Model) renderDumpAction() string {
	if !m.dumpAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Action\n")
	fmt.Fprintf(&output, "  %s\n", m.dumpAction.Title)

	for _, detail := range m.dumpAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) updateDumpActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelDumpAction()
	case tea.KeyEnter:
		return m, m.confirmDumpAction()
	}

	return m, nil
}

func (m Model) updateDumpEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.dumpEdit = DumpEditState{}
		m.statusMessage = ""

		return m, nil
	case tea.KeyCtrlS:
		m.saveDumpEdit()

		return m, nil
	}

	previous := m.dumpEdit.Input.Value()

	input, cmd := m.dumpEdit.Input.Update(msg)
	if previous != input.Value() {
		m.dumpEdit.OperationID = rand.Text()
	}

	m.dumpEdit.Input = input

	return m, cmd
}

func (m *Model) planDumpAction(action DumpActionID) {
	m.dumpAction = m.newDumpActionState(action)
	m.statusMessage = ""
}

func (m *Model) cancelDumpAction() {
	m.dumpAction = DumpActionState{}
	m.statusMessage = "Action cancelled."
}

func (m *Model) confirmDumpAction() tea.Cmd {
	request := m.dumpAction.Request
	m.dumpAction = DumpActionState{}

	if m.dumpExecutor == nil {
		m.statusMessage = "Dump action executor is unavailable."

		return nil
	}

	result, err := m.dumpExecutor.ExecuteDumpAction(request)
	if err != nil {
		m.statusMessage = "Action failed: " + err.Error()

		return nil
	}

	if result.EditorCommand != nil {
		m.statusMessage = "External editor opened. chr will resume when the editor process exits."

		return tea.ExecProcess(result.EditorCommand, func(err error) tea.Msg {
			return dumpEditorFinishedMsg{
				err:        err,
				path:       request.Path,
				command:    result.EditorCommandDescription,
				workingDir: result.EditorCommand.Dir,
			}
		})
	}

	if strings.TrimSpace(result.SelectedPath) != "" && result.Dumps == nil {
		m.startDumpEdit(result.SelectedPath, result.Content, result.Revision)

		return nil
	}

	m.applyDumpActionResult(result)

	return nil
}

func (m *Model) startDumpEdit(path, content, revision string) {
	input := textarea.New()
	input.SetValue(content)
	input.Focus()

	input.ShowLineNumbers = false
	input.Prompt = ""
	input.CharLimit = 0

	m.dumpEdit = DumpEditState{
		OperationID: rand.Text(), Revision: revision,
		Active: true,
		Path:   path,
		Input:  input,
	}
	m.selectedPane = 1
	m.maximizedPane = 1
	m.statusMessage = "Editing " + path
}

func (m *Model) saveDumpEdit() {
	if m.dumpExecutor == nil {
		m.statusMessage = "Dump action executor is unavailable."

		return
	}

	result, err := m.dumpExecutor.ExecuteDumpAction(DumpActionRequest{
		Action: DumpActionSave, OperationID: m.dumpEdit.OperationID, Revision: m.dumpEdit.Revision,
		Path:    m.dumpEdit.Path,
		Content: m.dumpEdit.Input.Value(),
	})
	if err != nil {
		m.statusMessage = "Save failed: " + err.Error()

		return
	}

	m.dumpEdit = DumpEditState{}
	m.applyDumpActionResult(result)
}

func (m *Model) applyDumpActionResult(result DumpActionResult) {
	if result.Dumps != nil {
		m.dumps = normalizeDumpViews(result.Dumps)
		m.selectDumpByPath(m.dumpSelectionPath(result.SelectedPath))
	}

	m.statusMessage = result.Message
}

func (m Model) dumpSelectionPath(path string) string {
	path = strings.TrimSpace(path)
	if path != "" {
		return path
	}

	dump, ok := m.selectedDumpView()
	if !ok {
		return ""
	}

	return dump.RelativePath
}

type dumpEditorFinishedMsg struct {
	err        error
	path       string
	command    string
	workingDir string
}

func (m *Model) completeDumpEdit(msg dumpEditorFinishedMsg) {
	if msg.err != nil {
		m.statusMessage = "Edit failed: " + externalEditorErrorMessage(msg)

		return
	}

	if m.dumpExecutor == nil {
		m.statusMessage = "Dump action executor is unavailable."

		return
	}

	result, err := m.dumpExecutor.ExecuteDumpAction(DumpActionRequest{
		Action: DumpActionRefresh,
		Path:   msg.path,
	})
	if err != nil {
		m.statusMessage = "Refresh failed: " + err.Error()

		return
	}

	m.applyDumpActionResult(result)
}

func externalEditorErrorMessage(msg dumpEditorFinishedMsg) string {
	parts := []string{}
	if strings.TrimSpace(msg.command) != "" {
		parts = append(parts, "command="+msg.command)
	}

	if strings.TrimSpace(msg.workingDir) != "" {
		parts = append(parts, "dir="+msg.workingDir)
	}

	parts = append(parts, msg.err.Error())

	return strings.Join(parts, " ")
}

func (m Model) defaultDumpScope() DumpScope {
	if m.SelectedView() != ViewCases {
		return DumpScope{}
	}

	selectedCase, ok := m.selectedCaseMetadata()
	if !ok {
		return DumpScope{}
	}

	return DumpScope{CaseSlug: selectedCase.Slug}
}

func (m Model) defaultDumpTitle() string {
	scope := m.defaultDumpScope()
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return "Project dump"
	}

	return scope.CaseSlug + " dump"
}

func (m Model) newDumpActionState(action DumpActionID) DumpActionState {
	switch action {
	case DumpActionCreate:
		return m.newCreateDumpActionState(action)
	case DumpActionAppend:
		return m.newSelectedDumpActionState(action, "Append dump")
	case DumpActionEdit:
		return m.newSelectedDumpActionState(action, "Edit dump")
	case DumpActionExternalEdit:
		return m.newSelectedDumpActionState(action, "Open dump in external editor")
	case DumpActionRefresh:
		return m.newSelectedDumpActionState(action, "Refresh dumps")
	default:
		return DumpActionState{}
	}
}

func (m Model) newCreateDumpActionState(action DumpActionID) DumpActionState {
	request := DumpActionRequest{
		Action: action, OperationID: rand.Text(),
		Title:   m.defaultDumpTitle(),
		Content: "Initial dump created from the TUI.",
		Scope:   m.defaultDumpScope(),
	}

	return DumpActionState{
		Pending: true,
		Request: request,
		Title:   "Create dump",
		Details: []string{
			"Title: " + request.Title,
			"Use Case: " + dumpScopeLabel(request.Scope),
		},
	}
}

func (m Model) newSelectedDumpActionState(action DumpActionID, title string) DumpActionState {
	dump, ok := m.selectedDumpView()
	if !ok {
		if action == DumpActionRefresh {
			return DumpActionState{
				Pending: true,
				Request: DumpActionRequest{
					Action: action,
				},
				Title: title,
				Details: []string{
					"Target: Dumps",
				},
			}
		}

		return DumpActionState{}
	}

	request := DumpActionRequest{
		Action: action, OperationID: rand.Text(), Revision: dump.Revision,
		Path: dump.RelativePath,
	}

	if action == DumpActionAppend {
		request.Content = "Appended dump note from the TUI."
	}

	return DumpActionState{
		Pending: true,
		Request: request,
		Title:   title,
		Details: []string{
			"Path: " + request.Path,
		},
	}
}
