package core

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeBudgetFixture(t *testing.T, limit *int64) (*WorkflowStore, WorkflowIntegrationEntry) {
	t.Helper()

	store, request := nativeEntryFixture(t)
	request.TokenLimit = limit

	return store, nativeEntryAdmit(t, store, request)
}

func nativeBudgetGuard(t *testing.T, store *WorkflowStore, id string) WorkflowIntegrationBudgetGuard {
	t.Helper()

	run := workflowTestRead(t, store, id)

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	file, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := store.readWorkflowIntegrationEntry(file, *file.Integration)
	if err != nil {
		t.Fatal(err)
	}

	return WorkflowIntegrationBudgetGuard{InvocationID: id, InvocationRevision: run.Revision, BudgetRevision: budget.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration}
}

func nativeBudgetReservation(entry WorkflowIntegrationEntry, id, role string, input, output int64) WorkflowIntegrationReservation {
	return WorkflowIntegrationReservation{ID: id, Kind: "agent_call", Role: role, EntryID: entry.ID, Unit: entry.Unit, ClaimGeneration: entry.ClaimGeneration, AttemptID: id, State: "reserved", Allowance: WorkflowTokenAllowance{Maximum: input + output, Input: input, Output: output, CapabilityRevision: contenthash.Sum([]byte("whole-call-meter/v1")), WholeCall: true, Deadline: time.Now().UTC().Add(time.Hour)}}
}

// TestNativeBudgetRace prevents concurrent roles from oversubscribing one ceiling
// and attaches retries without another charge, even with an old revision.
func TestNativeBudgetRace(t *testing.T) {
	limit := int64(150)
	store, entry := nativeBudgetFixture(t, &limit)
	guard := nativeBudgetGuard(t, store, entry.InvocationID)

	requests := make([]WorkflowIntegrationReserve, 8)
	for i := range requests {
		requests[i] = WorkflowIntegrationReserve{WorkflowIntegrationBudgetGuard: guard, Reservation: nativeBudgetReservation(entry, fmt.Sprint("call-", i), []string{"executor", "merger", "reviewer", "closing"}[i%4], 60, 40)}
	}

	results := make([]error, len(requests))

	var group sync.WaitGroup
	for i := range requests {
		group.Go(func() { _, results[i] = store.ReserveWorkflowIntegration(t.Context(), requests[i]) })
	}

	group.Wait()

	success := successfulNativeBudgetReservation(t, results)

	if success < 0 {
		t.Fatal("no call admitted")
	}

	budget, err := store.ReserveWorkflowIntegration(t.Context(), requests[success])
	if err != nil || budget.TokensReserved != 100 || len(budget.Reservations) != 1 {
		t.Fatal("retry recharged", err)
	}

	request := requests[(success+1)%len(requests)]

	request.WorkflowIntegrationBudgetGuard = nativeBudgetGuard(t, store, entry.InvocationID)
	if _, err := store.ReserveWorkflowIntegration(t.Context(), request); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("ceiling not enforced", err)
	}
}

// TestNativeBudgetUnknown keeps the full allowance after restart; only reliable
// whole-call evidence can settle once and free the unused part.
func TestNativeBudgetUnknown(t *testing.T) {
	limit := int64(100)
	store, entry := nativeBudgetFixture(t, &limit)

	reserve := WorkflowIntegrationReserve{WorkflowIntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: nativeBudgetReservation(entry, "merger-1", "merger", 70, 30)}
	if _, err := store.ReserveWorkflowIntegration(t.Context(), reserve); err != nil {
		t.Fatal(err)
	}

	request := WorkflowIntegrationSettlement{WorkflowIntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), ReservationID: reserve.Reservation.ID, State: "unknown"}
	if _, err := store.SettleWorkflowIntegration(t.Context(), request); err != nil {
		t.Fatal(err)
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)
	request.WorkflowIntegrationBudgetGuard = nativeBudgetGuard(t, restarted, entry.InvocationID)

	request.State = "released_before_launch"
	if _, err := restarted.SettleWorkflowIntegration(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("unknown call refunded", err)
	}

	proof := WorkflowTokenSettlement{SchemaVersion: 1, ReservationID: reserve.Reservation.ID, AttemptID: reserve.Reservation.AttemptID, CapabilityRevision: reserve.Reservation.Allowance.CapabilityRevision, WholeCall: true, Input: 45, Output: 15}
	ref := retainNativeRecordEvidence(t, restarted, entry.InvocationID, "meter", proof)
	request.State, request.Input, request.Output, request.Proof = "settled", 45, 15, &ref

	budget, err := restarted.SettleWorkflowIntegration(t.Context(), request)
	if err != nil || budget.TokensReserved != 0 || budget.TokensSettled != 60 || budget.UnitCounters[0].Merger != 1 {
		t.Fatal("settlement lost consumption", err)
	}

	nativeBudgetSettlementRetry(t, restarted, entry, request, budget)
}

