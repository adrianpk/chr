package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

func workflowApprovedAmendment(t *testing.T, store *WorkflowStore, set WorkflowRecord) WorkflowSetAmendmentRequest {
	t.Helper()

	plan := GeneratedWorkflowPlan{Title: "Approved changed delivery", PlanRef: "plan/approved-change.md", PlanHash: contenthash.Sum([]byte("approved amendment\n")), UseCase: set.Set.UseCase, UseCaseOrder: set.Set.UseCaseOrder, ClosingValidation: set.Set.ClosingValidation}
	if err := json.Unmarshal(set.Metadata["materialization_id"], &plan.MaterializationID); err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, descriptor := range set.Set.Slices {
		plan.Slices = append(plan.Slices, workflowApprovedSlice(t, store, descriptor, snapshot))
	}

	if err := os.WriteFile(filepath.Join(store.layout.Root(), plan.PlanRef), []byte("approved amendment\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	return WorkflowSetAmendmentRequest{SetPath: set.Path, ExpectedRevision: set.Revision, ExpectedSourceHash: set.SourceHash, Plan: orderedWorkflowPlan(plan), ReserveTicketKey: func(context.Context) (string, error) { return "PRJ-7", nil }}
}

func workflowApprovedSlice(t *testing.T, store *WorkflowStore, descriptor WorkflowSlice, snapshot WorkflowSnapshot) GeneratedWorkflowSlice {
	t.Helper()

	contract, err := store.readWorkflowSliceContract(descriptor)
	if err != nil {
		t.Fatal(err)
	}

	slice := GeneratedWorkflowSlice{ID: descriptor.ID, Label: fmt.Sprintf("Slice %d", descriptor.Order), Order: descriptor.Order, Title: descriptor.Title, Contract: contract.Contract, Validation: contract.Validation, DependsOn: descriptor.DependsOn}
	for _, record := range snapshot.Records {
		if record.Task == nil || record.Task.SliceID != descriptor.ID {
			continue
		}

		var task GeneratedWorkflowTask
		if err := json.Unmarshal(record.Domain["execution_contract"], &task); err != nil {
			t.Fatal(err)
		}

		slice.Tasks = append(slice.Tasks, task)
	}

	return slice
}

func workflowAppendPlanTask(plan *GeneratedWorkflowPlan) {
	slice := &plan.Slices[len(plan.Slices)-1]
	task := slice.Tasks[len(slice.Tasks)-1]
	task.ID, task.Label, task.Order, task.DependsOn = "task-later", "Later task", task.Order+1, nil
	slice.Tasks = append(slice.Tasks, task)
}

func workflowAppendPlanSlice(plan *GeneratedWorkflowPlan) {
	slice := plan.Slices[len(plan.Slices)-1]
	slice.ID, slice.Label, slice.Order, slice.DependsOn = "slice-later", "Later slice", slice.Order+1, nil
	task := slice.Tasks[0]
	task.ID, task.Label, task.Order, task.DependsOn = "task-later", "Later task", 1, nil
	slice.Tasks = []GeneratedWorkflowTask{task}
	plan.Slices = append(plan.Slices, slice)
}

// TestWorkflowAmendmentAdmission preserves initial images while explicitly
// admitting an approved later task, or reporting its default frozen exclusion.
func TestWorkflowAmendmentAdmission(t *testing.T) {
	for _, include := range []bool{false, true} {
		t.Run(fmt.Sprint(include), func(t *testing.T) {
			store, opts := workflowMaterializationFixture(t)

			hierarchy, err := store.MaterializePlan(t.Context(), opts)
			if err != nil {
				t.Fatal(err)
			}

			target, _ := workflowSelectionCommit(t, store)
			request := WorkflowStartRequest{RequestKey: "membership", Policy: WorkflowExecutionPolicy{IncludeLaterTickets: include}, Selection: WorkflowPreviewRequest{Selector: workflowClaimSelector(hierarchy, "slice_set"), Target: WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}}}
			run := workflowTestStart(t, store, request)
			amend := workflowApprovedAmendment(t, store, hierarchy.Set)
			workflowAppendPlanTask(&amend.Plan)

			changed, err := store.AmendWorkflowSet(t.Context(), amend)
			if err != nil {
				t.Fatal(err)
			}

			current, err := store.ReconcileWorkflowMembership(t.Context(), run.Invocation.ID)
			if err != nil {
				t.Fatal(err)
			}

			checkWorkflowAmendmentAdmission(t, store, run.Invocation, current, changed, include)

			attached, err := store.StartWorkflow(t.Context(), request)
			if err != nil || !attached.Attached || attached.Invocation.ID != current.ID || attached.Invocation.RequestHash != current.RequestHash {
				t.Fatal("later admission changed equivalent request identity", err)
			}
		})
	}
}

func checkWorkflowAmendmentAdmission(t *testing.T, store *WorkflowStore, initial, current WorkflowInvocation, changed WorkflowMaterializationResult, include bool) {
	t.Helper()
	checkWorkflowAmendmentIdentity(t, initial, current, changed)

	count, excluded := 6, 1
	if include {
		count, excluded = 7, 0
	}

	if len(current.Members) != count || current.Excluded["later"] != excluded {
		t.Fatal("amendment lost frozen policy, membership or revision guards")
	}

	if !reflect.DeepEqual(initial.Members, current.Members[:6]) || current.SetImages[0].SourceHash != changed.Set.SourceHash {
		t.Fatal("initial admission images were changed or own amendment was not accepted")
	}

	again, err := NewWorkflowStore(store.layout, store.projectID).ReconcileWorkflowMembership(t.Context(), current.ID)
	if err != nil || again.Revision != current.Revision {
		t.Fatal("unchanged membership generated another checkpoint", err)
	}
}

func checkWorkflowAmendmentIdentity(t *testing.T, initial, current WorkflowInvocation, changed WorkflowMaterializationResult) {
	t.Helper()

	if current.Policy != initial.Policy || current.RequestHash != initial.RequestHash || len(changed.Tasks) != 7 || changed.Set.Revision != 2 || len(changed.Set.Set.Amendments) != 1 {
		t.Fatal("amendment lost frozen policy, identity or revision guards")
	}
}

// TestWorkflowAmendmentGate preserves the exact approved replacement contract:
// removed commands and reasons cannot reappear from retained JSON fields.
func TestWorkflowAmendmentGate(t *testing.T) {
	required := WorkflowValidation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}
	waived := WorkflowValidation{WaiverReason: "Approved explicit closure waiver"}

	cases := []struct {
		name          string
		initial, next WorkflowValidation
	}{
		{"waive", required, waived},
		{"require", waived, required},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, opts := workflowMaterializationFixture(t)
			opts.Plan.ClosingValidation = tc.initial

			hierarchy, err := store.MaterializePlan(t.Context(), opts)
			if err != nil {
				t.Fatal(err)
			}

			request := workflowApprovedAmendment(t, store, hierarchy.Set)
			request.Plan.ClosingValidation = tc.next

			changed, err := store.AmendWorkflowSet(t.Context(), request)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(changed.Set.Set.ClosingValidation, tc.next) {
				t.Fatalf("replacement gate was changed by retained fields: %+v", changed.Set.Set.ClosingValidation)
			}
		})
	}
}

