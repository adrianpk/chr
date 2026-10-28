package workflow

import (
	"strconv"
	"strings"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// TestWorkflowLooseDelivery serially integrates three distinct ticket PRs. Each
// lane starts at its freshly hydrated target and exhaustion preserves receipts.
func TestWorkflowLooseDelivery(t *testing.T) {
	f := workflowLooseQueueFixture(t, false)
	workflowFinalizeFrontier(t, f)
	workflowAdvanceFrontier(t, f)

	for _, key := range []string{"PRJ-91", "PRJ-92"} {
		workflowLooseDeliverNext(t, &f, key)
		workflowAdvanceFrontier(t, f)
	}

	run, err := f.store.ReconcileWorkflowMembership(t.Context(), f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowLooseCompletion(t, run, f.provider.creates)
	checkWorkflowLooseArchive(t, run)

	before := run.Revision
	if again, err := f.store.ReconcileWorkflowMembership(t.Context(), run.ID); err != nil || again.Revision != before {
		t.Fatal("exhausted queue restarted or changed its evidence", err)
	}
}

func checkWorkflowLooseCompletion(t *testing.T, run Invocation, creates int) {
	t.Helper()

	if run.Status != "completed" || run.Outcome != "loose_exhausted" || run.Cursor != 3 || run.ConfirmedCursor != 3 || len(run.Deliveries) != 3 || creates != 3 {
		t.Fatal("queue did not exhaust after three distinct verified deliveries")
	}
}

func checkWorkflowLooseArchive(t *testing.T, run Invocation) {
	t.Helper()

	urls := map[string]bool{}
	for _, delivery := range run.Deliveries {
		if delivery.Unit.Kind != "ticket" || delivery.SetID != "" || urls[delivery.PullRequest.URL] {
			t.Fatal("independent units fabricated hierarchy or shared a PR")
		}

		urls[delivery.PullRequest.URL] = true
	}
}

func workflowLooseDeliverNext(t *testing.T, fixture *workflowIntegrationFixture, key string) {
	t.Helper()

	claim := workflowTestClaim(t, fixture.store, fixture.run.ID)
	if _, err := fixture.store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	workflowLoosePublishValidated(t, fixture, key)
}

func workflowLoosePublishValidated(t *testing.T, fixture *workflowIntegrationFixture, key string) {
	t.Helper()

	candidate, err := fixture.store.PrepareWorkflowTicketCandidate(t.Context(), fixture.run.ID)
	if err != nil || candidate.Ticket.Unit.ID != key || candidate.Lane.BaseCommit != fixture.target {
		t.Fatal("next independent unit was not selected from its confirmed target", err)
	}

	workflowIntegrationCommand(t, candidate.Lane.Worktree, "push", fixture.remote, candidate.Result.Commit+":"+candidate.Lane.Branch)

	number := strconv.Itoa(6 + workflowTestRead(t, fixture.store, fixture.run.ID).Cursor)
	fixture.provider.pr = PullRequestSnapshot{URL: fixture.provider.pr.Repository + "/pulls/" + number, Repository: fixture.provider.pr.Repository, HeadRepository: fixture.provider.pr.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", BaseCommit: fixture.target, HeadCommit: candidate.Result.Commit, State: "open"}
	fixture.provider.found = nil
	workflowPREnsure(t, fixture.store, fixture.run.ID, workflowFrontierProvider{fixture.provider, fixture.remote})
	workflowIntegrationCommand(t, fixture.actor, "fetch", "origin", candidate.Lane.Branch)
	workflowIntegrationCommand(t, fixture.actor, "merge", "--ff-only", "FETCH_HEAD")
	workflowIntegrationCommand(t, fixture.actor, "push", "origin", "dev")

	fixture.result = workflowIntegrationCommand(t, fixture.actor, "rev-parse", "HEAD")
	fixture.target = fixture.result
	fixture.provider.pr.State, fixture.provider.pr.Merged, fixture.provider.pr.ResultCommit = "closed", true, fixture.result
	fixture.provider.pr.BaseCommit = fixture.target
	workflowFinalizeFrontier(t, *fixture)
}

// TestWorkflowLooseExhaustion retains excluded admissions without claiming they
// integrated; only the actually delivered prefix contributes to its cursor.
func TestWorkflowLooseExhaustion(t *testing.T) {
	f := workflowLooseQueueFixture(t, false)
	for _, key := range []string{"PRJ-91", "PRJ-92"} {
		workflowLooseClassify(t, f.store, workflowLooseRecord(t, f.store, key), "normal", string(ticketcore.StatusBacklog))
	}

	workflowFinalizeFrontier(t, f)
	workflowAdvanceFrontier(t, f)

	run, err := f.store.ReconcileWorkflowMembership(t.Context(), f.run.ID)
	if err != nil || run.Status != "completed" || run.Cursor != 1 || run.ConfirmedCursor != 1 || len(run.Members) != 3 || len(run.LooseQueue.Excluded) != 2 || len(run.Deliveries) != 1 {
		t.Fatal("excluded work was hidden or counted as delivery", err)
	}
}
