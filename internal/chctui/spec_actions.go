package chctui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SpecActionID identifies a spec action awaiting TUI confirmation.
type SpecActionID string

// Spec action identifiers define supported spec commands.
const (
	SpecActionGenerate SpecActionID = "generate_spec"
	SpecActionRefresh  SpecActionID = "refresh_specs"
)

// SpecActionExecutor runs confirmed spec actions.
type SpecActionExecutor interface {
	ExecuteSpecAction(request SpecActionRequest) (SpecActionResult, error)
}

// SpecActionRequest carries selected spec action input.
type SpecActionRequest struct {
	Action SpecActionID
	Scope  QuizScope
}

// SpecActionResult contains refreshed spec state after an action.
type SpecActionResult struct {
	Message   string
	Documents []DocumentView
	Specs     []SpecView
	Scope     QuizScope
}

// SpecActionState stores pending confirmation state for a spec action.
type SpecActionState struct {
	Pending bool
	Request SpecActionRequest
	Title   string
	Details []string
}

type specActionFinishedMsg struct {
	Request SpecActionRequest
	Result  SpecActionResult
	Err     error
}

func (m Model) renderSpecAction() string {
	if !m.specAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Spec Action\n")
	output.WriteString(fmt.Sprintf("  %s\n", m.specAction.Title))

	for _, detail := range m.specAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) updateSpecsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "g":
		cmd := m.startSpecAction(SpecActionRequest{
			Action: SpecActionGenerate,
			Scope:  m.selectedSpecScope(),
		})

		return m, cmd
	case "r":
		cmd := m.startSpecAction(SpecActionRequest{
			Action: SpecActionRefresh,
			Scope:  m.selectedSpecScope(),
		})

		return m, cmd
	case "esc":
		m.selectPrimaryPane()
	default:
		return m.updateNavigationKey(msg)
	}

	return m, nil
}

func (m Model) updateSpecActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelSpecAction()
	case tea.KeyEnter:
		return m, m.confirmSpecAction()
	}

	return m, nil
}

func (m *Model) planSpecAction(action SpecActionID) {
	m.specAction = m.newSpecActionState(action)
	m.statusMessage = ""
}

func (m Model) newSpecActionState(action SpecActionID) SpecActionState {
	scope := m.selectedSpecScope()

	return SpecActionState{
		Pending: true,
		Request: SpecActionRequest{
			Action: action,
			Scope:  scope,
		},
		Title: specActionTitle(action),
		Details: []string{
			"Use Case: " + quizUseCaseLabel(scope),
			"Target: Specs",
		},
	}
}

func (m *Model) cancelSpecAction() {
	m.specAction = SpecActionState{}
	m.statusMessage = "Spec action cancelled."
}

func (m *Model) confirmSpecAction() tea.Cmd {
	request := m.specAction.Request
	m.specAction = SpecActionState{}

	return m.startSpecAction(request)
}

func (m Model) executeSpecActionCommand(request SpecActionRequest) tea.Cmd {
	executor := m.specExecutor

	return func() tea.Msg {
		if executor == nil {
			return specActionFinishedMsg{
				Request: request,
				Err:     fmt.Errorf("spec action executor is unavailable"),
			}
		}

		result, err := executor.ExecuteSpecAction(request)

		return specActionFinishedMsg{Request: request, Result: result, Err: err}
	}
}

func (m *Model) startSpecAction(request SpecActionRequest) tea.Cmd {
	if m.specExecutor == nil {
		m.statusMessage = "Spec action executor is unavailable."

		return nil
	}

	m.specBusy = true
	m.statusMessage = specBusyMessage(request)

	return m.executeSpecActionCommand(request)
}

func (m *Model) completeSpecAction(message specActionFinishedMsg) {
	m.specBusy = false

	if message.Err != nil {
		m.statusMessage = savedArtifactFailureStatus("Spec", "spec", message.Err)

		return
	}

	m.applySpecActionResult(message.Result)
}

func (m *Model) applySpecActionResult(result SpecActionResult) {
	if result.Documents != nil {
		m.documents = append([]DocumentView(nil), result.Documents...)
		m.selectedDocument = 0
	}

	if result.Specs != nil {
		m.specs = normalizeSpecViews(result.Specs)
		m.selectSpecByScope(result.Scope)
	}

	m.statusMessage = result.Message
}

func specActionTitle(action SpecActionID) string {
	switch action {
	case SpecActionGenerate:
		return "Generate spec"
	case SpecActionRefresh:
		return "Refresh specs"
	default:
		return "Spec action"
	}
}

func specBusyMessage(request SpecActionRequest) string {
	switch request.Action {
	case SpecActionGenerate:
		return "Generating spec..."
	case SpecActionRefresh:
		return "Refreshing specs..."
	default:
		return "Running spec action..."
	}
}
