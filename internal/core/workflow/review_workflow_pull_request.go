package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type workflowRemotePRJSON struct {
	URL          string                 `json:"html_url"`
	State        string                 `json:"state"`
	Merged       bool                   `json:"merged"`
	ResultCommit string                 `json:"merge_commit_sha"`
	Head         workflowRemotePRBranch `json:"head"`
	Base         workflowRemotePRBranch `json:"base"`
}

type workflowRemotePRBranch struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo struct {
		URL string `json:"html_url"`
	} `json:"repo"`
}

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

// FindWorkflowPullRequests observes all states and rejects incomplete pagination.
func (adapter RemoteReviewAdapter) FindWorkflowPullRequests(ctx context.Context, query PullRequestQuery) ([]workflowstate.PullRequestSnapshot, error) {
	resolvedReview, err := adapter.resolveWorkflowRepository(query.Repository)
	if err != nil {
		return nil, err
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	if !workflowstate.ValidBranch(query.HeadBranch) || !workflowstate.ValidBranch(query.BaseBranch) {
		return nil, workflowstate.ErrConflict
	}

	found := []workflowstate.PullRequestSnapshot{}
	count := 0

	for page := 1; page <= workflowstate.PageLimit; page++ {
		endpoint := workflowRemotePRListPath(reference) + "?state=all&limit=50&per_page=50&page=" + strconv.Itoa(page)

		items, err := adapter.readWorkflowPRPage(ctx, query.Repository, reference, config, endpoint)
		if err != nil {
			return nil, err
		}

		count += len(items)
		if count > workflowstate.RecordLimit {
			return nil, fmt.Errorf("workflow provider PR record bound exceeded")
		}

		if len(items) == 0 {
			return found, nil
		}

		found, err = appendWorkflowPRMatches(found, items, query.HeadBranch)
		if err != nil {
			return nil, err
		}

		if len(found) > 1 {
			return found, nil
		}
	}

	return nil, fmt.Errorf("workflow provider PR pagination incomplete")
}

func (adapter RemoteReviewAdapter) readWorkflowPRPage(ctx context.Context, repository string, reference ReviewProviderReference, config workspaceconfig.ReviewProviderTargetConfig, endpoint string) ([]workflowRemotePRJSON, error) {
	data, err := adapter.workflowProviderRunner().Output(ctx, config.CLI, workflowRemoteAPIArgs(repository, reference, "GET", endpoint, false)...)
	if err != nil {
		return nil, err
	}

	return decodeWorkflowPRList(data)
}

func decodeWorkflowPRList(data []byte) ([]workflowRemotePRJSON, error) {
	if len(data) > workflowstate.FileLimit {
		return nil, fmt.Errorf("workflow provider response exceeds file bounds")
	}

	var items []workflowRemotePRJSON
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	if items == nil || len(items) > workflowstate.PageLimit {
		return nil, fmt.Errorf("workflow provider PR page is invalid or oversized")
	}

	return items, nil
}

func appendWorkflowPRMatches(found []workflowstate.PullRequestSnapshot, items []workflowRemotePRJSON, branch string) ([]workflowstate.PullRequestSnapshot, error) {
	for _, item := range items {
		if item.Head.Ref != branch {
			continue
		}

		pr, err := workflowRemotePRSnapshot(item)
		if err != nil {
			return nil, err
		}

		found = append(found, pr)
	}

	return found, nil
}

// ReadWorkflowPullRequest retrieves provider metadata without review side effects.
func (adapter RemoteReviewAdapter) ReadWorkflowPullRequest(ctx context.Context, rawURL string) (workflowstate.PullRequestSnapshot, error) {
	repository, err := workflowPullRequestRepository(rawURL)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	resolvedReview, err := adapter.resolve(rawURL)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	endpoint := workflowRemotePRListPath(reference) + "/" + strconv.Itoa(reference.Number)

	data, err := adapter.workflowProviderRunner().Output(ctx, config.CLI, workflowRemoteAPIArgs(repository, reference, "GET", endpoint, false)...)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	return decodeWorkflowRemotePR(data)
}

// CreateWorkflowPullRequest creates a PR only; it never reviews or merges it.
func (adapter RemoteReviewAdapter) CreateWorkflowPullRequest(ctx context.Context, request PullRequestRequest) (workflowstate.PullRequestSnapshot, error) {
	resolvedReview, err := adapter.resolveWorkflowRepository(request.Repository)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	if !workflowstate.ValidBranch(request.HeadBranch) || !validWorkflowPullRequestContent(SlicePullRequestRequest{BaseBranch: request.BaseBranch, Title: request.Title, Body: request.Body}) {
		return workflowstate.PullRequestSnapshot{}, workflowstate.ErrConflict
	}

	payload, err := json.Marshal(map[string]string{"head": request.HeadBranch, "base": request.BaseBranch, "title": request.Title, "body": request.Body})
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	data, err := adapter.workflowProviderRunner().Input(ctx, config.CLI, payload, workflowRemoteAPIArgs(request.Repository, reference, "POST", workflowRemotePRListPath(reference), true)...)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	return decodeWorkflowRemotePR(data)
}

// ReadWorkflowBranch checks the actual canonical source before PR creation.
func (adapter RemoteReviewAdapter) ReadWorkflowBranch(ctx context.Context, repository, branch string) (string, error) {
	resolvedReview, err := adapter.resolveWorkflowRepository(repository)
	if err != nil {
		return "", err
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	if !workflowstate.ValidBranch(branch) {
		return "", workflowstate.ErrConflict
	}

	endpoint := workflowRemoteRepositoryPath(reference) + "/branches/" + url.PathEscape(branch)

	data, err := adapter.workflowProviderRunner().Output(ctx, config.CLI, workflowRemoteAPIArgs(repository, reference, "GET", endpoint, false)...)
	if err != nil {
		return "", err
	}

	return decodeWorkflowRemoteBranch(data)
}

func decodeWorkflowRemoteBranch(data []byte) (string, error) {
	if len(data) > workflowstate.FileLimit {
		return "", workflowstate.ErrConflict
	}

	var branch struct {
		Commit struct {
			SHA string `json:"sha"`
			ID  string `json:"id"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(data, &branch); err != nil {
		return "", err
	}

	if branch.Commit.SHA == "" {
		branch.Commit.SHA = branch.Commit.ID
	}

	if branch.Commit.ID != "" && branch.Commit.ID != branch.Commit.SHA {
		return "", workflowstate.ErrConflict
	}

	if !workflowstate.ValidCommitID(branch.Commit.SHA) {
		return "", workflowstate.ErrConflict
	}

	return branch.Commit.SHA, nil
}

func decodeWorkflowRemotePR(data []byte) (workflowstate.PullRequestSnapshot, error) {
	if len(data) > workflowstate.FileLimit {
		return workflowstate.PullRequestSnapshot{}, workflowstate.ErrConflict
	}

	var pr workflowRemotePRJSON
	if err := json.Unmarshal(data, &pr); err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	return workflowRemotePRSnapshot(pr)
}

func workflowRemotePRSnapshot(pr workflowRemotePRJSON) (workflowstate.PullRequestSnapshot, error) {
	repository, err := workflowstate.CanonicalRepository(pr.Base.Repo.URL)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	headRepository, err := workflowstate.CanonicalRepository(pr.Head.Repo.URL)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	return workflowstate.PullRequestSnapshot{URL: pr.URL, Repository: repository, HeadRepository: headRepository, HeadBranch: pr.Head.Ref, BaseBranch: pr.Base.Ref, HeadCommit: pr.Head.SHA, BaseCommit: pr.Base.SHA, State: pr.State, Merged: pr.Merged, ResultCommit: pr.ResultCommit}, nil
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

func workflowRemoteRepositoryPath(reference ReviewProviderReference) string {
	return "/repos/" + reference.Owner + "/" + reference.Repo
}

func workflowRemotePRListPath(reference ReviewProviderReference) string {
	return workflowRemoteRepositoryPath(reference) + "/pulls"
}

func workflowRemoteAPIArgs(repository string, reference ReviewProviderReference, method, endpoint string, input bool) []string {
	if reference.Provider == ReviewProviderGitHub {
		args := []string{"api", "--hostname", reference.Host, "-X", method, strings.TrimPrefix(endpoint, "/")}
		if input {
			args = append(args, "--input", "-")
		}

		return args
	}

	root := strings.TrimSuffix(repository, "/"+reference.Owner+"/"+reference.Repo)

	args := []string{"api", "-X", method, root + "/api/v1" + endpoint}
	if input {
		args = append(args, "-d", "@-")
	}

	return args
}
