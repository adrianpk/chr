package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
)

// PlanActionID identifies a plan action awaiting TUI confirmation.
type PlanActionID string

// Plan action identifiers define supported plan commands.
const (
	PlanActionGenerate            PlanActionID = "generate_missing_plan"
	PlanActionRegenerate          PlanActionID = "regenerate_plan"
	PlanActionRefresh             PlanActionID = "refresh_plans"
	PlanActionMaterializeTracking PlanActionID = "materialize_plan_tracking"
)

// PlanActionExecutor runs confirmed plan actions.
type PlanActionExecutor interface {
	ExecutePlanAction(request PlanActionRequest) (PlanActionResult, error)
}

// PlanActionRequest carries selected plan action input.
type PlanActionRequest struct {
	Action PlanActionID
	Scope  QuizScope
}

// PlanActionResult contains refreshed plan state after an action.
type PlanActionResult struct {
	Message   string
	Documents []DocumentView
	Plans     []PlanView
	Tickets   []core.IndexedTicket
	Scope     QuizScope
}

// PlanActionState stores pending confirmation state for a plan action.
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
	if m.planBusy {
		return strings.TrimRight("\n\nPlan Action Running\n  "+m.statusMessage+"\n\n  Please wait for this action to finish.", "\n")
	}

	if !m.planAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Plan Action\n")
	fmt.Fprintf(&output, "  %s\n", m.planAction.Title)

	for _, detail := range m.planAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	output.WriteString("\n  Enter: confirm\n")
	output.WriteString("  Esc: cancel\n")

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
		if m.planBusy {
			m.statusMessage = "Plan action already running."

			return m, nil
		}

		m.planPlanAction(PlanActionRegenerate)

		return m, nil
	case "alt+t":
		if m.planBusy {
			m.statusMessage = "Plan action already running."

			return m, nil
		}

		if !m.selectedPlanCanMaterializeTracking() {
			m.statusMessage = "Plan tracking requires an existing plan."

			return m, nil
		}

		m.planPlanAction(PlanActionMaterializeTracking)

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
	m.selectedPane = 1
	m.maximizedPane = 1
	m.statusMessage = "Confirm plan action with Enter, or cancel with Esc."
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
	m.maximizedPane = -1
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

	if planActionChangesContent(request.Action) {
		m.selectedPane = 1
		m.maximizedPane = 1
	}

	m.planBusy = true
	m.statusMessage = planBusyMessage(request)

	return m.executePlanActionCommand(request)
}

func (m *Model) completePlanAction(message planActionFinishedMsg) {
	m.planBusy = false

	if message.Err != nil {
		m.statusMessage = savedArtifactFailureStatus("Plan", "plan", message.Err)

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

	if result.Tickets != nil {
		m.tickets = NewTicketViews(result.Tickets)
		m.selectFirstFilteredTicket()
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
	case PlanActionMaterializeTracking:
		return "Create plan tracking"
	default:
		return "Plan action"
	}
}

func planActionChangesContent(action PlanActionID) bool {
	switch action {
	case PlanActionGenerate, PlanActionRegenerate, PlanActionMaterializeTracking:
		return true
	default:
		return false
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
	case PlanActionMaterializeTracking:
		return "Creating plan tracking..."
	default:
		return "Running plan action..."
	}
}
