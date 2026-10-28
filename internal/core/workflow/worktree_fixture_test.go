package workflow

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func outputWorktreeGit(ctx context.Context, projectDir string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", projectDir}, args...)

	output, err := exec.CommandContext(ctx, "git", commandArgs...).CombinedOutput()
	if err != nil {
		return "", worktreeGitCommandError(args, output, err)
	}

	return string(output), nil
}

func worktreeGitCommandError(args []string, output []byte, err error) error {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, detail)
}
