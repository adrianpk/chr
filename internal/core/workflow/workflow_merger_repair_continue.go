package workflow

import (
	"context"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ContinueWorkflowMergerRepair consumes a host-accepted repair and resumes its
// original replay step. It cannot run validation or grant candidate approval.
func (git OSWorkflowPreparationGit) ContinueWorkflowMergerRepair(ctx context.Context, state MergerRepairState, repair workflowstate.IntegrationTreeRepair) (MergerRepairState, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	reader, err := git.resumeMergerRepairReader(ctx, state)
	if err != nil {
		return state, err
	}

	proof := workflowMergerRepairProof(state)
	if len(state.Repairs) >= workflowstate.PageLimit || validateWorkflowTreeRepair(proof, repair) != nil {
		return state, workflowstate.ErrConflict
	}

	if err := reader.verifyMergerTreeRepair(ctx, repair, state.OriginalParent, state.PreparedParent, state.CurrentSource, state.Tree, state.Conflicts); err != nil {
		return state, err
	}

	if err := checkWorkflowRepairPredecessors(state, repair.Input); err != nil {
		return state, err
	}

	state.History = slices.Clone(state.History)
	state.Repairs = append(slices.Clone(state.Repairs), repair)

	if _, err := reader.Output(ctx, "read-tree", "--reset", "-u", repair.ResolvedTree); err != nil {
		return state, err
	}

	state.Tree, state.Conflicts = repair.ResolvedTree, nil
	if err := reader.commitMergerRepairStep(ctx, &state); err != nil {
		return state, err
	}

	err = reader.advanceMergerRepair(ctx, &state)

	return state, err
}

func checkWorkflowRepairPredecessors(state MergerRepairState, input workflowstate.MergerInput) error {
	if input.Method != "rebase" {
		return nil
	}

	checkpoint := input.HistoryCheckpoint
	if checkpoint == nil || checkpoint.ReplayIndex != state.ReplayIndex || len(checkpoint.Predecessors) != len(state.History) {
		return workflowstate.ErrConflict
	}

	for i, entry := range state.History {
		if checkpoint.Predecessors[i].Source != entry.Source || checkpoint.Predecessors[i].Commit != entry.Commit {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
