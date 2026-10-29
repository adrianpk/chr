package app

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

func serviceWorkflowPublishCoordinator(t *testing.T, project string) {
	t.Helper()

	if output, err := exec.CommandContext(t.Context(), "git", "-C", project, "add", "ops/default/coordinator.json").CombinedOutput(); err != nil {
		t.Fatalf("stage coordinator: %v: %s", err, output)
	}
}

// TestServiceWorkflowOwner exercises actual compatibility and TUI starts while
// a manual handoff owns the ticket; no entrypoint may create another attempt.
func TestServiceWorkflowOwner(t *testing.T) {
	operations, record, request, _ := serviceWorkNativeFixture(t)

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.PrepareWorkflowManualTask(t.Context(), claim, nil); err != nil {
		t.Fatal(err)
	}

	checkServiceWorkflowOwnerStarts(t, operations, record.Task.Key)

	retained, err := operations.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || len(retained.Attempts) != 1 || retained.Cursor != 0 || retained.Status != workflowstate.InvocationRunning {
		t.Fatal("competing entrypoint changed the manual owner", err)
	}

	if err := operations.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}
}

func checkServiceWorkflowOwnerStarts(t *testing.T, operations service, key string) {
	t.Helper()

	executor := newTUIWorkActionExecutor(operations)

	document, err := operations.ReadTicketDocument(key)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name       string
		start      func() error
		emptyQueue bool
	}{
		{"ticket", func() error { _, err := operations.StartAgentWorkTicket(key); return err }, false},
		{"ready", func() error { _, err := operations.StartReadyAgentWorkLoop(); return err }, true},
		{"all ready", func() error { _, err := operations.StartAllReadyAgentWork(); return err }, true},
		{"running selection", func() error {
			_, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{Statuses: []string{string(ticketcore.StatusInProgress)}})

			return err
		}, false},
		{"TUI ticket", func() error {
			_, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionStartTicket, TicketKey: key, TicketRevision: document.Revision})

			return err
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.start()
			if tc.emptyQueue {
				if err == nil || !strings.Contains(err.Error(), "no prepared unfinished tasks") {
					t.Fatal("ready selection did not exclude the running ticket", err)
				}

				return
			}

			if !errors.Is(err, workflowstate.ErrOwned) && !errors.Is(err, workflowcore.ErrWorkflowCapacity) {
				t.Fatal("competing start did not retain the live owner", err)
			}
		})
	}
}
