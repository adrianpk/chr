package workflow

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowOrderAmendment(t *testing.T, store *Store, set workflowstate.Record) SetAmendmentRequest {
	t.Helper()

	request := workflowApprovedAmendment(t, store, set)
	workflowAppendPlanTask(&request.Plan)

	binding := regenerateWorkflowOrder(t, store, request.Plan, "Approved additional future work.")
	request.Plan.PlanRevision = &binding
	request.Plan.PlanRef, request.Plan.PlanHash = binding.BodyRef, binding.BodyHash

	return request
}

// Explicit amendment retains existing ticket identities and historical maps,
// and atomically binds the approved new revision to its added canonical ticket.
func TestWorkflowPlanOrderAmendment(t *testing.T) {
	store, hierarchy, old := workflowOrderFixture(t)
	previous := workflowTrackerRead(t, filepath.Join(store.layout.Root(), old.PlanRevision.TrackerRef))
	request := workflowOrderAmendment(t, store, hierarchy.Set)

	amended, err := store.AmendWorkflowSet(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	if len(amended.Tasks) != 7 || amended.Tasks[0].Task.Key != hierarchy.Tasks[0].Task.Key || amended.Set.Set.PlanRevision.Revision != 2 {
		t.Fatal("amendment lost identities or revision")
	}

	if string(previous) != string(workflowTrackerRead(t, filepath.Join(store.layout.Root(), old.PlanRevision.TrackerRef))) {
		t.Fatal("amendment rewrote earlier revision map")
	}

	current, err := store.InspectWorkflowTracker(t.Context(), hierarchy.Set.ID)
	if err != nil || current.Status != "current" || current.SourceHash == old.SourceHash {
		t.Fatal(current, err)
	}
}

// An active claim prevents applying a new approved revision. Its draft can be
// generated independently, but amendment allocates nothing and retains history.
func TestWorkflowPlanOrderActiveAmendment(t *testing.T) {
	store, hierarchy, _ := workflowOrderFixture(t)
	request := workflowOrderAmendment(t, store, hierarchy.Set)
	commit, _ := workflowSelectionCommit(t, store)
	run := workflowTestStart(t, store, StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "active-order",
		Selection:  workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, "ticket"), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}})
	workflowTestClaim(t, store, run.Invocation.ID)

	observed, err := store.Read(t.Context(), hierarchy.Set.Path)
	if err != nil {
		t.Fatal(err)
	}

	request.ExpectedRevision, request.ExpectedSourceHash = observed.Revision, observed.SourceHash
	calls := 0

	request.ReserveTicketKey = func(context.Context) (string, error) { calls++; return "PRJ-7", nil }
	if _, err := store.AmendWorkflowSet(t.Context(), request); err == nil {
		t.Fatal("active set was amended")
	}

	if calls != 0 {
		t.Fatal("active amendment allocated tickets")
	}

	if _, err := store.requiredWorkflowOrderMap(t.Context(), hierarchy.Tasks[0]); err != nil {
		t.Fatal("draft or rejected amendment broke active map", err)
	}
}

// Direct record writes cannot replace a revision-bound set's plan or task order.
func TestWorkflowPlanOrderWriteGuard(t *testing.T) {
	store, hierarchy, _ := workflowOrderFixture(t)
	task := hierarchy.Tasks[0]

	task.Task.Order = 99
	if _, err := store.Write(t.Context(), task, task.Revision); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("task order bypassed amendment", err)
	}
}
