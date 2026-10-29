package app

import (
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Neutral options preserve canonical native pending restrictions and saved
// values. Both adapters can map these values without importing each other.
func TestTicketEditorOptions(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "native"}[native], func(t *testing.T) {
			provider := webTicketEditorProvider(t, native)
			operations := provider.service(t.Context())

			initial, err := operations.newTicketEditor("named-work")
			if err != nil || len(ticketEditorFieldOptions(initial, "status")) != 2 {
				t.Fatal("new ticket offers an execution state", initial, err)
			}

			body := "Retained editor content."

			saved, err := operations.CreateBoardTicket(workflowcore.TicketEditRequest{Body: &body, Fields: map[string]string{"title": "Editable record", "use_case": "named-work", "priority": "critical"}, Binding: workflowstate.MutationBinding{RequestID: "editor-options", Operator: "fixture"}})
			if err != nil {
				t.Fatal(err)
			}

			editor, err := operations.readTicketEditor(saved.Ticket.Key)

			want := 6
			if native {
				want = 2
			}

			if err != nil || len(ticketEditorStatusOptions(editor)) != want || editor.Fields["priority"] != "critical" || len(editor.Cases) != 1 {
				t.Fatal("metadata or canonical status restrictions changed", editor, err)
			}
		})
	}
}

// Repository tickets use their own established classification vocabulary and
// preserve existing values rather than silently selecting the first option.
func TestTicketRepositoryOptions(t *testing.T) {
	editor := ticketEditor{Key: "TKT-20261029101144", Fields: map[string]string{"priority": "custom-priority"}}

	for _, field := range []struct{ name, value string }{{"priority", "urgent"}, {"severity", "medium"}, {"type", "decision"}} {
		found := false
		for _, choice := range ticketEditorFieldOptions(editor, field.name) {
			found = found || choice.Value == field.value
		}

		if !found {
			t.Fatal("repository vocabulary changed", field)
		}
	}

	field := ticketEditorFieldValue(editor, ticketEditorField{Name: "priority"})
	if field.Options[len(field.Options)-1].Value != "custom-priority" {
		t.Fatal("unknown retained value was replaced", field)
	}
}
