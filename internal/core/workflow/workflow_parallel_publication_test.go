package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowParallelPublication excludes a linked process in another
// namespace while the first provider is live and after an uncertain response.
// A successful first lane invalidates the second lane's combined gate/review.
func TestWorkflowParallelPublication(t *testing.T) {
	for _, mode := range []string{"accepted", "lost", "unknown"} {
		t.Run(mode, func(t *testing.T) { testWorkflowParallelPublication(t, mode) })
	}
}

func testWorkflowParallelPublication(t *testing.T, mode string) {
	t.Helper()

	lanes := workflowParallelPublicationFixture(t, false)
	alias := workflowSubmissionAlias(t, lanes[0].fixture)
	release := filepath.Join(t.TempDir(), "release")
	first, firstSpec := workflowParallelPublisher(t, lanes[0], lanes[0].fixture.store, mode, t.TempDir(), release)
	waitWorkflowClaimFile(t, firstSpec.Entered)
	assertWorkflowParallelPublisherExcluded(t, lanes[1], alias)

	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	result := workflowParallelPublisherResultFor(t, first, firstSpec)
	if result.State != "awaiting_observation" || result.Unknown != (mode != "accepted") || (result.Error == "") != (mode == "accepted") {
		t.Fatalf("first publication outcome: %+v", result)
	}

	if mode == "unknown" {
		checkWorkflowParallelUnknown(t, lanes, alias, result.ID)

		return
	}

	finalizeWorkflowParallelPublication(t, lanes[0])
	checkWorkflowParallelRevalidation(t, lanes[1], alias)
}

func assertWorkflowParallelPublisherExcluded(t *testing.T, lane workflowParallelPublicationLane, alias *Store) {
	t.Helper()

	command, spec := workflowParallelPublisher(t, lane, alias, "accepted", t.TempDir(), "")
	result := workflowParallelPublisherResultFor(t, command, spec)

	if !result.Owned || result.ID != "" {
		t.Fatalf("competing lane bypassed target exclusion: %+v", result)
	}

	if _, err := os.Stat(spec.Entered); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("competing provider performed IO while another target owner remained", err)
	}
}

func checkWorkflowParallelUnknown(t *testing.T, lanes [2]workflowParallelPublicationLane, alias *Store, id string) {
	t.Helper()

	first := lanes[0]
	if _, err := first.fixture.store.StopWorkflow(t.Context(), first.fixture.run.ID, "cancel"); err != nil {
		t.Fatal(err)
	}

	assertWorkflowParallelPublisherExcluded(t, lanes[1], alias)
	first.provider.seal(id)

	sealed, err := first.fixture.store.ReconcileWorkflowSubmission(t.Context(), first.fixture.run.ID, first.provider, workflowSubmissionObserver(first.fixture))
	if err != nil || sealed.State != "unpublished" {
		t.Fatal("definitive unpublished observation did not release target", err)
	}

	submitWorkflowParallelPublication(t, lanes[1])
	finalizeWorkflowParallelPublication(t, lanes[1])
}

func assertWorkflowParallelPublisherStale(t *testing.T, lane workflowParallelPublicationLane, alias *Store) {
	t.Helper()

	command, spec := workflowParallelPublisher(t, lane, alias, "accepted", t.TempDir(), "")

	result := workflowParallelPublisherResultFor(t, command, spec)
	if result.Error == "" || result.Owned || result.ID != "" {
		t.Fatalf("target movement reused stale authorization: %+v", result)
	}

	if _, err := os.Stat(spec.Entered); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("stale combined result reached provider IO", err)
	}
}

func checkWorkflowParallelRevalidation(t *testing.T, lane workflowParallelPublicationLane, alias *Store) {
	t.Helper()
	assertWorkflowParallelPublisherStale(t, lane, alias)

	previous := workflowTestRead(t, lane.fixture.store, lane.fixture.run.ID)

	prepared, err := lane.fixture.store.PrepareWorkflowSliceIntegration(t.Context(), previous.ID, lane.provider, lane.fixture.git)
	if err != nil || prepared.Validation.Validation[0].Stdout != "combined" || prepared.Input.Git.BaseCommit == lane.fixture.base {
		t.Fatal("fresh preparation did not validate both independent results", err)
	}

	if _, err := lane.fixture.store.CheckWorkflowReviewGrant(t.Context(), previous.ID, lane.provider); err == nil {
		t.Fatal("old review authorized a new combined result")
	}

	fresh, err := lane.fixture.store.ReviewWorkflowSlice(t.Context(), previous.ID, lane.provider, nil)
	if err != nil || fresh.InputHash == previous.Review.InputHash {
		t.Fatal("target movement did not replace review binding", err)
	}

	submitWorkflowParallelPublication(t, lane)
	finalizeWorkflowParallelPublication(t, lane)
}

