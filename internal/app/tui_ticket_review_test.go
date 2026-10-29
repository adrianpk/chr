package app

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

// Reference attachment takes user input, validates it before writing, and keeps
// the selected ticket's description, checklist and discussion intact.
func TestTUITicketReviewForm(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "native"}[native], func(t *testing.T) {
			operations := webTicketEditorProvider(t, native).service(t.Context())
			original := createTUITicketSeed(t, operations)
			model := ticketFormModel(t, operations)
			runGuidedTUIKey(t, model, "v", false)
			ticketFormTabs(model, 1)
			typeTicketField(model, "forgejo")
			ticketFormTabs(model, 1)
			typeTicketField(model, "javascript:invalid")
			model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

			if !strings.Contains(model.View(), "HTTP URL") {
				t.Fatal("invalid URL accepted", model.View())
			}

			model.Update(tea.KeyMsg{Type: tea.KeyEsc})
			model.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
			typeTicketField(model, "https://forge.example.org/team/project/pulls/77")
			ticketFormTabs(model, 1)
			typeTicketField(model, "77")
			ticketFormTabs(model, 1)
			typeTicketField(model, "open")
			ticketFormTabs(model, 1)
			typeTicketField(model, "feat/user-entered")
			model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

			saved, err := operations.ReadTicketDocument(original.Ticket.Key)
			if err != nil || len(saved.Ticket.ReviewRefs) != 1 || saved.Description != original.Description || saved.Ticket.Branch != "feat/user-entered" {
				t.Fatal(saved.Ticket, saved.Fields, err, model.View())
			}

			checkEnteredTicketRef(t, saved.Ticket.ReviewRefs[0])
		})
	}
}

// A reference is a revision-bound edit and cannot bypass an active workflow
// claim. Neither stale metadata nor claimed targets are modified by the adapter.
func TestTUITicketReviewGuards(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)

	original, err := operations.ReadTicketDocument(hierarchy.Tasks[0].Task.Key)
	if err != nil {
		t.Fatal(err)
	}

	ref := &ticketmetadata.ReviewRef{Type: "pull_request", Provider: "forgejo", URL: "https://forge.example.org/team/project/pulls/9"}
	executor := newTUITicketActionExecutor(operations)
	request := tui.TicketActionRequest{Action: tui.TicketActionAttachReviewRef, Key: original.Ticket.Key, Revision: original.Revision, OperationID: "managed-ref", ReviewRef: ref}

	saved, err := executor.ExecuteTicketAction(request)
	if err != nil || saved.Editor == nil {
		t.Fatal("pending managed reference rejected", err)
	}

	document, err := operations.ReadTicketDocument(original.Ticket.Key)
	if err != nil || len(document.Ticket.ReviewRefs) != 1 || document.Description != original.Description {
		t.Fatal("native reference lost or contract changed", document, err)
	}

	if _, err := executor.ExecuteTicketAction(request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale reference accepted", err)
	}

	checkClaimedTicketRef(t, operations, hierarchy, request)
}

func checkClaimedTicketRef(t *testing.T, operations service, hierarchy workflowcore.MaterializationResult, request tui.TicketActionRequest) {
	t.Helper()

	executor := newTUITicketActionExecutor(operations)

	id := workflowCLIManualStart(t, operations, hierarchy)
	if _, err := operations.ClaimWorkflow(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	owned, err := operations.ReadTicketDocument(request.Key)
	if err != nil {
		t.Fatal(err)
	}

	request.Revision, request.OperationID = owned.Revision, "claimed-ref"
	if _, err := executor.ExecuteTicketAction(request); !errors.Is(err, workflowstate.ErrOwned) {
		t.Fatal("reference bypassed active claim", err)
	}
}

func checkEnteredTicketRef(t *testing.T, ref ticketmetadata.ReviewRef) {
	t.Helper()

	if ref.Provider != "forgejo" || ref.ID != "77" || ref.Status != "open" || !strings.HasSuffix(ref.URL, "/77") {
		t.Fatal(ref)
	}
}
