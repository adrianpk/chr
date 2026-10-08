package app

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestServiceWorkflowMaterialization exercises the operation service, native key
// sequence, canonical files and the existing SQLite projection without a model.
func TestServiceWorkflowMaterialization(t *testing.T) {
	project := newTestWorkspace(t)
	if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\n")
	writeServiceWorkspaceFile(t, filepath.Join(project, "ops", "default", "plan", "approved.md"), "approved\n")

	service := newTestService(t, project)
	plan := serviceWorkflowPlanFixture()

	result, err := service.MaterializeWorkflowPlan(t.Context(), plan)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Tasks) != 6 {
		t.Fatalf("task tickets = %d", len(result.Tasks))
	}

	checkServiceWorkflowProjection(t, service, project)
}

func serviceWorkflowPlanFixture() core.GeneratedWorkflowPlan {
	gate := core.WorkflowValidation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}

	plan := core.GeneratedWorkflowPlan{MaterializationID: "generation-a", Title: "Delivery", PlanRef: "plan/approved.md", PlanHash: core.WorkflowContentHash([]byte("approved\n")), ClosingValidation: gate}
	for i := 1; i <= 2; i++ {
		slice := core.GeneratedWorkflowSlice{ID: fmt.Sprintf("slice-%d", i), Label: fmt.Sprintf("Slice %d", i), Order: i, Title: "Integration", Contract: "Deliver one integration unit", Validation: gate}
		for j := 1; j <= 3; j++ {
			slice.Tasks = append(slice.Tasks, core.GeneratedWorkflowTask{ID: fmt.Sprintf("task-%d-%d", i, j), Label: fmt.Sprintf("T%d.%d", i, j), Order: j, Title: "Task", Body: "Original instructions", Outcome: "Deliver behavior", Scope: "Bounded task", Acceptance: []string{"Observed result"}, Validation: gate, Priority: "normal", Severity: "normal", Type: "task"})
		}

		plan.Slices = append(plan.Slices, slice)
	}

	return plan
}

func checkServiceWorkflowProjection(t *testing.T, service service, project string) {
	t.Helper()

	index, err := core.NewTicketIndex(t.Context(), service.env, service.assetsFS)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := index.Close(); err != nil {
			t.Error(err)
		}
	})

	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: project})
	if err != nil {
		t.Fatal(err)
	}

	workflow, ok := index.(core.WorkflowIndex)
	if !ok {
		t.Fatal("missing workflow projection")
	}

	page, err := workflow.QueryWorkflow(t.Context(), core.NewWorkflowStore(layout, service.projectContext.ProjectID), core.WorkflowQuery{Kind: "ticket", Limit: 128})
	if err != nil || len(page.Records) != 6 {
		t.Fatalf("indexed tasks = %d, %v", len(page.Records), err)
	}

	if page.Records[0].Task.Key == page.Records[0].ID {
		t.Fatal("native key replaced by artifact ID")
	}
}
