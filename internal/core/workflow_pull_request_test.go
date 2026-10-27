package core

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowPRProviderFixture struct {
	mu                              sync.Mutex
	pr                              WorkflowPullRequestSnapshot
	found                           []WorkflowPullRequestSnapshot
	createErr                       error
	beforeIO                        func(string)
	createHook                      func()
	creates, reads, finds, branches int
}

func (p *workflowPRProviderFixture) FindWorkflowPullRequests(_ context.Context, _ WorkflowPullRequestQuery) ([]WorkflowPullRequestSnapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.finds++
	if p.beforeIO != nil {
		p.beforeIO("find")
	}

	return p.found, nil
}

func (p *workflowPRProviderFixture) ReadWorkflowPullRequest(_ context.Context, _ string) (WorkflowPullRequestSnapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.reads++
	if p.beforeIO != nil {
		p.beforeIO("read")
	}

	return p.pr, nil
}

func (p *workflowPRProviderFixture) CreateWorkflowPullRequest(_ context.Context, _ WorkflowPullRequestRequest) (WorkflowPullRequestSnapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.creates++
	if p.beforeIO != nil {
		p.beforeIO("create")
	}

	p.found = []WorkflowPullRequestSnapshot{p.pr}
	if p.createHook != nil {
		p.createHook()
	}

	return p.pr, p.createErr
}

func (p *workflowPRProviderFixture) ReadWorkflowBranch(_ context.Context, _, _ string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.branches++

	return p.pr.HeadCommit, nil
}

// TestWorkflowSlicePRConcurrent serializes two independent store callers on the
// common repository lock and creates the canonical source PR exactly once.
func TestWorkflowSlicePRConcurrent(t *testing.T) {
	store, hierarchy, run, provider := workflowPRFixture(t)
	results := make(chan error, 2)
	start := make(chan struct{})

	for range 2 {
		go func() {
			<-start

			caller := NewWorkflowStore(store.layout, store.projectID)

			_, err := caller.EnsureWorkflowSlicePullRequest(t.Context(), run.ID, workflowPRContent(), provider)
			results <- err
		}()
	}

	close(start)

	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}

	if provider.creates != 1 || provider.reads != 1 {
		t.Fatal("concurrent callers created duplicate provider PRs")
	}

	checkWorkflowPRBinding(t, store, hierarchy, run.ID, provider.pr)
}

func workflowPRContent() WorkflowSlicePullRequestRequest {
	return WorkflowSlicePullRequestRequest{BaseBranch: "dev", Title: "Deliver approved slice", Body: "Deliver the aggregate candidate."}
}

func workflowPRFixture(t *testing.T) (*WorkflowStore, WorkflowMaterializationResult, WorkflowInvocation, *workflowPRProviderFixture) {
	t.Helper()

	return workflowPRFixtureFor(t, "slice")
}

func workflowPRFixtureFor(t *testing.T, extent string) (*WorkflowStore, WorkflowMaterializationResult, WorkflowInvocation, *workflowPRProviderFixture) {
	t.Helper()

	return workflowPRPolicyFixture(t, extent, workflowstate.ExecutionPolicy{})
}

func workflowPRPolicyFixture(t *testing.T, extent string, policy workflowstate.ExecutionPolicy) (*WorkflowStore, WorkflowMaterializationResult, WorkflowInvocation, *workflowPRProviderFixture) {
	t.Helper()

	return workflowPRPlanFixture(t, extent, policy, workflowMaterializationFixture)
}

