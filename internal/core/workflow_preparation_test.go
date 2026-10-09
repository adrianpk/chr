package core

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

type workflowPreparationProviderFixture struct {
	*workflowReviewProviderFixture
	guarded  bool
	revision string
}

func (p *workflowPreparationProviderFixture) WorkflowIntegrationCapability(_ context.Context, pr WorkflowPullRequestSnapshot, method string) (WorkflowPublicationCapability, error) {
	return WorkflowPublicationCapability{Repository: pr.Repository, PullRequestURL: pr.URL, Method: method, Mechanism: "fixture_atomic_source_base", Revision: workflowHash([]byte(p.revision)), AtomicSourceBase: p.guarded, ExactResult: p.guarded}, nil
}

type workflowPreparationFixture struct {
	store               *WorkflowStore
	hierarchy           WorkflowMaterializationResult
	run                 WorkflowInvocation
	provider            *workflowPreparationProviderFixture
	git                 OSWorkflowPreparationGit
	actor, remote, base string
}

func newWorkflowPreparationFixture(t *testing.T, method string, gate WorkflowValidation) workflowPreparationFixture {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)

	opts.Plan.Slices[0].Validation = gate
	for i := range opts.Plan.Slices[0].Tasks {
		opts.Plan.Slices[0].Tasks[i].Validation = WorkflowValidation{WaiverReason: "Preparation fixture validates the complete slice"}
	}

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	commit, _ := workflowSelectionCommit(t, store)
	request := WorkflowStartRequest{RequestKey: "slice-prepare", Selection: WorkflowPreviewRequest{Selector: workflowClaimSelector(hierarchy, "slice"), Target: WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}}}
	run := workflowTestStart(t, store, request).Invocation

	candidate, err := store.ExecuteWorkflowSlice(t.Context(), run.ID, workflowPRWorker())
	if err != nil {
		t.Fatal(err)
	}

	config := "review:\n  mode: trusted_auto_merge\nwork:\n  merge_strategy: " + method + "\n"
	writeWorkspaceConfig(t, store.layout.ProjectDir(), config)

	repo := request.Selection.Target.Repository
	pr := WorkflowPullRequestSnapshot{URL: repo + "/pulls/7", Repository: repo, HeadRepository: repo, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: commit, State: "open"}
	provider := &workflowPreparationProviderFixture{workflowReviewProviderFixture: &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}, target: commit}, guarded: true, revision: "guard-v1"}
	workflowPREnsure(t, store, run.ID, provider)

	root := t.TempDir()
	remote, actor := filepath.Join(root, "canonical.git"), filepath.Join(root, "target-writer")
	workflowIntegrationCommand(t, root, "init", "--bare", remote)
	workflowIntegrationCommand(t, candidate.Lane.Worktree, "push", remote, commit+":refs/heads/dev", candidate.Result.Commit+":"+candidate.Lane.Branch)
	workflowIntegrationCommand(t, root, "clone", "--no-checkout", remote, actor)
	workflowIntegrationCommand(t, actor, "config", "user.name", "Test")
	workflowIntegrationCommand(t, actor, "config", "user.email", "test@example.com")
	workflowIntegrationCommand(t, actor, "checkout", "-b", "dev", "origin/dev")

	base := workflowIntegrationCommit(t, actor, "target.txt", "Concurrent target content")
	workflowIntegrationCommand(t, actor, "push", "origin", "dev")

	provider.pr.BaseCommit, provider.target = base, base

	return workflowPreparationFixture{store, hierarchy, workflowTestRead(t, store, run.ID), provider, NewWorkflowPreparationGit(repo, remote), actor, remote, base}
}

func workflowPreparationGate(script string) WorkflowValidation {
	return WorkflowValidation{Required: true, Commands: [][]string{{"sh", "-c", script}}, TimeoutSeconds: 30}
}

