package workflow

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeBudgetGrants extends only an explicit positive ceiling, preserves
// consumed actions and attaches a repeated operator grant without adding it twice.
func TestNativeBudgetGrants(t *testing.T) {
	limit := int64(100)

	store, entry := nativeBudgetFixture(t, &limit)
	for i := range 3 {
		request := IntegrationReserve{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: nativeBudgetReservation(entry, fmt.Sprint("merger-", i), "merger", 10, 10)}
		if _, err := store.ReserveWorkflowIntegration(t.Context(), request); err != nil {
			t.Fatal(err)
		}
	}

	fourth := IntegrationReserve{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: nativeBudgetReservation(entry, "merger-4", "merger", 10, 10)}
	if _, err := store.ReserveWorkflowIntegration(t.Context(), fourth); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("merger ceiling bypassed", err)
	}

	grant := IntegrationGrantRequest{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Grant: workflowstate.IntegrationGrant{ID: "operator-extension", Kind: "merger", Amount: 1, Actor: "operator", Reason: "accepted additional reconciliation"}}

	budget, err := store.GrantWorkflowIntegration(t.Context(), grant)
	if err != nil || budget.UnitCounters[0].Merger != 3 || budget.LimitsSnapshot.Merger != 4 {
		t.Fatal("grant reset consumed merger calls", err)
	}

	if again, err := store.GrantWorkflowIntegration(t.Context(), grant); err != nil || !reflect.DeepEqual(budget, again) {
		t.Fatal("grant retried twice", err)
	}

	fourth.IntegrationBudgetGuard = nativeBudgetGuard(t, store, entry.InvocationID)
	if _, err := store.ReserveWorkflowIntegration(t.Context(), fourth); err != nil {
		t.Fatal(err)
	}

	nativeBudgetInvalidGrants(t, store, entry)
}

func nativeBudgetInvalidGrants(t *testing.T, store *Store, entry IntegrationEntry) {
	t.Helper()

	for _, amount := range []int64{0, -1, math.MaxInt64} {
		request := IntegrationGrantRequest{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Grant: workflowstate.IntegrationGrant{ID: fmt.Sprint("invalid-", amount), Kind: "tokens", Amount: amount, Actor: "operator", Reason: "explicit boundary check"}}
		if _, err := store.GrantWorkflowIntegration(t.Context(), request); err == nil {
			t.Fatal("invalid grant accepted")
		}
	}
}

// TestNativeBudgetDrift charges each accepted replacement base once and retains
// drift consumption through a new preparation generation and process restart.
func TestNativeBudgetDrift(t *testing.T) {
	store, entry := nativeBudgetFixture(t, nil)
	request := IntegrationReserve{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: IntegrationReservation{ID: "drift-1", Kind: "drift", EntryID: entry.ID, Unit: entry.Unit, AttemptID: "replacement-1", BaseCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", State: "reserved"}}

	budget, err := store.ReserveWorkflowIntegration(t.Context(), request)
	if err != nil || budget.UnitCounters[0].Drift != 1 {
		t.Fatal(err)
	}

	if again, err := store.ReserveWorkflowIntegration(t.Context(), request); err != nil || !reflect.DeepEqual(budget, again) {
		t.Fatal("same base recharged", err)
	}

	request.IntegrationBudgetGuard = nativeBudgetGuard(t, store, entry.InvocationID)

	request.Reservation.ID = "another-id"
	if _, err := store.ReserveWorkflowIntegration(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("same base gained another budget", err)
	}

	entry, err = store.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, store, entry, "preparing"))
	if err != nil {
		t.Fatal(err)
	}

	store = NewWorkflowStore(store.layout, store.projectID)

	budget, err = store.ReadWorkflowIntegrationBudget(t.Context(), entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Drift != 1 || entry.ClaimGeneration != 1 {
		t.Fatal("new generation reset drift", err)
	}
}

// TestNativeBudgetLegacy preserves real existing publication attempts at native
// admission and leaves the uncertain legacy journal byte-for-byte unchanged.
func TestNativeBudgetLegacy(t *testing.T) {
	f, provider := newWorkflowSubmissionFixture(t, "merge")

	submission, dispatch, err := f.store.beginWorkflowSubmission(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f))
	if err != nil || !dispatch {
		t.Fatal(err)
	}

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	path, err := workflowSubmissionPath(t.Context(), f.store.layout.ProjectDir(), run.Submission.JournalKey)
	if err != nil {
		t.Fatal(err)
	}

	before := mustReadNativeEntry(t, path)
	entry := nativeEntryAdmit(t, f.store, IntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: run.Lane.Unit, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/" + run.PullRequest.Request.BaseBranch})

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), entry.InvocationID)
	if err != nil || budget.UnitCounters[0].ExistingSubmissions != 1 || budget.UnitCounters[0].Submission != 1 || len(budget.Usage.Attempts) != len(run.Attempts) {
		t.Fatal("native admission refilled legacy usage", err)
	}

	nativeBudgetLegacyUncertainty(t, f, provider, entry, submission, path, before)
}

func nativeBudgetLegacyUncertainty(t *testing.T, f workflowPreparationFixture, provider *workflowSubmissionProviderFixture, entry IntegrationEntry, submission workflowstate.Submission, path string, before []byte) {
	t.Helper()

	if again, dispatch, err := f.store.BeginWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, entry.InvocationID), provider); !errors.Is(err, ErrWorkflowObservationRequired) || dispatch || again.ID != submission.ID {
		t.Fatal("legacy uncertainty converted to empty native intent", err)
	}

	if after := mustReadNativeEntry(t, path); string(after) != string(before) {
		t.Fatal("native admission rewrote legacy journal bytes")
	}
}
