package core

import (
	"context"
	"slices"
)

func importWorkflowMergerRepairTrees(ctx context.Context, git WorkflowMergerRepairGit, state WorkflowMergerRepairState, directory string, repair WorkflowIntegrationTreeRepair) error {
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

func (reader workflowIntegrationReader) verifyMergerRegenerationTrees(ctx context.Context, repair WorkflowIntegrationTreeRepair) error {
	if repair.ProposalTree != "" {
		if err := reader.checkMergerObservedPaths(ctx, repair.Input, repair.Proposal, repair.BaselineTree, repair.ProposalTree, false); err != nil {
			return err
		}

		if err := reader.checkMergerGeneratedEdits(ctx, repair.Input, repair.BaselineTree, repair.ProposalTree); err != nil {
			return err
		}
	}

	for _, evidence := range repair.Regeneration {
		if err := reader.checkMergerObservedPaths(ctx, repair.Input, repair.Proposal, evidence.InputTree, evidence.OutputTree, false); err != nil {
			return err
		}
	}

	return nil
}
