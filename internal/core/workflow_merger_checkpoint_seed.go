package core

import (
	"context"
	"path/filepath"
)

func (s *WorkflowStore) seedWorkflowMergerCheckpoint(ctx context.Context, input WorkflowMergerInput, lane WorkflowLane) (WorkflowLane, error) {
	state, err := s.readWorkflowMergerCheckpoint(input)
	if err != nil {
		return lane, err
	}

	reader := workflowIntegrationReader{dir: lane.Worktree, environment: workflowMergerEnvironment(filepath.Dir(lane.Worktree))}
	// Rebase and squash P need not reach the original source objects. Host
	// verification still needs both histories for semantic and side checks.
	for _, object := range []string{state.PreparedParent, state.Tree, state.Query.SourceCommit, state.Query.BaseCommit} {
		if _, err := reader.output(ctx, "fetch", "--quiet", "--no-tags", "--no-write-fetch-head", "--", state.Query.Lane.Worktree, object); err != nil {
			return lane, err
		}
	}

	if err := reader.materializeMergerTree(ctx, state.PreparedParent, state.Tree); err != nil {
		return lane, err
	}

	lane.BaseCommit, lane.HeadCommit = state.PreparedParent, state.PreparedParent

	return lane, nil
}
