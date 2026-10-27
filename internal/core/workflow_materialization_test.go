package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowMaterializationFixture(t *testing.T) (*WorkflowStore, WorkflowMaterializationOptions) {
	t.Helper()

	store := workflowTestStore(t)
	plan := workflowPlanFixture()

	path := filepath.Join(store.layout.Root(), plan.PlanRef)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("approved\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	number := 0

	return store, WorkflowMaterializationOptions{Plan: plan, ReserveTicketKey: func(ctx context.Context) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		number++

		return fmt.Sprintf("PRJ-%d", number), nil
	}}
}

// TestWorkflowMaterialization verifies real canonical files for six task tickets
// and two integration units, including resolution of a forward task reference.
func TestWorkflowMaterialization(t *testing.T) {
	store, opts := workflowMaterializationFixture(t)
	opts.Plan.Slices[0].Tasks[0].DependsOn = []workflowstate.Ref{{Kind: "ticket", ID: "task-1-3"}}

	result, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Tasks) != 6 || len(result.Set.Set.Slices) != 2 {
		t.Fatalf("wrong materialized hierarchy: %d tasks", len(result.Tasks))
	}

	if result.Tasks[0].Task.DependsOn[0].ID != result.Tasks[2].Task.Key {
		t.Fatal("forward dependency was not resolved to its ticket key")
	}

	checkMaterializedHierarchy(t, store, result)

	snapshot, err := store.Snapshot(t.Context())
	if err != nil || len(snapshot.Records) != 7 {
		t.Fatalf("snapshot = %d records, %v", len(snapshot.Records), err)
	}
}

func checkMaterializedHierarchy(t *testing.T, store *WorkflowStore, result WorkflowMaterializationResult) {
	t.Helper()

	for _, slice := range result.Set.Set.Slices {
		if slice.TaskInputHash != WorkflowTaskInputHash(result.Tasks, slice.ID) {
			t.Fatal("slice task fingerprint differs")
		}

		contract, err := os.ReadFile(filepath.Join(store.layout.Root(), slice.ContractRef))
		if err != nil || contenthash.Sum(contract) != slice.ContractHash {
			t.Fatalf("slice contract = %v", err)
		}
	}

	checkMaterializedTasks(t, store, result)
}

// TestWorkflowMaterializationRejects prevents partial durable writes for stale
// source images, unresolved references, cycles and foreign scoped ownership.
func TestWorkflowMaterializationRejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*generationcore.GeneratedWorkflowPlan)
	}{
		{"changed source", func(p *generationcore.GeneratedWorkflowPlan) { p.PlanHash = "changed" }},
		{"unknown task", func(p *generationcore.GeneratedWorkflowPlan) {
			p.Slices[0].Tasks[0].DependsOn = []workflowstate.Ref{{Kind: "ticket", ID: "absent"}}
		}},
		{"dependency cycle", func(p *generationcore.GeneratedWorkflowPlan) {
			p.Slices[0].Tasks[0].DependsOn = []workflowstate.Ref{{Kind: "slice", ID: p.Slices[1].ID}}
			p.Slices[1].DependsOn = []workflowstate.Ref{{Kind: "ticket", ID: p.Slices[0].Tasks[0].ID}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, opts := workflowMaterializationFixture(t)
			tc.mutate(&opts.Plan)

			if _, err := store.MaterializePlan(t.Context(), opts); err == nil {
				t.Fatal("invalid input accepted")
			}

			snapshot, err := store.Snapshot(t.Context())
			if err != nil || len(snapshot.Records) != 0 {
				t.Fatalf("partial durable workflow: %d, %v", len(snapshot.Records), err)
			}
		})
	}
}

func checkMaterializedTasks(t *testing.T, store *WorkflowStore, result WorkflowMaterializationResult) {
	t.Helper()

	for i, record := range result.Tasks {
		if record.Task.SetID != result.Set.ID || record.Task.Order != i%3+1 || record.Task.Phase != "pending" {
			t.Fatal("wrong task membership or phase")
		}

		persisted, err := store.Read(t.Context(), record.Path)
		if err != nil || persisted.Task.Key != record.Task.Key {
			t.Fatalf("persisted task = %v", err)
		}

		if persisted.Task.Key == persisted.ID {
			t.Fatal("artifact identity replaced native ticket key")
		}
	}
}
