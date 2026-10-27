package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Required generated input is validated before reserving any native key. Caller
// approval cannot substitute another scope, hierarchy or missing companion.
func TestWorkflowPlanOrderApproval(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testing.T, *WorkflowStore, *WorkflowMaterializationOptions)
	}{
		{"missing binding", func(_ *testing.T, _ *WorkflowStore, opts *WorkflowMaterializationOptions) {
			opts.Plan.PlanRevision = nil
		}},
		{"different scope", func(_ *testing.T, _ *WorkflowStore, opts *WorkflowMaterializationOptions) {
			opts.Plan.UseCase = "other-case"
		}},
		{"different task", func(_ *testing.T, _ *WorkflowStore, opts *WorkflowMaterializationOptions) {
			opts.Plan.Slices[0].Tasks[0].Title = "Changed task"
		}},
		{"missing map", func(t *testing.T, store *WorkflowStore, opts *WorkflowMaterializationOptions) {
			if err := os.Remove(filepath.Join(store.layout.Root(), opts.Plan.PlanRevision.TrackerRef)); err != nil {
				t.Fatal(err)
			}
		}},
		{"edited current map", func(t *testing.T, store *WorkflowStore, opts *WorkflowMaterializationOptions) {
			workflowTrackerWrite(t, filepath.Join(store.layout.Root(), opts.Plan.PlanRevision.CurrentTrackerRef), []byte("Operator notes"))
		}},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			store, opts := workflowOrderPlanFixture(t)
			calls := 0
			opts.ReserveTicketKey = func(context.Context) (string, error) { calls++; return "PRJ-1", nil }
			item.mutate(t, store, &opts)

			if _, err := store.MaterializePlan(t.Context(), opts); err == nil {
				t.Fatal("unbound or changed approval accepted")
			}

			if calls != 0 {
				t.Fatal("rejected approval allocated keys", calls)
			}
		})
	}
}

// Approved materialization binds canonical references once. A later draft
// keeps its own unbound map, while an old operation retry returns its original set.
func TestWorkflowPlanOrderHistory(t *testing.T) {
	store, opts := workflowOrderPlanFixture(t)

	first, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	bound := *first.Set.Set.PlanRevision

	before := workflowTrackerRead(t, filepath.Join(store.layout.Root(), bound.TrackerRef))
	checkBoundTicketOrder(t, before, first.Tasks[0])

	later := regenerateWorkflowOrder(t, store, opts.Plan, "Later draft, same intended decomposition.")
	if later.Revision != 2 || later.BodyHash == bound.BodyHash {
		t.Fatal(later, bound)
	}

	retry, err := store.MaterializePlan(t.Context(), opts)
	if err != nil || retry.Set.ID != first.Set.ID || retry.Tasks[0].Task.Key != first.Tasks[0].Task.Key {
		t.Fatal("retry duplicated binding", retry, err)
	}

	if string(before) != string(workflowTrackerRead(t, filepath.Join(store.layout.Root(), bound.TrackerRef))) {
		t.Fatal("new draft overwrote historical map")
	}

	if !strings.Contains(string(workflowTrackerRead(t, filepath.Join(store.layout.Root(), later.TrackerRef))), "Tickets are unbound") {
		t.Fatal("draft allocated tickets")
	}
}

func regenerateWorkflowOrder(t *testing.T, store *WorkflowStore, plan GeneratedWorkflowPlan, body string) PlanRevisionBinding {
	t.Helper()

	generated := orderedPlanCandidate()
	generated.Scope = QuizScope{ProjectID: store.projectID, CaseSlug: plan.UseCase}
	generated.Order = plannedWorkflowOrder(plan)

	generated.Body = body
	if _, err := WriteGeneratedPlan(t.Context(), store.layout.ProjectDir(), generated); err != nil {
		t.Fatal(err)
	}

	return readTestPlanBinding(t, store.layout.ProjectDir(), generated.Scope)
}

// The workflow journal retains binding, maps and records together. A new draft
// cannot race past an interrupted binding; retry recovers the same allocation.
func TestWorkflowPlanOrderRecovery(t *testing.T) {
	for _, stage := range []int{0, 3, 10} {
		t.Run("after image "+strconv.Itoa(stage), func(t *testing.T) { checkWorkflowPlanOrderRecovery(t, stage) })
	}
}

func checkWorkflowPlanOrderRecovery(t *testing.T, stage int) {
	t.Helper()

	store, opts := workflowOrderPlanFixture(t)
	calls := 0
	allocate := opts.ReserveTicketKey
	opts.ReserveTicketKey = func(ctx context.Context) (string, error) { calls++; return allocate(ctx) }

	store.afterReplace = func(index int) error {
		if index == stage {
			return errors.New("binding interrupted")
		}

		return nil
	}
	if _, err := store.MaterializePlan(t.Context(), opts); err == nil {
		t.Fatal("binding did not interrupt")
	}

	if _, err := PlanTargetRevision(t.Context(), store.layout.ProjectDir(), QuizScope{ProjectID: store.projectID}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("pending binding was readable", err)
	}

	reserved := calls
	restarted := NewWorkflowStore(store.layout, store.projectID)

	result, err := restarted.MaterializePlan(t.Context(), opts)
	if err != nil || calls != reserved {
		t.Fatal("recovery allocated duplicate work", calls, reserved, err)
	}

	status, err := restarted.InspectWorkflowTracker(t.Context(), result.Set.ID)
	if err != nil || status.Status != "current" {
		t.Fatal(status, err)
	}

	if _, err := restarted.requiredWorkflowOrderMap(t.Context(), result.Tasks[0]); err != nil {
		t.Fatal("recovered pair is unusable", err)
	}
}

func checkBoundTicketOrder(t *testing.T, before []byte, task WorkflowRecord) {
	t.Helper()

	if !strings.Contains(string(before), workflowRecordBodyRef(task)) || strings.Contains(string(before), "| unbound |") {
		t.Fatal("materialization did not bind canonical references")
	}
}
