package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrWorkflowAwaitingObservation retains uncertain PR creation without retrying it.
var ErrWorkflowAwaitingObservation = errors.New("workflow pull request awaits canonical observation")

// PullRequestQuery locates all PR states for one retained source branch.
type PullRequestQuery struct {
	Repository string `json:"repository"`
	HeadBranch string `json:"head_branch"`
	BaseBranch string `json:"base_branch"`
}

// PullRequestRequest supplies the initial canonical PR content.
type PullRequestRequest struct {
	PullRequestQuery
	Title string `json:"title"`
	Body  string `json:"body"`
}

// PullRequestProvider isolates canonical provider observation and PR creation.
// Find includes closed PRs and fork heads; callers must reject ambiguous identities.
// Implementations must respect context deadlines and bound their response buffers.
type PullRequestProvider interface {
	FindWorkflowPullRequests(context.Context, PullRequestQuery) ([]workflowstate.PullRequestSnapshot, error)
	ReadWorkflowPullRequest(context.Context, string) (workflowstate.PullRequestSnapshot, error)
	CreateWorkflowPullRequest(context.Context, PullRequestRequest) (workflowstate.PullRequestSnapshot, error)
	ReadWorkflowBranch(context.Context, string, string) (string, error)
}

// SlicePullRequestRequest freezes target branch and initial content.
type SlicePullRequestRequest struct {
	BaseBranch string
	Title      string
	Body       string
}

// PullRequestBinding retains creation intent across crash and candidate repair.
type PullRequestBinding struct {
	Schema           string             `json:"schema"`
	ID               string             `json:"id"`
	Unit             workflowstate.Ref  `json:"unit"`
	Request          PullRequestRequest `json:"request"`
	TargetCommit     string             `json:"target_commit"`
	CandidateID      string             `json:"candidate_id"`
	CandidateRef     string             `json:"candidate_ref"`
	CandidateHash    string             `json:"candidate_hash"`
	SourceCommit     string             `json:"source_commit"`
	SourceTree       string             `json:"source_tree"`
	URL              string             `json:"url"`
	State            string             `json:"state"`
	Reason           string             `json:"reason"`
	ObservationCount int                `json:"observation_count"`
	ObservationRef   string             `json:"observation_ref"`
	ObservationHash  string             `json:"observation_hash"`
}

// EnsureWorkflowSlicePullRequest creates or reuses one unit-owned canonical PR.
// Provider IO is bounded and serialized by the common repository mutation lock;
// it holds no worker capacity reservation. Uncertain creation never resubmits.
func (s *Store) EnsureWorkflowSlicePullRequest(ctx context.Context, id string, request SlicePullRequestRequest, provider PullRequestProvider) (result workflowstate.PullRequestSnapshot, err error) {
	if provider == nil || !validWorkflowPullRequestContent(request) {
		return result, fmt.Errorf("workflow PR provider, target branch and bounded content are required")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		observation, readErr := invocationLocation.Owner.prepareWorkflowPullRequest(ctx, &invocationLocation.Run, request)
		if readErr != nil {
			return readErr
		}

		result, readErr = invocationLocation.Owner.ensureWorkflowPullRequest(ctx, &invocationLocation.Run, observation, provider)

		return readErr
	})

	return result, err
}

func validWorkflowPullRequestContent(request SlicePullRequestRequest) bool {
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
