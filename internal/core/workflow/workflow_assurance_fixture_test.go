package workflow

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// The forge metadata is controlled; publication uses the production atomic Git
// transport. These composed scenarios do not claim live-provider certification.
type workflowAssuranceProvider struct {
	*nativePreparationProviderFixture
	t             *testing.T
	remote        string
	prs           map[string]PullRequestSnapshot
	creates       int
	calls         int
	lostPublisher GitPublisher
}

func (p *workflowAssuranceProvider) FindWorkflowPullRequests(_ context.Context, query PullRequestQuery) ([]PullRequestSnapshot, error) {
	var found []PullRequestSnapshot

	for _, pr := range p.prs {
		if pr.Repository == query.Repository && pr.HeadBranch == query.HeadBranch && pr.BaseBranch == query.BaseBranch {
			found = append(found, pr)
		}
	}

	return found, nil
}

func (p *workflowAssuranceProvider) CreateWorkflowPullRequest(ctx context.Context, request PullRequestRequest) (PullRequestSnapshot, error) {
	head, err := p.ReadWorkflowBranch(ctx, request.Repository, request.HeadBranch)
	if err != nil {
		return PullRequestSnapshot{}, err
	}

	base, err := p.ReadWorkflowBranch(ctx, request.Repository, request.BaseBranch)
	if err != nil {
		return PullRequestSnapshot{}, err
	}

	p.creates++
	p.pr = PullRequestSnapshot{URL: fmt.Sprintf("%s/pulls/%d", request.Repository, p.creates), Repository: request.Repository, HeadRepository: request.Repository, HeadBranch: request.HeadBranch, BaseBranch: request.BaseBranch, HeadCommit: head, BaseCommit: base, State: "open"}
	p.prs[p.pr.URL] = p.pr

	return p.pr, nil
}

func (p *workflowAssuranceProvider) ReadWorkflowBranch(ctx context.Context, _, branch string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "--git-dir="+p.remote, "rev-parse", "refs/heads/"+branch).Output()

	return strings.TrimSpace(string(output)), err
}

func (p *workflowAssuranceProvider) ReadWorkflowPullRequest(ctx context.Context, address string) (PullRequestSnapshot, error) {
	pr, found := p.prs[address]
	if !found {
		return pr, workflowstate.ErrConflict
	}

	var err error

	pr.BaseCommit, err = p.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)

	return pr, err
}

func (p *workflowAssuranceProvider) PublishWorkflowGit(ctx context.Context, request GitPublication) (GitPublicationResult, error) {
	p.calls++

	var publisher GitPublisher = NewWorkflowGitPublisher(request.Repository, p.remote)
	if p.lostPublisher != nil {
		publisher = p.lostPublisher
	}

	result, err := publisher.PublishWorkflowGit(ctx, request)
	if result.CommandsSent {
		err = errors.Join(err, p.retainPublicationFacts(ctx, request))
	}

	return result, err
}

func (p *workflowAssuranceProvider) retainPublicationFacts(ctx context.Context, request GitPublication) error {
	source, sourceErr := p.ReadWorkflowBranch(ctx, request.Repository, strings.TrimPrefix(request.SourceRef, "refs/heads/"))
	target, targetErr := p.ReadWorkflowBranch(ctx, request.Repository, strings.TrimPrefix(request.TargetRef, "refs/heads/"))

	if sourceErr != nil || targetErr != nil {
		return errors.Join(sourceErr, targetErr)
	}

	if source == request.Git.ResultCommit && target == request.Git.ResultCommit {
		for key, pr := range p.prs {
			if "refs/heads/"+pr.HeadBranch == request.SourceRef {
				pr.State, pr.Merged, pr.ResultCommit, pr.HeadCommit = "closed", true, request.Git.ResultCommit, request.Git.ResultCommit
				p.prs[key] = pr
			}
		}
	}

	return nil
}

