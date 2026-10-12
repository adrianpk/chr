package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestNativeProcessTargets shares target journals between admitted work and
// existing prepared publication owners. Two providers reach external IO before
// either is released; an unknown outcome excludes only its native target queue.
func TestNativeProcessTargets(t *testing.T) {
	lanes := workflowParallelPublicationTargets(t, false, "other")
	store := lanes[0].fixture.store
	_, pending := nativeRolePeer(t, store, "pending", "PRJ-92", "refs/heads/dev")

	eligible := nativeQueuePage(t, store, WorkflowIntegrationQueueQuery{TargetRef: pending.TargetRef, EligibleOnly: true, Limit: 10})
	if len(eligible.Entries) != 1 || eligible.Entries[0].Entry.ID != pending.ID {
		t.Fatal("native unit was not eligible before publication")
	}

	release := filepath.Join(t.TempDir(), "release")
	first, firstSpec := workflowParallelPublisher(t, lanes[0], lanes[0].fixture.store, "unknown", t.TempDir(), release)
	waitWorkflowClaimFile(t, firstSpec.Entered)

	second, secondSpec := workflowParallelPublisher(t, lanes[1], lanes[1].fixture.store, "accepted", t.TempDir(), release)
	waitWorkflowClaimFile(t, secondSpec.Entered)

	// Both provider checkpoints precede release of either publication process.
	// Each callback has already verified its own retained common-directory intent.
	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	one := workflowParallelPublisherResultFor(t, first, firstSpec)

	two := workflowParallelPublisherResultFor(t, second, secondSpec)
	if !one.Unknown || one.State != "awaiting_observation" || two.Error != "" || two.State != "awaiting_observation" {
		t.Fatal("independent publication lost its guarded outcome", one, two)
	}

	before := nativeQueueImages(t, store.layout.Root())

	unknown := nativeQueuePage(t, store, WorkflowIntegrationQueueQuery{TargetRef: "refs/heads/dev", EligibleOnly: true, Limit: 10})
	if len(unknown.Entries) != 0 {
		t.Fatal("unknown target remained eligible for another publication")
	}

	checkNativeIndependentPublication(t, lanes, before)
}

func checkNativeIndependentPublication(t *testing.T, lanes [2]workflowParallelPublicationLane, before map[string]string) {
	t.Helper()

	finalizeWorkflowParallelPublication(t, lanes[1])

	if workflowIntegrationCommand(t, lanes[1].fixture.remote, "rev-parse", "refs/heads/other") == lanes[1].fixture.base {
		t.Fatal("unrelated publication did not change its actual target")
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, lanes[0].fixture.store.layout.Root())) || workflowIntegrationCommand(t, lanes[0].fixture.remote, "rev-parse", "refs/heads/dev") != lanes[0].fixture.base {
		t.Fatal("unrelated delivery changed uncertain records or its target")
	}
}

// TestNativeProcessUnknown retains a genuine native dispatch with unknown
// application. A linked alias can read status and execute unrelated admitted
// work while the uncertain target journal and owning namespace stay unchanged.
func TestNativeProcessUnknown(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	submission := beginNativeSubmission(t, f)
	update := nativeSubmissionUpdate(t, f, submission)

	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	update.Facts.DispatchState, update.Facts.ApplicationState, update.Facts.DispatchIdentity = "dispatching", "unknown", &identity
	if _, err := f.store.RetainWorkflowNativeSubmission(t.Context(), update); err != nil {
		t.Fatal(err)
	}

	_, peer := nativeRolePeer(t, f.store, "peer", "PRJ-91", "refs/heads/dev")
	otherStore, other := nativeRolePeer(t, f.store, "independent", "PRJ-92", "refs/heads/other")
	linked := nativeCoordinatorLinked(t, f.store)

	common, err := workflowRepositoryDirectory(t.Context(), f.store.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	before := nativeQueueImages(t, f.store.layout.Root())
	journals := nativeQueueImages(t, filepath.Join(common, workflowPublicationDirectory))
	checkNativeUnknownQueries(t, linked, peer, other)

	runner := workflowParallelRunnerFixture(t, linked, other.InvocationID, "native")
	waitWorkflowClaimFile(t, runner.spec.Started)
	checkNativeUnknownQueries(t, linked, peer, other)
	nativeProcessFinish(t, runner)
	checkNativeProcessProof(t, otherStore, other.InvocationID)

	if !reflect.DeepEqual(before, nativeQueueImages(t, f.store.layout.Root())) || !reflect.DeepEqual(journals, nativeQueueImages(t, filepath.Join(common, workflowPublicationDirectory))) {
		t.Fatal("unrelated worker or status replaced uncertain publication evidence")
	}
}

func checkNativeUnknownQueries(t *testing.T, store *WorkflowStore, peer, other WorkflowIntegrationEntry) {
	t.Helper()

	queue := nativeQueuePage(t, store, WorkflowIntegrationQueueQuery{Repository: "https://FORGE.EXAMPLE:443/test/repo.git/", TargetRef: peer.TargetRef, EligibleOnly: true, Limit: 10})
	if len(queue.Entries) != 0 {
		t.Fatal("unknown native target escaped its alias barrier")
	}

	key := workflowHash([]byte(peer.Repository + "\n" + peer.TargetRef))
	if err := store.guardWorkflowSubmissionTarget(t.Context(), workflowHash([]byte("other intent")), key); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("uncertain native target admitted another publication", err)
	}

	otherKey := workflowHash([]byte(other.Repository + "\n" + other.TargetRef))
	if err := store.guardWorkflowSubmissionTarget(t.Context(), workflowHash([]byte("other intent")), otherKey); err != nil {
		t.Fatal("uncertain native publication blocked another target", err)
	}

	if _, err := store.QueryWorkflowIntegrationRoles(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10}); err != nil {
		t.Fatal("uncertain native dispatch blocked read-only role status", err)
	}
}
