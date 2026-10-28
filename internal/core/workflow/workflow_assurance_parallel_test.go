package workflow

import (
	"errors"
	"testing"
)

// TestWorkflowAssuranceParallel integrates one completed lane while an actual
// independent worker remains alive in another namespace and linked checkout.
// Publication cannot finish that worker or consume a third lane's reservation.
func TestWorkflowAssuranceParallel(t *testing.T) {
	stores, runs := workflowParallelRunnerLanes(t, 2)
	first := workflowParallelRunnerFixture(t, stores[0], runs[0].ID, "worker")
	waitWorkflowClaimFile(t, first.spec.Started)

	second := workflowParallelRunnerFixture(t, stores[1], runs[1].ID, "worker")
	waitWorkflowClaimFile(t, second.spec.Started)

	if _, err := stores[2].ClaimWorkflow(t.Context(), runs[2].ID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("composed third lane bypassed two actual owners", err)
	}

	if result := finishWorkflowParallelRunner(t, first); !result.Successful {
		t.Fatal("first composed worker failed", result)
	}

	ports, provider := workflowAssurancePorts(t, stores[0], runs[0].Selection.Target.Commit)
	// Preserve capacity two while changing only the fixture's publication policy.
	writeWorkspaceConfig(t, stores[0].layout.ProjectDir(), "review:\n  mode: trusted_auto_merge\nwork:\n  merge_strategy: rebase\n  max_active_workers: 2\n  integration:\n    publication_profile: git_atomic_dual_ref\n")

	completed := workflowAssuranceRun(t, stores[0], runs[0].ID, ports)
	if completed.DeliveryRef == "" || provider.creates != 1 || provider.calls != 1 {
		t.Fatal("first composed lane lacks its one canonical delivery")
	}

	other := workflowTestRead(t, stores[1], runs[1].ID)
	if other.Cursor != 0 || other.Attempts[0].Outcome != "running" || other.DeliveryRef != "" {
		t.Fatal("unrelated publication changed the still-running worker")
	}

	if result := finishWorkflowParallelRunner(t, second); !result.Successful {
		t.Fatal("canonical target movement stopped the independent source", result)
	}

	checkWorkflowAssuranceCapacity(t, stores[2], runs[2].ID)
}

func checkWorkflowAssuranceCapacity(t *testing.T, store *Store, id string) {
	t.Helper()

	claim := workflowTestClaim(t, store, id)
	if err := store.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal("capacity wait did not resume after owned workers finished", err)
	}
}
