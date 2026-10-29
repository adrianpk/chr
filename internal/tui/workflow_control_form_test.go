package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// Forms keep their original candidate/budget/question fences across refreshes.
// A rejected confirmed grant restores the exact user input for explicit retry.
func TestWorkflowControlForms(t *testing.T) {
	executor := &fakeWorkActionExecutor{}
	model := NewModel(Options{WorkExecutor: executor, Observability: ObservabilityView{Authority: workspaceconfig.WorkStateAuthorityTickets}})
	model.workflow = WorkflowView{ID: "run", Revision: 2, Integration: &workflowstate.ControlIntegrationGuard{EntryID: "entry", EntryRevision: 3, BudgetRevision: 4, CandidateHash: strings.Repeat("a", 64)}, GrantKinds: []string{"submission"}, DecisionRevision: 5, DecisionHash: strings.Repeat("b", 64), Questions: []workflowstate.MergerDecision{{ID: "question", Issue: "Which compatible interpretation?", Authority: "operator", Options: []string{"Preserve both requirements", "Defer to owner"}}}}
	model.openWorkflowControlForm(WorkActionWorkflowGrant)
	model.textForm.Fields[1].Input.SetValue("1")
	model.textForm.Fields[2].Input.SetValue("operator")
	model.textForm.Fields[3].Input.SetValue("Extend once")

	model.workflow.Integration = &workflowstate.ControlIntegrationGuard{BudgetRevision: 99}
	command := model.submitWorkflowControl()
	command()

	request := executor.request
	if request.Workflow.Integration.BudgetRevision != 4 {
		t.Fatal("refresh replaced form guard")
	}

	model.completeNativeWorkAction(workActionFinishedMsg{Request: request, Result: WorkActionResult{Confirmation: &WorkActionState{Pending: true, Request: request}}})

	confirmed := model.workAction.Request
	model.workAction = WorkActionState{}
	model.completeNativeWorkAction(workActionFinishedMsg{Request: confirmed, Err: errors.New("stale budget")})

	if !model.textForm.Active || model.textForm.Fields[3].value() != "Extend once" || model.textForm.WorkRequest.Workflow.Integration.BudgetRevision != 4 {
		t.Fatal("stale result discarded draft or silently refreshed its guard")
	}

	model.openWorkflowControlForm(WorkActionWorkflowAnswer)

	model.textForm.Focus = 2
	model.textForm.Fields[2].Input.SetValue("2")
	model.textForm.Fields[4].Input.SetValue("operator")
	model.textForm.Fields[5].Input.SetValue("Use retained option")
	model.submitWorkflowControl()()

	answer := executor.request.Workflow
	if answer.Fields["choice"] != "Defer to owner" || answer.Fields["authority"] != "operator" || answer.DecisionRevision != 5 {
		t.Fatal("answer lost captured question or authority")
	}
}

// Unavailable controls explain their current owner restriction before dispatch;
// links use the retained ticket identity rather than the queue's selected row.
func TestWorkflowControlLinks(t *testing.T) {
	tickets := &fakeTicketActionExecutor{result: TicketActionResult{Editor: &TicketEditorView{Key: "retained-ticket", CaseSlug: "retained-case"}}}
	model := NewModel(Options{TicketExecutor: tickets})

	model.workflow = WorkflowView{ID: "run", TicketKey: "retained-ticket", Actions: []WorkflowActionView{{Action: WorkActionWorkflowApprove, Reason: "No exact candidate awaits approval."}}}
	if model.workflowActionAvailable(WorkActionWorkflowApprove) || !strings.Contains(model.statusMessage, "No exact candidate") {
		t.Fatal("unavailable control did not explain why")
	}

	model.openWorkflowLink(true)

	if tickets.request.Key != "retained-ticket" || model.caseContext.CaseSlug != "retained-case" || model.SelectedView() != ViewCases {
		t.Fatal("case link followed queue selection")
	}

	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
}
