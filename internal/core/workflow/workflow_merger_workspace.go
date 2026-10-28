package workflow

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowMergerWorkspaceID(input workflowstate.MergerInput) string {
	return contenthash.Sum([]byte(input.InvocationID + "\x00" + input.EntryID + "\x00" + input.ID))
}

func (host workflowMergerHost) workspace(ctx context.Context, input workflowstate.MergerInput, seed string) (workflowMergerWorkspace, error) {
	common, err := projectcore.ResolveRepositoryDirectory(ctx, host.store.layout.ProjectDir())
	if err != nil {
		return workflowMergerWorkspace{}, err
	}

	parent := filepath.Join(common, "chr-merger-workspaces")
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return workflowMergerWorkspace{}, err
	}

	if resolved, err := filepath.EvalSymlinks(parent); err != nil || resolved != parent {
		return workflowMergerWorkspace{}, workflowstate.ErrConflict
	}

	root := filepath.Join(parent, workflowMergerWorkspaceID(input))
	if err := os.Mkdir(root, 0o700); err != nil {
		return workflowMergerWorkspace{}, errors.Join(workflowstate.ErrConflict, err)
	}

	// A private repository has no credential helpers, remotes, hooks or shared
	// writable object database. Neither worker nor user checkout is modified.
	worktree := filepath.Join(root, "checkout")

	format := "sha1"
	if len(seed) == 64 {
		format = "sha256"
	}

	commands := [][]string{{"init", "--quiet", "--object-format=" + format, worktree}, {"-C", worktree, "fetch", "--quiet", "--no-write-fetch-head", "--no-tags", common, seed}, {"-C", worktree, "checkout", "--quiet", "--detach", seed}}
	for _, args := range commands {
		command := exec.CommandContext(ctx, "git", args...)

		command.Env = workflowMergerEnvironment(root)
		if err := command.Run(); err != nil {
			return workflowMergerWorkspace{}, err
		}
	}

	return workflowMergerWorkspace{Lane: workflowstate.Lane{Unit: input.Unit, Branch: "refs/heads/chr/workflow/merger/" + input.OwnedWorkspaceID, Worktree: worktree, BaseCommit: seed, HeadCommit: seed}, Root: root}, nil
}

func workflowMergerEnvironment(root string) []string {
	return []string{"PATH=/usr/local/bin:/usr/bin:/bin", "HOME=" + root, "TMPDIR=" + root, "LANG=C.UTF-8", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0", "GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=core.hooksPath", "GIT_CONFIG_VALUE_0=/dev/null"}
}
