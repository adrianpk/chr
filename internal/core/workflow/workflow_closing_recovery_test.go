package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowClosingEmptyOwner(t *testing.T) (workflowIntegrationFixture, Record, Invocation) {
	t.Helper()

	return workflowClosingEmptyGate(t, workflowstate.Validation{WaiverReason: "Approved recovery fixture waiver"})
}

func workflowClosingEmptyGate(t *testing.T, gate workflowstate.Validation) (workflowIntegrationFixture, Record, Invocation) {
	t.Helper()

	fixture := workflowDeliveryFixture(t, "rebase")
	plan := workflowPlanFixture()

	set, err := fixture.store.Create(t.Context(), "Empty closing owner", Record{Kind: "plan", Set: &workflowstate.WorkflowSet{Schema: WorkflowSchema, Status: "ready", PlanRef: plan.PlanRef, PlanHash: plan.PlanHash, ClosingValidation: gate}})
	if err != nil {
		t.Fatal(err)
	}

	start := StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "closing-owner",
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice_set", Identity: set.ID}, Target: workflowstate.PreviewTarget{Repository: fixture.provider.pr.Repository, Commit: fixture.target}},
	}}
	run := workflowTestStart(t, fixture.store, start)

	return fixture, set, run.Invocation
}

func workflowClosingBeginTest(t *testing.T, fixture workflowIntegrationFixture, run Invocation, ports FrontierPorts) *workflowClosingClaim {
	t.Helper()

	image := run.SetImages[len(run.Closures)]
	request := SetClosureRequest{SetPath: image.Path, ExpectedRevision: image.Revision, ExpectedSourceHash: image.SourceHash, Repository: run.Selection.Target.Repository, Branch: "dev", InvocationID: run.ID}

	var claim *workflowClosingClaim

	err := fixture.store.withLock(t.Context(), func() error {
		var err error

		claim, _, err = fixture.store.beginWorkflowClosing(t.Context(), request, ports)

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.worker.releaseReservation(); err != nil {
			t.Error(err)
		}
	})

	return claim
}

