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
