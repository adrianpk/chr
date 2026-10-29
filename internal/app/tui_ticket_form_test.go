package app

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

// The terminal form persists entered content in either authority, reloads edits,
// and cancels without touching files; no generated default title or area leaks in.
func TestTUITicketForm(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "native"}[native], func(t *testing.T) {
			operations := webTicketEditorProvider(t, native).service(t.Context())
			model := ticketFormModel(t, operations)
			runGuidedTUIKey(t, model, "c", false)
			typeTicketField(model, "Entered ticket")
			model.Update(tea.KeyMsg{Type: tea.KeyTab})
			typeTicketField(model, "Full entered description.\n\nSecond paragraph.")
			ticketFormTabs(model, 3)
			model.Update(tea.KeyMsg{Type: tea.KeyRight})
			model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

			board, err := operations.ReadTicketBoard(0)
			if err != nil || len(board.Tickets) != 1 {
				t.Fatal(board, err, model.View())
			}

			document, err := operations.ReadTicketDocument(board.Tickets[0].Key)
			if err != nil || document.Fields["title"] != "Entered ticket" || document.Fields["priority"] != "high" || document.Fields["project_area"] != "" || document.Description != "Full entered description.\n\nSecond paragraph." {
				t.Fatal(document.Fields, document.Description, err, model.View())
			}

			editTicketForm(t, operations, model, document)
		})
	}
}

func ticketFormModel(t *testing.T, operations service) *tui.Model {
	t.Helper()

	cases, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	board, err := operations.ReadTicketBoard(0)
	if err != nil {
		t.Fatal(err)
	}

	model := tui.NewModel(tui.Options{Cases: cases, Tickets: board.Tickets, TicketExecutor: newTUITicketActionExecutor(operations)})
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 36})
	model.Update(tea.KeyMsg{Type: tea.KeyF6})

	return model
}

func typeTicketField(model *tui.Model, text string) {
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)})
}

func ticketFormTabs(model *tui.Model, count int) {
	for range count {
		model.Update(tea.KeyMsg{Type: tea.KeyTab})
	}
}

func editTicketForm(t *testing.T, operations service, model *tui.Model, original workflowcore.TicketDocument) {
	t.Helper()
	runGuidedTUIKey(t, model, "e", false)
	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	typeTicketField(model, " edited")
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	current, err := operations.ReadTicketDocument(original.Ticket.Key)
	if err != nil || current.Fields["title"] != "Entered ticket edited" || current.Revision == original.Revision || current.Description != original.Description {
		t.Fatal(current.Fields, err, model.View())
	}

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	runGuidedTUIKey(t, model, "e", false)
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("cancel changed ticket storage")
	}
}

// Stale saves keep every entered field until a confirmed reload. Cancelling
// reload retains the draft; a later reload observes the canonical writer's edit.
func TestTUITicketStaleForm(t *testing.T) {
	operations := webTicketEditorProvider(t, false).service(t.Context())
	body := "Original body"

	original, err := operations.CreateBoardTicket(workflowcore.TicketEditRequest{Body: &body, Fields: map[string]string{"title": "Original", "use_case": "named-work"}, Binding: workflowstate.MutationBinding{RequestID: "seed-form", Operator: "test"}})
	if err != nil {
		t.Fatal(err)
	}

	model := ticketFormModel(t, operations)
	runGuidedTUIKey(t, model, "e", false)
	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	typeTicketField(model, " retained draft")

	newer, err := operations.SaveTicketDocument(workflowcore.TicketEditRequest{Key: original.Ticket.Key, Revision: original.Revision, Fields: map[string]string{"title": "External edit"}, Binding: workflowstate.MutationBinding{RequestID: "external-form", Operator: "test"}})
	if err != nil {
		t.Fatal(err)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if !strings.Contains(model.View(), "Operation: stale") {
		t.Fatal(model.View())
	}

	runGuidedTUIKey(t, model, "r", false)
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !strings.Contains(model.View(), "retained draft") {
		t.Fatal("cancelled reload lost input", model.View())
	}

	runGuidedTUIKey(t, model, "alt+o", false)
	runGuidedTUIKey(t, model, "r", false)
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !strings.Contains(model.View(), "External edit") || strings.Contains(model.View(), "retained draft") {
		t.Fatal("reload did not observe latest", model.View())
	}

	current, err := operations.ReadTicketDocument(original.Ticket.Key)
	if err != nil || current.Revision != newer.Revision {
		t.Fatal("rejected form changed canonical storage", err)
	}
}

// Managed ticket restrictions are supplied by the owner, rendered to terminal
// users, and remain enforced when an adapter attempts to send a forbidden edit.
func TestTUITicketManagedForm(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	key := hierarchy.Tasks[0].Task.Key

	original, err := operations.ReadTicketDocument(key)
	if err != nil {
		t.Fatal(err)
	}

	executor := newTUITicketActionExecutor(operations)

	result, err := executor.ExecuteTicketAction(tui.TicketActionRequest{Action: tui.TicketActionRead, Key: key})
	if err != nil || result.Editor == nil || result.Editor.Phase != "pending" {
		t.Fatal(result, err)
	}

	checkManagedTicketFields(t, *result.Editor)

	body := "Unapproved replacement contract"

	_, err = executor.ExecuteTicketAction(tui.TicketActionRequest{Action: tui.TicketActionSave, Key: key, Revision: original.Revision, OperationID: "forbidden-body", Body: &body})
	if err == nil {
		t.Fatal("terminal bypassed managed description guard")
	}

	current, err := operations.ReadTicketDocument(key)
	if err != nil || current.Revision != original.Revision {
		t.Fatal("rejected edit changed a managed ticket", err)
	}
}

func checkManagedTicketFields(t *testing.T, editor tui.TicketEditorView) {
	t.Helper()

	for _, name := range []string{"description", "use_case"} {
		found := false

		for _, field := range editor.Fields {
			if field.Name == name {
				found = field.ReadOnly != ""
			}
		}

		if !found {
			t.Fatal("managed restriction was not mapped", name)
		}
	}
}
