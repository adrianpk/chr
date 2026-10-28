package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeClosingPublicationPeer(t *testing.T, f nativeClosingFixture) nativeSubmissionFixture {
	t.Helper()

	linked := nativeCoordinatorLinked(t, f.fixture.f.store)
	workflowSelectionGit(t, linked, "checkout", "--detach", f.receipt.Receipt.ResultCommit)
	writeWorkspaceConfig(t, linked.layout.ProjectDir(), "review:\n  mode: trusted_auto_merge\nwork:\n  merge_strategy: rebase\n  integration:\n    publication_profile: git_atomic_dual_ref\n")

	peer, entry := nativeRolePeer(t, linked, "closing-peer", "PRJ-91", "refs/heads/dev")
	target := workflowIntegrationCommand(t, linked.layout.ProjectDir(), "rev-parse", "HEAD")
	workflowIntegrationCommand(t, linked.layout.ProjectDir(), "push", f.fixture.remote, target+":refs/heads/dev")
	nativeReceiptFrontierPorts(t, f.fixture, target)

	claim := workflowTestClaim(t, peer, entry.InvocationID)
	if _, err := peer.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	candidate, err := peer.PrepareWorkflowTicketCandidate(t.Context(), entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	repo := f.receipt.Receipt.Repository
	pr := workflowstate.PullRequestSnapshot{URL: repo + "/pulls/8", Repository: repo, HeadRepository: repo, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: target, State: "open"}
	provider := &nativePreparationProviderFixture{workflowPreparationProviderFixture: &workflowPreparationProviderFixture{workflowReviewProviderFixture: &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}, target: target}, guarded: true, revision: "guard-v1"}, policyRevision: "policy-v1"}
	workflowPREnsure(t, peer, entry.InvocationID, provider)

	entry = nativeRoleTransition(t, peer, nativeReviewEntry(t, peer, entry.InvocationID), "preparing")
	nativeRoleTransition(t, peer, entry, "validating")

	prepared, err := peer.PrepareWorkflowSliceIntegration(t.Context(), entry.InvocationID, provider, NewWorkflowPreparationGit(repo, f.fixture.remote))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := peer.ReviewWorkflowSlice(t.Context(), entry.InvocationID, provider, nil); err != nil {
		t.Fatal(err)
	}

	entry = nativeRoleTransition(t, peer, nativeReviewEntry(t, peer, entry.InvocationID), "ready_to_publish")

	return nativeSubmissionFixture{peer, entry, prepared, provider}
}

// TestNativeClosingPublication uses a real native intent from another namespace.
// Durable publication exclusion must prevent admission of a set gate even when
// its own receipts and member fanout are complete and the remote did not move.
func TestNativeClosingPublication(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "gate")
	f := newNativeClosingFixture(t, "rebase", workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"touch", marker}}})
	ports := finalizeNativeClosingFixture(t, f)
	peer := nativeClosingPublicationPeer(t, f)
	beginNativeSubmission(t, peer)

	before := nativeQueueImages(t, f.fixture.f.store.layout.Root())

	status, err := f.fixture.f.store.QueryWorkflowIntegrationSetClosure(t.Context(), f.setPath)
	if err != nil || status.CanClose || !slices.Contains(status.Blockers, "publication_requires_observation") || !reflect.DeepEqual(before, nativeQueueImages(t, f.fixture.f.store.layout.Root())) {
		t.Fatal("read-only closure status hid or changed the shared intent", status, err)
	}

	if _, err := f.fixture.f.store.CloseWorkflowInvocationSet(t.Context(), f.fixture.f.entry.InvocationID, "dev", ports); !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("another namespace's intent did not block closure", err)
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("uncertain publication admitted a gate", err)
	}

	checkNativeClosingAbsent(t, f)
}

// TestNativeClosingPublicationRace retains a new native intent during the real
// closing command. Successful command output cannot confirm closure while that
// common-directory publication remains unresolved.
func TestNativeClosingPublicationRace(t *testing.T) {
	root := t.TempDir()
	marker, release := filepath.Join(root, "started"), filepath.Join(root, "release")
	gate := workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"sh", "-c", "touch \"$1\"; while [ ! -f \"$2\" ]; do sleep 0.02; done", "closing-gate", marker, release}}}
	f := newNativeClosingFixture(t, "rebase", gate)
	ports := finalizeNativeClosingFixture(t, f)
	peer := nativeClosingPublicationPeer(t, f)
	t.Cleanup(func() {
		if err := os.WriteFile(release, nil, 0o600); err != nil {
			t.Error(err)
		}
	})

	type closureResult struct {
		result SetClosureResult
		err    error
	}

	results := make(chan closureResult, 1)

	go func() {
		result, err := f.fixture.f.store.CloseWorkflowInvocationSet(t.Context(), f.fixture.f.entry.InvocationID, "dev", ports)
		results <- closureResult{result, err}
	}()

	waitWorkflowClosingMarker(t, marker)
	beginNativeSubmission(t, peer)

	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case closed := <-results:
		if !errors.Is(closed.err, ErrWorkflowObservationRequired) || closed.result.Outcome != "closing_validation_failed" || closed.result.Set.Set.ClosureRef != "" || len(closed.result.Receipt.Validation) != 1 || closed.result.Receipt.Validation[0].ExitCode != 0 {
			t.Fatal("successful gate ignored the new publication barrier", closed.err)
		}
	case <-t.Context().Done():
		t.Fatal(t.Context().Err())
	}
}
