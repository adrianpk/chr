package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestServiceWorkflowMembership uses native key allocation and the shared
// projection, retaining frozen exclusions through the central operation service.
func TestServiceWorkflowMembership(t *testing.T) {
	operations, hierarchy, target := serviceWorkflowSliceFixture(t)
	start := core.WorkflowStartRequest{RequestKey: "service-membership", Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice_set", Identity: hierarchy.Set.ID}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}}}

	run, err := operations.StartWorkflow(t.Context(), start)
	if err != nil {
		t.Fatal(err)
	}

	request := serviceWorkflowAmendmentRequest(t, operations, hierarchy)

	changed, err := operations.AmendWorkflowSet(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	if len(changed.Tasks) != 7 || changed.Tasks[6].Task.Key != "CHR-7" || len(changed.Set.Set.Amendments) != 1 {
		t.Fatal("service bypassed native key allocation or domain amendment")
	}

	current, err := operations.ReconcileWorkflowMembership(t.Context(), run.Invocation.ID)
	if err != nil || current.Excluded["later"] != 1 || len(current.Members) != 6 {
		t.Fatal("service silently broadened the frozen extent", err)
	}
}

func serviceWorkflowAmendmentRequest(t *testing.T, operations service, hierarchy core.WorkflowMaterializationResult) core.WorkflowSetAmendmentRequest {
	t.Helper()

	plan := serviceWorkflowPlanFixture()

	plan.PlanRef, plan.PlanHash = "plan/approved-change.md", core.WorkflowContentHash([]byte("approved amendment\n"))
	for i := range plan.Slices {
		data, err := os.ReadFile(filepath.Join(operations.projectContext.ProjectDir, "ops", "default", hierarchy.Set.Set.Slices[i].ContractRef))
		if err != nil {
			t.Fatal(err)
		}

		var contract struct {
			Validation workflowstate.Validation
		}
		if err := json.Unmarshal(data, &contract); err != nil {
			t.Fatal(err)
		}

		plan.Slices[i].Validation = contract.Validation
	}

	for i, record := range hierarchy.Tasks {
		var task generationcore.GeneratedWorkflowTask
		if err := json.Unmarshal(record.Domain["execution_contract"], &task); err != nil {
			t.Fatal(err)
		}

		plan.Slices[i/3].Tasks[i%3] = task
	}

	task := plan.Slices[1].Tasks[2]
	task.ID, task.Label, task.Order = "task-later", "Later task", 4
	plan.Slices[1].Tasks = append(plan.Slices[1].Tasks, task)
	writeServiceWorkspaceFile(t, filepath.Join(operations.projectContext.ProjectDir, "ops", "default", plan.PlanRef), "approved amendment\n")

	return core.WorkflowSetAmendmentRequest{SetPath: hierarchy.Set.Path, ExpectedRevision: hierarchy.Set.Revision, ExpectedSourceHash: hierarchy.Set.SourceHash, Plan: plan, ReserveTicketKey: func(context.Context) (string, error) { return "", fmt.Errorf("adapter allocator must not execute") }}
}
