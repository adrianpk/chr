package workflow

import (
	"errors"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeMergerUnmetered preserves finite attempts and runtime without
// inventing an accounting capability. A finite invocation rejects this mode.
func TestNativeMergerUnmetered(t *testing.T) {
	limit := int64(100)
	for _, tokens := range []*int64{nil, &limit} {
		store, entry := nativeBudgetFixture(t, tokens)
		request := IntegrationReserve{IntegrationBudgetGuard: nativeBudgetGuard(t, store, entry.InvocationID), Reservation: IntegrationReservation{ID: "unmetered", Kind: "agent_call", Role: "merger", EntryID: entry.ID, Unit: entry.Unit, ClaimGeneration: entry.ClaimGeneration, AttemptID: "unmetered", State: "reserved", RuntimeDeadline: time.Now().UTC().Add(time.Minute)}}

		budget, err := store.ReserveWorkflowIntegration(t.Context(), request)
		if tokens != nil {
			if !errors.Is(err, ErrWorkflowIntegrationBudget) {
				t.Fatal("finite invocation accepted unmetered call", err)
			}

			continue
		}

		if err != nil || budget.UnitCounters[0].Merger != 1 || budget.TokensReserved != 0 {
			t.Fatal("unmetered call lost its finite action charge", err)
		}

		again, err := store.ReserveWorkflowIntegration(t.Context(), request)
		if err != nil || !workflowNativeJSONEqual(budget, again) {
			t.Fatal("retry charged another automatic attempt", err)
		}
	}
}

// TestNativeMergerSettlement accepts accounting only from the certified driver,
// retains the exact proof and reuses the existing budget verifier after restart.
func TestNativeMergerSettlement(t *testing.T) {
	f, request, harness := nativeMergerExecution(t, "merger-metered")

	var certificate MergerHarnessCertificate
	if err := f.store.readWorkflowNativeEvidence(f.run.ID, harness.Certificate, &certificate); err != nil {
		t.Fatal(err)
	}

	certificate.EnforcesWholeCall, certificate.ReportsWholeCall = true, true
	certificate.Accounting = workflowTokenAccounting
	harness.Certificate = retainNativeRecordEvidence(t, f.store, f.run.ID, "metered-harness", certificate)
	request.Input.TokenAllowance = &workflowstate.TokenAllowance{Maximum: 100, Input: 60, Output: 40, CapabilityRevision: certificate.CapabilityRevision, WholeCall: true, Deadline: request.Input.Deadline}

	var err error

	request.Input, err = SealWorkflowMergerInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	attempt, err := f.store.ExecuteWorkflowMerger(t.Context(), request, harness, f.provider)
	if err != nil || attempt.TrustedUsage == nil {
		t.Fatal("compatible driver accounting was lost", err)
	}

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	budget, err := restarted.ReadWorkflowIntegrationBudget(t.Context(), f.run.ID)
	if err != nil || budget.TokensReserved != 0 || budget.TokensSettled != 30 || budget.UnitCounters[0].Merger != 1 {
		t.Fatal("settled charge was not retained", err)
	}

	if err := restarted.checkWorkflowTokenSettlement(f.run.ID, budget.Reservations[0]); err != nil {
		t.Fatal(err)
	}
}

// TestNativeMergerMeterGuards prevents supplementary model prose, partial usage
// or a changed capability from refunding a whole-call reservation.
func TestNativeMergerMeterGuards(t *testing.T) {
	_, input, _ := nativeMergerInput(t)
	revision := contenthash.Sum([]byte("meter/v1"))
	input.TokenAllowance = &workflowstate.TokenAllowance{Maximum: 100, Input: 60, Output: 40, WholeCall: true, CapabilityRevision: revision, Deadline: input.Deadline}
	certificate := MergerHarnessCertificate{ReportsWholeCall: true, CapabilityRevision: revision}

	cases := map[string]workflowstate.MergerUsage{
		"partial":        {Input: 20, Output: 10, WholeCall: false, CapabilityRevision: revision},
		"revision":       {Input: 20, Output: 10, WholeCall: true, CapabilityRevision: contenthash.Sum([]byte("meter/v2"))},
		"input ceiling":  {Input: 61, Output: 10, WholeCall: true, CapabilityRevision: revision},
		"output ceiling": {Input: 10, Output: 41, WholeCall: true, CapabilityRevision: revision},
	}
	for name, usage := range cases {
		t.Run(name, func(t *testing.T) {
			if err := checkWorkflowMergerHarnessUsage(input, certificate, &usage); err == nil {
				t.Fatal("untrusted accounting accepted")
			}
		})
	}

	certificate.ReportsWholeCall = false

	usage := workflowstate.MergerUsage{Input: 20, Output: 10, WholeCall: true, CapabilityRevision: revision}
	if err := checkWorkflowMergerHarnessUsage(input, certificate, &usage); err == nil {
		t.Fatal("uncertified accounting accepted")
	}
}
