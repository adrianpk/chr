package app

import (
	"fmt"
	"maps"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

// Real retained runs paginate independently of queue tickets. Read-only status
// keeps terminal outcomes and member revisions without admitting or resuming work.
func TestTUIWorkflowPages(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	executor := newTUIWorkActionExecutor(operations)

	checkEmptyWorkflowPage(t, executor)

	createWorkflowBrowserRuns(t, operations, hierarchy)

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	first, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowPage, Workflow: &tui.WorkflowActionRequest{}})
	if err != nil || len(first.WorkflowPage.Items) != 16 || first.WorkflowPage.NextOffset != 16 {
		t.Fatal(first, err)
	}

	last, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowPage, Workflow: &tui.WorkflowActionRequest{Offset: 16}})
	if err != nil || len(last.WorkflowPage.Items) != 1 || last.WorkflowPage.NextOffset != -1 {
		t.Fatal(last, err)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("workflow browsing changed canonical state")
	}

	checkWorkflowPageMember(t, executor, first.WorkflowPage.Items[0])
}

func checkWorkflowPageMember(t *testing.T, executor tuiWorkActionExecutor, item tui.WorkflowView) {
	t.Helper()

	result, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowStatus, Workflow: &tui.WorkflowActionRequest{ID: item.ID, ExpectedRevision: item.Revision, Offset: 0}})
	if err != nil || result.Workflow == nil || result.Workflow.State != "cancelled" || len(result.Workflow.Members) != 1 {
		t.Fatal(result, err)
	}

	for _, action := range result.Workflow.Actions {
		if action.Reason == "" {
			t.Fatal("terminal action advertised as eligible", action)
		}
	}
}

func createWorkflowBrowserRuns(t *testing.T, operations service, hierarchy workflowcore.MaterializationResult) {
	t.Helper()

	target := serviceWorkflowDependencyGit(t, operations.projectContext.ProjectDir, "rev-parse", "HEAD")
	for index := range 17 {
		run, err := operations.StartWorkflow(t.Context(), workflowcore.StartRequest{StartInput: workflowstate.StartInput{RequestKey: fmt.Sprintf("browser-%d", index), Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}}}})
		if err != nil {
			t.Fatal(err)
		}

		if _, err := operations.ControlWorkExtent(t.Context(), WorkflowControlRequest{Action: "cancel", ID: run.Invocation.ID, ExpectedRevision: run.Invocation.Revision}); err != nil {
			t.Fatal(err)
		}
	}
}

func checkEmptyWorkflowPage(t *testing.T, executor tuiWorkActionExecutor) {
	t.Helper()

	empty, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowPage, Workflow: &tui.WorkflowActionRequest{}})
	if err != nil || empty.WorkflowPage == nil || len(empty.WorkflowPage.Items) != 0 {
		t.Fatal(empty, err)
	}
}
