package core

import (
	"context"
	"os"
	"path/filepath"
)

// CheckWorkflowMergerCheckpoint verifies actual Git objects, the replay prefix
// and private ownership independently of the retained checkpoint's assertions.
func (git OSWorkflowPreparationGit) CheckWorkflowMergerCheckpoint(ctx context.Context, state WorkflowMergerRepairState) error {
	if err := checkWorkflowMergerPrivateRepository(ctx, state.Query.Lane.Worktree); err != nil {
		return err
	}

	reader, err := git.resumeMergerRepairReader(ctx, state)
	if err != nil {
		return err
	}

	head, err := reader.output(ctx, "rev-parse", "HEAD")
	if err != nil || head != state.PreparedParent {
		return ErrWorkflowConflict
	}

	return nil
}

func checkWorkflowMergerPrivateRepository(ctx context.Context, directory string) error {
	if !filepath.IsAbs(directory) {
		return ErrWorkflowConflict
	}

	resolved, err := filepath.EvalSymlinks(directory)
	if err != nil || resolved != directory {
		return ErrWorkflowConflict
	}

	info, err := os.Lstat(filepath.Join(directory, ".git"))
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ErrWorkflowConflict
	}

	reader := workflowIntegrationReader{dir: directory, environment: workflowMergerEnvironment(filepath.Dir(directory))}

	remotes, err := reader.output(ctx, "remote")
	if err != nil || remotes != "" {
		return ErrWorkflowConflict
	}

	return nil
}

// ImportWorkflowMergerTree imports the host snapshot into the private repair
// object store. It does not change a branch or publish a temporary commit.
func (git OSWorkflowPreparationGit) ImportWorkflowMergerTree(ctx context.Context, state WorkflowMergerRepairState, directory, tree string) error {
	if !workflowCommitID(tree) || directory == state.Query.Lane.Worktree {
		return ErrWorkflowConflict
	}

	reader, err := git.resumeMergerRepairReader(ctx, state)
	if err != nil {
		return err
	}

	_, err = reader.output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--no-write-fetch-head", "--", directory, tree)

	return err
}
