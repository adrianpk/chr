package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// GuidedExecutor maps terminal requests to the shared guided application owner.
type GuidedExecutor interface {
	ExecuteGuided(GuidedRequest) (GuidedView, error)
}

// GuidedRequest freezes a selected run, input revision and operation before IO.
type GuidedRequest struct {
	Action, Document, CaseSlug string
	QuizAction                 QuizActionID
	ReloadQuiz                 bool
	Command                    authoringcore.GuidedUseCaseCommand
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
	Quiz                                    *QuizView
	Session                                 *authoringcore.Session
	Runs                                    []authoringcore.Session
	Cases                                   []authoringcore.CaseMetadata
	Policy, Problems, Progress              []string
	Document                                *GuidedDocument
	Originals                               []GuidedDocument
}

type guidedState struct {
	Active, Busy bool
	Sequence     uint64
	View         GuidedView
	Pending      *GuidedRequest
	Problem      string
}

type guidedFinishedMsg struct {
	Sequence uint64
	Request  GuidedRequest
	View     GuidedView
	Err      error
}

func (m *Model) openGuided() tea.Cmd {
	request := GuidedRequest{Action: "open", Command: authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{ProjectID: m.projectID, RunID: m.caseContext.RunID}}}
	if metadata, ok := m.selectedCaseMetadata(); ok {
		request.CaseSlug = metadata.Slug
	}

	m.guided = guidedState{Active: true}

	return m.startGuidedRequest(request)
}

func (m *Model) startGuidedRequest(request GuidedRequest) tea.Cmd {
	if m.guided.Busy || m.guidedExecutor == nil || !m.wizardRequestCurrent(request) {
		return nil
	}

	m.guidedSequence++

	m.guided.Sequence, m.guided.Busy = m.guidedSequence, true
	if request.Action == "input" || request.Action == "start" {
		m.guided.Pending = &request
		m.guided.Problem = ""
	}

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
	if !m.guidedReplyCurrent(message) {
		return
	}

	if message.Err != nil {
		m.failGuidedRequest(message)

		return
	}

	if retainGuidedOriginal(message.Request, m.guided.View, message.View) {
		message.View.Document = m.guided.View.Document
	}

	m.completeWizardDraft(message)

	m.guided.View = message.View
	m.observeGuidedCase(message.View)

	m.finishGuidedRequest(message.Request)

	m.navigateGuidedReply(message.Request)
}

func (m *Model) observeGuidedCase(view GuidedView) {
	if view.Cases != nil && !m.caseDirty && !m.caseEdit.Active {
		m.cases = append([]authoringcore.CaseMetadata(nil), view.Cases...)
		if caseSliceContainsSlug(m.cases, m.caseContext.CaseSlug) {
			m.selectedCase = m.caseIndexForSlug(m.caseContext.CaseSlug)
		} else {
			m.clampSelectedCase()
		}
	}

	if session := view.Session; session != nil && session.Guided != nil {
		m.bindCaseContext(session.Guided.CaseSlug, session.ID)
	}
}

func retainGuidedOriginal(request GuidedRequest, before, after GuidedView) bool {
	return request.Action == "inspect" && sameGuidedTerminalRun(before, after) && before.Document != nil && before.Document.Commit != ""
}

func sameGuidedTerminalRun(before, after GuidedView) bool {
	return before.Session != nil && after.Session != nil && before.Session.ID == after.Session.ID && before.Session.ProjectID == after.Session.ProjectID && before.Session.Namespace == after.Session.Namespace
}

func (m *Model) guidedRequest(action string) GuidedRequest {
	command := authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{ProjectID: m.projectID, Namespace: m.guided.View.Namespace}}

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

func (m *Model) navigateGuidedReply(request GuidedRequest) {
	if request.ReloadQuiz && m.wizardQuizCurrent() {
		m.startQuizAnswer()
	}

	if request.Action == "start" || request.Action == "open" || request.Action == "input" {
		m.selectView(m.wizardStageView())
	}

	if request.Action == "document" {
		m.selectView(wizardDocumentView(request.Document))
	}
}
