package workflow

import (
	"context"
	"path/filepath"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (host workflowMergerHost) seedCheckpoint(ctx context.Context, input workflowstate.MergerInput, lane workflowstate.Lane) (workflowstate.Lane, error) {
	state, err := host.store.readWorkflowMergerCheckpoint(input)
	if err != nil {
		return lane, err
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: lane.Worktree, Environment: workflowMergerEnvironment(filepath.Dir(lane.Worktree))}}
	// Rebase and squash P need not reach the original source objects. Host
	// verification still needs both histories for semantic and side checks.
	for _, object := range []string{state.PreparedParent, state.Tree, state.Query.SourceCommit, state.Query.BaseCommit} {
		if _, err := reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-write-fetch-head", "--", state.Query.Lane.Worktree, object); err != nil {
			return lane, err
		}
	}

	if err := reader.materializeMergerTree(ctx, state.PreparedParent, state.Tree); err != nil {
		return lane, err
	}

	lane.BaseCommit, lane.HeadCommit = state.PreparedParent, state.PreparedParent

	return lane, nil
}
