package workflow

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowFrontierProvider struct {
	*workflowPRProviderFixture
	remote string
}

func (p workflowFrontierProvider) ReadWorkflowBranch(ctx context.Context, repository, branch string) (string, error) {
	if repository != p.pr.Repository {
		return "", workflowstate.ErrConflict
	}

	output, err := outputWorktreeGit(ctx, p.remote, "rev-parse", "--verify", "refs/heads/"+branch+"^{commit}")

	return strings.TrimSpace(output), err
}

func workflowFrontierPorts(fixture workflowIntegrationFixture) FrontierPorts {
	return FrontierPorts{Provider: workflowFrontierProvider{fixture.provider, fixture.remote}, Observer: fixture.git, Target: fixture.git}
}

func workflowFinalizeFrontier(t *testing.T, fixture workflowIntegrationFixture) Record {
	t.Helper()

	if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, "rebase", fixture.provider, fixture.git); err != nil {
		t.Fatal(err)
	}

	receipt, err := fixture.store.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, fixture.provider, fixture.git)
	if err != nil {
		t.Fatal(err)
	}

	return receipt
}

// TestWorkflowFrontierOrder runs two distinct canonical PR identities against a
// real remote. Neither task validation nor PR merge can release the receipt
// barrier, and the next lane starts from freshly fetched canonical objects.
func TestWorkflowFrontierOrder(t *testing.T) {
	for _, extent := range []string{"slice_set", "use_case"} {
		t.Run(extent, func(t *testing.T) {
			fixture := workflowDeliveryExtentFixture(t, "rebase", extent)
			checkWorkflowFrontierExclusions(t, fixture, extent)
			checkWorkflowFrontierBarrier(t, fixture)

			receipt := workflowFinalizeFrontier(t, fixture)
			first := workflowAdvanceFrontier(t, fixture)
			checkWorkflowFrontierNext(t, fixture, first, receipt)
			workflowDeliverNextFrontier(t, &fixture)

			second := workflowAdvanceFrontier(t, fixture)
			if second.Status != "waiting_integration" || second.Outcome != "set_closure_required" || second.ConfirmedCursor != 6 || len(second.Deliveries) != 2 || fixture.provider.creates != 2 {
				t.Fatal("last slice bypassed set closure or duplicated a PR")
			}

			if second.Deliveries[0].PullRequest.URL == second.Deliveries[1].PullRequest.URL || second.Deliveries[0].EvidenceHash != first.Deliveries[0].EvidenceHash {
				t.Fatal("continuation replaced the previous PR or delivery image")
			}

			checkWorkflowFrontierBarrier(t, fixture)
		})
	}
}

func checkWorkflowFrontierExclusions(t *testing.T, fixture workflowIntegrationFixture, extent string) {
	t.Helper()

	if extent != "use_case" {
		return
	}

	preview := workflowTestPreview(t, fixture.store, extent, "ordered-case")
	if preview.Total != 6 || preview.Excluded["unsliced"] != 1 || len(fixture.run.Members) != 6 {
		t.Fatal("use-case continuation admitted unsliced work")
	}
}

func workflowAdvanceFrontier(t *testing.T, fixture workflowIntegrationFixture) Invocation {
	t.Helper()

	store := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

	run, err := store.AdvanceWorkflowFrontier(t.Context(), fixture.run.ID, workflowFrontierPorts(fixture))
	if err != nil {
		t.Fatal(err)
	}

	return workflowTestRead(t, store, run.ID)
}

