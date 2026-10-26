package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
)

// TestGuidedUseCaseTUIControls freezes the displayed native revision and keeps
// an unsent draft when a separate control completes; approval reuses its owner.
func TestGuidedUseCaseTUIControls(t *testing.T) {
	m := guidedTerminalFixture()
	m.guided.View.NativeRevision = 17
	m.guided.View.Controls = []string{"pause", "cancel"}
	m.editGuidedInput("answer")
	m.guided.Input.SetValue("Unsent original answer.")

	m.guided.Editing = false

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

	if request.Command.NativeRevision != 17 || request.Command.Action != "pause" || m.guided.Input.Value() != "Unsent original answer." {
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
	original := GuidedRequest{Action: "input", Command: core.GuidedUseCaseCommand{OperationID: "original-control", Action: "resume", ExpectedRevision: 4, NativeRevision: 17}}
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
	m.guided.View.Session = nil

	m.guided.View.Runs = []core.AuthoringSession{retained}
	if !strings.Contains(m.View(), "inspect results") || strings.Contains(m.View(), "completed") {
		t.Fatal("catalog claimed completion without inspecting receipts")
	}
}