// TestWorkflowPreparationMethods validates actual combined trees and ordered
// method histories on an isolated lane, while source and canonical target stay
// unchanged. Restart reuses immutable host evidence without running the gate.
func TestWorkflowPreparationMethods(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) { testWorkflowPreparationMethod(t, method) })
	}
}

func testWorkflowPreparationMethod(t *testing.T, method string) {
	t.Helper()

	gate := workflowPreparationGate("test -f result.txt && if test -f target.txt; then printf combined; else printf source; fi")
	f := newWorkflowPreparationFixture(t, method, gate)

	old, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, f.provider, nil)
	if err != nil {
		t.Fatal(err)
	}

	candidate, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowPreparationMethodFacts(t, f, candidate, method)
	assertWorkflowPreparationSource(t, f)
	checkWorkflowPreparationReview(t, f, old, candidate)

	before := workflowTestRead(t, f.store, f.run.ID)
	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	again, err := restarted.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git)
	if err != nil || again.Validation.AttemptID != candidate.Validation.AttemptID || workflowTestRead(t, restarted, f.run.ID).Revision != before.Revision {
		t.Fatal("preparation retry repeated validation or changed evidence")
	}
}

func checkWorkflowPreparationMethodFacts(t *testing.T, f workflowPreparationFixture, candidate WorkflowIntegrationCandidate, method string) {
	t.Helper()

	if candidate.Input.Git.Method != method || candidate.Input.Git.BaseCommit != f.base || candidate.Validation.Validation[0].Stdout != "combined" || candidate.Lane.Worktree == f.run.Lane.Worktree {
		t.Fatal("host gate did not validate the configured combined result")
	}

	if method == "rebase" && len(candidate.Input.Git.History) != len(f.run.Members) {
		t.Fatal("rebase discarded ordered source changes")
	}
}

func checkWorkflowPreparationReview(t *testing.T, f workflowPreparationFixture, old WorkflowReviewBinding, candidate WorkflowIntegrationCandidate) {
	t.Helper()

	if _, err := f.store.CheckWorkflowReviewGrant(t.Context(), f.run.ID, f.provider); err == nil {
		t.Fatal("source-only review authorized new combined content")
	}

	fresh, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, f.provider, nil)
	if err != nil || fresh.InputHash == old.InputHash {
		t.Fatalf("fresh combined review: %+v, %v", fresh, err)
	}

	guard, err := f.store.CheckWorkflowPublicationGuard(t.Context(), f.run.ID, f.provider)
	if err != nil || guard.Git.ResultCommit != candidate.Input.Git.ResultCommit || guard.PreparationHash == "" {
		t.Fatalf("guard: %+v, %v", guard, err)
	}
}

func assertWorkflowPreparationSource(t *testing.T, f workflowPreparationFixture) {
	t.Helper()

	if workflowIntegrationCommand(t, f.remote, "rev-parse", "refs/heads/dev") != f.base || workflowIntegrationCommand(t, f.run.Lane.Worktree, "rev-parse", "HEAD") != f.run.Lane.HeadCommit {
		t.Fatal("preparation published a target or changed the source")
	}

	if status := workflowIntegrationCommand(t, f.run.Lane.Worktree, "status", "--porcelain"); status != "" {
		t.Fatal("source lane was modified")
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Cursor != f.run.Cursor || run.CandidateRef != f.run.CandidateRef || run.DeliveryRef != "" || f.provider.creates != 1 {
		t.Fatal("preparation repeated work or finalized delivery")
	}

	checkWorkflowPreparationTaskProofs(t, f.store, run)
}

func checkWorkflowPreparationTaskProofs(t *testing.T, store *WorkflowStore, run WorkflowInvocation) {
	t.Helper()

	for _, member := range run.Members {
		record, err := store.Read(t.Context(), member.Admission.Path)
		if err != nil || record.Task.Status == TicketStatusDone || len(record.Task.Evidence) == 0 {
			t.Fatal("task proof was lost or finalized")
		}
	}
}
