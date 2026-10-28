package workflow

import (
	"context"
	"os"
	"path/filepath"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckWorkflowMergerCheckpoint verifies actual Git objects, the replay prefix
// and private ownership independently of the retained checkpoint's assertions.
func (git OSWorkflowPreparationGit) CheckWorkflowMergerCheckpoint(ctx context.Context, state MergerRepairState) error {
	if err := checkWorkflowMergerPrivateRepository(ctx, state.Query.Lane.Worktree); err != nil {
		return err
	}

	reader, err := git.resumeMergerRepairReader(ctx, state)
	if err != nil {
		return err
	}

	head, err := reader.Output(ctx, "rev-parse", "HEAD")
	if err != nil || head != state.PreparedParent {
		return workflowstate.ErrConflict
	}

	return nil
}

func checkWorkflowMergerPrivateRepository(ctx context.Context, directory string) error {
	if !filepath.IsAbs(directory) {
		return workflowstate.ErrConflict
	}

	resolved, err := filepath.EvalSymlinks(directory)
	if err != nil || resolved != directory {
		return workflowstate.ErrConflict
	}

	info, err := os.Lstat(filepath.Join(directory, ".git"))
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return workflowstate.ErrConflict
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: directory, Environment: workflowMergerEnvironment(filepath.Dir(directory))}}

	remotes, err := reader.Output(ctx, "remote")
	if err != nil || remotes != "" {
		return workflowstate.ErrConflict
	}

	return nil
}

// ImportWorkflowMergerTree imports the host snapshot into the private repair
// object store. It does not change a branch or publish a temporary commit.
func (git OSWorkflowPreparationGit) ImportWorkflowMergerTree(ctx context.Context, state MergerRepairState, directory, tree string) error {
	if !workflowCommitID(tree) || directory == state.Query.Lane.Worktree {
		return workflowstate.ErrConflict
	}

	reader, err := git.resumeMergerRepairReader(ctx, state)
	if err != nil {
		return err
	}

	_, err = reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--no-write-fetch-head", "--", directory, tree)

	return err
}
