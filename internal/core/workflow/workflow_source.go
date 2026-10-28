package workflow

import (
	"context"
	"fmt"
	"time"
)

// SourceQuery identifies only an owned validated source branch. It has
// no target ref; PreviousCommit permits repair of a retained PR source only.
type SourceQuery struct {
	Repository     string
	Directory      string
	Branch         string
	Commit         string
	Tree           string
	PreviousCommit string
}

// SourceGit publishes an exact candidate using trusted host transport.
type SourceGit interface {
	PublishWorkflowSource(context.Context, SourceQuery) error
}

// PublishWorkflowSource checks immutable evidence before and after source IO.
// Crash recovery reuses the retained candidate and observes the remote branch;
// source publication never completes a ticket or publishes an integration ref.
func (s *Store) PublishWorkflowSource(ctx context.Context, id string, publisher SourceGit) error {
	if publisher == nil {
		return fmt.Errorf("workflow source publisher is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	return s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		if !workflowSourcePublishable(run) {
			return ErrWorkflowOwned
		}

		observation, err := owner.observeWorkflowSlice(ctx, run)
		if err != nil {
			return err
		}

		candidate, err := owner.readWorkflowSliceCandidate(ctx, run, observation)
		if err != nil {
			return err
		}

		query := SourceQuery{Repository: run.ExecutionSelection().Target.Repository, Directory: candidate.Lane.Worktree, Branch: candidate.Lane.Branch, Commit: candidate.Result.Commit, Tree: candidate.Result.Tree}
		if run.PullRequest != nil {
			query.PreviousCommit = run.PullRequest.SourceCommit
		}

		if err := publisher.PublishWorkflowSource(ctx, query); err != nil {
			return err
		}

		_, err = owner.readWorkflowSliceCandidate(ctx, run, observation)

		return err
	})
}

func workflowSourcePublishable(run workflowInvocationFile) bool {
	if run.StopIntent != "" || run.Status == "paused" || workflowInvocationTerminal(run.Status) {
		return false
	}

	return run.PullRequest == nil || run.PullRequest.State == "prepared" || run.PullRequest.State == "bound"
}
