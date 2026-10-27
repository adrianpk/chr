package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestNativeTargetAliases binds actual admission, selection, worker capacity and
// preparation limits across URL spellings, namespaces and linked checkouts.
func TestNativeTargetAliases(t *testing.T) {
	store, record, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("true"))
	request.Selection.Target.Repository = "https://FORGE.EXAMPLE:443/test/repo.git/"
	run := workflowTestStart(t, store, request).Invocation
	first := nativeEntryAdmit(t, store, WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: WorkflowRef{Kind: "ticket", ID: record.Task.Key}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"})

	peer, second := nativeRolePeer(t, store, "peer", "PRJ-91", "refs/heads/dev")
	if first.Repository != second.Repository {
		t.Fatal("URL spelling created a separate canonical repository")
	}

	linked := nativeCoordinatorLinked(t, store)
	q := WorkflowIntegrationScheduleQuery{Repository: request.Selection.Target.Repository, TargetRef: first.TargetRef, Limit: 10}

	page, err := linked.QueryWorkflowIntegrationSchedule(t.Context(), q)
	if err != nil || len(page.Items) != 2 || page.Items[0].Entry.ID != first.ID {
		t.Fatal("linked alias did not share the ordered frontier", page, err)
	}

	claim := nativeScheduleClaim(t, linked, q)
	if _, err := peer.ClaimWorkflow(t.Context(), second.InvocationID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("namespace bypassed the linked checkout reservation", err)
	}

	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal("selected owner could not execute its real worker", err)
	}

	for _, owner := range []*WorkflowStore{store, linked} {
		writeWorkflowTestConfigValue(t, owner.layout.ProjectDir(), "work", "integration.max_prepared_per_target", "1")
	}

	nativeRoleTransition(t, store, first, "preparing")

	before := nativeQueueImages(t, peer.layout.Root())
	if _, err := peer.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, peer, second, "preparing")); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("URL alias bypassed the shared prepared bound", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, peer.layout.Root())) {
		t.Fatal("rejected alias preparation changed accepted work")
	}
}

func nativeCoordinatorLinked(t *testing.T, store *WorkflowStore) *WorkflowStore {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "linked")
	workflowSelectionGit(t, store, "worktree", "add", "-qb", "native-linked", dir, "HEAD")

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: dir, Namespace: "alias"})
	if err != nil {
		t.Fatal(err)
	}

	return NewWorkflowStore(layout, store.projectID)
}

// TestNativeTargetBarrier uses a real guarded native intent. A URL alias and
// linked checkout cannot escape its unresolved target; another target remains
// eligible and observations leave every authoritative image unchanged.
func TestNativeTargetBarrier(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	submission := beginNativeSubmission(t, f)
	_, peer := nativeRolePeer(t, f.store, "peer", "PRJ-91", "refs/heads/dev")
	_, other := nativeRolePeer(t, f.store, "other", "PRJ-92", "refs/heads/other")
	linked := nativeCoordinatorLinked(t, f.store)
	pr := submission.Guard.PullRequest
	pr.Repository = "https://FORGE.EXAMPLE:443/test/repo.git/"

	key, err := workflowSubmissionTargetKey(pr)
	if err != nil || key != contenthash.Sum([]byte(peer.Repository+"\n"+peer.TargetRef)) {
		t.Fatal("publication and queue aliases have different target keys", err)
	}

	before := nativeQueueImages(t, f.store.layout.Root())
	if err := linked.guardWorkflowSubmissionTarget(t.Context(), contenthash.Sum([]byte("another intent")), key); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("alias bypassed the actual publication barrier", err)
	}

	nativeCoordinatorBarrierQueries(t, linked, pr.Repository, peer, other)

	if !reflect.DeepEqual(before, nativeQueueImages(t, f.store.layout.Root())) {
		t.Fatal("barrier observation changed authoritative images")
	}
}

