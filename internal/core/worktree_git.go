package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// WorktreeGit defines a package boundary interface.
type WorktreeGit interface {
	FetchRemote(ctx context.Context, projectDir, remote string) error
	ResolveRef(ctx context.Context, projectDir, ref string) (string, error)
	BranchExists(ctx context.Context, projectDir, branch string) (bool, error)
	BranchContainsRef(ctx context.Context, projectDir, branch, ref string) (bool, error)
	CreateBranch(ctx context.Context, projectDir, branch, baseRef string) error
	AddWorktree(ctx context.Context, projectDir, path, branch string) error
	RemoveWorktree(ctx context.Context, projectDir, path string) error
	WorktreeList(ctx context.Context, projectDir string) ([]WorktreeGitEntry, error)
	WorktreeStatus(ctx context.Context, path string) ([]string, error)
}

// OSWorktreeGit describes a package boundary value.
type OSWorktreeGit struct{}

// WorktreeGitEntry describes a package boundary value.
type WorktreeGitEntry struct {
	Path     string
	Head     string
	Branch   string
	Detached bool
	Bare     bool
	Prunable bool
}

// FetchRemote executes the corresponding package operation.
func (OSWorktreeGit) FetchRemote(ctx context.Context, projectDir, remote string) error {
	return runWorktreeGit(ctx, projectDir, "fetch", remote)
}

// ResolveRef executes the corresponding package operation.
func (OSWorktreeGit) ResolveRef(ctx context.Context, projectDir, ref string) (string, error) {
	output, err := outputWorktreeGit(ctx, projectDir, "rev-parse", "--verify", ref)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// BranchExists executes the corresponding package operation.
func (OSWorktreeGit) BranchExists(ctx context.Context, projectDir, branch string) (bool, error) {
	return gitRefExists(ctx, projectDir, "refs/heads/"+branch)
}

// BranchContainsRef executes the corresponding package operation.
func (OSWorktreeGit) BranchContainsRef(ctx context.Context, projectDir, branch, ref string) (bool, error) {
	err := exec.CommandContext(ctx, "git", "-C", projectDir, "merge-base", "--is-ancestor", ref, branch).Run()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, err
}

// CreateBranch executes the corresponding package operation.
func (OSWorktreeGit) CreateBranch(ctx context.Context, projectDir, branch, baseRef string) error {
	return runWorktreeGit(ctx, projectDir, "branch", branch, baseRef)
}

// AddWorktree executes the corresponding package operation.
func (OSWorktreeGit) AddWorktree(ctx context.Context, projectDir, path, branch string) error {
	return runWorktreeGit(ctx, projectDir, "worktree", "add", path, branch)
}

// RemoveWorktree executes the corresponding package operation.
func (OSWorktreeGit) RemoveWorktree(ctx context.Context, projectDir, path string) error {
	return runWorktreeGit(ctx, projectDir, "worktree", "remove", path)
}

// WorktreeList executes the corresponding package operation.
func (OSWorktreeGit) WorktreeList(ctx context.Context, projectDir string) ([]WorktreeGitEntry, error) {
	output, err := outputWorktreeGit(ctx, projectDir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return parseWorktreeGitList(output), nil
}

// WorktreeStatus executes the corresponding package operation.
func (OSWorktreeGit) WorktreeStatus(ctx context.Context, path string) ([]string, error) {
	output, err := outputWorktreeGit(ctx, path, "status", "--porcelain")
	if err != nil {
		return nil, err
	}

	return workReadinessNonEmptyLines(output), nil
}

func parseWorktreeGitList(output string) []WorktreeGitEntry {
	entries := make([]WorktreeGitEntry, 0)
	current := WorktreeGitEntry{}

	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			if current.Path != "" {
				entries = append(entries, current)
				current = WorktreeGitEntry{}
			}

			continue
		}

		current = applyWorktreeGitListLine(current, line)
	}

	if current.Path != "" {
		entries = append(entries, current)
	}

	return entries
}

func applyWorktreeGitListLine(entry WorktreeGitEntry, line string) WorktreeGitEntry {
	key, value, _ := strings.Cut(line, " ")
	switch key {
	case "worktree":
		entry.Path = strings.TrimSpace(value)
	case "HEAD":
		entry.Head = strings.TrimSpace(value)
	case "branch":
		entry.Branch = strings.TrimPrefix(strings.TrimSpace(value), "refs/heads/")
	case "detached":
		entry.Detached = true
	case "bare":
		entry.Bare = true
	case "prunable":
		entry.Prunable = true
	}

	return entry
}

func outputWorktreeGit(ctx context.Context, projectDir string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", projectDir}, args...)

	output, err := exec.CommandContext(ctx, "git", commandArgs...).CombinedOutput()
	if err != nil {
		return "", worktreeGitCommandError(args, output, err)
	}

	return string(output), nil
}

func runWorktreeGit(ctx context.Context, projectDir string, args ...string) error {
	_, err := outputWorktreeGit(ctx, projectDir, args...)

	return err
}

func worktreeGitCommandError(args []string, output []byte, err error) error {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, detail)
}
