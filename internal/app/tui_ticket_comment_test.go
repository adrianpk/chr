package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

// Distinct free-text comments round trip without replacing the description or
// older discussion. Empty and stale submissions leave the complete draft intact.
func TestTUITicketComments(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "native"}[native], func(t *testing.T) {
			operations := webTicketEditorProvider(t, native).service(t.Context())
			original := createTUITicketSeed(t, operations)
			model := ticketFormModel(t, operations)
			runGuidedTUIKey(t, model, "a", false)
			model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

			if !strings.Contains(model.View(), "non-empty comment") {
				t.Fatal("empty comment accepted", model.View())
			}

			model.Update(tea.KeyMsg{Type: tea.KeyEsc})
			typeTicketField(model, "First entered observation.")
			model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
			runGuidedTUIKey(t, model, "a", false)
			model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

			if !strings.Contains(model.View(), "First entered observation.") {
				t.Fatal("history missing", model.View())
			}

			model.Update(tea.KeyMsg{Type: tea.KeyTab})
			typeTicketField(model, "Second distinct observation.")
			model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

			saved, err := operations.ReadTicketDocument(original.Ticket.Key)
			if err != nil || saved.Description != original.Description || strings.Count(saved.Discussion, "First entered observation.") != 1 || strings.Count(saved.Discussion, "Second distinct observation.") != 1 {
				t.Fatal(saved.Description, saved.Discussion, err)
			}

			staleTicketComment(t, operations, model, saved)
		})
	}
}

func createTUITicketSeed(t *testing.T, operations service) workflowcore.TicketDocument {
	t.Helper()

	body := "Original description.\n\n- [ ] Retained checklist item"

	document, err := operations.CreateBoardTicket(workflowcore.TicketEditRequest{Body: &body, Fields: map[string]string{"title": "Selected target", "use_case": "named-work"}, Binding: workflowstate.MutationBinding{RequestID: "ticket-seed", Operator: "maintainer"}})
	if err != nil {
		t.Fatal(err)
	}

	return document
}

func staleTicketComment(t *testing.T, operations service, model *tui.Model, before workflowcore.TicketDocument) {
	t.Helper()
	runGuidedTUIKey(t, model, "a", false)
	typeTicketField(model, "Retained stale comment")

	newer, err := operations.SaveTicketDocument(workflowcore.TicketEditRequest{Key: before.Ticket.Key, Revision: before.Revision, Comment: "External comment", Binding: workflowstate.MutationBinding{RequestID: "external-comment", Operator: "maintainer"}})
	if err != nil {
		t.Fatal(err)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if !strings.Contains(model.View(), "Operation: stale") {
		t.Fatal(model.View())
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !strings.Contains(model.View(), "Retained stale comment") {
		t.Fatal("stale rejection lost draft", model.View())
	}

	current, err := operations.ReadTicketDocument(before.Ticket.Key)
	if err != nil || current.Revision != newer.Revision || strings.Contains(current.Discussion, "Retained stale comment") {
		t.Fatal("stale comment changed the source", err)
	}

	runGuidedTUIKey(t, model, "alt+o", false)
	runGuidedTUIKey(t, model, "r", false)
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !strings.Contains(model.View(), "Ticket Discussion") || strings.Contains(model.View(), "Retained stale comment") {
		t.Fatal("reload changed form mode or retained rejected input", model.View())
	}
}
