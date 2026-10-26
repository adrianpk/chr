package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
)

type guidedTestExecutor func(GuidedRequest) (GuidedView, error)

func (execute guidedTestExecutor) ExecuteGuided(request GuidedRequest) (GuidedView, error) {
	return execute(request)
}

func guidedTerminalFixture() *Model {
	m := NewModel(Options{ProjectID: "project"})
	session := &core.AuthoringSession{ID: "guided-one", ProjectID: "project", Namespace: "default", Revision: 2, Topic: "Bounded audit", Guided: &core.GuidedUseCaseState{Stage: "quiz", Status: "waiting_input", Quiz: &core.GuidedQuizState{QuestionID: "first", InputDigest: "original"}}}
	m.guided = guidedState{Active: true, View: GuidedView{Session: session, Project: "Example", Namespace: "default", InputAction: "answer", Prompt: "What should remain bounded?"}}

	return m
}

// TestGuidedUseCaseTUIStaleInput freezes the question before editing; a newer
// observation cannot rebind a draft, and rejection keeps its complete text.
func TestGuidedUseCaseTUIStaleInput(t *testing.T) {
	m := guidedTerminalFixture()

	var submitted GuidedRequest

	m.guidedExecutor = guidedTestExecutor(func(request GuidedRequest) (GuidedView, error) {
		submitted = request

		return GuidedView{}, core.ErrWorkflowConflict
	})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = guidedTestModel(t, updated)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Original unsent answer.")})
	m = guidedTestModel(t, updated)
	newer := *m.guided.View.Session
	newer.Revision++
	m.completeGuided(guidedFinishedMsg{Sequence: m.guided.Sequence, Request: GuidedRequest{Action: "inspect"}, View: GuidedView{Session: &newer, Namespace: "default", InputAction: "answer"}})

	updated, command := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	m = guidedTestModel(t, updated)
	if command == nil || !m.guided.Busy {
		t.Fatal("input was not frozen before IO")
	}

	updated, _ = m.Update(command())

	m = guidedTestModel(t, updated)
	if submitted.Command.ExpectedRevision != 2 || submitted.Command.QuestionID != "first" || submitted.Command.Content != "Original unsent answer." {
		t.Fatal("refresh rebound the unsent draft", submitted)
	}

	if !m.guided.Editing || m.guided.Input.Value() != submitted.Command.Content || m.guided.Pending == nil {
		t.Fatal("failed input discarded its draft or original operation")
	}
}

// TestGuidedUseCaseTUIReplay retries the exact failed request and ignores an
// earlier read after the guided screen has closed and reopened.
func TestGuidedUseCaseTUIReplay(t *testing.T) {
	m := guidedTerminalFixture()
	m.guided.View.Controls = []string{"finish"}

	var requests []GuidedRequest

	m.guidedExecutor = guidedTestExecutor(func(request GuidedRequest) (GuidedView, error) {
		requests = append(requests, request)
		if len(requests) == 1 {
			return GuidedView{}, errors.New("lost acknowledgement")
		}

		return m.guided.View, nil
	})
	command := m.submitGuidedStep("finish")
	m.completeGuided(guidedTestFinished(t, command()))

	original := *m.guided.Pending
	updated, retry := m.guidedActionKey("y")

	m = guidedTestModel(t, updated)
	if retry == nil {
		t.Fatal("original request cannot retry")
	}

	m.completeGuided(guidedTestFinished(t, retry()))

	if len(requests) != 2 || requests[1] != original || m.guided.Pending != nil {
		t.Fatal("retry replaced its operation or arguments")
	}

	old := m.openGuided()
	oldResult := old()
	current := m.openGuided()
	m.completeGuided(guidedTestFinished(t, oldResult))

	if !m.guided.Busy {
		t.Fatal("obsolete read replaced the reopened screen")
	}

	m.completeGuided(guidedTestFinished(t, current()))
}

// TestGuidedUseCaseTUIOriginalScroll exposes the final bytes of an original
// larger than the viewport without clipping the stored document or active input.
func TestGuidedUseCaseTUIOriginalScroll(t *testing.T) {
	m := guidedTerminalFixture()
	content := strings.Repeat("Full original line.\n", 100) + "Final original bytes."
	m.guided.View.Document = &GuidedDocument{Kind: "dump", Commit: strings.Repeat("a", 40), Content: content}

	m.guided.Scroll = 10000
	if !strings.Contains(m.renderGuidedWorkspace(90, 25), "Final original bytes.") || m.guided.View.Document.Content != content {
		t.Fatal("the complete original cannot be inspected")
	}
}

func guidedTestModel(t *testing.T, model tea.Model) *Model {
	t.Helper()

	value, ok := model.(*Model)
	if !ok {
		t.Fatal("unexpected terminal model")
	}

	return value
}

func guidedTestFinished(t *testing.T, message tea.Msg) guidedFinishedMsg {
	t.Helper()

	finished, ok := message.(guidedFinishedMsg)
	if !ok {
		t.Fatal("missing guided completion")
	}

	return finished
}