func workflowPRPlanFixture(t *testing.T, extent string, policy workflowstate.ExecutionPolicy, planFixture func(*testing.T) (*WorkflowStore, WorkflowMaterializationOptions)) (*WorkflowStore, WorkflowMaterializationResult, WorkflowInvocation, *workflowPRProviderFixture) {
	t.Helper()

	store, opts := planFixture(t)
	waiver := workflowstate.Validation{WaiverReason: "PR lifecycle fixture uses retained immutable host waiver"}

	for i := range opts.Plan.Slices {
		opts.Plan.Slices[i].Validation = waiver
		for j := range opts.Plan.Slices[i].Tasks {
			opts.Plan.Slices[i].Tasks[j].Validation = waiver
		}
	}

	if extent == "use_case" {
		opts.Plan.UseCase = "ordered-case"
	}

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	if extent == "use_case" {
		independent := workflowTestTask("PRJ-99")
		independent.Task.UseCase = opts.Plan.UseCase

		if _, err := store.Create(t.Context(), "Excluded unsliced work", independent); err != nil {
			t.Fatal(err)
		}
	}

	commit, _ := workflowSelectionCommit(t, store)
	request := WorkflowStartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "slice-pr",
		Selection:  workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, extent), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}}
	request.Policy = policy
	run := workflowTestStart(t, store, request)

	candidate, err := store.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, workflowPRWorker())
	if err != nil {
		t.Fatal(err)
	}

	pr := WorkflowPullRequestSnapshot{URL: request.Selection.Target.Repository + "/pulls/7", Repository: request.Selection.Target.Repository, HeadRepository: request.Selection.Target.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: commit, State: "open"}

	return store, hierarchy, run.Invocation, &workflowPRProviderFixture{pr: pr}
}

func workflowPRWorker() []string {
	return []string{"sh", "-c", "printf 'Task result\\n' >> result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task result'"}
}

func workflowPREnsure(t *testing.T, store *WorkflowStore, id string, provider WorkflowPullRequestProvider) WorkflowPullRequestSnapshot {
	t.Helper()

	pr, err := store.EnsureWorkflowSlicePullRequest(t.Context(), id, workflowPRContent(), provider)
	if err != nil {
		t.Fatal(err)
	}

	return pr
}

// TestWorkflowSlicePRReuse freezes intent before IO, shares one PR across all
// tasks and reuses it from a new store without rerunning work or finalizing tasks.
func TestWorkflowSlicePRReuse(t *testing.T) {
	store, hierarchy, run, provider := workflowPRFixture(t)
	provider.beforeIO = func(operation string) {
		retained, err := store.readWorkflowInvocation(run.ID)
		if err != nil || retained.PullRequest == nil {
			t.Fatal("provider IO preceded durable PR intent")
		}

		if operation == "create" && retained.PullRequest.State != "creating" {
			t.Fatal("PR creation preceded durable creation intent")
		}
	}
	pr := workflowPREnsure(t, store, run.ID, provider)

	restarted := NewWorkflowStore(store.layout, store.projectID)
	if again := workflowPREnsure(t, restarted, run.ID, provider); again != pr || provider.creates != 1 || provider.finds != 1 || provider.reads != 1 {
		t.Fatal("restart replaced its canonical PR")
	}

	checkWorkflowPRBinding(t, store, hierarchy, run.ID, pr)
}

func checkWorkflowPRBinding(t *testing.T, store *WorkflowStore, hierarchy WorkflowMaterializationResult, id string, pr WorkflowPullRequestSnapshot) {
	t.Helper()

	run, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || run.Status != "waiting_review" || run.PullRequest.URL != pr.URL || run.PullRequest.ObservationCount != 1 || len(run.Attempts) != 4 {
		t.Fatalf("PR binding: %+v, %v", run, err)
	}

	checkWorkflowPRSet(t, store, hierarchy.Set.Path, run)
	checkWorkflowPRTasks(t, store, hierarchy.Tasks[:3])

	if _, err := store.ReadWorkflowSliceCandidate(t.Context(), id); err != nil {
		t.Fatal("review status invalidated immutable aggregate host proof")
	}
}

func checkWorkflowPRSet(t *testing.T, store *WorkflowStore, path string, run WorkflowInvocation) {
	t.Helper()

	set, err := store.Read(t.Context(), path)
	if err != nil || set.Set.Slices[0].PullRequest != run.PullRequest.URL || set.Set.Slices[0].CandidateRef != run.CandidateRef || set.Set.Slices[0].Branch != run.Lane.Branch {
		t.Fatal("set lacks common branch/candidate/PR binding")
	}
}

func checkWorkflowPRTasks(t *testing.T, store *WorkflowStore, tasks []WorkflowRecord) {
	t.Helper()

	for _, task := range tasks {
		record, err := store.Read(t.Context(), task.Path)
		if err != nil || record.Task.Phase != "waiting_slice" || record.Task.Status != string(ticketcore.StatusReadyForReview) || record.Task.DeliveryRef != "" {
			t.Fatal("PR binding changed task completion or failed to share review status")
		}
	}
}

