package app

import (
	"errors"
	"fmt"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
	"forge.adrianpk.com/chr/internal/webui"
)

// A terminal move cannot use a legacy review alias to bypass Web's canonical
// review-evidence guard. Both paths operate on real owned ticket documents.
func TestTUITicketMove(t *testing.T) {
	provider, original := webTicketWriteFixture(t)
	operations := provider.service(t.Context())
	executor := newTUITicketActionExecutor(operations)
	before := readWebTicketFile(t, provider, original.Key)

	_, err := executor.ExecuteTicketAction(tui.TicketActionRequest{Action: tui.TicketActionMoveStatus, Key: original.Key, Status: "reviewing"})
	if err == nil || readWebTicketFile(t, provider, original.Key) != before {
		t.Fatal("terminal accepted a legacy review move without canonical evidence", err)
	}
}

// Terminal moves and Web moves share canonical destinations, revision guards,
// native phase restrictions and review/integration evidence requirements.
func TestTUIBoardMoves(t *testing.T) {
	cases := []struct {
		status         string
		legacy, native bool
	}{
		{"backlog", true, true},
		{"ready_for_dev", true, true},
		{"in_progress", true, false},
		{"ready_for_review", false, false},
		{"ready_for_merge", false, false},
		{"done", false, false},
		{"ready", false, false},
		{"reviewing", false, false},
	}

	for _, native := range []bool{false, true} {
		for _, tc := range cases {
			allowed := tc.legacy
			if native {
				allowed = tc.native
			}

			for _, adapter := range []string{"web", "tui"} {
				t.Run(fmt.Sprintf("native=%t/%s/%s", native, tc.status, adapter), func(t *testing.T) {
					checkTUIBoardMove(t, native, tc.status, adapter, allowed)
				})
			}
		}
	}
}

func checkTUIBoardMove(t *testing.T, native bool, status, adapter string, allowed bool) {
	t.Helper()

	provider := webTicketEditorProvider(t, native)
	operations := provider.service(t.Context())
	body := "A guarded move target."

	document, err := operations.CreateBoardTicket(workflowcore.TicketEditRequest{Body: &body, Fields: map[string]string{"title": "Move target", "status": "ready_for_dev"}, Binding: workflowstate.MutationBinding{RequestID: "create-move-target", Operator: "maintainer"}})
	if err != nil {
		t.Fatal(err)
	}

	before, err := operations.ReadTicketDocument(document.Ticket.Key)
	if err != nil {
		t.Fatal(err)
	}

	if applied := runTUIBoardMove(t, provider, before, status, adapter); applied != allowed {
		t.Fatal("adapter move disagreed with the canonical guard", adapter)
	}

	after, err := operations.ReadTicketDocument(before.Ticket.Key)
	if err != nil || (allowed && string(after.Ticket.Status) != status) || (!allowed && after.Revision != before.Revision) {
		t.Fatal("move lost canonical status or changed rejected data", err)
	}
}

func runTUIBoardMove(t *testing.T, provider webViewProvider, before workflowcore.TicketDocument, status, adapter string) bool {
	t.Helper()

	if adapter == "web" {
		response, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{Action: webActionTicketMove, ID: before.Ticket.Key, Revision: before.Revision, Confirm: true, Fields: map[string]string{"status": status}})
		if err != nil {
			t.Fatal(err)
		}

		return response.Status == webui.WriteStatusOK
	}

	request := tui.NewTicketActionState(tui.TicketActionMoveStatus, tui.NewTicketViews([]ticketcore.Indexed{before.Ticket})[0], status).Request
	_, err := newTUITicketActionExecutor(provider.service(t.Context())).ExecuteTicketAction(request)

	return err == nil
}

// A confirmed terminal move cannot overwrite an edit made after selection.
func TestTUIBoardStale(t *testing.T) {
	provider, original := webTicketWriteFixture(t)
	before := readWebTicketFile(t, provider, original.Key)

	_, err := newTUITicketActionExecutor(provider.service(t.Context())).ExecuteTicketAction(tui.TicketActionRequest{Action: tui.TicketActionMoveStatus, Key: original.Key, Revision: original.Revision, Status: "backlog"})
	if !errors.Is(err, workflowstate.ErrConflict) || readWebTicketFile(t, provider, original.Key) != before {
		t.Fatal("stale move changed a newer ticket", err)
	}
}

// A live claim keeps status under its workflow owner in both adapters.
func TestTUIBoardOwned(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)

	id := workflowCLIManualStart(t, operations, hierarchy)
	if _, err := operations.ClaimWorkflow(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	document, err := operations.ReadTicketDocument(hierarchy.Tasks[0].Task.Key)
	if err != nil {
		t.Fatal(err)
	}

	request := tui.TicketActionRequest{Action: tui.TicketActionMoveStatus, Key: document.Ticket.Key, Revision: document.Revision, Status: "backlog"}
	if _, err := newTUITicketActionExecutor(operations).ExecuteTicketAction(request); !errors.Is(err, workflowstate.ErrOwned) {
		t.Fatal("TUI move bypassed the workflow owner", err)
	}

	response := writeWebTicket(operations, webui.WriteActionRequest{Action: webActionTicketMove, ID: document.Ticket.Key, Revision: document.Revision, Fields: map[string]string{"status": "backlog"}})
	if response.Status == webui.WriteStatusOK {
		t.Fatal("Web move bypassed the workflow owner")
	}

	after, err := operations.ReadTicketDocument(document.Ticket.Key)
	if err != nil || after.Revision != document.Revision {
		t.Fatal("rejected move changed the owned document", err)
	}
}
