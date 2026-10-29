package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// Preview is a read before the separate execution confirmation. Escape and
// changed query context cannot dispatch the returned executable request.
func TestDashPreviewCancel(t *testing.T) {
	executor := &fakeWorkActionExecutor{result: WorkActionResult{Confirmation: &WorkActionState{
		Pending: true,
		Request: WorkActionRequest{Action: WorkActionWorkflowTicket, Workflow: &WorkflowActionRequest{}}, Title: "Exact policy",
	}}}
	model := NewModel(Options{
		Tickets: dashSelectionTickets("CHR-1"), WorkExecutor: executor,
		Observability: ObservabilityView{Authority: workspaceconfig.WorkStateAuthorityTickets},
	})
	model.selectView(ViewDash)

	command, handled := model.handleDashKey("v")
	if !handled || command == nil {
		t.Fatal("Preview Work is unreachable")
	}

	model.Update(command())

	if !executor.request.Workflow.Preview || !model.workAction.Pending {
		t.Fatal("preview executed work")
	}

	executor.called = false

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if executor.called || model.workAction.Pending {
		t.Fatal("cancel dispatched work")
	}

	model.planWorkAction(WorkActionWorkflowTicket)
	model.cycleAgentWorkOrder()

	if command := model.confirmWorkAction(); command != nil {
		t.Fatal("changed scope retained confirmation")
	}
}

// A response from a preview for an earlier queue scope cannot reopen its
// execution confirmation, even when the same ticket remains selected.
func TestDashPreviewReply(t *testing.T) {
	executor := &fakeWorkActionExecutor{result: WorkActionResult{Confirmation: &WorkActionState{Pending: true}}}
	model := NewModel(Options{
		Tickets: dashSelectionTickets("CHR-1"), WorkExecutor: executor,
		Observability: ObservabilityView{Authority: workspaceconfig.WorkStateAuthorityTickets},
	})
	model.selectView(ViewDash)

	command, _ := model.handleDashKey("v")
	model.cycleAgentWorkOrder()
	model.Update(command())

	if model.workAction.Pending {
		t.Fatal("late preview restored obsolete confirmation")
	}
}

// Workflow details use retained read-only status and reveal the detail pane;
// unsupported authority is rejected before a confirmation or service call.
func TestDashWorkflowDetails(t *testing.T) {
	executor := &fakeWorkActionExecutor{}
	model := NewModel(Options{WorkExecutor: executor, Observability: ObservabilityView{Authority: workspaceconfig.WorkStateAuthorityTickets}})
	model.selectView(ViewDash)

	model.workflow = WorkflowView{ID: "retained", Revision: 7}

	command, _ := model.handleDashKey("w")
	if command == nil || model.selectedPane != 1 {
		t.Fatal("Workflow details are unreachable")
	}

	model.Update(command())

	if executor.request.Action != WorkActionWorkflowStatus || executor.request.Workflow.ID != "retained" || executor.request.Workflow.ExpectedRevision != 7 {
		t.Fatal(executor.request)
	}

	model.observability.Authority = workspaceconfig.WorkStateAuthorityTrackerCompat
	for _, key := range []string{"v", "w"} {
		if command, _ := model.handleDashKey(key); command != nil || model.workAction.Pending {
			t.Fatal("unsupported authority invited work")
		}
	}
}

// Returning to the same queue values still begins a new interaction. A preview
// from the previous generation cannot reopen executable confirmation.
func TestDashPreviewGeneration(t *testing.T) {
	executor := &fakeWorkActionExecutor{result: WorkActionResult{Confirmation: &WorkActionState{Pending: true}}}
	model := NewModel(Options{
		Tickets: dashSelectionTickets("CHR-1"), WorkExecutor: executor,
		Observability: ObservabilityView{Authority: workspaceconfig.WorkStateAuthorityTickets},
	})
	model.selectView(ViewDash)
	model.cycleAgentWorkOrder()

	command, _ := model.handleDashKey("v")
	for range 5 {
		model.cycleAgentWorkOrder()
	}

	model.Update(command())

	if model.workAction.Pending {
		t.Fatal("previous generation reopened confirmation after returning to the same order")
	}
}
