package app

import (
	"os/exec"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestServiceWorkflowPreview proves the shared operation observes files without SQL or tracker authority.
func TestServiceWorkflowPreview(t *testing.T) {
	project := newTestWorkspace(t)
	if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\n")
	writeServiceWorkspaceFile(t, filepath.Join(project, "ops", "default", "plan", "approved.md"), "approved\n")

	operations := newTestService(t, project)

	result, err := operations.MaterializeWorkflowPlan(t.Context(), serviceWorkflowPlanFixture())
	if err != nil {
		t.Fatal(err)
	}

	preview, err := operations.PreviewWorkflow(t.Context(), workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice_set", Identity: result.Set.ID}, Limit: 2})
	if err != nil || preview.Total != 6 || preview.Authority != "tickets" || preview.NextOffset != 2 {
		t.Fatalf("preview = %+v, %v", preview, err)
	}

	if _, err := operations.PreviewWorkflow(t.Context(), workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket"}, Limit: 1}); err == nil {
		t.Fatal("missing exact identity became queue")
	}
}
