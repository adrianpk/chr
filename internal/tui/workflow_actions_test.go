package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestModelWorkflowBinding freezes service-supplied identity, revision, snapshot
// and approval hash through refresh and keyboard confirmation.
func TestModelWorkflowBinding(t *testing.T) {
	for _, action := range []WorkActionID{WorkActionWorkflowSlice, WorkActionWorkflowApprove} {
		t.Run(string(action), func(t *testing.T) {
			native := &WorkflowActionRequest{ID: "run-selected", ExpectedRevision: 7, BindingHash: strings.Repeat("a", 64), Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice", Identity: "selected-slice"}, SnapshotHash: strings.Repeat("b", 64)}, Selected: &workflowstate.SelectionGuard{Key: "CHR-1", Revision: 3}}
			executor := &fakeWorkActionExecutor{}
			model := NewModel(Options{WorkExecutor: executor})
			model.workAction = WorkActionState{Pending: true, Request: WorkActionRequest{Action: action, Workflow: native}}
			model.workflow = WorkflowView{ID: "run-refreshed", Revision: 99, ApprovalHash: strings.Repeat("c", 64)}
			updated, command := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			current := requireUpdatedModel(t, updated)
			if command == nil {
				t.Fatal("confirmation did not dispatch")
			}

			message := command()

			if !reflect.DeepEqual(executor.request.Workflow, native) {
				t.Fatal("confirmation replaced the displayed binding", executor.request)
			}

			current.Update(message)
		})
	}
}

// TestModelWorkflowConflict renders the service failure and refreshed status,
// without losing retained progress or converting a failure into confirmation.
func TestModelWorkflowConflict(t *testing.T) {
	model := NewModel(Options{})
	model.workflowExecutionBusy = true
	request := WorkActionRequest{Action: WorkActionWorkflowContinue, Workflow: &WorkflowActionRequest{ID: "run-one", ExpectedRevision: 2}}
	current := WorkflowView{ID: "run-one", Revision: 3, Lines: []string{"State: waiting_capacity", "Host-validated tasks: 1/3; confirmed tickets: 0", "Next action owner: current_owner"}}
	updated, _ := model.Update(workActionFinishedMsg{Request: request, Result: WorkActionResult{Workflow: &current}, Err: errors.New("workflow revision conflict")})

	model = requireUpdatedModel(t, updated)
	if model.workflowExecutionBusy || model.workAction.Pending || model.workflow.Revision != 3 || !strings.Contains(model.statusMessage, "workflow revision conflict") {
		t.Fatal("failed action hid stale conflict or changed confirmation state")
	}

	view := model.renderNativeWorkflow()
	for _, text := range current.Lines {
		if !strings.Contains(view, text) {
			t.Fatal("native status omitted progress or wait", view)
		}
	}
}

// TestModelWorkflowStopAvailable permits status and stop requests during an
// execution, but never queues a second continuation in the TUI.
func TestModelWorkflowStopAvailable(t *testing.T) {
	executor := &fakeWorkActionExecutor{}
	model := NewModel(Options{WorkExecutor: executor})
	model.workflowExecutionBusy = true

	for _, action := range []WorkActionID{WorkActionWorkflowPause, WorkActionWorkflowStatus, WorkActionWorkflowContinue} {
		request := WorkActionRequest{Action: action, Workflow: &WorkflowActionRequest{ID: "run-one", ExpectedRevision: 1}}

		command := model.startWorkAction(request)
		if action == WorkActionWorkflowContinue {
			if command != nil {
				t.Fatal("TUI queued another continuation")
			}
		} else if command == nil || model.workBusy {
			t.Fatal("execution locked status or stop controls")
		}
	}
}

// TestModelWorkflowSelection keeps legacy selection behavior and routes native
// ticket and case choices through read-only preview before execution.
func TestModelWorkflowSelection(t *testing.T) {
	model := NewModel(Options{Observability: ObservabilityView{Authority: core.WorkStateAuthorityTickets}, Tickets: []core.IndexedTicket{{IndexedTicketIdentity: core.IndexedTicketIdentity{Key: "CHR-1"}, IndexedTicketClassification: core.IndexedTicketClassification{Status: core.TicketStatusReadyForDev}}}, Cases: []core.CaseMetadata{{Slug: "delivery"}}})
	model.selectView(ViewTickets)
	model.planWorkAction(WorkActionStartTicket)

	if model.workAction.Request.Action != WorkActionWorkflowTicket || model.workAction.Request.TicketKey != "CHR-1" || !model.workAction.Request.Workflow.Preview {
		t.Fatal("native ticket selection bypassed preview")
	}

	model.selectView(ViewCases)
	model.planWorkAction(WorkActionWorkflowCase)

	if model.workAction.Request.Workflow.Selection.Selector.Identity != "delivery" || model.workAction.Request.TicketKey != "" {
		t.Fatal("selected case identity was replaced")
	}
}

// TestModelWorkflowRefresh prevents an older asynchronous read from replacing
// a newer control result, while keeping the selected retained page.
func TestModelWorkflowRefresh(t *testing.T) {
	model := NewModel(Options{})
	model.workflow = WorkflowView{ID: "run-one", Revision: 5}
	page := WorkflowPageView{Offset: 16, Items: []WorkflowView{{ID: "run-one", Revision: 4}}, NextOffset: -1}
	older := WorkflowView{ID: "run-one", Revision: 4}
	model.completeAgentWorkRefresh(agentWorkRefreshFinishedMsg{result: AgentWorkRefreshResult{Workflow: &older, Observability: ObservabilityView{WorkflowPage: page}}})

	if model.workflow.Revision != 5 || model.workflowPage.Offset != 16 {
		t.Fatal("old refresh replaced the selected control revision or page")
	}

	newer := WorkflowView{ID: "run-one", Revision: 6, Lines: []string{"State: waiting_review"}}
	model.completeAgentWorkRefresh(agentWorkRefreshFinishedMsg{result: AgentWorkRefreshResult{Workflow: &newer, Observability: ObservabilityView{WorkflowPage: page}}})

	if model.workflow.Revision != 6 || !strings.Contains(model.renderNativeWorkflow(), "waiting_review") {
		t.Fatal("newer retained progress was not displayed")
	}
}
