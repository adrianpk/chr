package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrWorkflowAwaitingObservation retains uncertain PR creation without retrying it.
var ErrWorkflowAwaitingObservation = errors.New("workflow pull request awaits canonical observation")

// PullRequestProvider isolates canonical provider observation and PR creation.
// Find includes closed PRs and fork heads; callers must reject ambiguous identities.
// Implementations must respect context deadlines and bound their response buffers.
type PullRequestProvider interface {
	FindWorkflowPullRequests(context.Context, workflowstate.PullRequestQuery) ([]workflowstate.PullRequestSnapshot, error)
	ReadWorkflowPullRequest(context.Context, string) (workflowstate.PullRequestSnapshot, error)
	CreateWorkflowPullRequest(context.Context, workflowstate.PullRequestRequest) (workflowstate.PullRequestSnapshot, error)
	ReadWorkflowBranch(context.Context, string, string) (string, error)
}

// PullRequestBinding retains creation intent across crash and candidate repair.
type PullRequestBinding struct {
	Schema           string                           `json:"schema"`
	ID               string                           `json:"id"`
	Unit             workflowstate.Ref                `json:"unit"`
	Request          workflowstate.PullRequestRequest `json:"request"`
	TargetCommit     string                           `json:"target_commit"`
	CandidateID      string                           `json:"candidate_id"`
	CandidateRef     string                           `json:"candidate_ref"`
	CandidateHash    string                           `json:"candidate_hash"`
	SourceCommit     string                           `json:"source_commit"`
	SourceTree       string                           `json:"source_tree"`
	URL              string                           `json:"url"`
	State            string                           `json:"state"`
	Reason           string                           `json:"reason"`
	ObservationCount int                              `json:"observation_count"`
	ObservationRef   string                           `json:"observation_ref"`
	ObservationHash  string                           `json:"observation_hash"`
}

// EnsureWorkflowSlicePullRequest creates or reuses one unit-owned canonical PR.
// Provider IO is bounded and serialized by the common repository mutation lock;
// it holds no worker capacity reservation. Uncertain creation never resubmits.
func (s *Store) EnsureWorkflowSlicePullRequest(ctx context.Context, id string, request workflowstate.PullRequestContent, provider PullRequestProvider) (result workflowstate.PullRequestSnapshot, err error) {
	if provider == nil || !workflowforge.ValidPullRequestContent(request) {
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
