package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowContinuationFixture(t *testing.T) (*Store, StartRequest, workflowstate.Invocation) {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)

	waiver := workflowstate.Validation{WaiverReason: "Continuation fixture observes retained host proof"}
	for i := range opts.Plan.Slices {
		opts.Plan.Slices[i].Validation = waiver
		for j := range opts.Plan.Slices[i].Tasks {
			opts.Plan.Slices[i].Tasks[j].Validation = waiver
		}
	}

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	commit, _ := workflowSelectionCommit(t, store)
	request := StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "bounded-call",
		Policy:     workflowstate.ExecutionPolicy{CheckpointStepLimit: 2},
		Selection:  workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, "slice_set"), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}}

	return store, request, workflowTestStart(t, store, request).Invocation
}

func workflowContinueOne(t *testing.T, store *Store, id string, ports ContinuationPorts) ContinuationProgress {
	t.Helper()

	progress, err := ContinueWorkflow(t.Context(), store, id, ContinuationRequest{Worker: workflowPRWorker(), PullRequest: workflowPRContent(), StepLimit: 1}, ports)
	if err != nil {
		t.Fatal(err)
	}

	return progress
}

// TestWorkflowContinuationTasks restarts after every accepted task and aggregate
// transition, preserving original request identity, attempts and cumulative work.
// Each bounded call executes only one task; its immutable checkpoint chains back.
func TestWorkflowContinuationTasks(t *testing.T) {
	store, request, run := workflowContinuationFixture(t)
	for step := 1; step <= 4; step++ {
		restarted := NewWorkflowStore(store.layout, store.projectID)
		progress := workflowContinueOne(t, restarted, run.ID, ContinuationPorts{})
		checkWorkflowContinuationStep(t, progress, step)
		checkWorkflowContinuationIdentity(t, progress.Invocation.Continuation, run, step)

		request.RequestKey = "equivalent-restart"

		attached := workflowTestStart(t, restarted, request)
		if !attached.Attached || attached.Invocation.ID != run.ID || attached.Invocation.Policy.CheckpointStepLimit != 2 {
			t.Fatal("equivalent request reset scope or frozen checkpoint policy")
		}
	}

	retained := workflowTestRead(t, store, run.ID)

	data, err := os.ReadFile(filepath.Join(retained.Lane.Worktree, "result.txt"))
	if err != nil || string(data) != "Task result\nTask result\nTask result\n" || retained.CandidateRef == "" || retained.ConfirmedCursor != 0 {
		t.Fatal("restart repeated a task or treated validation as integration")
	}
}

func checkWorkflowContinuationStep(t *testing.T, progress ContinuationProgress, step int) {
	t.Helper()

	if progress.Steps != 1 || progress.Reason != "step_limit" || len(progress.Invocation.Attempts) != step || progress.Invocation.Cursor != min(step, 3) {
		t.Fatal("bounded restart duplicated work or exceeded its call limit")
	}
}

func checkWorkflowContinuationIdentity(t *testing.T, checkpoint *workflowstate.ContinuationCheckpoint, run workflowstate.Invocation, step int) {
	t.Helper()

	if checkpoint == nil || checkpoint.Sequence != step || checkpoint.RequestHash != run.RequestHash || checkpoint.MemberCount != 6 || checkpoint.AttemptCount != step {
		t.Fatal("checkpoint lost frozen identity, extent or consumed attempts")
	}
}

// TestWorkflowContinuationBounds rejects per-call expansion and changed frozen
// selectors before task admission, preserving the equivalent invocation.
func TestWorkflowContinuationBounds(t *testing.T) {
	store, request, run := workflowContinuationFixture(t)
	for _, limit := range []int{-1, 3, 129} {
		if _, err := ContinueWorkflow(t.Context(), store, run.ID, ContinuationRequest{StepLimit: limit}, ContinuationPorts{}); err == nil {
			t.Fatal("invalid or broadened call budget accepted")
		}
	}

	request.Selection.Selector = workflowstate.Selector{Extent: "use_case", Identity: "broader"}
	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("retry key broadened the original extent", err)
	}

	retained := workflowTestRead(t, store, run.ID)
	if retained.Revision != run.Revision || len(retained.Attempts) != 0 || retained.Continuation != nil {
		t.Fatal("invalid continuation changed accepted state")
	}
}

// TestWorkflowContinuationDelivery restarts at a bound canonical PR, observes
// actual integration once, then releases the finalized frontier on the next call
// without repeating a worker, PR creation or receipt fan-out.
func TestWorkflowContinuationDelivery(t *testing.T) {
	fixture := workflowDeliveryExtentFixture(t, "rebase", "slice_set")
	workflowReviewConfig(t, fixture.store, workflowstate.ReviewHuman)

	ports := ContinuationPorts{Publication: PublicationPorts{Provider: workflowFrontierProvider{fixture.provider, fixture.remote}, Observer: fixture.git}, Target: fixture.git}
	restarted := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)
	first := workflowContinueOne(t, restarted, fixture.run.ID, ports)
	checkWorkflowContinuationReceipt(t, first, fixture)

	second := workflowContinueOne(t, restarted, fixture.run.ID, ports)
	if second.Invocation.Lane != nil || len(second.Invocation.Deliveries) != 1 || second.Invocation.Deliveries[0].DeliveryRef != first.Invocation.DeliveryRef || second.Invocation.Cursor != 3 || fixture.provider.creates != 1 {
		t.Fatal("integration restart replaced its receipt or repeated provider creation")
	}

	third := workflowContinueOne(t, restarted, fixture.run.ID, ports)
	if third.Invocation.Cursor != 4 || len(third.Invocation.Attempts) != 5 || third.Invocation.Lane.BaseCommit != fixture.target {
		t.Fatal("next call repeated old work or used the old canonical base")
	}
}

func checkWorkflowContinuationReceipt(t *testing.T, progress ContinuationProgress, fixture workflowIntegrationFixture) {
	t.Helper()

	if progress.Invocation.DeliveryRef == "" || progress.Invocation.ConfirmedCursor != 3 || len(progress.Invocation.Attempts) != 4 || fixture.provider.creates != 1 {
		t.Fatal("PR restart repeated validated work or skipped receipt finalization")
	}
}