// TestWorkflowSlicePRObservation retains a lost create response and reuses only
// canonical observation; absence after uncertainty never authorizes another create.
func TestWorkflowSlicePRObservation(t *testing.T) {
	store, _, run, provider := workflowPRFixture(t)

	provider.createErr = errors.New("lost provider response")
	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), run.ID, workflowPRContent(), provider); !errors.Is(err, ErrWorkflowAwaitingObservation) {
		t.Fatal("creation uncertainty was not retained")
	}

	provider.found = nil
	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), run.ID, workflowPRContent(), provider); !errors.Is(err, ErrWorkflowAwaitingObservation) || provider.creates != 1 {
		t.Fatal("unobserved creation was blindly retried")
	}

	provider.found = []WorkflowPullRequestSnapshot{provider.pr}

	pr := workflowPREnsure(t, NewWorkflowStore(store.layout, store.projectID), run.ID, provider)
	if pr.URL != provider.pr.URL || provider.creates != 1 {
		t.Fatal("lost response recovery created another PR")
	}
}

// TestWorkflowSlicePRBlock rejects closed, ambiguous, foreign and changed heads
// while preserving the offending observation and never replacing the shared PR.
func TestWorkflowSlicePRBlock(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*workflowPRProviderFixture)
	}{
		{"closed unmerged", func(p *workflowPRProviderFixture) { p.pr.State = "closed" }},
		{"foreign repository", func(p *workflowPRProviderFixture) { p.pr.Repository = "https://forge.example/other/repo" }},
		{"fork head", func(p *workflowPRProviderFixture) { p.pr.HeadRepository = "https://forge.example/fork/repo" }},
		{"wrong source branch", func(p *workflowPRProviderFixture) { p.pr.HeadBranch = "other" }},
		{"wrong target branch", func(p *workflowPRProviderFixture) { p.pr.BaseBranch = "main" }},
		{"wrong candidate", func(p *workflowPRProviderFixture) { p.pr.HeadCommit = strings.Repeat("f", 40) }},
		{"wrong PR URL", func(p *workflowPRProviderFixture) { p.pr.URL = "https://forge.example/test/repo/pulls/8" }},
		{"merged without result", func(p *workflowPRProviderFixture) { p.pr.Merged, p.pr.State = true, "closed" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, _, run, provider := workflowPRFixture(t)
			workflowPREnsure(t, store, run.ID, provider)
			tc.mutate(provider)
			assertWorkflowPRBlocked(t, store, run.ID, provider)
		})
	}
}

func assertWorkflowPRBlocked(t *testing.T, store *WorkflowStore, id string, provider *workflowPRProviderFixture) {
	t.Helper()

	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), id, workflowPRContent(), provider); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("mismatched canonical PR accepted")
	}

	run, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || run.Status != "blocked" || run.PullRequest.State != "blocked" || run.PullRequest.ObservationRef == "" || run.Blocker == nil {
		t.Fatal("PR blocker or observation was not durable")
	}

	reads := provider.reads
	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), id, workflowPRContent(), provider); !errors.Is(err, workflowstate.ErrConflict) || provider.creates > 1 || provider.reads != reads {
		t.Fatal("ordinary retry bypassed explicit reconciliation")
	}
}

// TestWorkflowSlicePRExisting adopts only a matching unique existing PR; closed
// or duplicate source PRs block before any creation or remote branch mutation.
func TestWorkflowSlicePRExisting(t *testing.T) {
	for _, name := range []string{"open", "closed", "ambiguous"} {
		t.Run(name, func(t *testing.T) {
			store, _, run, provider := workflowPRFixture(t)
			if name == "closed" {
				provider.pr.State = "closed"
			}

			provider.found = []WorkflowPullRequestSnapshot{provider.pr}
			if name == "ambiguous" {
				provider.found = append(provider.found, provider.pr)
			}

			if name == "open" {
				workflowPREnsure(t, store, run.ID, provider)
			} else {
				assertWorkflowPRBlocked(t, store, run.ID, provider)
			}

			if provider.creates != 0 || provider.branches != 0 {
				t.Fatal("existing source PR was replaced")
			}
		})
	}
}

