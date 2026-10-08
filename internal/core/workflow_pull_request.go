package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrWorkflowAwaitingObservation retains uncertain PR creation without retrying it.
var ErrWorkflowAwaitingObservation = errors.New("workflow pull request awaits canonical observation")

// WorkflowPullRequestQuery locates all PR states for one retained source branch.
type WorkflowPullRequestQuery struct {
	Repository string `json:"repository"`
	HeadBranch string `json:"head_branch"`
	BaseBranch string `json:"base_branch"`
}

// WorkflowPullRequestRequest supplies the initial canonical PR content.
type WorkflowPullRequestRequest struct {
	WorkflowPullRequestQuery
	Title string `json:"title"`
	Body  string `json:"body"`
}

// WorkflowPullRequestSnapshot contains provider facts, never worker completion claims.
type WorkflowPullRequestSnapshot struct {
	URL            string `json:"url"`
	Repository     string `json:"repository"`
	HeadRepository string `json:"head_repository"`
	HeadBranch     string `json:"head_branch"`
	BaseBranch     string `json:"base_branch"`
	HeadCommit     string `json:"head_commit"`
	BaseCommit     string `json:"base_commit"`
	State          string `json:"state"`
	Merged         bool   `json:"merged"`
	ResultCommit   string `json:"result_commit"`
}

// WorkflowPullRequestProvider isolates canonical provider observation and PR creation.
// Find includes closed PRs and fork heads; callers must reject ambiguous identities.
// Implementations must respect context deadlines and bound their response buffers.
type WorkflowPullRequestProvider interface {
	FindWorkflowPullRequests(context.Context, WorkflowPullRequestQuery) ([]WorkflowPullRequestSnapshot, error)
	ReadWorkflowPullRequest(context.Context, string) (WorkflowPullRequestSnapshot, error)
	CreateWorkflowPullRequest(context.Context, WorkflowPullRequestRequest) (WorkflowPullRequestSnapshot, error)
	ReadWorkflowBranch(context.Context, string, string) (string, error)
}

// WorkflowSlicePullRequestRequest freezes target branch and initial content.
type WorkflowSlicePullRequestRequest struct {
	BaseBranch string
	Title      string
	Body       string
}

// WorkflowPullRequestBinding retains creation intent across crash and candidate repair.
type WorkflowPullRequestBinding struct {
	Schema           string                     `json:"schema"`
	ID               string                     `json:"id"`
	Unit             WorkflowRef                `json:"unit"`
	Request          WorkflowPullRequestRequest `json:"request"`
	TargetCommit     string                     `json:"target_commit"`
	CandidateID      string                     `json:"candidate_id"`
	CandidateRef     string                     `json:"candidate_ref"`
	CandidateHash    string                     `json:"candidate_hash"`
	SourceCommit     string                     `json:"source_commit"`
	SourceTree       string                     `json:"source_tree"`
	URL              string                     `json:"url"`
	State            string                     `json:"state"`
	Reason           string                     `json:"reason"`
	ObservationCount int                        `json:"observation_count"`
	ObservationRef   string                     `json:"observation_ref"`
	ObservationHash  string                     `json:"observation_hash"`
}

// EnsureWorkflowSlicePullRequest creates or reuses one slice-owned canonical PR.
// Provider IO is bounded and serialized by the common repository mutation lock;
// it holds no worker capacity reservation. Uncertain creation never resubmits.
func (s *WorkflowStore) EnsureWorkflowSlicePullRequest(ctx context.Context, id string, request WorkflowSlicePullRequestRequest, provider WorkflowPullRequestProvider) (result WorkflowPullRequestSnapshot, err error) {
	if provider == nil || !validWorkflowPullRequestContent(request) {
		return result, fmt.Errorf("workflow PR provider, target branch and bounded content are required")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		observation, readErr := owner.prepareWorkflowPullRequest(ctx, &run, request)
		if readErr != nil {
			return readErr
		}

		result, readErr = owner.ensureWorkflowPullRequest(ctx, &run, observation, provider)

		return readErr
	})

	return result, err
}

func validWorkflowPullRequestContent(request WorkflowSlicePullRequestRequest) bool {
	return validWorkflowPRBranch(request.BaseBranch) && strings.TrimSpace(request.Title) != "" && len(request.Title) <= 256 && !strings.ContainsAny(request.Title, "\x00\r\n") && len(request.Body) <= 16*1024
}

func validWorkflowPRBranch(branch string) bool {
	if !validWorkflowPRBranchText(branch) {
		return false
	}

	for _, invalid := range []string{"..", "@{", "//"} {
		if strings.Contains(branch, invalid) {
			return false
		}
	}

	return validWorkflowPRBranchParts(branch)
}

func validWorkflowPRBranchText(branch string) bool {
	return branch != "" && branch != "@" && len(branch) <= 256 && !strings.HasPrefix(branch, "-") && !strings.HasPrefix(branch, "refs/") && !strings.ContainsAny(branch, " ~^:?*[\\\x7f") && !strings.ContainsFunc(branch, func(r rune) bool { return r < 32 })
}

func validWorkflowPRBranchParts(branch string) bool {
	for _, part := range strings.Split(branch, "/") {
		if part == "" || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".") || strings.HasSuffix(part, ".lock") {
			return false
		}
	}

	return true
}
