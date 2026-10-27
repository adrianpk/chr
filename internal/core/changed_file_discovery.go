package core

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// ChangedFileDiscoverer defines a package boundary interface.
type ChangedFileDiscoverer interface {
	DiscoverChangedFiles(ctx context.Context, opts DiscoverChangedFilesOptions) ([]executioncore.ResultChangedFile, error)
}

// DiscoverChangedFilesOptions describes a package boundary value.
type DiscoverChangedFilesOptions struct {
	ProjectDir   string
	WorkspaceDir string
}

type gitChangedFileDiscoverer struct{}

// NewChangedFileDiscoverer constructs the corresponding package service or value.
func NewChangedFileDiscoverer() ChangedFileDiscoverer {
	return &gitChangedFileDiscoverer{}
}

// DiscoverChangedFiles executes the corresponding package operation.
func (d *gitChangedFileDiscoverer) DiscoverChangedFiles(
	ctx context.Context,
	opts DiscoverChangedFilesOptions,
) ([]executioncore.ResultChangedFile, error) {
	workspaceDir := changedFileDiscoveryWorkspaceDir(opts)
	if strings.TrimSpace(workspaceDir) == "" {
		return nil, fmt.Errorf("project directory is required")
	}

	output, err := exec.CommandContext(
		ctx,
		"git",
		"-C",
		workspaceDir,
		"status",
		"--porcelain=v1",
		"--untracked-files=all",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("discover changed files: %w", err)
	}

	return parseGitStatusChangedFiles(string(output)), nil
}

func changedFileDiscoveryWorkspaceDir(opts DiscoverChangedFilesOptions) string {
	if strings.TrimSpace(opts.WorkspaceDir) != "" {
		return opts.WorkspaceDir
	}

	return opts.ProjectDir
}

func parseGitStatusChangedFiles(output string) []executioncore.ResultChangedFile {
	changedFiles := []executioncore.ResultChangedFile{}
	seen := map[string]struct{}{}

	for _, line := range strings.Split(output, "\n") {
		path := changedFilePathFromGitStatusLine(line)
		if path == "" || operationalChangedFile(path) {
			continue
		}

		_, exists := seen[path]
		if exists {
			continue
		}

		seen[path] = struct{}{}
		changedFiles = append(changedFiles, executioncore.ResultChangedFile{Path: path})
	}

	slices.SortFunc(changedFiles, func(left, right executioncore.ResultChangedFile) int {
		if left.Path < right.Path {
			return -1
		}

		if left.Path > right.Path {
			return 1
		}

		return 0
	})

	return changedFiles
}

func changedFilePathFromGitStatusLine(line string) string {
	if len(line) < 4 {
		return ""
	}

	path := strings.TrimSpace(line[3:])
	if strings.Contains(path, " -> ") {
		_, path, _ = strings.Cut(path, " -> ")
	}

	return normalizeChangedFilePath(path)
}

func normalizeChangedFilePath(path string) string {
	path = filepath.ToSlash(strings.Trim(path, `"`))
	if path == "." || filepath.IsAbs(path) || strings.HasPrefix(path, "../") {
		return ""
	}

	return path
}

func operationalChangedFile(path string) bool {
	runtimePath := workspacecore.UserRelativePath("runtime")

	return path == runtimePath || strings.HasPrefix(path, runtimePath+"/")
}