func expireWorkflowClosingTest(t *testing.T, store *Store, claim *workflowClosingClaim) {
	t.Helper()

	claim.state.LeaseUntil = time.Now().Add(-time.Second)

	if err := store.withLock(t.Context(), func() error { return store.commitWorkflowClosingState(t.Context(), claim) }); err != nil {
		t.Fatal(err)
	}

	// Simulate host loss: close both inherited capacity descriptors while leaving
	// durable operation locks for the next fenced closing generation to adopt.
	claim.worker.locks = nil
	if err := claim.worker.releaseReservation(); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowClosingRecovery exercises two expired generations, including their
// distinct operation locks. Central recovery keeps the closure budget and never
// creates task attempts or confirms a gate that was interrupted.
func TestWorkflowClosingRecovery(t *testing.T) {
	fixture, set, run := workflowClosingEmptyOwner(t)
	for generation := 1; generation <= 2; generation++ {
		claim := workflowClosingBeginTest(t, fixture, run, workflowFrontierPorts(fixture))
		if _, err := fixture.store.ReconcileWorkflow(t.Context(), run.ID); !errors.Is(err, ErrWorkflowOwned) {
			t.Fatal("live closing owner was adopted", err)
		}

		expireWorkflowClosingTest(t, fixture.store, claim)

		restarted := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

		current, err := restarted.ReconcileWorkflow(t.Context(), run.ID)
		if err != nil {
			t.Fatal("expired closing generation could not adopt its operation lock", err)
		}

		run = current
		checkWorkflowClosingRecovered(t, restarted, set.Path, run, generation)
	}
}

func checkWorkflowClosingRecovered(t *testing.T, store *Store, path string, run Invocation, generation int) {
	t.Helper()

	set, err := store.Read(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}

	if set.Set.Status != "blocked" || set.Set.ClosureRef != "" || len(set.Set.Closing.Attempts) != generation {
		t.Fatal("recovery lost the lifetime budget or confirmed interrupted closure")
	}

	if run.ClosingSet != "" || run.Outcome != "interrupted" || len(run.Attempts) != 0 {
		t.Fatal("recovery retained a live closing role or created task work")
	}
}

// TestWorkflowClosingJournal interrupts each changed image in the initial and
// terminal transitions. Restart must retain one attempt and one terminal receipt,
// with the set and owning invocation agreeing on whether closure was confirmed.
func TestWorkflowClosingJournal(t *testing.T) {
	for _, stage := range []string{"start", "finish"} {
		for _, position := range []int{0, 2, 3, 4} {
			t.Run(fmt.Sprintf("%s %d", stage, position), func(t *testing.T) { testWorkflowClosingJournal(t, stage, position) })
		}
	}
}

func testWorkflowClosingJournal(t *testing.T, stage string, position int) {
	t.Helper()

	fixture, set, run := workflowClosingEmptyOwner(t)
	ports := workflowFrontierPorts(fixture)

	fault := func(index int) error {
		if index == position {
			return errors.New("interrupted closing journal")
		}

		return nil
	}
	if stage == "start" {
		fixture.store.afterReplace = fault
	} else {
		ports.Target = &workflowClosingTargetHook{TargetGit: ports.Target, hook: func() { fixture.store.afterReplace = fault }}
	}

	if _, err := fixture.store.CloseWorkflowInvocationSet(t.Context(), run.ID, "dev", ports); err == nil {
		t.Fatal("closing transition was not interrupted")
	}

	restarted := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

	current, err := restarted.Read(t.Context(), set.Path)
	if err != nil {
		t.Fatal(err)
	}

	retained := workflowTestRead(t, restarted, run.ID)
	if len(current.Set.Closing.Attempts) != 1 || len(retained.Attempts) != 0 {
		t.Fatal("journal recovery lost or duplicated the closing attempt")
	}

	checkWorkflowClosingJournalState(t, stage, set.ID, current, retained)
}

// TestWorkflowClosingPostObservation rejects expiry or member edits made during
// the final remote observation, after the first confirmation checks already ran.
func TestWorkflowClosingPostObservation(t *testing.T) {
	for _, change := range []string{"deadline", "member"} {
		t.Run(change, func(t *testing.T) { testWorkflowClosingPostObservation(t, change) })
	}
}

func testWorkflowClosingPostObservation(t *testing.T, change string) {
	t.Helper()

	fixture, ports := workflowClosingFixture(t)
	workflowClosingContract(t, fixture, workflowstate.Validation{WaiverReason: "Approved boundary fixture waiver"})

	claim := workflowClosingBeginTest(t, fixture, workflowTestRead(t, fixture.store, fixture.run.ID), ports)
	hook := &workflowClosingTargetHook{TargetGit: ports.Target, calls: 1, hook: func() {
		if change == "deadline" {
			claim.state.LeaseUntil = time.Now().Add(-time.Second)
		} else {
			path := filepath.Join(fixture.store.layout.Root(), filepath.Dir(fixture.hierarchy.Tasks[0].Path), "artifact.md")
			if err := os.WriteFile(path, []byte("changed during final observation\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}}
	// The first observation is admission; this direct confirmation is the second.
	ports.Target = hook
	if err := fixture.store.checkWorkflowClosingConfirmation(context.Background(), claim, ports); err == nil {
		t.Fatal("post-observation change retained closing authority")
	}
}

func checkWorkflowClosingJournalState(t *testing.T, stage, id string, current Record, retained Invocation) {
	t.Helper()

	if stage == "start" {
		if current.Set.Status != "closing" || retained.ClosingSet != id || current.Set.ClosureRef != "" {
			t.Fatal("partial admission was not replayed as a fenced closing role")
		}
	} else if current.Set.Status != "completed" || retained.Status != "completed" || len(retained.Closures) != 1 {
		t.Fatal("terminal replay did not confirm set, invocation and receipt together")
	}
}
