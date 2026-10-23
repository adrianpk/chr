package tui

import (
	"crypto/rand"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
)

// GuidedExecutor maps terminal requests to the shared guided application owner.
type GuidedExecutor interface {
	ExecuteGuided(GuidedRequest) (GuidedView, error)
}

// GuidedRequest freezes a selected run, input revision and operation before IO.
type GuidedRequest struct {
	Action, Document string
	Command          core.GuidedUseCaseCommand
}

// GuidedDocument is a complete original, verified by the application reader.
type GuidedDocument struct {
	Kind, Reference, Commit, Content string
}

// GuidedView contains shared observations and the next available human input.
type GuidedView struct {
	Project, Namespace, Prompt, InputAction string
	Status                                  string
	NativeRevision                          uint64
	Controls                                []string
	Retry                                   *GuidedRequest
	Native                                  *WorkflowView
	CanRetainIdea, CanApprove               bool
	Session                                 *core.AuthoringSession
	Runs                                    []core.AuthoringSession
	Policy, Problems, Progress              []string
	Document                                *GuidedDocument
}

type guidedState struct {
	Active, Busy, Editing bool
	Selected, Scroll      int
	Sequence              uint64
	View                  GuidedView
	Input                 textarea.Model
	Bound                 GuidedRequest
	Pending               *GuidedRequest
	Problem               string
}

type guidedFinishedMsg struct {
	Sequence uint64
	Request  GuidedRequest
	View     GuidedView
	Err      error
}

func (m *Model) openGuided() tea.Cmd {
	m.guided = guidedState{Active: true}

	return m.startGuidedRequest(GuidedRequest{Action: "inspect", Command: core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{ProjectID: m.projectID}}})
}

func (m *Model) startGuidedRequest(request GuidedRequest) tea.Cmd {
	if m.guided.Busy || m.guidedExecutor == nil {
		return nil
	}

	m.guidedSequence++
	m.guided.Sequence, m.guided.Busy = m.guidedSequence, true
	m.guided.Pending = &request
	m.guided.Problem = ""
	sequence, executor := m.guided.Sequence, m.guidedExecutor

	return func() tea.Msg {
		view, err := executor.ExecuteGuided(request)

		return guidedFinishedMsg{Sequence: sequence, Request: request, View: view, Err: err}
	}
}

func (m *Model) completeGuided(message guidedFinishedMsg) {
	if !m.guided.Active || message.Sequence != m.guided.Sequence {
		return
	}

	m.guided.Busy = false
	if message.Err != nil {
		m.guided.Problem = message.Err.Error()

		return
	}

	if retainGuidedOriginal(message.Request, m.guided.View, message.View) {
		message.View.Document = m.guided.View.Document
	}

	m.guided.View = message.View

	m.guided.Pending, m.guided.Problem = nil, ""
	if message.Request.Command.OperationID == m.guided.Bound.Command.OperationID && message.Request.Action == "input" {
		m.guided.Editing = false
		m.guided.Input = textarea.Model{}
	}

	if message.Request.Action != "inspect" {
		m.guided.Scroll = 0
	}
}

func retainGuidedOriginal(request GuidedRequest, before, after GuidedView) bool {
	return request.Action == "inspect" && sameGuidedTerminalRun(before, after) && before.Document != nil && before.Document.Commit != ""
}

func sameGuidedTerminalRun(before, after GuidedView) bool {
	return before.Session != nil && after.Session != nil && before.Session.ID == after.Session.ID && before.Session.ProjectID == after.Session.ProjectID && before.Session.Namespace == after.Session.Namespace
}

func (m Model) guidedRequest(action string) GuidedRequest {
	command := core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{ProjectID: m.projectID, Namespace: m.guided.View.Namespace}}

	command.NativeRevision = m.guided.View.NativeRevision
	if session := m.guided.View.Session; session != nil {
		command.RunID, command.ExpectedRevision = session.ID, session.Revision
		if quiz := session.Guided.Quiz; quiz != nil {
			command.QuestionID, command.InputDigest = quiz.QuestionID, quiz.InputDigest
		}
	}

	return GuidedRequest{Action: action, Command: command}
}

func (m *Model) refreshGuided() tea.Cmd {
	if m.guided.Busy || m.guided.Problem != "" {
		return nil
	}

	return m.startGuidedRequest(m.guidedRequest("inspect"))
}

func (m *Model) startGuidedRun() tea.Cmd {
	request := m.guidedRequest("start")
	request.Command.RunID = "guided-" + strings.ToLower(rand.Text())
	request.Command.ExpectedRevision = 0
	m.guided.Editing = false

	return m.startGuidedRequest(request)
}
