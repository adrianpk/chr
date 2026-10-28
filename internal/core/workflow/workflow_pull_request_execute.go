package workflow

import (
	"context"
	"errors"
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// workflowPullRequestHost sequences provider IO inside the Store-owned lock.
// Intent, current-candidate checks and complete observation writes stay in Store.
type workflowPullRequestHost struct{ store *Store }

func (host workflowPullRequestHost) ensure(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, provider PullRequestProvider) (workflowstate.PullRequestSnapshot, error) {
	if run.PullRequest.URL != "" {
		pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
		if err != nil {
			return pr, err
		}

		return pr, host.store.acceptWorkflowPullRequest(ctx, run, observation, []workflowstate.PullRequestSnapshot{pr})
	}

	found, err := provider.FindWorkflowPullRequests(ctx, run.PullRequest.Request.PullRequestQuery)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	if len(found) > 0 {
		err := host.store.acceptWorkflowPullRequest(ctx, run, observation, found)

		return found[0], err
	}

	if run.PullRequest.State != "prepared" {
		return workflowstate.PullRequestSnapshot{}, host.store.awaitWorkflowPullRequest(run, nil)
	}

	return host.create(ctx, run, observation, provider)
}

func (host workflowPullRequestHost) create(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, provider PullRequestProvider) (workflowstate.PullRequestSnapshot, error) {
	binding := run.PullRequest

	commit, err := provider.ReadWorkflowBranch(ctx, binding.Request.Repository, binding.Request.HeadBranch)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	if commit != binding.SourceCommit {
		return workflowstate.PullRequestSnapshot{}, fmt.Errorf("%w: canonical source branch differs from validated candidate", workflowstate.ErrConflict)
	}

	if err := host.store.beginWorkflowPullRequestCreation(ctx, run, observation); err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	pr, err := provider.CreateWorkflowPullRequest(ctx, binding.Request)
	if err != nil {
		return pr, errors.Join(err, host.store.awaitWorkflowPullRequest(run, []workflowstate.PullRequestSnapshot{pr}))
	}

	return pr, host.store.acceptWorkflowPullRequest(ctx, run, observation, []workflowstate.PullRequestSnapshot{pr})
}
