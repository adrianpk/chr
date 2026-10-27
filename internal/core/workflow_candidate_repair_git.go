package core

import (
	"context"
	"path/filepath"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PrepareWorkflowRepairedIntegration independently verifies retained method
// history and imports exactly the host-accepted P into a fresh gate checkout.
func (git OSWorkflowPreparationGit) PrepareWorkflowRepairedIntegration(ctx context.Context, query WorkflowPreparationQuery, state WorkflowMergerRepairState) (WorkflowIntegrationProof, error) {
	if !validWorkflowRepairedPreparationQuery(git, query, state) {
		return WorkflowIntegrationProof{}, workflowstate.ErrConflict
	}

	proof := *state.Result
	if state.Query.SourceCommit != query.SourceCommit || state.Query.OriginCommit != query.OriginCommit || state.Query.SourceTree != query.SourceTree || checkWorkflowPreparedProof(query, proof) != nil {
		return proof, workflowstate.ErrConflict
	}

	if err := checkWorkflowMergerPrivateRepository(ctx, state.Query.Lane.Worktree); err != nil {
		return proof, err
	}

	reader := workflowIntegrationReader{dir: state.Query.Lane.Worktree, signing: git.signing, environment: workflowMergerEnvironment(filepath.Dir(state.Query.Lane.Worktree))}
	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return proof, err
	}

	if err := checkWorkflowRepairedCheckout(ctx, reader, proof); err != nil {
		return proof, err
	}

	return proof, attachWorkflowPreparation(ctx, reader, query, proof)
}

func checkWorkflowRepairedCheckout(ctx context.Context, reader workflowIntegrationReader, proof WorkflowIntegrationProof) error {
	head, err := reader.output(ctx, "rev-parse", "HEAD")
	if err != nil || head != proof.ResultCommit {
		return errorsOrConflict(err)
	}

	status, err := reader.output(ctx, "status", "--porcelain", "--untracked-files=all")
	if err != nil || status != "" {
		return errorsOrConflict(err)
	}

	tree, err := reader.output(ctx, "write-tree")
	if err != nil || tree != proof.PreparedTree {
		return errorsOrConflict(err)
	}

	return nil
}

func validWorkflowRepairedPreparationQuery(git OSWorkflowPreparationGit, query WorkflowPreparationQuery, state WorkflowMergerRepairState) bool {
	return state.Result != nil && git.validateQuery(query) == nil && state.Query.Lane.Unit == query.Lane.Unit
}
