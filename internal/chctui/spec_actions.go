package chctui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type SpecActionID string

const (
	SpecActionGenerate SpecActionID = "generate_spec"
	SpecActionRefresh  SpecActionID = "refresh_specs"
)

type SpecActionExecutor interface {
	ExecuteSpecAction(request SpecActionRequest) (SpecActionResult, error)
}

type SpecActionRequest struct {
	Action SpecActionID
	Scope  QuizScope
}

type SpecActionResult struct {
	Message   string
	Documents []DocumentView
	Specs     []SpecView
	Scope     QuizScope
}

type SpecActionState struct {
	Pending bool
	Request SpecActionRequest
	Title   string
	Details []string
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
		m.planSpecAction(SpecActionGenerate)

		return m, m.confirmSpecAction()
	case "r":
		m.planSpecAction(SpecActionRefresh)

		return m, m.confirmSpecAction()
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

	if m.specExecutor == nil {
		m.statusMessage = "Spec action executor is unavailable."

		return nil
	}

	result, err := m.specExecutor.ExecuteSpecAction(request)
	if err != nil {
		m.statusMessage = "Spec action failed: " + err.Error()

		return nil
	}

	if result.Documents != nil {
		m.documents = append([]DocumentView(nil), result.Documents...)
		m.selectedDocument = 0
	}

	if result.Specs != nil {
		m.specs = normalizeSpecViews(result.Specs)
		m.selectSpecByScope(result.Scope)
	}

	m.statusMessage = result.Message

	return nil
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
