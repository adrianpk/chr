package app

import (
	"errors"
	"maps"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

// The adapter must not replace an observed revision with a newer snapshot on
// preview, and native confirmation must reject later changes before admission.
func TestTUIDashRevision(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	executor := newTUIWorkActionExecutor(operations)
	selected := hierarchy.Tasks[0]
	request := tui.WorkActionRequest{Action: tui.WorkActionWorkflowTicket, TicketKey: selected.Task.Key, TicketRevision: selected.SourceHash, Workflow: &tui.WorkflowActionRequest{Preview: true, Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket"}}}}
	preview, err := executor.ExecuteWorkAction(request)
	if err != nil || preview.Confirmation == nil {
		t.Fatal("initial preview failed", err)
	}

	if _, err := operations.SaveTicketDocument(workflowcore.TicketEditRequest{Key: selected.Task.Key, Revision: selected.SourceHash, Comment: "Concurrent update", Binding: workflowstate.MutationBinding{RequestID: "concurrent-comment", Operator: "maintainer"}}); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	for _, tc := range []struct {
		name    string
		request tui.WorkActionRequest
	}{
		{"preview", request},
		{"confirmation", preview.Confirmation.Request},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := executor.ExecuteWorkAction(tc.request); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("stale Dash action accepted", err)
			}
		})
	}

	if after := workflowCLIImage(t, operations.projectContext.ProjectDir); !maps.Equal(before, after) {
		t.Fatal("stale Dash action mutated retained work")
	}
}