func checkWorkflowFrontierBarrier(t *testing.T, fixture workflowIntegrationFixture) {
	t.Helper()

	before := workflowTestRead(t, fixture.store, fixture.run.ID)
	if before.DeliveryRef == "" && before.Lane != nil {
		if _, err := fixture.store.AdvanceWorkflowFrontier(t.Context(), before.ID, workflowFrontierPorts(fixture)); !errors.Is(err, ErrWorkflowObservationRequired) {
			t.Fatalf("unfinalized frontier advanced: %v", err)
		}
	}

	if _, err := fixture.store.ClaimWorkflow(t.Context(), before.ID); err == nil {
		t.Fatal("receipt or set closure barrier admitted the next task")
	}

	after := workflowTestRead(t, fixture.store, before.ID)
	if after.Revision != before.Revision || len(after.Attempts) != len(before.Attempts) {
		t.Fatal("barrier check mutated retained work")
	}
}

func checkWorkflowFrontierNext(t *testing.T, fixture workflowIntegrationFixture, run Invocation, receipt Record) {
	t.Helper()
	checkWorkflowFrontierCleared(t, run)

	if !reflect.DeepEqual(run.Selection, fixture.run.Selection) || run.RequestHash != fixture.run.RequestHash || run.ExecutionSelection().Target.Commit != fixture.target || run.Deliveries[0].DeliveryRef != receipt.Path {
		t.Fatal("continuation changed admission identity or lost verified delivery")
	}

	if host := workflowIntegrationCommand(t, fixture.store.layout.ProjectDir(), "rev-parse", "HEAD"); host != run.Selection.Target.Commit {
		t.Fatal("hydration moved the host checkout")
	}

	again := workflowAdvanceFrontier(t, fixture)
	if again.Revision != run.Revision || len(again.Deliveries) != 1 {
		t.Fatal("frontier retry repeated a confirmed checkpoint")
	}
}

func checkWorkflowFrontierCleared(t *testing.T, run Invocation) {
	t.Helper()

	if run.Status != "pending" || run.Lane != nil || run.PullRequest != nil || run.DeliveryRef != "" || run.CandidateRef != "" || run.ConfirmedCursor != 3 || len(run.Deliveries) != 1 {
		t.Fatal("restart resurrected bindings or discarded the confirmed frontier")
	}
}

func workflowDeliverNextFrontier(t *testing.T, fixture *workflowIntegrationFixture) {
	t.Helper()

	candidate, err := fixture.store.ExecuteWorkflowSlice(t.Context(), fixture.run.ID, workflowPRWorker())
	if err != nil {
		t.Fatal(err)
	}

	if candidate.Lane.BaseCommit != fixture.target || candidate.Lane.Unit.ID != fixture.hierarchy.Set.Set.Slices[1].ID || len(candidate.Slice.Tasks) != 3 {
		t.Fatal("next slice used the old target or replayed integrated tasks")
	}

	workflowIntegrationCommand(t, candidate.Lane.Worktree, "push", fixture.remote, candidate.Lane.HeadCommit+":"+candidate.Lane.Branch)

	fixture.provider.pr = PullRequestSnapshot{URL: fixture.provider.pr.Repository + "/pulls/8", Repository: fixture.provider.pr.Repository, HeadRepository: fixture.provider.pr.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", BaseCommit: fixture.target, HeadCommit: candidate.Result.Commit, State: "open"}
	fixture.provider.found = nil
	workflowPREnsure(t, fixture.store, fixture.run.ID, workflowFrontierProvider{fixture.provider, fixture.remote})
	checkWorkflowFrontierBarrier(t, *fixture)
	workflowIntegrationCommand(t, fixture.actor, "fetch", "origin", candidate.Lane.Branch)
	workflowIntegrationCommand(t, fixture.actor, "merge", "--ff-only", "FETCH_HEAD")
	workflowIntegrationCommand(t, fixture.actor, "push", "origin", "dev")

	fixture.result = workflowIntegrationCommand(t, fixture.actor, "rev-parse", "HEAD")
	fixture.target = fixture.result
	fixture.provider.pr.State, fixture.provider.pr.Merged, fixture.provider.pr.ResultCommit = "closed", true, fixture.result
	fixture.provider.pr.BaseCommit = fixture.target
	workflowFinalizeFrontier(t, *fixture)
}
