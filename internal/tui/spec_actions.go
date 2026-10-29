package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
)

// SpecActionID identifies a spec action awaiting TUI confirmation.
type SpecActionID string

// Spec action identifiers define supported spec commands.
const (
	SpecActionGenerate        SpecActionID = "generate_spec"
	SpecActionGenerateFeature SpecActionID = "generate_acceptance_feature"
	SpecActionRefresh         SpecActionID = "refresh_specs"
	SpecActionCreate          SpecActionID = "create_spec"
	SpecActionEdit            SpecActionID = "edit_spec"
	SpecActionSave            SpecActionID = "save_spec"
)

// SpecActionExecutor runs confirmed spec actions.
type SpecActionExecutor interface {
	ExecuteSpecAction(request SpecActionRequest) (SpecActionResult, error)
}

// SpecActionRequest carries selected spec action input.
type SpecActionRequest struct {
	Action                                              SpecActionID
	Scope                                               QuizScope
	OperationID, Path, Revision, Title, Content, Status string
}

// SpecActionResult contains refreshed spec state after an action.
type SpecActionResult struct {
	Outcome                 operation.Result
	Message                 string
	Documents               []DocumentView
	Specs                   []SpecView
	Scope                   QuizScope
	SelectedPath            string
	Document                *SpecView
	Committed               bool
	AcceptanceFeatureStatus string
	AcceptanceFeaturePath   string
	AcceptanceFeatureRound  string
	AcceptanceFeatureReason string
}

// SpecActionState stores pending confirmation state for a spec action.
type SpecActionState struct {
	Pending bool
	Request SpecActionRequest
	Title   string
	Details []string
}

type specActionFinishedMsg struct {
	Binding operationBinding
	Request SpecActionRequest
	Result  SpecActionResult
	Err     error
}

func (m *Model) renderSpecAction() string {
	if !m.specAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Spec Action\n")
	fmt.Fprintf(&output, "  %s\n", m.specAction.Title)

	for _, detail := range m.specAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) updateSpecsKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "c":
		m.openSpecCreateForm()

		return nil
	case "e":
		m.openSpecEditForm()

		return nil
	case "g":
		cmd := m.startSpecAction(SpecActionRequest{
			Action: SpecActionGenerate,
			Scope:  m.selectedSpecScope(),
		})

		return cmd
	case "f":
		cmd := m.startSpecAction(SpecActionRequest{
			Action: SpecActionGenerateFeature,
			Scope:  m.selectedSpecScope(),
		})

		return cmd
	case "r":
		cmd := m.startSpecAction(SpecActionRequest{
			Action: SpecActionRefresh,
			Scope:  m.selectedSpecScope(),
		})

		return cmd
	case "esc":
		m.selectPrimaryPane()
	default:
		return m.updateNavigationKey(msg)
	}

	return nil
}

func (m *Model) updateSpecActionKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelSpecAction()
	case tea.KeyEnter:
		return m.confirmSpecAction()
	}

	return nil
}

func (m *Model) planSpecAction(action SpecActionID) {
	switch action {
	case SpecActionCreate:
		m.openSpecCreateForm()

		return
	case SpecActionEdit:
		m.openSpecEditForm()

		return
	}

	m.specAction = m.newSpecActionState(action)
	m.statusMessage = ""
}

func (m *Model) newSpecActionState(action SpecActionID) SpecActionState {
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

func (m *Model) executeSpecActionCommand(request SpecActionRequest) tea.Cmd {
	binding := m.beginOperation("spec")
	executor := m.specExecutor

	return func() tea.Msg {
		if executor == nil {
			return specActionFinishedMsg{
				Binding: binding,
				Request: request,
				Err:     fmt.Errorf("spec action executor is unavailable"),
			}
		}

		result, err := executor.ExecuteSpecAction(request)

		return specActionFinishedMsg{Binding: binding, Request: request, Result: result, Err: err}
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
	owned, apply := m.finishOperation(message.Binding)
	if !owned {
		return
	}

	m.specBusy = false

	if !apply {
		return
	}

	m.recordOperation(message.Result.Outcome)

	if message.Err != nil {
		m.statusMessage = specActionFailureStatus(message.Request.Action, message.Err)

		return
	}

	m.applySpecActionResult(message.Result)

	if message.Request.Action == SpecActionGenerate {
		m.offerGenerationResult(ViewSpecs, message.Result.SelectedPath, message.Result.Scope, false)
	}
}

func (m *Model) applySpecActionResult(result SpecActionResult) {
	if result.Documents != nil {
		m.documents = append([]DocumentView(nil), result.Documents...)
		m.selectedDocument = 0
	}

	if result.Specs != nil {
		selected, _ := m.selectedSpecView()

		path := result.SelectedPath
		if path == "" {
			path = selected.RelativePath
		}

		m.specs = normalizeSpecViews(result.Specs)
		if !m.selectSpecByPath(path) {
			m.selectSpecByScope(result.Scope)
		}
	}

	m.statusMessage = specActionResultMessage(result)
}

func specActionTitle(action SpecActionID) string {
	switch action {
	case SpecActionGenerate:
		return "Generate spec"
	case SpecActionGenerateFeature:
		return "Generate acceptance feature"
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
	case SpecActionGenerateFeature:
		return "Generating acceptance feature..."
	case SpecActionRefresh:
		return "Refreshing specs..."
	default:
		return "Running spec action..."
	}
}

func specActionFailureStatus(action SpecActionID, err error) string {
	switch action {
	case SpecActionGenerateFeature:
		return savedArtifactFailureStatus("Acceptance feature", "acceptance feature", err)
	default:
		return savedArtifactFailureStatus("Spec", "spec", err)
	}
}

func specActionResultMessage(result SpecActionResult) string {
	if result.AcceptanceFeatureStatus == "" {
		return result.Message
	}

	parts := []string{result.Message, "status: " + result.AcceptanceFeatureStatus}
	if result.AcceptanceFeaturePath != "" {
		parts = append(parts, "feature: "+result.AcceptanceFeaturePath)
	}

	if result.AcceptanceFeatureRound != "" {
		parts = append(parts, "feature round: "+result.AcceptanceFeatureRound)
	}

	if result.AcceptanceFeatureReason != "" {
		parts = append(parts, "reason: "+result.AcceptanceFeatureReason)
	}

	return strings.Join(parts, "  ")
}
