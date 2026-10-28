package workflow

import (
	"context"
	"path/filepath"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) observeWorkflowMergerRepair(ctx context.Context, claim *Claim, prepared workflowPreparedTask, provider PullRequestProvider, state MergerRepairState, attempt *MergerAttempt) (workflowstate.IntegrationTreeRepair, error) {
	input, proposal := attempt.Input, *attempt.Proposal
	repair := workflowstate.IntegrationTreeRepair{Input: input, Proposal: proposal, OriginalParent: state.OriginalParent, PreparedParent: state.PreparedParent, Source: state.CurrentSource, BaselineTree: state.Tree}
	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: prepared.input.Lane.Worktree, Environment: workflowMergerEnvironment(filepath.Dir(prepared.input.Lane.Worktree))}}

	tree, err := workflowMergerSnapshot(ctx, reader, state)
	if err != nil {
		return repair, err
	}

	repair.ProposalTree, repair.ResolvedTree = tree, tree
	if err := reader.checkMergerObservedPaths(ctx, input, proposal, state.Tree, tree, false); err != nil {
		return repair, err
	}

	if err := reader.checkMergerGeneratedEdits(ctx, input, state.Tree, tree); err != nil {
		return repair, err
	}

	if err := reader.checkMergerSideSelection(ctx, state, input, tree); err != nil {
		return repair, err
	}

	evidence, err := s.regenerateWorkflowMerger(ctx, claim, prepared, provider, reader, state, repair)

	repair.Regeneration = evidence
	if err != nil {
		return repair, err
	}

	if len(evidence) > 0 {
		repair.ResolvedTree = evidence[len(evidence)-1].OutputTree
	}

	repair.CompletedAt = time.Now().UTC()
	if err := reader.checkMergerObservedPaths(ctx, input, proposal, state.Tree, repair.ResolvedTree, true); err != nil {
		return repair, err
	}

	if err := reader.verifyMergerTreeRepair(ctx, repair, state.OriginalParent, state.PreparedParent, state.CurrentSource, state.Tree, state.Conflicts); err != nil {
		return repair, err
	}

	return repair, validateWorkflowTreeRepair(workflowMergerRepairProof(state), repair)
}

func (reader workflowIntegrationReader) checkMergerGeneratedEdits(ctx context.Context, input workflowstate.MergerInput, baseline, proposal string) error {
	for _, conflict := range input.Conflicts {
		if !workflowmerge.WorkflowMergerProcedureRequired(conflict) {
			continue
		}

		before, err := reader.mergerPathObject(ctx, baseline, conflict.Path)
		if err != nil {
			return err
		}

		after, err := reader.mergerPathObject(ctx, proposal, conflict.Path)
		if err != nil || before != after {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