func nativeCoordinatorBarrierQueries(t *testing.T, store *WorkflowStore, repository string, peer, other WorkflowIntegrationEntry) {
	t.Helper()

	queue, err := store.QueryWorkflowIntegrationQueue(t.Context(), WorkflowIntegrationQueueQuery{Repository: repository, TargetRef: peer.TargetRef, EligibleOnly: true, Limit: 10})
	if err != nil || len(queue.Entries) != 0 {
		t.Fatal("alias queue ignored the unresolved target", queue, err)
	}

	page, err := store.QueryWorkflowIntegrationSchedule(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10})
	if err != nil || len(page.Items) != 2 || page.Items[0].Entry.ID != peer.ID || page.Items[1].Entry.ID != other.ID {
		t.Fatal("uncertain publication blocked eligible implementation", page, err)
	}
}

// TestNativeCoordinatorReplay copies an interrupted real admission graph into
// another clone. Missing, copied and foreign-host registrations must reject
// mutation before replaying even one journal image or invoking a provider.
func TestNativeCoordinatorReplay(t *testing.T) {
	for _, registration := range []string{"missing", "copied", "foreign host", "stripped marker"} {
		t.Run(registration, func(t *testing.T) { nativeCoordinatorReplay(t, registration) })
	}
}

func nativeCoordinatorReplay(t *testing.T, registration string) {
	t.Helper()

	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)
	transition := nativeEntryTransition(t, store, entry, "preparing")
	crash := errors.New("interrupted native transition")

	store.afterReplace = func(int) error { return crash }
	if _, err := store.TransitionWorkflowIntegration(t.Context(), transition); !errors.Is(err, crash) {
		t.Fatal("fixture did not retain a real transition", err)
	}

	workflowSelectionGit(t, store, "add", "ops")
	workflowSelectionCommit(t, store)

	clone := workflowCoordinatorClone(t, store)
	nativeCoordinatorRegistration(t, store, clone, registration)

	before := nativeQueueImages(t, clone.layout.Root())
	if _, err := clone.TransitionWorkflowIntegration(t.Context(), transition); !errors.Is(err, ErrWorkflowCoordinator) {
		t.Fatal("unregistered clone replayed a native transition", err)
	}

	if _, _, err := clone.BeginWorkflowNativeSubmission(t.Context(), WorkflowIntegrationBudgetGuard{InvocationID: admission.InvocationID}, nil); !errors.Is(err, ErrWorkflowCoordinator) {
		t.Fatal("unregistered clone entered publication", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, clone.layout.Root())) {
		t.Fatal("rejected coordinator changed durable workflow images")
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)

	retained := nativeEntryAdmit(t, restarted, admission)
	if retained.State != "preparing" || retained.QueueSequence != entry.QueueSequence || retained.Revision != entry.Revision+1 {
		t.Fatal("registered owner could not replay its own transition")
	}
}

func nativeCoordinatorRegistration(t *testing.T, source, clone *WorkflowStore, mode string) {
	t.Helper()

	if mode == "stripped marker" {
		if err := os.Remove(filepath.Join(clone.layout.Root(), workflowCoordinatorPath)); err != nil {
			t.Fatal(err)
		}

		return
	}

	if mode == "missing" {
		return
	}

	common, err := workflowRepositoryDirectory(t.Context(), source.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	data := nativeCoordinatorImage(t, common, mode)

	common, err = workflowRepositoryDirectory(t.Context(), clone.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(common, "chr-workflow-coordinator.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func nativeCoordinatorImage(t *testing.T, common, mode string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(common, "chr-workflow-coordinator.json"))
	if err != nil {
		t.Fatal(err)
	}

	if mode == "foreign host" {
		var registration workflowCoordinatorRegistration
		if err := json.Unmarshal(data, &registration); err != nil {
			t.Fatal(err)
		}

		registration.Owner = contenthash.Sum([]byte("another machine and common directory"))

		data, err = json.Marshal(registration)
		if err != nil {
			t.Fatal(err)
		}
	}

	return data
}
