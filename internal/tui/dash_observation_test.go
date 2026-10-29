package tui

import (
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/app/work"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// A locally ready-looking ticket cannot override application-owned policy.
// Its blocked reason and exclusions remain visible; Enter never queues work.
func TestDashPolicy(t *testing.T) {
	model := NewModel(Options{
		Tickets:         []ticketcore.Indexed{{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-1"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev}}},
		WorkObservation: &work.Observation{Queue: []work.QueueItem{{Key: "CHR-1", Readiness: "blocked", Reason: "Policy excludes this task."}}, QueueTotal: 1, QueueNext: -1, Excluded: map[string]int{"policy": 2}},
	})
	model.selectView(ViewDash)

	for _, text := range []string{"Eligible: 0 / 1 queued", "[blocked] Policy excludes this task.", "2 excluded: policy"} {
		if !strings.Contains(model.View(), text) {
			t.Fatal(text, model.View())
		}
	}

	model.planWorkAction(WorkActionStartTicket)

	if model.workAction.Pending || !strings.Contains(model.statusMessage, "Policy excludes") {
		t.Fatal(model.workAction, model.statusMessage)
	}

	model.planWorkAction(WorkActionRunAllReady)

	if model.workAction.Pending {
		t.Fatal("empty eligible queue advertised execution")
	}
}

// Timeline rows retain invocation ownership and separate host validation from
// confirmed delivery; oversized/empty snapshots remain bounded and scrollable.
func TestDashActivity(t *testing.T) {
	observation := &work.Observation{
		Current: []work.Context{
			{ID: "one:attempt", Kind: "native", TicketKey: "CHR-1", Phase: "validating", Message: "Running go", Validated: 1, Confirmed: 0, Remaining: 2},
			{ID: "two:attempt", Kind: "native", TicketKey: "CHR-2", Phase: "working"},
		},
		Results: []work.Context{{ID: "execution:done", Kind: "execution", Status: "succeeded", TicketKey: "CHR-3"}},
	}
	model := NewModel(Options{WorkObservation: observation})
	model.selectView(ViewDash)

	for _, text := range []string{"one:attempt; phase: validating", "two:attempt; phase: working", "Running go", "Host-validated tasks: 1; confirmed tickets: 0; remaining: 2", "Harness succeeded"} {
		if !strings.Contains(model.View(), text) {
			t.Fatal(text, model.View())
		}
	}

	for range 1000 {
		observation.Events = append(observation.Events, work.Event{ContextID: "bound-context", ID: "bound-event"})
	}

	lines := model.renderObservedActivity()
	if len(lines) > 270 || !strings.Contains(strings.Join(lines, "\n"), "+872 more events") {
		t.Fatal("unbounded event rendering", len(lines))
	}

	model.selectedPane = 2
	model.scrollDetail(10)

	if model.detailScroll != 10 {
		t.Fatal("Dash activity cannot scroll")
	}

	observation.Current, observation.Results, observation.Events = nil, nil, nil

	for _, text := range []string{"No current work.", "No retained results.", "No recent activity."} {
		if !strings.Contains(model.View(), text) {
			t.Fatal(text, model.View())
		}
	}
}
