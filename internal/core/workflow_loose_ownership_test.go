package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowLooseOwnership excludes ready grouped work discovered after
// delivery even when its use case matches an accepted later-admission filter.
func TestWorkflowLooseOwnership(t *testing.T) {
	f := workflowLooseConfiguredQueue(t, true, func(store *WorkflowStore, request *WorkflowStartRequest) {
		request.Selection.Filters.UseCases = []string{"example"}

		for _, key := range []string{"PRJ-90", "PRJ-91", "PRJ-92"} {
			record := workflowLooseRecord(t, store, key)

			record.Task.UseCase = "example"
			if _, err := store.Write(t.Context(), record, record.Revision); err != nil {
				t.Fatal(err)
			}
		}
	})
	plan := workflowLooseApprovedPlan(t, f.store)
	plan.UseCase = "example"
	number := 200

	grouped, err := f.store.MaterializePlan(t.Context(), WorkflowMaterializationOptions{Plan: plan, ReserveTicketKey: func(context.Context) (string, error) {
		number++

		return fmt.Sprintf("PRJ-%d", number), nil
	}})
	if err != nil {
		t.Fatal(err)
	}

	workflowFinalizeFrontier(t, f)
	workflowAdvanceFrontier(t, f)

	run, err := f.store.ReconcileWorkflowMembership(t.Context(), f.run.ID)
	if err != nil || len(run.Members) != 3 || run.Excluded["slice_owned"] != len(grouped.Tasks) || run.Excluded["later"] != 0 {
		t.Fatal("matching grouped work entered the independent queue", err)
	}
}

func workflowLooseApprovedPlan(t *testing.T, store *WorkflowStore) GeneratedWorkflowPlan {
	t.Helper()

	plan := workflowPlanFixture()

	path := filepath.Join(store.layout.Root(), plan.PlanRef)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("approved\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	return plan
}
