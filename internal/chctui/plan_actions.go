package chctui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type PlanActionID string

const (
	PlanActionGenerate   PlanActionID = "generate_missing_plan"
	PlanActionRegenerate PlanActionID = "regenerate_plan"
	PlanActionRefresh    PlanActionID = "refresh_plans"
)

type PlanActionExecutor interface {
	ExecutePlanAction(request PlanActionRequest) (PlanActionResult, error)
}

type PlanActionRequest struct {
	Action PlanActionID
	Scope  QuizScope
}

type PlanActionResult struct {
	Message   string
	Documents []DocumentView
	Plans     []PlanView
	Scope     QuizScope
}

type PlanActionState struct {
	Pending bool
	Request PlanActionRequest
	Title   string
	Details []string
}

type planActionFinishedMsg struct {
	Request PlanActionRequest
	Result  PlanActionResult
	Err     error
}

func (m Model) renderPlanAction() string {
	if !m.planAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Plan Action\n")
	output.WriteString(fmt.Sprintf("  %s\n", m.planAction.Title))

	for _, detail := range m.planAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) updatePlansKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "g":
		cmd := m.startPlanAction(PlanActionRequest{
			Action: PlanActionGenerate,
			Scope:  m.selectedPlanScope(),
		})

		return m, cmd
	case "G":
		m.planPlanAction(PlanActionRegenerate)

		return m, nil
	case "r":
		cmd := m.startPlanAction(PlanActionRequest{
			Action: PlanActionRefresh,
			Scope:  m.selectedPlanScope(),
		})

		return m, cmd
	case "esc":
		m.selectPrimaryPane()
	default:
		return m.updateNavigationKey(msg)
	}

	return m, nil
}

func (m Model) updatePlanActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelPlanAction()
	case tea.KeyEnter:
		return m, m.confirmPlanAction()
	}

	return m, nil
}

func (m *Model) planPlanAction(action PlanActionID) {
	m.planAction = m.newPlanActionState(action)
	m.statusMessage = ""
}

func (m Model) newPlanActionState(action PlanActionID) PlanActionState {
	scope := m.selectedPlanScope()

	return PlanActionState{
		Pending: true,
		Request: PlanActionRequest{
			Action: action,
			Scope:  scope,
		},
		Title: planActionTitle(action),
		Details: []string{
			"Use Case: " + quizUseCaseLabel(scope),
			"Target: Plans",
		},
	}
}

func (m *Model) cancelPlanAction() {
	m.planAction = PlanActionState{}
	m.statusMessage = "Plan action cancelled."
}

func (m *Model) confirmPlanAction() tea.Cmd {
	request := m.planAction.Request
	m.planAction = PlanActionState{}

	return m.startPlanAction(request)
}

func (m Model) executePlanActionCommand(request PlanActionRequest) tea.Cmd {
	executor := m.planExecutor

	return func() tea.Msg {
		if executor == nil {
			return planActionFinishedMsg{
				Request: request,
				Err:     fmt.Errorf("plan action executor is unavailable"),
			}
		}

		result, err := executor.ExecutePlanAction(request)

		return planActionFinishedMsg{Request: request, Result: result, Err: err}
	}
}

func (m *Model) startPlanAction(request PlanActionRequest) tea.Cmd {
	if m.planBusy {
		m.statusMessage = "Plan action already running."

		return nil
	}

	if m.planExecutor == nil {
		m.statusMessage = "Plan action executor is unavailable."

		return nil
	}

	if request.Action == PlanActionGenerate && m.planExistsForScope(request.Scope) {
		m.statusMessage = "Plan already exists. Use regenerate plan to replace it."

		return nil
	}

	m.planBusy = true
	m.statusMessage = planBusyMessage(request)

	return m.executePlanActionCommand(request)
}

func (m *Model) completePlanAction(message planActionFinishedMsg) {
	m.planBusy = false

	if message.Err != nil {
		m.statusMessage = "Plan action failed; showing last saved plan if available: " + message.Err.Error()

		return
	}

	m.applyPlanActionResult(message.Result)
}

func (m *Model) applyPlanActionResult(result PlanActionResult) {
	if result.Documents != nil {
		m.documents = append([]DocumentView(nil), result.Documents...)
		m.selectedDocument = 0
	}

	if result.Plans != nil {
		m.plans = normalizePlanViews(result.Plans)
		m.selectPlanByScope(result.Scope)
	}

	m.statusMessage = result.Message
}

func planActionTitle(action PlanActionID) string {
	switch action {
	case PlanActionGenerate:
		return "Generate plan"
	case PlanActionRegenerate:
		return "Regenerate plan"
	case PlanActionRefresh:
		return "Refresh plans"
	default:
		return "Plan action"
	}
}

func planBusyMessage(request PlanActionRequest) string {
	switch request.Action {
	case PlanActionGenerate:
		return "Generating plan..."
	case PlanActionRegenerate:
		return "Regenerating plan..."
	case PlanActionRefresh:
		return "Refreshing plans..."
	default:
		return "Running plan action..."
	}
}
