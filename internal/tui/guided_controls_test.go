package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// TestGuidedUseCaseTUIControls freezes the displayed native revision and keeps
// an unsent draft when a separate control completes; approval reuses its owner.
func TestGuidedUseCaseTUIControls(t *testing.T) {
	m := guidedTerminalFixture()
	m.guided.View.NativeRevision = 17
	m.guided.View.Controls = []string{"pause", "cancel"}
	m.editGuidedInput("answer")
	m.quizAnswer.Input.SetValue("Unsent original answer.")

	m.quizAnswer.Active = false

	var request GuidedRequest

	m.guidedExecutor = guidedTestExecutor(func(input GuidedRequest) (GuidedView, error) {
		request = input

		return m.guided.View, nil
	})
	updated, command := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})

	m = guidedTestModel(t, updated)
	if command == nil {
		t.Fatal("available pause did not dispatch")
	}

	m.completeGuided(guidedTestFinished(t, command()))

	if request.Command.NativeRevision != 17 || request.Command.Action != "pause" || m.quizAnswer.Input.Value() != "Unsent original answer." {
		t.Fatal("control replaced its native binding or unsent draft")
	}

	m.guided.View.CanApprove = true
	m.guided.View.Native = &WorkflowView{ID: "retained-native", Revision: 19, ApprovalHash: "displayed-review"}
	m.guidedExtraKey("a")

	approval := m.workAction.Request.Workflow
	if m.guided.Active || approval == nil || approval.ID != "retained-native" || approval.ExpectedRevision != 19 || approval.BindingHash != "displayed-review" {
		t.Fatal("guided approval bypassed the existing exact native confirmation")
	}
}

// TestGuidedUseCaseTUIRetainedStep submits the canonical pending operation
// unchanged rather than allocating another command or refilling its allowance.
func TestGuidedUseCaseTUIRetainedStep(t *testing.T) {
	m := guidedTerminalFixture()
	original := GuidedRequest{Action: "input", Command: authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{RunID: m.guided.View.Session.ID, ProjectID: m.projectID}, OperationID: "original-control", Action: "resume", ExpectedRevision: 4, NativeRevision: 17}}
	m.guided.View.Retry = &original

	var received GuidedRequest

	m.guidedExecutor = guidedTestExecutor(func(request GuidedRequest) (GuidedView, error) {
		received = request

		return m.guided.View, nil
	})

	command := m.guidedExtraKey("c")
	if command == nil {
		t.Fatal("canonical retry was not offered")
	}

	m.completeGuided(guidedTestFinished(t, command()))

	if received != original {
		t.Fatal("canonical retry refreshed its original command")
	}
}

// Catalog inspection cannot label unverified terminal state as delivered work.
func TestGuidedUseCaseTUICatalogResult(t *testing.T) {
	m := guidedTerminalFixture()
	retained := *m.guided.View.Session
	retained.Guided.Status = "completed"
	retained.Guided.Stage = "results"
	m.guided.View.Session = &retained
	m.guided.View.Status = "completed"
	m.selectView(ViewBoard)

	if !strings.Contains(m.View(), "inspect results") || strings.Contains(m.View(), "completed") {
		t.Fatal("catalog claimed completion without inspecting receipts")
	}
}
