package core

import (
	"context"
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) advanceMergerRepair(ctx context.Context, state *WorkflowMergerRepairState) error {
	sources, err := reader.mergerRepairSources(ctx, *state)
	if err != nil || state.ReplayIndex < 0 || state.ReplayIndex > len(sources) {
		return errorsOrConflict(err)
	}

	for state.ReplayIndex < len(sources) {
		state.CurrentSource = sources[state.ReplayIndex]

		baseline, err := reader.mergerTree(ctx, state.OriginalParent, state.PreparedParent, state.CurrentSource)
		if err != nil {
			return err
		}

		state.Tree, state.Conflicts = baseline.tree, baseline.conflicts
		if len(state.Conflicts) > 0 {
			state.Preparation = *reader.preparation

			return reader.materializeMergerTree(ctx, state.PreparedParent, state.Tree)
		}

		if err := reader.commitMergerRepairStep(ctx, state); err != nil {
			return err
		}
	}

	return reader.finishMergerRepair(ctx, state)
}

func (reader workflowIntegrationReader) mergerRepairSources(ctx context.Context, state WorkflowMergerRepairState) ([]string, error) {
	if state.Query.Method != "rebase" {
		return []string{state.Query.SourceCommit}, nil
	}

	return reader.linearSource(ctx, workflowMergerRepairProof(state))
}

func (reader workflowIntegrationReader) commitMergerRepairStep(ctx context.Context, state *WorkflowMergerRepairState) error {
	var (
		commit string
		err    error
	)

	if state.Query.Method == "rebase" {
		commit, err = reader.replayCommit(ctx, state.Tree, state.PreparedParent, state.CurrentSource)
	} else {
		parents := []string{state.Query.BaseCommit}
		if state.Query.Method == "merge" {
			parents = append(parents, state.Query.SourceCommit)
		}

		commit, err = reader.preparationCommit(ctx, state.Tree, parents)
	}

	if err != nil {
		return err
	}

	state.History = append(state.History, WorkflowIntegrationCommit{Source: state.CurrentSource, Commit: commit, Tree: state.Tree, Parent: state.PreparedParent})
	state.OriginalParent, state.PreparedParent = state.CurrentSource, commit
	state.ReplayIndex++

	return nil
}

func (reader workflowIntegrationReader) finishMergerRepair(ctx context.Context, state *WorkflowMergerRepairState) error {
	baseTree, err := reader.tree(ctx, state.Query.BaseCommit)
	if err != nil {
		return err
	}

	if state.Tree == baseTree || state.PreparedParent == state.Query.SourceCommit {
		return fmt.Errorf("%w: empty result or source identity collision requires operator resolution", workflowstate.ErrConflict)
	}

	proof := workflowMergerRepairProof(*state)
	proof.ResultCommit, proof.PreparedTree, proof.VerifiedTree = state.PreparedParent, state.Tree, state.Tree
	proof.History = state.History

	proof.Provenance, err = reader.integrationProvenance(ctx, proof)
	if err != nil {
		return err
	}

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil || validateWorkflowIntegrationProof(proof.WorkflowIntegrationQuery, proof) != nil {
		return errorsOrConflict(err)
	}

	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return err
	}

	if err := reader.materializeMergerTree(ctx, proof.ResultCommit, proof.PreparedTree); err != nil {
		return err
	}

	state.Preparation, state.Conflicts, state.Result = *reader.preparation, nil, &proof

	return nil
}