func (p *workflowAssuranceProvider) InspectWorkflowGitPublication(ctx context.Context, request GitPublication) (GitPublicationAdvertisement, error) {
	return NewWorkflowGitPublisher(request.Repository, p.remote).InspectWorkflowGitPublication(ctx, request)
}

func (p *workflowAssuranceProvider) ObserveWorkflowNativePublication(ctx context.Context, binding NativePublicationBinding, git IntegrationGit) (NativePublicationObservation, error) {
	pr, err := p.ReadWorkflowPullRequest(ctx, binding.PullRequest.URL)
	if err != nil {
		return NativePublicationObservation{}, err
	}

	runner := &workflowPRCommandFixture{fakeRemoteReviewRunner: newFakeRemoteReviewRunner()}
	forge := nativeCapabilityFixture{adapter: NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner), runner: runner, request: NativeCapabilityRequest{PullRequest: pr}}
	forge.prefix = nativePolicyCommand(p.t, forge, "")
	setNativePRFixture(p.t, forge, pr)

	forge.runner.outputs[forge.prefix+"/pulls/"+pr.URL[strings.LastIndex(pr.URL, "/")+1:]] = forge.runner.outputs[forge.prefix+"/pulls/7"]
	forge.runner.outputs[forge.prefix+"/branches/"+url.PathEscape(pr.HeadBranch)] = []byte(`{"commit":{"sha":"` + pr.HeadCommit + `"}}`)

	return forge.adapter.ObserveWorkflowNativePublication(ctx, binding, git)
}

func workflowAssurancePorts(t *testing.T, store *Store, target string) (ContinuationPorts, *workflowAssuranceProvider) {
	t.Helper()

	repository := "https://forge.example/test/repo"
	remote := filepath.Join(t.TempDir(), "canonical.git")
	workflowIntegrationCommand(t, store.layout.ProjectDir(), "init", "--bare", remote)
	workflowIntegrationCommand(t, store.layout.ProjectDir(), "push", remote, target+":refs/heads/dev")

	base := &workflowPRProviderFixture{}
	review := &workflowReviewProviderFixture{workflowPRProviderFixture: base, target: target}
	preparation := &workflowPreparationProviderFixture{workflowReviewProviderFixture: review, guarded: true, revision: "assurance-v1"}
	provider := &workflowAssuranceProvider{nativePreparationProviderFixture: &nativePreparationProviderFixture{preparation, "assurance-policy-v1"}, t: t, remote: remote, prs: map[string]PullRequestSnapshot{}}
	observer := NewWorkflowIntegrationGit(repository, remote)
	ports := ContinuationPorts{Publication: PublicationPorts{Provider: provider, Preparer: NewWorkflowPreparationGit(repository, remote), Observer: observer, NativePublisher: provider, NativeObserver: provider}, Target: observer, Source: observer}

	writeWorkspaceConfig(t, store.layout.ProjectDir(), "review:\n  mode: trusted_auto_merge\nwork:\n  merge_strategy: rebase\n  integration:\n    publication_profile: git_atomic_dual_ref\n")

	return ports, provider
}

func workflowAssuranceContinue(t *testing.T, store *Store, id string, ports ContinuationPorts) ContinuationProgress {
	t.Helper()

	restarted := NewWorkflowStore(store.layout, store.projectID)
	request := ContinuationRequest{Worker: workflowTaskCommand(t, "worker"), PullRequest: workflowPRContent(), StepLimit: 1}

	progress, err := restarted.ContinueWorkflow(t.Context(), id, request, ports)
	if err != nil {
		t.Fatalf("composed continuation (%s, cursor %d, confirmed %d, outcome %s): %v", progress.Reason, progress.Invocation.Cursor, progress.Invocation.ConfirmedCursor, progress.Invocation.Outcome, err)
	}

	if progress.Steps > 1 {
		t.Fatal("composed call exceeded its frozen checkpoint limit")
	}

	return progress
}