// TestNativeBudgetRecovery replays each interrupted authoritative replacement
// and returns the original reservation rather than resetting counters.
func TestNativeBudgetRecovery(t *testing.T) {
	for _, failure := range []int{0, 1} {
		t.Run(fmt.Sprint(failure), func(t *testing.T) {
			store, entry := nativeBudgetFixture(t, nil)
			request := WorkflowIntegrationReserve{WorkflowIntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: nativeBudgetReservation(entry, "merger-1", "merger", 10, 10)}
			injected := errors.New("interrupted budget replacement")

			store.afterReplace = func(index int) error {
				if index == failure {
					return injected
				}

				return nil
			}
			if _, err := store.ReserveWorkflowIntegration(t.Context(), request); !errors.Is(err, injected) {
				t.Fatal("fault not reached", err)
			}

			store.afterReplace = nil

			budget, err := store.ReserveWorkflowIntegration(t.Context(), request)
			if err != nil || budget.TokensReserved != 20 || budget.UnitCounters[0].Merger != 1 || len(budget.Reservations) != 1 {
				t.Fatal("replay reset or recharged", err)
			}
		})
	}
}

// TestNativeBudgetBounds rejects unsupported metering, invalid arithmetic and
// stale callbacks without changing either the ledger or its invocation seal.
func TestNativeBudgetBounds(t *testing.T) {
	cases := []struct {
		name   string
		change func(*WorkflowIntegrationReservation)
	}{
		{"output-only", func(r *WorkflowIntegrationReservation) { r.Allowance.WholeCall = false }},
		{"negative", func(r *WorkflowIntegrationReservation) { r.Allowance.Input = -1 }},
		{"overflow", func(r *WorkflowIntegrationReservation) { r.Allowance.Input = math.MaxInt64; r.Allowance.Output = 1 }},
		{"capability", func(r *WorkflowIntegrationReservation) { r.Allowance.CapabilityRevision = "unsupported" }},
		{"deadline", func(r *WorkflowIntegrationReservation) { r.Allowance.Deadline = time.Now().Add(-time.Hour) }},
		{"foreign", func(r *WorkflowIntegrationReservation) { r.EntryID = contenthash.Sum([]byte("foreign")) }},
	}
	store, entry := nativeBudgetFixture(t, nil)

	before := workflowTestRead(t, store, entry.InvocationID)
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := WorkflowIntegrationReserve{WorkflowIntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: nativeBudgetReservation(entry, "invalid", "executor", 10, 10)}
			test.change(&request.Reservation)

			if _, err := store.ReserveWorkflowIntegration(t.Context(), request); err == nil {
				t.Fatal("invalid reservation accepted")
			}

			if after := workflowTestRead(t, store, entry.InvocationID); !reflect.DeepEqual(before, after) {
				t.Fatal("rejection changed owner")
			}
		})
	}
}

func successfulNativeBudgetReservation(t *testing.T, results []error) int {
	t.Helper()

	success := -1

	for i, err := range results {
		if err == nil {
			if success >= 0 {
				t.Fatal("concurrent calls oversubscribed")
			}

			success = i
		} else if !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal(err)
		}
	}

	return success
}

func nativeBudgetSettlementRetry(t *testing.T, restarted *WorkflowStore, entry WorkflowIntegrationEntry, request WorkflowIntegrationSettlement, budget WorkflowIntegrationBudget) {
	t.Helper()

	if again, err := restarted.SettleWorkflowIntegration(t.Context(), request); err != nil || !reflect.DeepEqual(budget, again) {
		t.Fatal("settlement retry changed accounting", err)
	}

	request.Input = 40

	request.WorkflowIntegrationBudgetGuard = nativeBudgetGuard(t, restarted, entry.InvocationID)
	if _, err := restarted.SettleWorkflowIntegration(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("settled usage rewritten", err)
	}
}
