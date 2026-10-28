package workflow

import (
	"context"
	"path/filepath"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PrepareWorkflowRepairedIntegration independently verifies retained method
// history and imports exactly the host-accepted P into a fresh gate checkout.
func (git OSWorkflowPreparationGit) PrepareWorkflowRepairedIntegration(ctx context.Context, query PreparationQuery, state MergerRepairState) (workflowstate.IntegrationProof, error) {
	if !validWorkflowRepairedPreparationQuery(git, query, state) {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
	}

	proof := *state.Result
	if state.Query.SourceCommit != query.SourceCommit || state.Query.OriginCommit != query.OriginCommit || state.Query.SourceTree != query.SourceTree || checkWorkflowPreparedProof(query, proof) != nil {
		return proof, workflowstate.ErrConflict
	}

	if err := checkWorkflowMergerPrivateRepository(ctx, state.Query.Lane.Worktree); err != nil {
		return proof, err
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: state.Query.Lane.Worktree, Environment: workflowMergerEnvironment(filepath.Dir(state.Query.Lane.Worktree))}, signing: git.signing}
	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return proof, err
	}

	if err := checkWorkflowRepairedCheckout(ctx, reader, proof); err != nil {
		return proof, err
	}

	return proof, attachWorkflowPreparation(ctx, reader, query, proof)
}

func checkWorkflowRepairedCheckout(ctx context.Context, reader workflowIntegrationReader, proof workflowstate.IntegrationProof) error {
	head, err := reader.Output(ctx, "rev-parse", "HEAD")
	if err != nil || head != proof.ResultCommit {
		return errorsOrConflict(err)
	}

	status, err := reader.Output(ctx, "status", "--porcelain", "--untracked-files=all")
	if err != nil || status != "" {
		return errorsOrConflict(err)
	}

	tree, err := reader.Output(ctx, "write-tree")
	if err != nil || tree != proof.PreparedTree {
		return errorsOrConflict(err)
	}

	return nil
}

func validWorkflowRepairedPreparationQuery(git OSWorkflowPreparationGit, query PreparationQuery, state MergerRepairState) bool {
	return state.Result != nil && git.validateQuery(query) == nil && state.Query.Lane.Unit == query.Lane.Unit
}
