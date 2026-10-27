package execution

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WorkReadinessGit defines a package boundary interface.
type WorkReadinessGit interface {
	CurrentBranch(ctx context.Context, projectDir string) (string, error)
	WorktreeStatus(ctx context.Context, projectDir string) ([]string, error)
	BranchExists(ctx context.Context, projectDir, branch string) (bool, error)
	RefExists(ctx context.Context, projectDir, ref string) (bool, error)
	BranchContainsRef(ctx context.Context, projectDir, branch, ref string) (bool, error)
	GitStateInProgress(ctx context.Context, projectDir string) (GitOperationState, error)
}

// GitOperationState preserves the observed operation and its independent presence.
type GitOperationState struct {
	Name       string
	InProgress bool
}

// WorkReadinessRepairGit defines a package boundary interface.
type WorkReadinessRepairGit interface {
	WorkReadinessGit
	ExcludeLocalPattern(ctx context.Context, projectDir, pattern string) (bool, error)
	FetchRemote(ctx context.Context, projectDir, remote string) error
	FastForwardBranch(ctx context.Context, projectDir, branch, ref string) error
	SwitchBranch(ctx context.Context, projectDir, branch string) error
	CreateBranch(ctx context.Context, projectDir, branch string) error
}

// OSWorkReadinessGit describes a package boundary value.
type OSWorkReadinessGit struct{}

// CurrentBranch executes the corresponding package operation.
func (OSWorkReadinessGit) CurrentBranch(ctx context.Context, projectDir string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "branch", "--show-current").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// WorktreeStatus executes the corresponding package operation.
func (OSWorkReadinessGit) WorktreeStatus(ctx context.Context, projectDir string) ([]string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "status", "--porcelain").Output()
	if err != nil {
		return nil, err
	}

	return workReadinessNonEmptyLines(string(output)), nil
}

// BranchExists executes the corresponding package operation.
func (OSWorkReadinessGit) BranchExists(ctx context.Context, projectDir, branch string) (bool, error) {
	return gitRefExists(ctx, projectDir, "refs/heads/"+branch)
}

// RefExists executes the corresponding package operation.
func (OSWorkReadinessGit) RefExists(ctx context.Context, projectDir, ref string) (bool, error) {
	return gitRefExists(ctx, projectDir, ref)
}

// BranchContainsRef executes the corresponding package operation.
func (OSWorkReadinessGit) BranchContainsRef(ctx context.Context, projectDir, branch, ref string) (bool, error) {
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

// GitStateInProgress executes the corresponding package operation.
func (OSWorkReadinessGit) GitStateInProgress(ctx context.Context, projectDir string) (GitOperationState, error) {
	for _, state := range []string{"MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "rebase-merge", "rebase-apply"} {
		path, err := gitPath(ctx, projectDir, state)
		if err != nil {
			return GitOperationState{}, err
		}

		_, err = os.Stat(path)
		if err == nil {
			return GitOperationState{Name: state, InProgress: true}, nil
		}

		if !os.IsNotExist(err) {
			return GitOperationState{}, err
		}
	}

	return GitOperationState{}, nil
}

// ExcludeLocalPattern executes the corresponding package operation.
func (OSWorkReadinessGit) ExcludeLocalPattern(ctx context.Context, projectDir, pattern string) (bool, error) {
	err := ctx.Err()
	if err != nil {
		return false, err
	}

	path, err := gitPath(ctx, projectDir, "info/exclude")
	if err != nil {
		return false, err
	}

	return appendLineIfMissing(path, pattern)
}

// FetchRemote executes the corresponding package operation.
func (OSWorkReadinessGit) FetchRemote(ctx context.Context, projectDir, remote string) error {
	return runWorkReadinessGit(ctx, projectDir, "fetch", remote)
}

// FastForwardBranch executes the corresponding package operation.
func (git OSWorkReadinessGit) FastForwardBranch(ctx context.Context, projectDir, branch, ref string) error {
	current, err := git.CurrentBranch(ctx, projectDir)
	if err != nil {
		return err
	}

	if current != branch {
		return fmt.Errorf("current branch is %s, expected %s", current, branch)
	}

	return runWorkReadinessGit(ctx, projectDir, "merge", "--ff-only", ref)
}

// SwitchBranch executes the corresponding package operation.
func (OSWorkReadinessGit) SwitchBranch(ctx context.Context, projectDir, branch string) error {
	return runWorkReadinessGit(ctx, projectDir, "switch", branch)
}

// CreateBranch executes the corresponding package operation.
func (OSWorkReadinessGit) CreateBranch(ctx context.Context, projectDir, branch string) error {
	return runWorkReadinessGit(ctx, projectDir, "switch", "-c", branch)
}

func gitRefExists(ctx context.Context, projectDir, ref string) (bool, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--verify", ref).CombinedOutput()
	if err == nil {
		return true, nil
	}

	if gitCommandOutputLooksUnavailable(output) {
		return false, fmt.Errorf("verify git ref %s: %w", ref, err)
	}

	return false, nil
}

func gitPath(ctx context.Context, projectDir, path string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--git-path", path).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve git path %s: %w", path, err)
	}

	resolvedPath := strings.TrimSpace(string(output))
	if filepath.IsAbs(resolvedPath) {
		return resolvedPath, nil
	}

	return filepath.Join(projectDir, resolvedPath), nil
}

func gitCommandOutputLooksUnavailable(output []byte) bool {
	return strings.Contains(string(output), "not a git repository")
}

func runWorkReadinessGit(ctx context.Context, projectDir string, args ...string) error {
	commandArgs := append([]string{"-C", projectDir}, args...)

	output, err := exec.CommandContext(ctx, "git", commandArgs...).CombinedOutput()
	if err == nil {
		return nil
	}

	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, detail)
}

func appendLineIfMissing(path, value string) (bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return false, nil
	}

	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read %s: %w", path, err)
	}

	if lineExists(string(content), value) {
		return false, nil
	}

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return false, fmt.Errorf("create %s parent: %w", path, err)
	}

	next := string(content)
	if next != "" && !strings.HasSuffix(next, "\n") {
		next += "\n"
	}

	next += value + "\n"

	err = os.WriteFile(path, []byte(next), 0o644)
	if err != nil {
		return false, fmt.Errorf("write %s: %w", path, err)
	}

	return true, nil
}

func lineExists(content, value string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == value {
			return true
		}
	}

	return false
}

func workReadinessNonEmptyLines(value string) []string {
	lines := make([]string, 0)

	for _, line := range strings.Split(value, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func pathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
