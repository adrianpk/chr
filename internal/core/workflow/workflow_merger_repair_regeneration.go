package workflow

import (
	"context"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func importWorkflowMergerRepairTrees(ctx context.Context, git MergerRepairGit, state MergerRepairState, directory string, repair workflowstate.IntegrationTreeRepair) error {
	trees := []string{repair.ProposalTree, repair.ResolvedTree}
	for _, evidence := range repair.Regeneration {
		trees = append(trees, evidence.InputTree, evidence.OutputTree)
	}

	slices.Sort(trees)

	for _, tree := range slices.Compact(trees) {
		if err := git.ImportWorkflowMergerTree(ctx, state, directory, tree); err != nil {
			return err
		}
	}

	return nil
}

func (reader workflowIntegrationReader) verifyMergerRegenerationTrees(ctx context.Context, repair workflowstate.IntegrationTreeRepair) error {
	if repair.ProposalTree != "" {
		if err := reader.checkMergerObservedPaths(ctx, repair.Input, repair.Proposal, workflowMergerPathCheck{Before: repair.BaselineTree, After: repair.ProposalTree, Mode: workflowMergerDeclaredPaths}); err != nil {
			return err
		}

		if err := reader.checkMergerGeneratedEdits(ctx, repair.Input, repair.BaselineTree, repair.ProposalTree); err != nil {
			return err
		}
	}

	for _, evidence := range repair.Regeneration {
		if err := reader.checkMergerObservedPaths(ctx, repair.Input, repair.Proposal, workflowMergerPathCheck{Before: evidence.InputTree, After: evidence.OutputTree, Mode: workflowMergerDeclaredPaths}); err != nil {
			return err
		}
	}

	return nil
}
