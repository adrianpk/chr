package app

import (
	"maps"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/tui"
)

// The scope form reaches the real preview owner, freezes policy/filter/order,
// and cancels without creating an invocation, starting a worker or writing files.
func TestTUIWorkScopeForm(t *testing.T) {
	operations, record, _, provider := serviceWorkNativeFixture(t)
	model := tui.NewModel(tui.Options{Tickets: []ticketcore.Indexed{mustTUIWorkflowTicket(t, operations, record.Task.Key)}, TicketExecutor: newTUITicketActionExecutor(operations), WorkExecutor: newTUIWorkActionExecutor(operations), Observability: tui.ObservabilityView{Authority: workspaceconfig.WorkStateAuthorityTickets}})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	model.Update(tea.KeyMsg{Type: tea.KeyF7})

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	runGuidedTUIKey(t, model, "b", false)

	if !strings.Contains(model.View(), "Preview Work Scope") {
		t.Fatal(model.View())
	}

	for range 4 {
		model.Update(tea.KeyMsg{Type: tea.KeyRight})
	}

	ticketFormTabs(model, 6)
	model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model.Update(tea.KeyMsg{Type: tea.KeyRight})

	_, command := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if command == nil {
		t.Fatal("scope form did not dispatch preview")
	}

	model.Update(command())

	if !strings.Contains(model.View(), "Confirm native workflow extent") || !strings.Contains(model.View(), "order priority") || !strings.Contains(model.View(), "Include later tickets:") {
		t.Fatal(model.View())
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) || provider.creates != 0 {
		t.Fatal("preview or cancel changed canonical work")
	}
}

func mustTUIWorkflowTicket(t *testing.T, operations service, key string) ticketcore.Indexed {
	t.Helper()

	document, err := operations.ReadTicketDocument(key)
	if err != nil {
		t.Fatal(err)
	}

	return document.Ticket
}

// A bounded member projection cannot replace the canonical selected ticket's
// source fence. Fresh and stale anchors are checked independently of that window.
func TestTUIWorkAnchorWindow(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	selected := hierarchy.Tasks[len(hierarchy.Tasks)-1]
	request := tui.WorkActionRequest{Action: tui.WorkActionWorkflowSet, TicketKey: selected.Task.Key, TicketRevision: selected.SourceHash}

	bound, err := operations.PreviewSelectedWork(t.Context(), workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice_set"}}, selected.Task.Key)
	if err != nil {
		t.Fatal(err)
	}

	bound.View.Preview.Members = bound.View.Preview.Members[:1]

	executor := newTUIWorkActionExecutor(operations)
	if !executor.tuiWorkObservationMatches(request, bound) {
		t.Fatal("bounded member window hid the canonical anchor")
	}

	request.TicketRevision = "obsolete"
	if executor.tuiWorkObservationMatches(request, bound) {
		t.Fatal("stale anchor accepted")
	}
}
