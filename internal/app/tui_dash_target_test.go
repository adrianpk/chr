package app

import (
	"errors"
	"maps"
	"strings"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/tui"
)

// Keyboard selection is carried through the real native adapter, with a
// different persisted backlog ticket still selected on the Board.
func TestTUIDashTarget(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	body := "Keep Board selection independent."

	backlog, err := operations.CreateBoardTicket(workflowcore.TicketEditRequest{Body: &body, Fields: map[string]string{"title": "Board target", "status": "backlog"}, Binding: workflowstate.MutationBinding{RequestID: "board-target", Operator: "maintainer"}})
	if err != nil {
		t.Fatal(err)
	}

	tickets := []ticketcore.Indexed{backlog.Ticket}

	for _, record := range hierarchy.Tasks[:2] {
		document, err := operations.ReadTicketDocument(record.Task.Key)
		if err != nil {
			t.Fatal(err)
		}

		tickets = append(tickets, document.Ticket)
	}

	model := tui.NewModel(tui.Options{Tickets: tickets, Observability: tui.ObservabilityView{Authority: workspaceconfig.WorkStateAuthorityTickets}, WorkExecutor: newTUIWorkActionExecutor(operations)})
	runGuidedTUIKey(t, model, "f7", false)
	runGuidedTUIKey(t, model, "down", false)
	runGuidedTUIKey(t, model, "enter", false)
	runGuidedTUIKey(t, model, "enter", true)

	if !strings.Contains(model.View(), "Selected ticket: "+tickets[2].Key+" revision") {
		t.Fatal("native preview did not bind the displayed Dash key", model.View())
	}
}

// The adapter must not replace an observed revision with a newer snapshot on
// preview, and native confirmation must reject later changes before admission.
func TestTUIDashRevision(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	executor := newTUIWorkActionExecutor(operations)
	selected := hierarchy.Tasks[0]
	request := tui.WorkActionRequest{Action: tui.WorkActionWorkflowTicket, TicketKey: selected.Task.Key, TicketRevision: selected.SourceHash, Workflow: &tui.WorkflowActionRequest{Preview: true, Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket"}}}}

	preview, err := executor.ExecuteWorkAction(request)
	if err != nil || preview.Confirmation == nil {
		t.Fatal("initial preview failed", err)
	}

	if _, err := operations.SaveTicketDocument(workflowcore.TicketEditRequest{Key: selected.Task.Key, Revision: selected.SourceHash, Comment: "Concurrent update", Binding: workflowstate.MutationBinding{RequestID: "concurrent-comment", Operator: "maintainer"}}); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	for _, tc := range []struct {
		name    string
		request tui.WorkActionRequest
	}{
		{"preview", request},
		{"confirmation", preview.Confirmation.Request},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := executor.ExecuteWorkAction(tc.request); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("stale Dash action accepted", err)
			}
		})
	}

	if after := workflowCLIImage(t, operations.projectContext.ProjectDir); !maps.Equal(before, after) {
		t.Fatal("stale Dash action mutated retained work")
	}
}
