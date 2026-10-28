package workflow

import (
	"context"
	"slices"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// BeginWorkflowMergerCombinedRepair freezes a clean merge/squash baseline for
// semantic repair. Rebase repair must name its original replay step instead.
func (git OSWorkflowPreparationGit) BeginWorkflowMergerCombinedRepair(ctx context.Context, query workflowstate.PreparationQuery, proof workflowstate.IntegrationProof) (MergerRepairState, error) {
	if !validWorkflowCombinedRepairQuery(git, query, proof) {
		return MergerRepairState{}, workflowstate.ErrConflict
	}

	reader, err := git.newMergerRepairReader(ctx, query)
	if err != nil {
		return MergerRepairState{}, err
	}

	if _, err := reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-write-fetch-head", "--", query.SourceDirectory, proof.ResultCommit); err != nil {
		return MergerRepairState{}, err
	}

	reader.preparation = proof.Provenance.Preparation
	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return MergerRepairState{}, err
	}

	state := MergerRepairState{SchemaVersion: 1, Query: query, Preparation: *reader.preparation, OriginalParent: query.OriginCommit, PreparedParent: query.BaseCommit, CurrentSource: query.SourceCommit}

	baseline, err := reader.mergerTree(ctx, state.OriginalParent, state.PreparedParent, state.CurrentSource)
	if err != nil || len(baseline.conflicts) != 0 {
		return state, errorsOrConflict(err)
	}

	state.Tree = baseline.tree

	return state, reader.materializeMergerTree(ctx, state.PreparedParent, state.Tree)
}

func validWorkflowCombinedRepairQuery(git OSWorkflowPreparationGit, query workflowstate.PreparationQuery, proof workflowstate.IntegrationProof) bool {
	return slices.Contains([]string{"merge", "squash"}, query.Method) && git.validateQuery(query) == nil && workflowintegrate.CheckWorkflowPreparedProof(query, proof, workflowIntegrationRepairs{}) == nil && proof.Provenance != nil && proof.Provenance.Preparation != nil
}
