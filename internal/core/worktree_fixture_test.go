package core

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func testCleanupWorktreeWorkspace() workspacecore.ExecutionConfig {
	workspace := workspacecore.DefaultExecutionConfig()
	workspace.Mode = workspacecore.ExecutionWorktree

	return workspace
}

func writeWorktreeTestMetadata(t *testing.T, plan executioncore.WorktreePlan) {
	t.Helper()

	path, err := executioncore.WorktreeMetadataPath(plan)
	if err != nil {
		t.Fatalf("WorktreeMetadataPath returned error: %v", err)
	}

	if err := executioncore.WriteWorktreeMetadata(path, executioncore.NewWorktreeMetadata(plan, time.Now())); err != nil {
		t.Fatalf("WriteWorktreeMetadata returned error: %v", err)
	}
}

func assertWorktreeMetadataMissing(t *testing.T, plan executioncore.WorktreePlan) {
	t.Helper()

	path, err := executioncore.WorktreeMetadataPath(plan)
	if err != nil {
		t.Fatalf("WorktreeMetadataPath returned error: %v", err)
	}

	_, err = os.Stat(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("metadata path %s stat error = %v, want missing", path, err)
	}
}

type fakeWorktreeGit struct {
	fetched           []string
	refs              map[string]string
	branches          map[string]bool
	branchContainsRef map[string]bool
	createdBranches   map[string]string
	addedWorktrees    []fakeAddedWorktree
	removedWorktrees  []string
	entries           []executioncore.WorktreeGitEntry
	statuses          map[string][]string
	fetchErr          error
}

type fakeAddedWorktree struct {
	path   string
	branch string
}

func newFakeWorktreeGit() *fakeWorktreeGit {
	return &fakeWorktreeGit{
		refs:              map[string]string{"origin/dev": "base-sha"},
		branches:          map[string]bool{},
		branchContainsRef: map[string]bool{},
		createdBranches:   map[string]string{},
		statuses:          map[string][]string{},
	}
}

func (git *fakeWorktreeGit) FetchRemote(_ context.Context, _, remote string) error {
	if git.fetchErr != nil {
		return git.fetchErr
	}

	git.fetched = append(git.fetched, remote)

	return nil
}

func (git *fakeWorktreeGit) ResolveRef(_ context.Context, _, ref string) (string, error) {
	sha, ok := git.refs[ref]
	if !ok {
		return "", errors.New("missing ref")
	}

	return sha, nil
}

func (git *fakeWorktreeGit) BranchExists(_ context.Context, _, branch string) (bool, error) {
	return git.branches[branch], nil
}

func (git *fakeWorktreeGit) BranchContainsRef(_ context.Context, _, branch, ref string) (bool, error) {
	contains, ok := git.branchContainsRef[branchContainsRefKey(branch, ref)]
	if !ok {
		return true, nil
	}

	return contains, nil
}

func (git *fakeWorktreeGit) CreateBranch(_ context.Context, _, branch, baseRef string) error {
	git.branches[branch] = true
	git.createdBranches[branch] = baseRef

	return nil
}

func (git *fakeWorktreeGit) AddWorktree(_ context.Context, _, path, branch string) error {
	git.addedWorktrees = append(git.addedWorktrees, fakeAddedWorktree{path: path, branch: branch})

	return nil
}

func (git *fakeWorktreeGit) RemoveWorktree(_ context.Context, _, path string) error {
	git.removedWorktrees = append(git.removedWorktrees, path)

	return nil
}

func (git *fakeWorktreeGit) WorktreeList(context.Context, string) ([]executioncore.WorktreeGitEntry, error) {
	return append([]executioncore.WorktreeGitEntry(nil), git.entries...), nil
}

func (git *fakeWorktreeGit) WorktreeStatus(_ context.Context, path string) ([]string, error) {
	return append([]string(nil), git.statuses[path]...), nil
}

func branchContainsRefKey(branch, ref string) string {
	return branch + "\x00" + ref
}