// TestWorkflowSlicePRRepair retains the same branch and PR identity, removes
// aggregate review readiness and rebinds a freshly validated candidate once.
func TestWorkflowSlicePRRepair(t *testing.T) {
	store, hierarchy, run, provider := workflowPRFixture(t)
	pr := workflowPREnsure(t, store, run.ID, provider)

	before, err := store.ReadWorkflow(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	repaired, err := store.RepairWorkflowSlice(t.Context(), run.ID, hierarchy.Tasks[1].Task.Key)
	if err != nil || repaired.PullRequest.URL != pr.URL || repaired.CandidateRef != "" {
		t.Fatalf("PR repair: %+v, %v", repaired, err)
	}

	checkWorkflowPRRepairSet(t, store, hierarchy.Set.Path, pr.URL)

	candidate, err := store.ExecuteWorkflowSlice(t.Context(), run.ID, workflowPRWorker())
	if err != nil {
		t.Fatal(err)
	}

	provider.pr.HeadCommit = candidate.Result.Commit
	if again := workflowPREnsure(t, store, run.ID, provider); again.URL != pr.URL || provider.creates != 1 {
		t.Fatal("repair created a replacement PR")
	}

	checkWorkflowPRRepairBinding(t, store, run.ID, before)
}

func checkWorkflowPRRepairSet(t *testing.T, store *WorkflowStore, path, prURL string) {
	t.Helper()

	set, err := store.Read(t.Context(), path)
	if err != nil || set.Set.Slices[0].CandidateRef != "" || set.Set.Slices[0].PullRequest != prURL || set.Set.Slices[0].Phase != "developing" {
		t.Fatal("repair retained stale review candidate or discarded PR identity")
	}
}

func checkWorkflowPRRepairBinding(t *testing.T, store *WorkflowStore, id string, before WorkflowInvocation) {
	t.Helper()

	after, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || after.PullRequest.ID != before.PullRequest.ID || after.PullRequest.CandidateID == before.PullRequest.CandidateID || after.Lane.Branch != before.Lane.Branch {
		t.Fatal("repair lost creation identity or reused stale candidate")
	}
}

// TestWorkflowSlicePRSource refuses to create from an unpublished candidate and
// refuses changed target/content intent; provider merge metadata never marks done.
func TestWorkflowSlicePRSource(t *testing.T) {
	store, hierarchy, run, provider := workflowPRFixture(t)
	source := provider.pr.HeadCommit

	provider.pr.HeadCommit = run.Selection.Target.Commit
	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), run.ID, workflowPRContent(), provider); !errors.Is(err, workflowstate.ErrConflict) || provider.creates != 0 {
		t.Fatal("remote source mismatch created a PR")
	}

	provider.pr.HeadCommit = source
	workflowPREnsure(t, store, run.ID, provider)

	changed := workflowPRContent()

	changed.BaseBranch = "main"
	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), run.ID, changed, provider); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("target branch binding changed")
	}

	provider.pr.Merged, provider.pr.State, provider.pr.ResultCommit = true, "closed", source
	workflowPREnsure(t, store, run.ID, provider)

	retained, err := store.ReadWorkflow(t.Context(), run.ID)
	if err != nil || retained.Status != "waiting_integration" {
		t.Fatal("provider merge observation skipped delivery verification")
	}

	checkWorkflowPRMergedRepair(t, store, run.ID, hierarchy.Tasks[0].Task.Key)

	for _, task := range hierarchy.Tasks[:3] {
		record, err := store.Read(t.Context(), task.Path)
		if err != nil || record.Task.Status == string(ticketcore.StatusDone) || record.Task.DeliveryRef != "" {
			t.Fatal("worker/provider metadata finalized task without host integration proof")
		}
	}
}

func checkWorkflowPRMergedRepair(t *testing.T, store *WorkflowStore, id, task string) {
	t.Helper()

	if _, err := store.RepairWorkflowSlice(t.Context(), id, task); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("repair discarded an observed integration before verification")
	}
}
