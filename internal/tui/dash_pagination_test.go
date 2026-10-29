package tui

import (
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/app/work"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Page/filter/order changes are sent to observation, never to a work executor;
// a reordered page keeps its selected key and an empty page clears it safely.
func TestDashPagination(t *testing.T) {
	executor := &dashPageExecutor{observation: work.Observation{Queue: []work.QueueItem{{Key: "CHR-2"}, {Key: "CHR-1"}}, QueueNext: 128}}
	model := NewModel(Options{WorkObservation: &executor.observation, AgentWorkExecutor: executor})
	model.selectView(ViewDash)
	model.moveDashSelection(1)

	for _, key := range []string{"o", "p", "]", "["} {
		command, handled := model.handleDashKey(key)
		if !handled || command == nil {
			t.Fatal(key, "queue command is unreachable")
		}

		model.Update(command())

		if selected, ok := model.selectedDashTicketView(); !ok || selected.Key != "CHR-1" {
			t.Fatal("query changed selected key", selected)
		}
	}

	expected := []work.Request{
		{Order: workflowstate.TicketOrderPriority},
		{Priority: "critical", Order: workflowstate.TicketOrderPriority},
		{Offset: 128, Priority: "critical", Order: workflowstate.TicketOrderPriority},
		{Priority: "critical", Order: workflowstate.TicketOrderPriority},
	}
	if !reflect.DeepEqual(executor.requests, expected) {
		t.Fatal(executor.requests)
	}

	executor.observation.Queue = nil
	command, _ := model.handleDashKey("o")
	model.Update(command())

	if _, ok := model.selectedDashTicketView(); ok || model.workAction.Pending {
		t.Fatal("empty page retained a target")
	}
}

type dashPageExecutor struct {
	observation work.Observation
	requests    []work.Request
}

func (e *dashPageExecutor) RefreshAgentWork(request AgentWorkRefreshRequest) (AgentWorkRefreshResult, error) {
	e.requests = append(e.requests, request.Queue)
	copy := e.observation

	return AgentWorkRefreshResult{Observation: &copy}, nil
}
