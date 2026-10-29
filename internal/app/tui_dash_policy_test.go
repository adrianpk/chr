package app

import (
	"errors"
	"maps"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

// A policy change after preview must reject the original confirmation before
// admission; unchanged source tickets cannot make an old policy safe to execute.
func TestTUIDashPolicyBinding(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	operations.workflowWorker = []string{"sh", "-c", "exit 77"}
	selected := hierarchy.Tasks[0]
	executor := newTUIWorkActionExecutor(operations)

	preview, err := executor.ExecuteWorkAction(tui.WorkActionRequest{
		Action:    tui.WorkActionWorkflowTicket,
		TicketKey: selected.Task.Key, TicketRevision: selected.SourceHash,
		Workflow: &tui.WorkflowActionRequest{Preview: true, Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket"}}},
	})
	if err != nil || preview.Confirmation == nil {
		t.Fatal(preview, err)
	}

	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n  worker_timeout_seconds: 60\n")

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	_, err = executor.ExecuteWorkAction(preview.Confirmation.Request)
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed policy accepted", err)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("policy conflict changed retained work")
	}
}

// Changing the extent after preview cannot reuse another scope's confirmation.
func TestTUIDashScopeBinding(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	operations.workflowWorker = []string{"sh", "-c", "exit 77"}
	selected := hierarchy.Tasks[0]
	executor := newTUIWorkActionExecutor(operations)

	preview, err := executor.ExecuteWorkAction(tui.WorkActionRequest{
		Action:    tui.WorkActionWorkflowTicket,
		TicketKey: selected.Task.Key, TicketRevision: selected.SourceHash,
		Workflow: &tui.WorkflowActionRequest{Preview: true, Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket"}}},
	})
	if err != nil || preview.Confirmation == nil {
		t.Fatal(preview, err)
	}

	request := preview.Confirmation.Request
	request.Workflow.Selection.Selector = workflowstate.Selector{Extent: "slice_set", Identity: hierarchy.Set.ID}
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	_, err = executor.ExecuteWorkAction(request)
	if !errors.Is(err, workflowstate.ErrConflict) || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("changed scope admitted work", err)
	}
}