// TestWorkflowAmendmentRejects leaves all records unchanged for stale, reordered,
// reassociated, removed or rewritten input, and for an implicit reopening.
func TestWorkflowAmendmentRejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*WorkflowSetAmendmentRequest)
	}{
		{"stale revision", func(r *WorkflowSetAmendmentRequest) { r.ExpectedRevision++ }},
		{"stale image", func(r *WorkflowSetAmendmentRequest) { r.ExpectedSourceHash = "changed" }},
		{"unchanged plan", func(r *WorkflowSetAmendmentRequest) { r.Plan.PlanHash = contenthash.Sum([]byte("approved\n")) }},
		{"task contract", func(r *WorkflowSetAmendmentRequest) { r.Plan.Slices[0].Tasks[0].Body = "changed" }},
		{"task order", func(r *WorkflowSetAmendmentRequest) { r.Plan.Slices[0].Tasks[0].Order = 9 }},
		{"slice order", func(r *WorkflowSetAmendmentRequest) { r.Plan.Slices[0].Order = 9 }},
		{"removed task", func(r *WorkflowSetAmendmentRequest) { r.Plan.Slices[0].Tasks = r.Plan.Slices[0].Tasks[1:] }},
		{"removed slice", func(r *WorkflowSetAmendmentRequest) { r.Plan.Slices = r.Plan.Slices[1:] }},
		{"reassociation", func(r *WorkflowSetAmendmentRequest) { r.Plan.UseCase = "other-case" }},
		{"active reopening", func(r *WorkflowSetAmendmentRequest) { r.Reopen = true }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, hierarchy := workflowSelectionFixture(t)
			request := workflowApprovedAmendment(t, store, hierarchy.Set)
			tc.mutate(&request)

			if _, err := store.AmendWorkflowSet(t.Context(), request); err == nil {
				t.Fatal("unsafe amendment accepted")
			}

			current, err := store.Read(t.Context(), hierarchy.Set.Path)
			if err != nil || current.SourceHash != hierarchy.Set.SourceHash {
				t.Fatal("rejected amendment mutated the observed set", err)
			}
		})
	}
}

// TestWorkflowAmendmentLiveOwner rejects a live capacity reservation even before
// the worker starts, and a checkpointed claim whose owner needs reconciliation.
func TestWorkflowAmendmentLiveOwner(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)

	claim, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.releaseReservation(); err != nil {
			t.Error(err)
		}
	})

	set, err := store.Read(t.Context(), hierarchy.Set.Path)
	if err != nil {
		t.Fatal(err)
	}

	amend := workflowApprovedAmendment(t, store, set)
	workflowAppendPlanSlice(&amend.Plan)

	if _, err := store.AmendWorkflowSet(t.Context(), amend); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatalf("live reservation bypassed: %v", err)
	}

	if err := store.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	if _, err := store.AmendWorkflowSet(t.Context(), amend); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatalf("checkpointed owner bypassed: %v", err)
	}
}
