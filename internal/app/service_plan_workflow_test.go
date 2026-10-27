package app

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
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

	operation, err := service.MaterializePlanTracking(t.Context(), PlanTrackingRequest{Workflow: &plan})
	if err != nil {
		t.Fatal(err)
	}

	if operation.Workflow == nil || len(operation.Workflow.Tasks) != 6 {
		t.Fatal("missing task workflow result")
	}

	checkServiceWorkflowProjection(t, service, project)
}

func serviceWorkflowPlanFixture() generationcore.GeneratedWorkflowPlan {
	gate := workflowstate.Validation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}

	plan := generationcore.GeneratedWorkflowPlan{MaterializationID: "generation-a", Title: "Delivery", PlanRef: "plan/approved.md", PlanHash: authoringcore.WorkflowContentHash([]byte("approved\n")), ClosingValidation: gate}
	for i := 1; i <= 2; i++ {
		slice := generationcore.GeneratedWorkflowSlice{ID: fmt.Sprintf("slice-%d", i), Label: fmt.Sprintf("Slice %d", i), Order: i, Title: "Integration", Contract: "Deliver one integration unit", Validation: gate}
		for j := 1; j <= 3; j++ {
			slice.Tasks = append(slice.Tasks, generationcore.GeneratedWorkflowTask{ID: fmt.Sprintf("task-%d-%d", i, j), Label: fmt.Sprintf("T%d.%d", i, j), Order: j, Title: "Task", Body: "Original instructions", Outcome: "Deliver behavior", Scope: "Bounded task", Acceptance: []string{"Observed result"}, Validation: gate, Priority: "normal", Severity: "normal", Type: "task"})
		}

		plan.Slices = append(plan.Slices, slice)
	}

	return plan
}

func checkServiceWorkflowProjection(t *testing.T, service service, project string) {
	t.Helper()

	index, err := core.NewWorkflowTicketIndex(t.Context(), service.env, service.assetsFS)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := index.Close(); err != nil {
			t.Error(err)
		}
	})

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: project})
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

// TestServiceWorkflowAdoptionOwnership uses the real legacy execution database
// to distinguish unfinished ownership from completed historical attempts.
func TestServiceWorkflowAdoptionOwnership(t *testing.T) {
	cases := []struct {
		name    string
		status  core.ExecutionStatus
		blocked bool
	}{
		{"pending", core.ExecutionStatusPending, true},
		{"running", core.ExecutionStatusRunning, true},
		{"failed history", core.ExecutionStatusFailed, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := newTestService(t, newTestWorkspace(t))

			repository, err := core.NewExecutionRepository(t.Context(), service.env, service.assetsFS)
			if err != nil {
				t.Fatal(err)
			}

			now := time.Now().UTC()

			record := core.ExecutionRecord{ID: "legacy-attempt", ProjectID: service.projectContext.ProjectID, TicketKey: "PRJ-1", Status: tc.status, CreatedAt: now, UpdatedAt: now}
			if err := repository.CreateExecution(t.Context(), record); err != nil {
				t.Fatal(err)
			}

			if err := repository.Close(); err != nil {
				t.Fatal(err)
			}

			err = service.checkWorkflowAdoptionExecutions(t.Context(), []string{"PRJ-1"})
			if (err != nil) != tc.blocked {
				t.Fatalf("ownership observation = %v, blocked = %v", err, tc.blocked)
			}
		})
	}
}
