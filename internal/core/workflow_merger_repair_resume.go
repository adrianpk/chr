package core

import (
	"context"
	"slices"
	"strconv"
	"strings"
)

func (git OSWorkflowPreparationGit) resumeMergerRepairReader(ctx context.Context, state WorkflowMergerRepairState) (workflowIntegrationReader, error) {
	query := state.Query

	query.SourceDirectory, query.ResultCommit = query.Lane.Worktree, query.BaseCommit
	if !validWorkflowMergerRepairStateIdentity(state) || git.transport.validateQuery(query.WorkflowIntegrationQuery) != nil {
		return workflowIntegrationReader{}, ErrWorkflowConflict
	}

	reader, err := git.preparationReader(query.Lane.Worktree, query)
	if err != nil || !validWorkflowRepairPreparation(state.Preparation, *reader.preparation) {
		return reader, errorsOrConflict(err)
	}

	reader.preparation = &state.Preparation
	if err := reader.verifyMergerRepairPrefix(ctx, state); err != nil {
		return reader, err
	}

	baseline, err := reader.mergerTree(ctx, state.OriginalParent, state.PreparedParent, state.CurrentSource)
	if err != nil || baseline.tree != state.Tree || !slices.Equal(baseline.conflicts, state.Conflicts) {
		return reader, errorsOrConflict(err)
	}

	return reader, nil
}

func validWorkflowRepairPreparation(p, policy WorkflowIntegrationPreparation) bool {
	if !validWorkflowCommitDate(p.InitialCommitterDate) || !validWorkflowCommitDate(p.CommitterDate) || p.RequiredSignatures != policy.RequiredSignatures || !workflowNativeJSONEqual(p.Signing, policy.Signing) {
		return false
	}

	initial, err := strconv.ParseInt(strings.Fields(p.InitialCommitterDate)[0], 10, 64)
	if err != nil || initial > 1<<63-1-workflowPreparationCollisionLimit || p.NoOpCollisions < 0 || p.NoOpCollisions > workflowPreparationCollisionLimit {
		return false
	}

	return p.CommitterDate == strconv.FormatInt(initial+int64(p.NoOpCollisions), 10)+" +0000"
}

func (reader workflowIntegrationReader) verifyMergerRepairPrefix(ctx context.Context, state WorkflowMergerRepairState) error {
	sources, err := reader.mergerRepairSources(ctx, state)
	if err != nil || !validWorkflowMergerRepairPosition(state, sources) {
		return errorsOrConflict(err)
	}

	original, parent := state.Query.OriginCommit, state.Query.BaseCommit

	proof := workflowMergerRepairProof(state)
	for i, entry := range state.History {
		if entry.Source != sources[i] || entry.Parent != parent {
			return ErrWorkflowConflict
		}

		tree, err := reader.expectedIntegrationTree(ctx, proof, original, parent, entry.Source)
		if err != nil || tree != entry.Tree || reader.verifyMergerReplayObject(ctx, entry, state.Preparation) != nil {
			return errorsOrConflict(err)
		}

		original, parent = entry.Source, entry.Commit
	}

	if !workflowMergerRepairEndMatches(state, original, parent) {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowMergerRepairStateIdentity(state WorkflowMergerRepairState) bool {
	return state.SchemaVersion == 1 && state.Result == nil && validWorkflowLane(state.Query.Lane) && slices.Contains([]string{"merge", "rebase", "squash"}, state.Query.Method) && state.Query.Lane.BaseCommit == state.Query.BaseCommit
}

func validWorkflowMergerRepairPosition(state WorkflowMergerRepairState, sources []string) bool {
	return state.ReplayIndex >= 0 && state.ReplayIndex < len(sources) && len(state.History) == state.ReplayIndex && state.CurrentSource == sources[state.ReplayIndex]
}

func workflowMergerRepairEndMatches(state WorkflowMergerRepairState, original, parent string) bool {
	return original == state.OriginalParent && parent == state.PreparedParent
}

func (reader workflowIntegrationReader) verifyMergerReplayObject(ctx context.Context, entry WorkflowIntegrationCommit, p WorkflowIntegrationPreparation) error {
	source, err := reader.commitAttribution(ctx, entry.Source)
	if err != nil {
		return err
	}

	prepared, err := reader.commitAttribution(ctx, entry.Commit)
	if err != nil || prepared.Tree != entry.Tree || !slices.Equal(prepared.Parents, []string{entry.Parent}) || !workflowReplayAttributionMatches(source, prepared) || !workflowPreparationCommitterMatches(prepared, p, true) {
		return errorsOrConflict(err)
	}

	if p.RequiredSignatures {
		_, err = reader.verifyPreparationSignature(ctx, entry.Commit, p.Signing)
	}

	return err
}
