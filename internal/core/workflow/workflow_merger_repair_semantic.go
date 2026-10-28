package workflow

import (
	"context"
	"slices"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// BeginWorkflowMergerSemanticRepair rewinds an existing verified rebase to the
// owning original step. The unchanged prefix stays exact and the suffix replays.
func (git OSWorkflowPreparationGit) BeginWorkflowMergerSemanticRepair(ctx context.Context, query workflowstate.PreparationQuery, proof workflowstate.IntegrationProof, original string) (MergerRepairState, error) {
	if git.validateQuery(query) != nil || !validWorkflowMergerSemanticProof(query, proof) {
		return MergerRepairState{}, workflowstate.ErrConflict
	}

	expected := query.IntegrationQuery

	expected.SourceDirectory, expected.ResultCommit = "", proof.ResultCommit
	if workflowintegrate.ValidateWorkflowIntegrationProof(expected, proof, workflowIntegrationRepairs{}) != nil || proof.BaseCommit != query.BaseCommit {
		return MergerRepairState{}, workflowstate.ErrConflict
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	reader, err := git.newMergerRepairReader(ctx, query)
	if err != nil {
		return MergerRepairState{}, err
	}

	if _, err := reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-write-fetch-head", "--", query.SourceDirectory, proof.ResultCommit); err != nil {
		return MergerRepairState{}, err
	}

	if !validWorkflowRepairPreparation(*proof.Provenance.Preparation, *reader.preparation) {
		return MergerRepairState{}, workflowstate.ErrConflict
	}

	reader.preparation = proof.Provenance.Preparation
	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return MergerRepairState{}, err
	}

	return reader.semanticMergerCheckpoint(ctx, query, proof, original)
}

func validWorkflowMergerSemanticProof(query workflowstate.PreparationQuery, proof workflowstate.IntegrationProof) bool {
	return query.Method == "rebase" && proof.Provenance != nil && proof.Provenance.Preparation != nil && len(proof.Provenance.IntegrationFixes) == 0
}

func (reader workflowIntegrationReader) semanticMergerCheckpoint(ctx context.Context, query workflowstate.PreparationQuery, proof workflowstate.IntegrationProof, original string) (MergerRepairState, error) {
	index := slices.IndexFunc(proof.History, func(entry workflowstate.IntegrationCommit) bool { return entry.Source == original })
	if index < 0 {
		return MergerRepairState{}, workflowstate.ErrConflict
	}

	parent := query.OriginCommit
	if index > 0 {
		parent = proof.History[index-1].Source
	}

	state := MergerRepairState{SchemaVersion: 1, Query: query, Preparation: *reader.preparation, ReplayIndex: index, OriginalParent: parent, PreparedParent: proof.History[index].Parent, CurrentSource: original, History: slices.Clone(proof.History[:index])}
	for _, repair := range proof.Provenance.Repairs {
		if slices.ContainsFunc(state.History, func(entry workflowstate.IntegrationCommit) bool { return entry.Source == repair.Source }) {
			state.Repairs = append(state.Repairs, repair)
		}
	}

	baseline, err := reader.mergerTree(ctx, parent, state.PreparedParent, original)
	if err != nil {
		return state, err
	}

	state.Tree, state.Conflicts = baseline.tree, baseline.conflicts

	return state, reader.materializeMergerTree(ctx, state.PreparedParent, state.Tree)
}

// PrepareWorkflowMergerIntegrationFix creates an explicitly attributed host fix
// on a verified private result. Its finding is retained in method provenance.
func (git OSWorkflowPreparationGit) PrepareWorkflowMergerIntegrationFix(ctx context.Context, state MergerRepairState, tree string, finding workflowstate.ReviewFinding) (MergerRepairState, error) {
	if state.Result == nil || state.SchemaVersion != 1 || state.Query.Repository != git.transport.repository {
		return state, workflowstate.ErrConflict
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: state.Query.Lane.Worktree}, signing: git.signing}

	proof, err := reader.prepareRebaseFix(ctx, *state.Result, tree, finding)
	if err != nil {
		return state, err
	}

	if err := reader.materializeMergerTree(ctx, proof.ResultCommit, tree); err != nil {
		return state, err
	}

	state.Result, state.Tree, state.PreparedParent = &proof, tree, proof.ResultCommit

	return state, nil
}
