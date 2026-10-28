package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type pullRequestJSON struct {
	URL          string            `json:"html_url"`
	State        string            `json:"state"`
	Merged       bool              `json:"merged"`
	ResultCommit string            `json:"merge_commit_sha"`
	Head         pullRequestBranch `json:"head"`
	Base         pullRequestBranch `json:"base"`
}

type pullRequestBranch struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo struct {
		URL string `json:"html_url"`
	} `json:"repo"`
}

// Find observes all states and rejects incomplete pagination.
func (adapter Client) Find(ctx context.Context, query workflowstate.PullRequestQuery) ([]workflowstate.PullRequestSnapshot, error) {
	if !workflowstate.ValidBranch(query.HeadBranch) || !workflowstate.ValidBranch(query.BaseBranch) {
		return nil, workflowstate.ErrConflict
	}

	found := []workflowstate.PullRequestSnapshot{}
	count := 0

	for page := 1; page <= workflowstate.PageLimit; page++ {
		endpoint := adapter.Endpoint.RepositoryPath() + "/pulls?state=all&limit=50&per_page=50&page=" + strconv.Itoa(page)

		items, err := adapter.readPRPage(ctx, endpoint)
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

		found, err = appendPRMatches(found, items, query.HeadBranch)
		if err != nil {
			return nil, err
		}

		if len(found) > 1 {
			return found, nil
		}
	}

	return nil, fmt.Errorf("workflow provider PR pagination incomplete")
}

func (adapter Client) readPRPage(ctx context.Context, endpoint string) ([]pullRequestJSON, error) {
	data, err := adapter.Request(ctx, Request{Method: "GET", Path: endpoint})
	if err != nil {
		return nil, err
	}

	return decodePRList(data)
}

func decodePRList(data []byte) ([]pullRequestJSON, error) {
	if len(data) > workflowstate.FileLimit {
		return nil, fmt.Errorf("workflow provider response exceeds file bounds")
	}

	var items []pullRequestJSON
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	if items == nil || len(items) > workflowstate.PageLimit {
		return nil, fmt.Errorf("workflow provider PR page is invalid or oversized")
	}

	return items, nil
}

func appendPRMatches(found []workflowstate.PullRequestSnapshot, items []pullRequestJSON, branch string) ([]workflowstate.PullRequestSnapshot, error) {
	for _, item := range items {
		if item.Head.Ref != branch {
			continue
		}

		pr, err := pullRequestSnapshot(item)
		if err != nil {
			return nil, err
		}

		found = append(found, pr)
	}

	return found, nil
}

// Read retrieves provider metadata without review side effects.
func (adapter Client) Read(ctx context.Context) (workflowstate.PullRequestSnapshot, error) {
	endpoint := adapter.Endpoint.RepositoryPath() + "/pulls/" + strconv.Itoa(adapter.Endpoint.Number)

	data, err := adapter.Request(ctx, Request{Method: "GET", Path: endpoint})
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	return decodePullRequest(data)
}

// Create creates a PR only; it never reviews or merges it.
func (adapter Client) Create(ctx context.Context, request workflowstate.PullRequestRequest) (workflowstate.PullRequestSnapshot, error) {
	if !workflowstate.ValidBranch(request.HeadBranch) || !ValidPullRequestContent(workflowstate.PullRequestContent{BaseBranch: request.BaseBranch, Title: request.Title, Body: request.Body}) {
		return workflowstate.PullRequestSnapshot{}, workflowstate.ErrConflict
	}

	payload, err := json.Marshal(map[string]string{"head": request.HeadBranch, "base": request.BaseBranch, "title": request.Title, "body": request.Body})
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	data, err := adapter.Request(ctx, Request{Method: "POST", Path: adapter.Endpoint.RepositoryPath() + "/pulls", Body: payload})
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	return decodePullRequest(data)
}

// Branch checks the actual canonical source before PR creation.
func (adapter Client) Branch(ctx context.Context, branch string) (string, error) {
	if !workflowstate.ValidBranch(branch) {
		return "", workflowstate.ErrConflict
	}

	endpoint := adapter.Endpoint.RepositoryPath() + "/branches/" + url.PathEscape(branch)

	data, err := adapter.Request(ctx, Request{Method: "GET", Path: endpoint})
	if err != nil {
		return "", err
	}

	return decodeBranch(data)
}

func decodeBranch(data []byte) (string, error) {
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

func decodePullRequest(data []byte) (workflowstate.PullRequestSnapshot, error) {
	if len(data) > workflowstate.FileLimit {
		return workflowstate.PullRequestSnapshot{}, workflowstate.ErrConflict
	}

	var pr pullRequestJSON
	if err := json.Unmarshal(data, &pr); err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	return pullRequestSnapshot(pr)
}

func pullRequestSnapshot(pr pullRequestJSON) (workflowstate.PullRequestSnapshot, error) {
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

// ValidPullRequestContent bounds the retained target and initial PR content.
func ValidPullRequestContent(request workflowstate.PullRequestContent) bool {
	return workflowstate.ValidBranch(request.BaseBranch) && strings.TrimSpace(request.Title) != "" && len(request.Title) <= 256 && !strings.ContainsAny(request.Title, "\x00\r\n") && len(request.Body) <= 16*1024
}
