package project

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResolveRepositoryDirectory resolves the checkout's shared Git metadata directory.
func ResolveRepositoryDirectory(ctx context.Context, projectDir string) (string, error) {
	command := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--path-format=absolute", "--git-common-dir")

	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("resolve workflow repository coordinator: %w", err)
	}

	return filepath.EvalSymlinks(strings.TrimSpace(string(output)))
}
