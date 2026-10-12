package core

import (
	"errors"
	"testing"
	"time"
)

// TestNativeCandidateRepairGuard rejects changed host findings and policies
// before consuming a merger call. Task repair cannot bypass the native route.
func TestNativeCandidateRepairGuard(t *testing.T) {
	for _, change := range []string{"findings", "policy", "scope"} {
		t.Run(change, func(t *testing.T) { nativeCandidateRepairGuard(t, change) })
	}
}

func nativeCandidateRepairGuard(t *testing.T, change string) {
	t.Helper()

	f, request, harness := nativeCandidateRepairFixture(t, "rebase", "combined_gate_regression")
	store, id := f.f.store, f.f.run.ID
	before := workflowTestRead(t, store, id)

	switch change {
	case "findings":
		request.Input.FindingRefs = []WorkflowEvidenceRef{request.Input.ValidationContractRef}
	case "policy":
		f.provider.mu.Lock()
		f.provider.policyRevision = "changed-policy"
		f.provider.mu.Unlock()
	case "scope":
		request.Input.ScopeHash = workflowHash([]byte("expanded-scope"))
	}

	var err error

	request.Input, err = SealWorkflowMergerInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.ExecuteWorkflowMergerRepair(t.Context(), request, harness, f.provider, f.f.git); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("changed admission acquired merger authority", err)
	}

	if after := workflowTestRead(t, store, id); after.Revision != before.Revision || len(after.Attempts) != len(before.Attempts) {
		t.Fatal("rejected admission consumed an attempt or changed retained evidence")
	}

	key := before.Members[len(before.Members)-1].Admission.Key
	if _, err := store.RepairWorkflowSlice(t.Context(), id, key); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("native integration failure rewound source tasks", err)
	}

	assertWorkflowPreparationSource(t, f.f)
}

// TestNativeCandidateClean proves a passing candidate has no implicit merger
// call, repair binding or charged merger allowance.
func TestNativeCandidateClean(t *testing.T) {
	f, candidate := nativeReviewFixture(t, ReviewWorkflowModeTrustedAutoMerge)
	store, id := f.f.store, f.f.run.ID
	nativeReviewAuthorize(t, f)

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || len(budget.Reservations) != 0 || len(candidate.Input.Git.Provenance.Repairs) != 0 || workflowTestRead(t, store, id).CandidateRepair != nil {
		t.Fatal("clean result launched or charged a merger", err)
	}
}
