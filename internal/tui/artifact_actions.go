package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// ArtifactActionID identifies an artifact action awaiting TUI confirmation.
type ArtifactActionID string

// Artifact action identifiers define the supported artifact commands.
const (
	ArtifactActionCreateAuthoringSession ArtifactActionID = "create_authoring_session"
	ArtifactActionCreateContextQuiz      ArtifactActionID = "create_context_quiz"
)

// ArtifactActionExecutor runs confirmed artifact actions.
type ArtifactActionExecutor interface {
	ExecuteArtifactAction(action ArtifactActionID) (ArtifactActionResult, error)
}

// ArtifactActionResult contains refreshed artifact-facing state after an action.
type ArtifactActionResult struct {
	Outcome           operation.Result
	Message           string
	Tickets           []ticketcore.Indexed
	Documents         []DocumentView
	AuthoringSessions []authoringcore.Session
}

// ArtifactActionState stores the pending confirmation state for an artifact
// action.
type ArtifactActionState struct {
	Pending bool
	Action  ArtifactActionID
	Title   string
	Details []string
}

// NewArtifactActionState builds the confirmation details for an artifact
// action.
func NewArtifactActionState(action ArtifactActionID) ArtifactActionState {
	switch action {
	case ArtifactActionCreateAuthoringSession:
		return ArtifactActionState{
			Pending: true,
			Action:  action,
			Title:   "Create authoring session",
			Details: []string{
				"Topic: TUI artifact creation",
				"Target: spec",
				"Mode: tui",
			},
		}
	default:
		return ArtifactActionState{}
	}
}

func (m *Model) renderArtifactAction() string {
	if !m.artifactAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Action\n")
	fmt.Fprintf(&output, "  %s\n", m.artifactAction.Title)

	for _, detail := range m.artifactAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) updateArtifactActionKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelArtifactAction()
	case tea.KeyEnter:
		m.confirmArtifactAction()
	}

	return nil
}

func (m *Model) planArtifactAction(action ArtifactActionID) {
	if action == ArtifactActionCreateContextQuiz {
		m.artifactAction = ArtifactActionState{}
		m.recordOperation(operation.Result{Status: operation.Blocked, Summary: "Context-aware authoring quiz is unavailable in this TUI.", BlockedReason: "Use chr author quiz new for a contextual authoring document, or F3 to start project or case questions."})

		return
	}

	m.artifactAction = NewArtifactActionState(action)
	m.statusMessage = ""
}

func (m *Model) cancelArtifactAction() {
	m.artifactAction = ArtifactActionState{}
	m.statusMessage = "Action cancelled."
}

func (m *Model) confirmArtifactAction() {
	action := m.artifactAction.Action
	m.artifactAction = ArtifactActionState{}

	if m.actionExecutor == nil {
		m.statusMessage = "Artifact action executor is unavailable."

		return
	}

	result, err := m.actionExecutor.ExecuteArtifactAction(action)
	m.recordOperation(result.Outcome)

	if err != nil {
		m.statusMessage = "Action failed: " + err.Error()

		return
	}

	m.applyArtifactActionResult(result)
}

func (m *Model) applyArtifactActionResult(result ArtifactActionResult) {
	if result.Tickets != nil {
		m.replaceTickets(result.Tickets)
	}

	if result.Documents != nil {
		m.documents = append([]DocumentView(nil), result.Documents...)
		m.selectedDocument = 0
	}

	if result.AuthoringSessions != nil {
		m.authoringSessions = NewAuthoringSessionViews(result.AuthoringSessions)
		m.selectedAuthoringSession = 0
	}

	m.statusMessage = result.Message
}
