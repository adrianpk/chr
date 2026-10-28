package workflow

import (
	"context"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowReviewProviderFixture struct {
	*workflowPRProviderFixture
	target string
}

func (p *workflowReviewProviderFixture) ReadWorkflowBranch(_ context.Context, _, branch string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.branches++
	if branch == p.pr.BaseBranch {
		return p.target, nil
	}

	return p.pr.HeadCommit, nil
}

func workflowReviewFixture(t *testing.T, mode workflowstate.ReviewMode) (*Store, MaterializationResult, Invocation, *workflowReviewProviderFixture) {
	t.Helper()

	store, hierarchy, run, provider := workflowPRPolicyFixture(t, "slice", workflowstate.ExecutionPolicy{MaxWorkerAttempts: 128})
	workflowReviewConfig(t, store, mode)

	observed := &workflowReviewProviderFixture{workflowPRProviderFixture: provider, target: provider.pr.BaseCommit}
	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), run.ID, workflowPRContent(), observed); err != nil {
		t.Fatal(err)
	}

	return store, hierarchy, workflowTestRead(t, store, run.ID), observed
}

func workflowReviewConfig(t *testing.T, store *Store, mode workflowstate.ReviewMode) {
	t.Helper()

	writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "review", "mode", string(mode))
	writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "work", "merge_strategy", "rebase")
}

func workflowReviewWorker(decision, key, before string) []string {
	script := "import json,sys\ni=json.load(sys.stdin)\n" + before + "\nf=[]\n"
	if decision == "needs_changes" {
		script += "f=[{'category':'correctness','title':'Broken result','detail':'Repair the retained task result.'}]\n"
	}

	script += "print(json.dumps({'binding_hash':i['review_hash'],'result':{'decision':sys.argv[1],'summary':'Inspect the retained candidate and required evidence.','blocking_findings':f},'repair_task_key':sys.argv[2]}))"

	return []string{"python3", "-c", script, decision, key}
}

// TestWorkflowReviewModes proves the observable policy stops with real owned
// subprocesses. The source worker and canonical PR are never repeated by review.
func TestWorkflowReviewModes(t *testing.T) {
	cases := []struct {
		name                    string
		mode                    workflowstate.ReviewMode
		decision, state, status string
	}{
		{"human", workflowstate.ReviewHuman, "", "waiting_approval", "waiting_review"},
		{"trusted", workflowstate.ReviewTrustedAutoMerge, "", "authorized", "waiting_integration"},
		{"agent pass", workflowstate.ReviewAgentic, "pass", "authorized", "waiting_integration"},
		{"agent evidence", workflowstate.ReviewAgentic, "insufficient_evidence", "insufficient_evidence", "waiting_review"},
		{"agent human", workflowstate.ReviewAgentic, "human_required", "human_required", "waiting_review"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			testWorkflowReviewMode(t, test.mode, test.decision, test.state, test.status)
		})
	}
}

// TestWorkflowReviewApproval requires the displayed hash, retains both agentic
// human-required evidence and explicit approval, and does not rewrite retries.
func TestWorkflowReviewApproval(t *testing.T) {
	for _, mode := range []workflowstate.ReviewMode{workflowstate.ReviewHuman, workflowstate.ReviewAgentic} {
		t.Run(string(mode), func(t *testing.T) {
			testWorkflowReviewApproval(t, mode)
		})
	}
}

func testWorkflowReviewMode(t *testing.T, mode workflowstate.ReviewMode, decision, state, status string) {
	t.Helper()

	store, hierarchy, run, provider := workflowReviewFixture(t, mode)

	binding, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, workflowReviewWorker(decision, "", ""))
	if err != nil || binding.State != state {
		t.Fatalf("review: %+v, %v", binding, err)
	}

	retained := workflowTestRead(t, store, run.ID)
	if retained.Status != status || provider.creates != 1 || retained.CandidateRef != run.CandidateRef {
		t.Fatal("review repeated work or crossed its policy stop")
	}

	checkWorkflowReviewProjection(t, store, hierarchy.Tasks[:len(run.Members)], binding)

	again, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, []string{"false"})
	if err != nil || again != binding || workflowTestRead(t, store, run.ID).Revision != retained.Revision {
		t.Fatal("review retry repeated a decision")
	}

	checkWorkflowReviewAuthorized(t, store, run.ID, provider, binding)
}

func testWorkflowReviewApproval(t *testing.T, mode workflowstate.ReviewMode) {
	t.Helper()

	store, _, run, provider := workflowReviewFixture(t, mode)

	binding, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, workflowReviewWorker("human_required", "", ""))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.ApproveWorkflowSlice(t.Context(), run.ID, "another binding", provider); err == nil {
		t.Fatal("wrong approval accepted")
	}

	if _, err := store.CheckWorkflowReviewGrant(t.Context(), run.ID, provider); err == nil {
		t.Fatal("unapproved review granted integration")
	}

	approved, err := store.ApproveWorkflowSlice(t.Context(), run.ID, binding.InputHash, provider)
	if err != nil || approved.State != "authorized" {
		t.Fatalf("approval: %+v, %v", approved, err)
	}

	checkWorkflowApprovalRetry(t, store, run.ID, provider, binding)

	if mode == workflowstate.ReviewAgentic && approved.DecisionCount != 2 {
		t.Fatal("human approval discarded the agentic stop")
	}
}

func checkWorkflowReviewProjection(t *testing.T, store *Store, tasks []Record, binding workflowstate.ReviewBinding) {
	t.Helper()

	for _, task := range tasks {
		record, err := store.Read(t.Context(), task.Path)
		if err != nil || record.Task.Status == string(ticketcore.StatusDone) {
			t.Fatal("review completed a task without integration")
		}

		if (record.Task.Status == string(ticketcore.StatusReadyForMerge)) != (binding.State == "authorized") {
			t.Fatal("task projection differs from review grant")
		}
	}
}

func checkWorkflowReviewAuthorized(t *testing.T, store *Store, id string, provider *workflowReviewProviderFixture, binding workflowstate.ReviewBinding) {
	t.Helper()

	if binding.State == "authorized" {
		if _, err := store.CheckWorkflowReviewGrant(t.Context(), id, provider); err != nil {
			t.Fatal(err)
		}
	}
}

func checkWorkflowApprovalRetry(t *testing.T, store *Store, id string, provider *workflowReviewProviderFixture, binding workflowstate.ReviewBinding) {
	t.Helper()

	retained := workflowTestRead(t, store, id)
	if _, err := store.ApproveWorkflowSlice(t.Context(), id, binding.InputHash, provider); err != nil {
		t.Fatal(err)
	}

	if workflowTestRead(t, store, id).Revision != retained.Revision {
		t.Fatal("approval retry rewrote the immutable decision")
	}

	if _, err := store.CheckWorkflowReviewGrant(t.Context(), id, provider); err != nil {
		t.Fatal(err)
	}
}
