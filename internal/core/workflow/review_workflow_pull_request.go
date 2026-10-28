package workflow

import (
	"context"
	"fmt"
	"strings"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// CheckWorkflowRepositoryProvider checks local provider prerequisites without
// reading or creating a pull request. Remote authentication remains untested.
func (adapter RemoteReviewAdapter) CheckWorkflowRepositoryProvider(ctx context.Context, repository string) error {
	resolvedReview, err := adapter.resolveWorkflowRepository(repository)
	if err != nil {
		return err
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	diagnostic := ReviewProviderDiagnostic{}
	diagnostic.addProviderChecks(ctx, OSReviewProviderDiagnosticRunner{}, reference, config, true)

	if reviewProviderDiagnosticStatus(diagnostic.Checks) != ReviewProviderDiagnosticStatusReady {
		return fmt.Errorf("workflow repository provider setup is incomplete")
	}

	return nil
}

// FindWorkflowPullRequests resolves enabled host configuration before provider IO.
func (adapter RemoteReviewAdapter) FindWorkflowPullRequests(ctx context.Context, query workflowstate.PullRequestQuery) ([]workflowstate.PullRequestSnapshot, error) {
	resolved, err := adapter.resolveWorkflowRepository(query.Repository)
	if err != nil {
		return nil, err
	}

	client := adapter.workflowForgeClient(query.Repository, resolved)

	return client.Find(ctx, query)
}

// CreateWorkflowPullRequest resolves enabled host configuration before provider IO.
func (adapter RemoteReviewAdapter) CreateWorkflowPullRequest(ctx context.Context, request workflowstate.PullRequestRequest) (workflowstate.PullRequestSnapshot, error) {
	resolved, err := adapter.resolveWorkflowRepository(request.Repository)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	client := adapter.workflowForgeClient(request.Repository, resolved)

	return client.Create(ctx, request)
}

// ReadWorkflowBranch resolves enabled host configuration before provider IO.
func (adapter RemoteReviewAdapter) ReadWorkflowBranch(ctx context.Context, repository, branch string) (string, error) {
	resolved, err := adapter.resolveWorkflowRepository(repository)
	if err != nil {
		return "", err
	}

	client := adapter.workflowForgeClient(repository, resolved)

	return client.Branch(ctx, branch)
}

// ReadWorkflowPullRequest validates the canonical URL before provider IO.
func (adapter RemoteReviewAdapter) ReadWorkflowPullRequest(ctx context.Context, rawURL string) (workflowstate.PullRequestSnapshot, error) {
	repository, err := workflowPullRequestRepository(rawURL)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	resolved, err := adapter.resolve(rawURL)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	client := adapter.workflowForgeClient(repository, resolved)

	return client.Read(ctx)
}

func (adapter RemoteReviewAdapter) resolveWorkflowRepository(repository string) (reviewProviderEndpoint, error) {
	canonical, err := workflowstate.CanonicalRepository(repository)
	if err != nil {
		return reviewProviderEndpoint{}, err
	}

	marker := "/pulls/1"
	if strings.HasPrefix(canonical, "https://github.com/") {
		marker = "/pull/1"
	}

	return adapter.resolve(canonical + marker)
}

func (adapter RemoteReviewAdapter) workflowForgeClient(repository string, resolved reviewProviderEndpoint) workflowforge.Client {
	return workflowforge.Client{Endpoint: workflowForgeEndpoint(repository, resolved.reference, resolved.config), Commands: adapter.workflowProviderRunner()}
}

func workflowForgeEndpoint(repository string, reference ReviewProviderReference, config workspaceconfig.ReviewProviderTargetConfig) workflowforge.Endpoint {
	return workflowforge.Endpoint{Repository: repository, Provider: string(reference.Provider), Host: reference.Host, Owner: reference.Owner, Repo: reference.Repo, Number: reference.Number, Command: config.CLI}
}
