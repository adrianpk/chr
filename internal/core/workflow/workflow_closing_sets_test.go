package workflow

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowClosingLaterSet(t *testing.T, fixture workflowIntegrationFixture) MaterializationResult {
	t.Helper()

	plan := workflowPlanFixture()
	plan.MaterializationID, plan.Title, plan.UseCase = "closing-later", "Later associated set", "ordered-case"
	plan.ClosingValidation = workflowstate.Validation{WaiverReason: "Approved later set closing waiver"}
	plan.Slices = plan.Slices[:1]
	plan.Slices[0].ID += "-later"

	plan.Slices[0].Validation = plan.ClosingValidation
	for i := range plan.Slices[0].Tasks {
		plan.Slices[0].Tasks[i].ID += "-later"
		plan.Slices[0].Tasks[i].Validation = plan.ClosingValidation
	}

	number := 100

	later, err := fixture.store.MaterializePlan(t.Context(), MaterializationOptions{Plan: plan, ReserveTicketKey: func(context.Context) (string, error) { number++; return fmt.Sprintf("PRJ-%d", number), nil }})
	if err != nil {
		t.Fatal(err)
	}

	return later
}

// TestWorkflowClosingSetAdvance delivers two associated sets with three distinct
// real PR histories. The second set cannot claim work before closure of the
// first, and its lane uses the confirmed closing target without replaying work.
func TestWorkflowClosingSetAdvance(t *testing.T) {
	fixture, ports := workflowClosingExtentFixture(t, "use_case", workflowstate.ExecutionPolicy{IncludeLaterTickets: true})
	workflowClosingContract(t, fixture, workflowstate.Validation{WaiverReason: "Approved first set closing waiver"})

	later := workflowClosingLaterSet(t, fixture)
	if _, err := fixture.store.ReconcileWorkflowMembership(t.Context(), fixture.run.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := fixture.store.ClaimWorkflow(t.Context(), fixture.run.ID); err == nil {
		t.Fatal("later associated set bypassed closing authority")
	}

	first, err := fixture.store.CloseWorkflowInvocationSet(t.Context(), fixture.run.ID, "dev", ports)
	if err != nil {
		t.Fatal(err)
	}

	before := workflowReadTestBytes(t, fixture.store, first.ReceiptRef)

	candidate, err := fixture.store.ExecuteWorkflowSlice(t.Context(), fixture.run.ID, workflowPRWorker())
	if err != nil {
		t.Fatal(err)
	}

	if candidate.Lane.BaseCommit != first.Receipt.Target.Commit || candidate.Lane.Unit.ID != later.Set.Set.Slices[0].ID || len(candidate.Slice.Tasks) != 3 {
		t.Fatal("later set did not use the confirmed closing base and next unit")
	}

	workflowClosingIntegrateLater(t, &fixture, candidate, ports)
	workflowAdvanceFrontier(t, fixture)

	closed, err := fixture.store.CloseWorkflowInvocationSet(t.Context(), fixture.run.ID, "dev", ports)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowClosingSetAdvance(t, fixture, first, closed, before)
}

func workflowClosingIntegrateLater(t *testing.T, fixture *workflowIntegrationFixture, candidate SliceCandidate, ports FrontierPorts) {
	t.Helper()

	workflowIntegrationCommand(t, candidate.Lane.Worktree, "push", fixture.remote, candidate.Lane.HeadCommit+":"+candidate.Lane.Branch)

	fixture.provider.pr = PullRequestSnapshot{URL: fixture.provider.pr.Repository + "/pulls/9", Repository: fixture.provider.pr.Repository, HeadRepository: fixture.provider.pr.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", BaseCommit: fixture.target, HeadCommit: candidate.Result.Commit, State: "open"}
	fixture.provider.found = nil
	workflowPREnsure(t, fixture.store, fixture.run.ID, workflowFrontierProvider{fixture.provider, fixture.remote})
	workflowIntegrationCommand(t, fixture.actor, "fetch", "origin", candidate.Lane.Branch)
	workflowIntegrationCommand(t, fixture.actor, "merge", "--ff-only", "FETCH_HEAD")
	workflowIntegrationCommand(t, fixture.actor, "push", "origin", "dev")

	fixture.target = workflowIntegrationCommand(t, fixture.actor, "rev-parse", "HEAD")
	fixture.provider.pr.State, fixture.provider.pr.Merged, fixture.provider.pr.ResultCommit = "closed", true, fixture.target
	fixture.provider.pr.BaseCommit = fixture.target

	provider, ok := ports.Provider.(workflowClosingProvider)
	if !ok {
		t.Fatal("closing fixture lost its canonical PR history")
	}

	provider.prs[fixture.provider.pr.URL] = fixture.provider.pr
	workflowFinalizeFrontier(t, *fixture)
}

func checkWorkflowClosingSetAdvance(t *testing.T, fixture workflowIntegrationFixture, first, closed SetClosureResult, before []byte) {
	t.Helper()

	run := workflowTestRead(t, fixture.store, fixture.run.ID)
	if run.Status != "completed" || run.Cursor != 9 || len(run.Closures) != 2 || len(run.Deliveries) != 3 {
		t.Fatal("associated set continuation lost ordered progress")
	}

	if closed.Set.ID == first.Set.ID || !bytes.Equal(before, workflowReadTestBytes(t, fixture.store, first.ReceiptRef)) {
		t.Fatal("later closure replaced the previous immutable receipt")
	}
}
