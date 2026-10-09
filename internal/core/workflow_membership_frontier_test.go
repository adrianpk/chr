package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// TestWorkflowMembershipFrontier admits a later slice only after finalized
// delivery and fresh target hydration, without replaying its integrated prefix.
func TestWorkflowMembershipFrontier(t *testing.T) {
	fixture := workflowDeliveryPolicyFixture(t, "rebase", "slice_set", WorkflowExecutionPolicy{IncludeLaterTickets: true})
	workflowFinalizeFrontier(t, fixture)

	before := workflowAdvanceFrontier(t, fixture)

	set, err := fixture.store.Read(t.Context(), fixture.hierarchy.Set.Path)
	if err != nil {
		t.Fatal(err)
	}

	amend := workflowApprovedAmendment(t, fixture.store, set)
	workflowAppendPlanSlice(&amend.Plan)

	changed, err := fixture.store.AmendWorkflowSet(t.Context(), amend)
	if err != nil {
		t.Fatal(err)
	}

	current, err := fixture.store.ReconcileWorkflowMembership(t.Context(), before.ID)
	if err != nil {
		t.Fatal("later admission lost the confirmed frontier", err)
	}

	checkWorkflowMembershipFrontier(t, fixture, current, changed, set)

	claim, err := fixture.store.ClaimWorkflow(t.Context(), before.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.releaseReservation(); err != nil {
			t.Error(err)
		}
	})

	if claim.Generation == 0 || workflowTestRead(t, fixture.store, before.ID).Attempts[len(before.Attempts)].TaskKey != before.Members[3].Admission.Key {
		t.Fatal("admission replayed integrated work")
	}

	if err := fixture.store.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}
}

func checkWorkflowMembershipFrontier(t *testing.T, fixture workflowIntegrationFixture, current WorkflowInvocation, changed WorkflowMaterializationResult, previous WorkflowRecord) {
	t.Helper()

	if len(current.Members) != 7 || current.Cursor != 3 || current.ConfirmedCursor != 3 || len(current.Deliveries) != 1 || current.ExecutionSelection().Target.Commit != fixture.target {
		t.Fatal("later admission lost the confirmed frontier")
	}

	if !reflect.DeepEqual(changed.Set.Set.Slices[0], previous.Set.Slices[0]) {
		t.Fatal("amendment replaced integrated slice projections")
	}
}

// TestWorkflowMembershipLaterSet distinguishes future associated sets from
// retroactive imports before the confirmed use-case frontier.
func TestWorkflowMembershipLaterSet(t *testing.T) {
	for _, retroactive := range []bool{false, true} {
		t.Run(map[bool]string{false: "later", true: "retroactive"}[retroactive], func(t *testing.T) {
			fixture := workflowDeliveryPolicyFixture(t, "rebase", "use_case", WorkflowExecutionPolicy{IncludeLaterTickets: true})
			workflowFinalizeFrontier(t, fixture)

			before := workflowAdvanceFrontier(t, fixture)

			later := workflowMaterializeLaterSet(t, fixture)
			if retroactive {
				workflowImportEarlierSet(t, fixture.store, later.Set, fixture.hierarchy.Set.CreatedAt)
			}

			current, err := fixture.store.ReconcileWorkflowMembership(t.Context(), before.ID)
			checkWorkflowMembershipLaterSet(t, before, current, later, retroactive, err)
		})
	}
}

func checkWorkflowMembershipLaterSet(t *testing.T, before, current WorkflowInvocation, later WorkflowMaterializationResult, retroactive bool, failure error) {
	t.Helper()

	if retroactive {
		if !errors.Is(failure, ErrWorkflowConflict) || current.Revision != before.Revision {
			t.Fatal("retroactive use-case work crossed confirmed progress", failure)
		}

		return
	}

	if failure != nil {
		t.Fatal(failure)
	}

	checkWorkflowMembershipLaterScope(t, current, later)
}

func checkWorkflowMembershipLaterScope(t *testing.T, current WorkflowInvocation, later WorkflowMaterializationResult) {
	t.Helper()

	if len(current.Members) != 12 || len(current.SetImages) != 2 || current.Members[6].Admission.SetID != later.Set.ID || current.Cursor != 3 || current.Excluded["unsliced"] != 1 {
		t.Fatal("later set changed frozen scope, ordering or exclusions")
	}
}

func workflowMaterializeLaterSet(t *testing.T, fixture workflowIntegrationFixture) WorkflowMaterializationResult {
	t.Helper()

	plan := workflowPlanFixture()

	plan.UseCase, plan.MaterializationID, plan.Title = "ordered-case", "generation-later", "Later set"
	for i := range plan.Slices {
		plan.Slices[i].ID += "-later"
		for j := range plan.Slices[i].Tasks {
			plan.Slices[i].Tasks[j].ID += "-later"
		}
	}

	number := 100

	result, err := fixture.store.MaterializePlan(t.Context(), WorkflowMaterializationOptions{Plan: plan, ReserveTicketKey: func(context.Context) (string, error) { number++; return fmt.Sprintf("PRJ-%d", number), nil }})
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func workflowImportEarlierSet(t *testing.T, store *WorkflowStore, set WorkflowRecord, first time.Time) {
	t.Helper()

	// An imported set can have an older immutable creation time. Its existing
	// ticket timestamps do not determine use-case set order.
	set.CreatedAt = first.Add(-time.Hour)

	data, err := encodeWorkflowRecord(set)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(store.layout.Root(), set.Path), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowMembershipExternalOrder prevents a raw external edit from being
// accepted as an owned amendment, even with explicit later admission enabled.
func TestWorkflowMembershipExternalOrder(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice_set")
	request.Policy.IncludeLaterTickets = true
	run := workflowTestStart(t, store, request)
	task := hierarchy.Tasks[0]
	task.Task.Order = 9

	data, err := encodeWorkflowRecord(task)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(store.layout.Root(), task.Path), data, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ReconcileWorkflowMembership(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("external reorder was silently reconciled", err)
	}
}
