package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type workflowParallelPublicationLane struct {
	fixture  workflowPreparationFixture
	provider *workflowParallelPublicationProvider
}

// Each PR has its own server observation; sharing the canonical branch must not
// make one lane read another lane's PR after the first publication.
type workflowParallelPublicationProvider struct {
	*workflowSubmissionProviderFixture
}

func (p *workflowParallelPublicationProvider) observationPath() string {
	return filepath.Join(p.remote, "pr-"+workflowHash([]byte(p.pr.URL))+".json")
}

func (p *workflowParallelPublicationProvider) ReadWorkflowPullRequest(ctx context.Context, url string) (WorkflowPullRequestSnapshot, error) {
	pr := p.pr
	if url != pr.URL {
		return pr, ErrWorkflowConflict
	}

	data, err := os.ReadFile(p.observationPath())
	if err == nil {
		err = json.Unmarshal(data, &pr)
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return pr, err
	}

	pr.BaseCommit, err = p.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)

	return pr, err
}

func (p *workflowParallelPublicationProvider) SubmitWorkflowIntegration(ctx context.Context, submission WorkflowSubmission) (WorkflowSubmissionResponse, error) {
	response, err := p.workflowSubmissionProviderFixture.SubmitWorkflowIntegration(ctx, submission)
	if !response.Accepted {
		return response, err
	}

	pr := submission.Guard.PullRequest
	pr.State, pr.Merged, pr.ResultCommit = "closed", true, response.ResultCommit

	data, encodeErr := json.Marshal(pr)
	if encodeErr != nil {
		return response, errors.Join(err, encodeErr)
	}

	return response, errors.Join(err, atomicWorkflowFile(p.observationPath(), data))
}

func workflowParallelPublicationFixture(t *testing.T, conflict bool) [2]workflowParallelPublicationLane {
	t.Helper()

	return workflowParallelPublicationTargets(t, conflict, "dev")
}

func workflowParallelPublicationTargets(t *testing.T, conflict bool, target string) [2]workflowParallelPublicationLane {
	t.Helper()

	gate := workflowPreparationGate("test -f alpha.txt || test -f beta.txt || exit 1; if test -f alpha.txt && test -f beta.txt; then printf combined; else printf single; fi")
	paths := [2]string{"alpha.txt", "beta.txt"}

	if conflict {
		gate, paths = workflowPreparationGate("test -f shared.txt"), [2]string{"shared.txt", "shared.txt"}
	}

	store, first, request := workflowTicketFixture(t, "ticket", gate)
	other := workflowParallelPublicationStore(t, store)
	second := workflowLooseCreate(t, other, first, "PRJ-91", "normal", nil)
	workflowSelectionGit(t, store, "add", "ops/other/coordinator.json")

	request.Selection.Target.Commit, _ = workflowSelectionCommit(t, store)

	remote := filepath.Join(t.TempDir(), "canonical.git")
	workflowIntegrationCommand(t, filepath.Dir(remote), "init", "--bare", remote)
	workflowIntegrationCommand(t, store.layout.ProjectDir(), "push", remote, request.Selection.Target.Commit+":refs/heads/dev")

	if target != "dev" {
		workflowIntegrationCommand(t, store.layout.ProjectDir(), "push", remote, request.Selection.Target.Commit+":refs/heads/"+target)
	}

	return [2]workflowParallelPublicationLane{
		workflowParallelPublicationSource(t, store, first, request, paths[0], remote, 7),
		workflowParallelPublicationTargetSource(t, other, second, request, paths[1], remote, 8, target),
	}
}

func workflowParallelPublicationStore(t *testing.T, original *WorkflowStore) *WorkflowStore {
	t.Helper()

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: original.layout.ProjectDir(), Namespace: "other"})
	if err != nil {
		t.Fatal(err)
	}

	return NewWorkflowStore(layout, original.projectID)
}

func workflowParallelPublicationSource(t *testing.T, store *WorkflowStore, record WorkflowRecord, request WorkflowStartRequest, path, remote string, number int) workflowParallelPublicationLane {
	t.Helper()

	return workflowParallelPublicationTargetSource(t, store, record, request, path, remote, number, "dev")
}

func workflowParallelPublicationTargetSource(t *testing.T, store *WorkflowStore, record WorkflowRecord, request WorkflowStartRequest, path, remote string, number int, target string) workflowParallelPublicationLane {
	t.Helper()

	request.RequestKey, request.Selection.Selector.Identity = record.Task.Key, record.Task.Key
	run := workflowTestStart(t, store, request).Invocation
	claim := workflowTestClaim(t, store, run.ID)
	worker := []string{"sh", "-c", "printf '%s\\n' \"$2\" > \"$1\" && git add -- \"$1\" && git -c user.name=Test -c user.email=test@example.com commit -qm 'Independent result'", "source", path, record.Task.Key}

	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, worker); err != nil {
		t.Fatal(err)
	}

	candidate, err := store.PrepareWorkflowTicketCandidate(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	workflowTicketReviewConfig(t, store, "rebase", ReviewWorkflowModeTrustedAutoMerge)
	workflowIntegrationCommand(t, store.layout.ProjectDir(), "push", remote, candidate.Result.Commit+":"+candidate.Lane.Branch)

	pr := WorkflowPullRequestSnapshot{URL: request.Selection.Target.Repository + "/pulls/" + strconv.Itoa(number), Repository: request.Selection.Target.Repository, HeadRepository: request.Selection.Target.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: target, HeadCommit: candidate.Result.Commit, BaseCommit: request.Selection.Target.Commit, State: "open"}

	return workflowParallelPublicationPrepare(t, store, record, run.ID, remote, pr)
}

func workflowParallelPublicationPrepare(t *testing.T, store *WorkflowStore, record WorkflowRecord, id, remote string, pr WorkflowPullRequestSnapshot) workflowParallelPublicationLane {
	t.Helper()

	provider := workflowParallelPublicationProviderFor(pr, remote)
	content := workflowPRContent()

	content.BaseBranch = pr.BaseBranch
	if _, err := store.EnsureWorkflowSlicePullRequest(t.Context(), id, content, provider); err != nil {
		t.Fatal(err)
	}

	fixture := workflowPreparationFixture{store: store, hierarchy: WorkflowMaterializationResult{Tasks: []WorkflowRecord{record}}, run: workflowTestRead(t, store, id), provider: provider.workflowPreparationProviderFixture, git: NewWorkflowPreparationGit(pr.Repository, remote), remote: remote, base: pr.BaseCommit}

	if _, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, provider, fixture.git); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ReviewWorkflowSlice(t.Context(), id, provider, nil); err != nil {
		t.Fatal(err)
	}

	return workflowParallelPublicationLane{fixture, provider}
}

func workflowParallelPublicationProviderFor(pr WorkflowPullRequestSnapshot, remote string) *workflowParallelPublicationProvider {
	preparation := &workflowPreparationProviderFixture{workflowReviewProviderFixture: &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}, target: pr.BaseCommit}, guarded: true, revision: "guard-v1"}

	return &workflowParallelPublicationProvider{&workflowSubmissionProviderFixture{workflowPreparationProviderFixture: preparation, remote: remote, sealed: map[string]bool{}}}
}