func submitWorkflowParallelPublication(t *testing.T, lane workflowParallelPublicationLane) {
	t.Helper()

	submission, err := lane.fixture.store.SubmitWorkflowSliceIntegration(t.Context(), lane.fixture.run.ID, lane.provider, workflowSubmissionObserver(lane.fixture))
	if err != nil || submission.Response != "accepted" {
		t.Fatal("eligible independent lane failed guarded publication", err)
	}
}

func finalizeWorkflowParallelPublication(t *testing.T, lane workflowParallelPublicationLane) IntegrationProgress {
	t.Helper()

	progress, err := lane.fixture.store.AdvanceWorkflowSliceIntegration(t.Context(), lane.fixture.run.ID, workflowPublicationPorts(lane.fixture, lane.provider))
	if err != nil || progress.Receipt == nil || progress.Invocation.Status != "completed" {
		t.Fatal("canonical publication was not finalized", err)
	}

	return progress
}

// TestWorkflowParallelPublicationConflict blocks only the conflicting lane
// after the first result is integrated, preserving its receipt and target tree.
func TestWorkflowParallelPublicationConflict(t *testing.T) {
	lanes := workflowParallelPublicationFixture(t, true)
	submitWorkflowParallelPublication(t, lanes[0])

	first := finalizeWorkflowParallelPublication(t, lanes[0])
	target := workflowIntegrationCommand(t, lanes[0].fixture.remote, "rev-parse", "refs/heads/dev")

	_, err := lanes[1].fixture.store.PrepareWorkflowSliceIntegration(t.Context(), lanes[1].fixture.run.ID, lanes[1].provider, lanes[1].fixture.git)
	if err == nil || workflowTestRead(t, lanes[1].fixture.store, lanes[1].fixture.run.ID).Status != "blocked" {
		t.Fatal("conflicting independent lane did not block", err)
	}

	if current := workflowTestRead(t, lanes[0].fixture.store, lanes[0].fixture.run.ID); current.Revision != first.Invocation.Revision || current.DeliveryRef != first.Receipt.Path || workflowIntegrationCommand(t, lanes[0].fixture.remote, "rev-parse", "refs/heads/dev") != target {
		t.Fatal("conflicting lane altered the integrated lane or target")
	}
}

// TestWorkflowParallelPublicationHuman releases worker capacity during human
// approval while retaining both candidates, without acquiring a target journal.
func TestWorkflowParallelPublicationHuman(t *testing.T) {
	lanes := workflowParallelPublicationFixture(t, false)
	first := lanes[0]
	workflowTicketReviewConfig(t, first.fixture.store, "rebase", workflowstate.ReviewHuman)

	if _, err := first.fixture.store.PrepareWorkflowSliceIntegration(t.Context(), first.fixture.run.ID, first.provider, first.fixture.git); err != nil {
		t.Fatal(err)
	}

	waiting, err := first.fixture.store.ReviewWorkflowSlice(t.Context(), first.fixture.run.ID, first.provider, nil)
	if err != nil || waiting.State != "waiting_approval" {
		t.Fatal("human approval was not retained", err)
	}

	lock, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: first.fixture.store.layout.ProjectDir(), Operation: AgentOperationDevelopment})
	if err != nil {
		t.Fatal("human wait retained worker capacity", err)
	}

	workflowTestReleaseLegacy(t, lock)

	current := workflowTestRead(t, first.fixture.store, first.fixture.run.ID)
	if current.Status != "waiting_review" || current.CandidateRef != first.fixture.run.CandidateRef || current.Submission != nil || workflowTestRead(t, lanes[1].fixture.store, lanes[1].fixture.run.ID).Submission != nil {
		t.Fatal("human wait changed a candidate or acquired publication exclusion")
	}
}
