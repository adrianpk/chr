package chc

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type RuntimeAssets interface {
	Init(ctx context.Context, opts RuntimeInitOptions) (RuntimeInitResult, error)
}

type runtimeAssets struct {
	workspace *workspace
}

type RuntimeInitOptions struct {
	ProjectDir string
	DryRun     bool
}

type RuntimeInitResult struct {
	Root    string
	Created []string
	Skipped []string
}

func NewRuntimeAssets(assetsFS fs.FS) RuntimeAssets {
	return &runtimeAssets{workspace: &workspace{assetsFS: assetsFS}}
}

func (r *runtimeAssets) Init(ctx context.Context, opts RuntimeInitOptions) (RuntimeInitResult, error) {
	err := ctx.Err()
	if err != nil {
		return RuntimeInitResult{}, err
	}

	absProjectDir, err := resolveProjectDir(opts.ProjectDir)
	if err != nil {
		return RuntimeInitResult{}, err
	}

	result := RuntimeInitResult{Root: UserPath(absProjectDir, "runtime")}

	files, err := r.workspace.profileFiles(DefaultProfile, workspaceFile)
	if err != nil {
		return RuntimeInitResult{}, err
	}

	for _, file := range files {
		err = ctx.Err()
		if err != nil {
			return RuntimeInitResult{}, err
		}

		if !isRuntimeAsset(file.path) {
			continue
		}

		rel := filepath.Join(DirectoryName, UserDirectoryName, filepath.FromSlash(file.path))
		target := filepath.Join(absProjectDir, rel)

		created, err := createFile(target, file.content, opts.DryRun)
		if err != nil {
			return RuntimeInitResult{}, fmt.Errorf("init runtime asset %s: %w", file.path, err)
		}

		recordRuntimePath(&result, rel, created)
	}

	runsPath := filepath.Join(DirectoryName, UserDirectoryName, "runtime", "runs")

	created, err := createDirectory(filepath.Join(absProjectDir, runsPath), opts.DryRun)
	if err != nil {
		return RuntimeInitResult{}, fmt.Errorf("init runtime runs directory: %w", err)
	}

	recordRuntimePath(&result, runsPath, created)

	return result, nil
}

func isRuntimeAsset(path string) bool {
	slashPath := filepath.ToSlash(path)

	return slashPath == "runtime" || strings.HasPrefix(slashPath, "runtime/")
}

func recordRuntimePath(result *RuntimeInitResult, rel string, created bool) {
	path := filepath.ToSlash(rel)
	if created {
		result.Created = append(result.Created, path)

		return
	}

	result.Skipped = append(result.Skipped, path)
}

func createDirectory(path string, dryRun bool) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}

	if !os.IsNotExist(err) {
		return false, err
	}

	if dryRun {
		return true, nil
	}

	err = os.MkdirAll(path, 0o755)
	if err != nil {
		return false, err
	}

	return true, nil
}
