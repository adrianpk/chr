package core

import (
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func materializeSelectionSet(t *testing.T, store *WorkflowStore, opts WorkflowMaterializationOptions, suffix, useCase string, order int) WorkflowMaterializationResult {
	t.Helper()

	opts.Plan.Slices = append([]generationcore.GeneratedWorkflowSlice(nil), opts.Plan.Slices...)
	for i := range opts.Plan.Slices {
		opts.Plan.Slices[i].Tasks = append([]generationcore.GeneratedWorkflowTask(nil), opts.Plan.Slices[i].Tasks...)
	}

	opts.Plan.MaterializationID = "generation-" + suffix

	opts.Plan.UseCase, opts.Plan.UseCaseOrder = useCase, order
	for i := range opts.Plan.Slices {
		slice := &opts.Plan.Slices[i]

		slice.ID += "-" + suffix
		for j := range slice.Tasks {
			slice.Tasks[j].ID += "-" + suffix
		}
	}

	result, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

// TestWorkflowSelectionHierarchy ignores interleaved key creation and display ranks.
func TestWorkflowSelectionHierarchy(t *testing.T) {
	store, opts := workflowMaterializationFixture(t)
	second := materializeSelectionSet(t, store, opts, "a2", "alpha", 2)
	materializeSelectionSet(t, store, opts, "b1", "beta", 1)

	first := materializeSelectionSet(t, store, opts, "a1", "alpha", 1)

	preview := workflowTestPreview(t, store, "use_case", "alpha")
	if preview.SetOrderRule != "use_case_order" || len(preview.Members) != 12 {
		t.Fatalf("wrong set order: %+v", preview)
	}

	for i, member := range preview.Members {
		expected := first.Tasks
		if i >= 6 {
			expected = second.Tasks
		}

		if member.Key != expected[i%6].Task.Key {
			t.Fatalf("interleaved keys reordered hierarchy at %d", i)
		}
	}
}

// TestWorkflowSelectionLocalOrder uses positive gapped task and slice orders, never labels or keys.
func TestWorkflowSelectionLocalOrder(t *testing.T) {
	store, opts := workflowMaterializationFixture(t)

	opts.Plan.Slices[0].Order, opts.Plan.Slices[1].Order = 20, 10
	for i := range opts.Plan.Slices {
		for j := range opts.Plan.Slices[i].Tasks {
			opts.Plan.Slices[i].Tasks[j].Order = []int{30, 10, 20}[j]
		}
	}

	result := materializeSelectionSet(t, store, opts, "gaps", "alpha", 0)

	preview := workflowTestPreview(t, store, "slice_set", result.Set.ID)
	for i, member := range preview.Members {
		if member.TaskOrder != []int{10, 20, 30}[i%3] {
			t.Fatal("task orders were renumbered or ignored")
		}

		if i < 3 && member.SliceID != opts.Plan.Slices[1].ID+"-gaps" {
			t.Fatal("slice order ignored")
		}
	}
}

// TestWorkflowSelectionSetOrder covers immutable-time fallback, mixed orders and stale membership.
func TestWorkflowSelectionSetOrder(t *testing.T) {
	cases := []struct {
		name          string
		first, second int
		fail          bool
	}{
		{"creation fallback", 0, 0, false}, {"explicit", 2, 1, false}, {"mixed", 0, 1, true}, {"duplicate", 1, 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, first, second, err := selectionOrderedPair(t, tc.first, tc.second)
			if (err != nil) != tc.fail {
				t.Fatalf("ordering contract = %v", err)
			}

			if tc.fail {
				return
			}

			preview := workflowTestPreview(t, store, "use_case", "alpha")

			expected := first.Set.ID
			if tc.first > tc.second {
				expected = second.Set.ID
			}

			if preview.Members[0].SetID != expected {
				t.Fatal("wrong set frontier")
			}
		})
	}
}

// TestWorkflowSelectionMembership prevents stale slice fingerprints from admitting changed tasks.
func TestWorkflowSelectionMembership(t *testing.T) {
	store, result := workflowSelectionFixture(t)
	task := result.Tasks[0]
	task.Task.Order = 9
	workflowTestWrite(t, store, task)

	if _, err := store.Preview(t.Context(), workflowstate.WorkflowPreviewRequest{Selector: workflowstate.WorkflowSelector{Extent: "slice_set", Identity: result.Set.ID}, Limit: 128}); err == nil {
		t.Fatalf("changed membership admitted for %s", result.Set.ID)
	}
}

func selectionOrderedPair(t *testing.T, firstOrder, secondOrder int) (*WorkflowStore, WorkflowMaterializationResult, WorkflowMaterializationResult, error) {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)
	first := materializeSelectionSet(t, store, opts, "first", "alpha", firstOrder)
	opts.Plan.MaterializationID = "generation-second"

	opts.Plan.UseCase, opts.Plan.UseCaseOrder = "alpha", secondOrder
	for i := range opts.Plan.Slices {
		opts.Plan.Slices[i].ID += "-second"
		for j := range opts.Plan.Slices[i].Tasks {
			opts.Plan.Slices[i].Tasks[j].ID += "-second"
		}
	}

	second, err := store.MaterializePlan(t.Context(), opts)

	return store, first, second, err
}
