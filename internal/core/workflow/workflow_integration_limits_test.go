package workflow

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeLimitFixture(t *testing.T, policy workflowstate.ExecutionPolicy) (*Store, IntegrationEntry) {
	t.Helper()

	store, record, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("true"))
	request.Policy = policy
	run := workflowTestStart(t, store, request).Invocation
	entry := nativeEntryAdmit(t, store, IntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: workflowstate.Ref{Kind: "ticket", ID: record.Task.Key}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"})

	return store, entry
}

// TestNativeLimitPolicy freezes default and explicit ceilings at admission and
// refuses a changed policy as authority for the already retained budget.
func TestNativeLimitPolicy(t *testing.T) {
	cases := []struct {
		name          string
		merger, drift int
		wantMerger    int64
		wantDrift     int64
	}{
		{"defaults", 0, 0, 3, 5},
		{"bounded", 2, 1, 2, 1},
		{"maximum", 10, 10, 10, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, entry := nativeLimitFixture(t, workflowstate.ExecutionPolicy{MaxMergerAttempts: tc.merger, MaxDriftRevisions: tc.drift})
			store = NewWorkflowStore(store.layout, store.projectID)

			budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), entry.InvocationID)
			if err != nil || budget.AdmissionLimits.Merger != tc.wantMerger || budget.AdmissionLimits.Drift != tc.wantDrift || budget.LimitsSnapshot != budget.AdmissionLimits {
				t.Fatal("restart changed the admitted ceilings", err)
			}

			run, err := store.readWorkflowInvocation(entry.InvocationID)
			if err != nil {
				t.Fatal(err)
			}

			run.Policy.MaxMergerAttempts = int(tc.wantMerger + 1)
			if _, err := store.readWorkflowIntegrationBudget(run); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("changed execution policy replaced frozen budget authority", err)
			}
		})
	}
}

// TestNativeLimitExhaustion retains charged counters across queue generation
// replacement and restart; a retry cannot turn a consumed identity into a refill.
func TestNativeLimitExhaustion(t *testing.T) {
	for _, kind := range []string{"merger", "drift", "submission"} {
		t.Run(kind, func(t *testing.T) { testNativeLimitExhaustion(t, kind) })
	}
}

func testNativeLimitExhaustion(t *testing.T, kind string) {
	t.Helper()

	store, entry := nativeLimitFixture(t, workflowstate.ExecutionPolicy{MaxMergerAttempts: 2, MaxDriftRevisions: 1, MaxIntegrationAttempts: 2})

	limit := 2
	if kind == "drift" {
		limit = 1
	}

	for i := range limit {
		r := nativeLimitReservation(entry, kind, i)

		request := IntegrationReserve{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: r}
		if _, err := store.ReserveWorkflowIntegration(t.Context(), request); err != nil {
			t.Fatal(err)
		}

		if _, err := store.ReserveWorkflowIntegration(t.Context(), request); err != nil {
			t.Fatal("same identity was recharged", err)
		}
	}

	var err error

	entry, err = store.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, store, entry, "preparing"))
	if err != nil {
		t.Fatal(err)
	}

	store = NewWorkflowStore(store.layout, store.projectID)

	request := IntegrationReserve{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: nativeLimitReservation(entry, kind, limit)}
	if _, err := store.ReserveWorkflowIntegration(t.Context(), request); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("replacement generation refilled consumed budget", err)
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), entry.InvocationID)
	if err != nil || workflowIntegrationCounterValue(budget.UnitCounters[0], kind) != int64(limit) || len(budget.Reservations) != limit {
		t.Fatal("exhausted budget lost retained history", err)
	}
}

func nativeLimitReservation(entry IntegrationEntry, kind string, number int) IntegrationReservation {
	id := fmt.Sprintf("%s-%d", kind, number)
	r := IntegrationReservation{ID: id, Kind: kind, EntryID: entry.ID, Unit: entry.Unit, ClaimGeneration: entry.ClaimGeneration, AttemptID: id, State: "reserved"}

	switch kind {
	case "merger":
		r.Kind, r.Role, r.RuntimeDeadline = "agent_call", "merger", time.Now().UTC().Add(time.Minute)
	case "drift":
		r.BaseCommit = fmt.Sprintf("%040x", number+1)
	}

	return r
}

// TestNativeDeadlineRetry retains an accepted deadline across restart and
// rejects an extended retry without changing the original reserved charge.
func TestNativeDeadlineRetry(t *testing.T) {
	store, entry := nativeLimitFixture(t, workflowstate.ExecutionPolicy{WorkerTimeoutSeconds: 60})
	r := nativeLimitReservation(entry, "merger", 0)
	r.RuntimeDeadline = time.Now().UTC().Add(30 * time.Second)
	request := IntegrationReserve{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: r}

	before, err := store.ReserveWorkflowIntegration(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	store = NewWorkflowStore(store.layout, store.projectID)
	if again, err := store.ReserveWorkflowIntegration(t.Context(), request); err != nil || !reflect.DeepEqual(before, again) {
		t.Fatal("restart changed the reserved deadline or consumption", err)
	}

	request.Reservation.RuntimeDeadline = request.Reservation.RuntimeDeadline.Add(time.Second)
	if _, err := store.ReserveWorkflowIntegration(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("same call acquired a later deadline", err)
	}

	after, err := store.ReadWorkflowIntegrationBudget(t.Context(), entry.InvocationID)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("deadline mutation changed the retained reservation", err)
	}
}

// TestNativeBudgetDeadline rejects oversized runtime before charging any role;
// metered and unmetered calls share the frozen worker timeout after restart.
func TestNativeBudgetDeadline(t *testing.T) {
	cases := []struct {
		name, role string
		metered    bool
	}{
		{"executor", "executor", true},
		{"reviewer", "reviewer", true},
		{"closing", "closing", true},
		{"merger", "merger", true},
		{"unmetered merger", "merger", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, entry := nativeLimitFixture(t, workflowstate.ExecutionPolicy{WorkerTimeoutSeconds: 5})
			store = NewWorkflowStore(store.layout, store.projectID)

			before, err := store.ReadWorkflowIntegrationBudget(t.Context(), entry.InvocationID)
			if err != nil {
				t.Fatal(err)
			}

			r := nativeBudgetReservation(entry, "oversized-call", tc.role, 10, 10)
			r.Allowance.Deadline = time.Now().UTC().Add(time.Minute)

			if !tc.metered {
				r.RuntimeDeadline, r.Allowance = r.Allowance.Deadline, TokenAllowance{}
			}

			request := IntegrationReserve{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: r}
			if _, err := store.ReserveWorkflowIntegration(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("call extended the retained worker timeout", err)
			}

			after, err := store.ReadWorkflowIntegrationBudget(t.Context(), entry.InvocationID)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("rejected runtime consumed or changed budget", err)
			}
		})
	}
}
