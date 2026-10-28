package workflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowAmendmentRecovery interrupts the approval, ticket, set and run
// projections. Restart replays a single membership transition without reallocation.
func TestWorkflowAmendmentRecovery(t *testing.T) {
	for _, stop := range []int{0, 2, 4, 6} {
		t.Run(fmt.Sprint(stop), func(t *testing.T) {
			store, hierarchy, request := workflowClaimFixture(t)
			request.Selection.Selector = workflowClaimSelector(hierarchy, "slice_set")
			run := workflowTestStart(t, store, request)
			amend := workflowApprovedAmendment(t, store, hierarchy.Set)
			workflowAppendPlanTask(&amend.Plan)

			store.afterReplace = func(index int) error {
				if index == stop {
					return errors.New("owned amendment interruption")
				}

				return nil
			}

			if _, err := store.AmendWorkflowSet(t.Context(), amend); err == nil {
				t.Fatal("interruption was not reached")
			}

			restarted := NewWorkflowStore(store.layout, store.projectID)

			set, err := restarted.Read(t.Context(), hierarchy.Set.Path)
			if err != nil || set.Revision != hierarchy.Set.Revision+1 || len(set.Set.Amendments) != 1 {
				t.Fatal("journal mixed amendment images", err)
			}

			checkWorkflowAmendmentRestart(t, restarted, run.Invocation.ID, set, amend)
		})
	}
}

func checkWorkflowAmendmentRestart(t *testing.T, store *Store, id string, set workflowstate.Record, amend SetAmendmentRequest) {
	t.Helper()

	current, err := store.ReconcileWorkflowMembership(t.Context(), id)
	if err != nil || len(current.Members) != 6 || current.Excluded["later"] != 1 || current.SetImages[0].SourceHash != set.SourceHash {
		t.Fatal("restart lost frozen scope or accepted set image", err)
	}

	if _, err := store.AmendWorkflowSet(t.Context(), amend); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale amendment allocated another transition", err)
	}
}

// TestWorkflowAmendmentEvidence rejects edits to the retained approval before
// later admission can turn it into executable work.
func TestWorkflowAmendmentEvidence(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice_set")
	request.Policy.IncludeLaterTickets = true
	run := workflowTestStart(t, store, request)
	amend := workflowApprovedAmendment(t, store, hierarchy.Set)
	workflowAppendPlanTask(&amend.Plan)

	result, err := store.AmendWorkflowSet(t.Context(), amend)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(store.layout.Root(), result.ReceiptRef), []byte("changed approval"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ReconcileWorkflowMembership(t.Context(), run.Invocation.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed approved amendment became executable", err)
	}
}
