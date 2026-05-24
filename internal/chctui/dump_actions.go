package chctui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type DumpActionID string

const (
	DumpActionCreate DumpActionID = "create_dump"
	DumpActionAppend DumpActionID = "append_dump"
	DumpActionEdit   DumpActionID = "edit_dump"
)

type DumpActionExecutor interface {
	ExecuteDumpAction(request DumpActionRequest) (DumpActionResult, error)
}

type DumpActionRequest struct {
	Action  DumpActionID
	Title   string
	Content string
	Scope   DumpScope
	Path    string
}

type DumpActionResult struct {
	Message string
	Dumps   []DumpView
}

type DumpActionState struct {
	Pending bool
	Request DumpActionRequest
	Title   string
	Details []string
}

func (m Model) renderDumpAction() string {
	if !m.dumpAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Action\n")
	output.WriteString(fmt.Sprintf("  %s\n", m.dumpAction.Title))

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
		m.confirmDumpAction()
	}

	return m, nil
}

func (m *Model) planDumpAction(action DumpActionID) {
	m.dumpAction = m.newDumpActionState(action)
	m.statusMessage = ""
}

func (m *Model) cancelDumpAction() {
	m.dumpAction = DumpActionState{}
	m.statusMessage = "Action cancelled."
}

func (m *Model) confirmDumpAction() {
	request := m.dumpAction.Request
	m.dumpAction = DumpActionState{}

	if m.dumpExecutor == nil {
		m.statusMessage = "Dump action executor is unavailable."

		return
	}

	result, err := m.dumpExecutor.ExecuteDumpAction(request)
	if err != nil {
		m.statusMessage = "Action failed: " + err.Error()

		return
	}

	m.applyDumpActionResult(result)
}

func (m *Model) applyDumpActionResult(result DumpActionResult) {
	if result.Dumps != nil {
		m.dumps = normalizeDumpViews(result.Dumps)
		m.selectedDump = 0
	}

	m.statusMessage = result.Message
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
		request := DumpActionRequest{
			Action:  action,
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
				"Scope: " + dumpScopeLabel(request.Scope),
			},
		}
	case DumpActionAppend:
		dump, ok := m.selectedDumpView()
		if !ok {
			return DumpActionState{}
		}

		request := DumpActionRequest{
			Action:  action,
			Path:    dump.RelativePath,
			Content: "Appended dump note from the TUI.",
		}

		return DumpActionState{
			Pending: true,
			Request: request,
			Title:   "Append dump",
			Details: []string{
				"Path: " + request.Path,
			},
		}
	case DumpActionEdit:
		dump, ok := m.selectedDumpView()
		if !ok {
			return DumpActionState{}
		}

		request := DumpActionRequest{
			Action: action,
			Path:   dump.RelativePath,
		}

		return DumpActionState{
			Pending: true,
			Request: request,
			Title:   "Edit dump",
			Details: []string{
				"Path: " + request.Path,
			},
		}
	default:
		return DumpActionState{}
	}
}
