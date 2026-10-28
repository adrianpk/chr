package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// BeginWorkflowMergerRepair prepares an owned private checkpoint. Conflicts are
// retained at their original source step, without touching canonical refs.
func (git OSWorkflowPreparationGit) BeginWorkflowMergerRepair(ctx context.Context, query PreparationQuery) (MergerRepairState, error) {
	if err := git.validateQuery(query); err != nil {
		return MergerRepairState{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	reader, err := git.newMergerRepairReader(ctx, query)
	if err != nil {
		return MergerRepairState{}, err
	}

	state := MergerRepairState{SchemaVersion: 1, Query: query, Preparation: *reader.preparation, OriginalParent: query.OriginCommit, PreparedParent: query.BaseCommit}
	if err := reader.advanceMergerRepair(ctx, &state); err != nil {
		return state, err
	}

	return state, nil
}

func (git OSWorkflowPreparationGit) newMergerRepairReader(ctx context.Context, query PreparationQuery) (workflowIntegrationReader, error) {
	reader, err := git.preparationReader(query.Lane.Worktree, query)
	if err != nil {
		return reader, err
	}

	parent := filepath.Dir(reader.Directory)
	if resolved, err := filepath.EvalSymlinks(parent); err != nil || resolved != parent {
		return reader, workflowstate.ErrConflict
	}

	if err := os.Mkdir(reader.Directory, 0o700); err != nil {
		return reader, err
	}

	format := "sha1"
	if len(query.SourceCommit) == 64 {
		format = "sha256"
	}

	if _, err := reader.Output(ctx, "init", "--quiet", "--object-format="+format); err != nil {
		return reader, err
	}

	if err := git.importMergerRepairObjects(ctx, reader, query); err != nil {
		return reader, err
	}

	return reader, nil
}

func (git OSWorkflowPreparationGit) importMergerRepairObjects(ctx context.Context, reader workflowIntegrationReader, query PreparationQuery) error {
	for _, fetch := range [][]string{{query.SourceDirectory, query.SourceCommit}, {git.transport.transport, "refs/heads/" + query.TargetBranch + ":refs/heads/target"}} {
		if _, err := reader.Output(ctx, append([]string{"fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--no-write-fetch-head", "--"}, fetch...)...); err != nil {
			return err
		}
	}

	base, err := reader.Output(ctx, "rev-parse", "refs/heads/target^{commit}")
	if err != nil || base != query.BaseCommit {
		return fmt.Errorf("%w: canonical base moved during repair preparation", workflowstate.ErrConflict)
	}

	tree, err := reader.tree(ctx, query.SourceCommit)
	if err != nil || tree != query.SourceTree {
		return workflowstate.ErrConflict
	}

	proof := workflowMergerRepairProof(MergerRepairState{Query: query})
	if err := reader.verifyBase(ctx, proof); err != nil {
		return err
	}

	_, err = reader.Output(ctx, "merge-base", "--is-ancestor", query.OriginCommit, query.SourceCommit)

	return err
}

func workflowMergerRepairProof(state MergerRepairState) IntegrationProof {
	proof := IntegrationProof{IntegrationQuery: state.Query.IntegrationQuery, BaseCommit: state.Query.BaseCommit, ObservedTarget: state.Query.BaseCommit}

	proof.SourceDirectory = ""
	if len(state.Repairs) > 0 {
		proof.Provenance = &IntegrationProvenance{Repairs: state.Repairs}
	}

	return proof
}
