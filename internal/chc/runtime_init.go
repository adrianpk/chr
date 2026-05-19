package chc

import (
	"context"
	"fmt"
	"io/fs"
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
	if err := ctx.Err(); err != nil {
		return RuntimeInitResult{}, err
	}

	absProjectDir, err := resolveProjectDir(opts.ProjectDir)
	if err != nil {
		return RuntimeInitResult{}, err
	}

	result := RuntimeInitResult{Root: filepath.Join(absProjectDir, DirectoryName, "runtime")}

	files, err := r.workspace.profileFiles(DefaultProfile, workspaceFile)
	if err != nil {
		return RuntimeInitResult{}, err
	}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return RuntimeInitResult{}, err
		}

		if !isRuntimeAsset(file.path) {
			continue
		}

		rel := filepath.Join(DirectoryName, filepath.FromSlash(file.path))
		target := filepath.Join(absProjectDir, rel)

		created, err := createFile(target, file.content, opts.DryRun)
		if err != nil {
			return RuntimeInitResult{}, fmt.Errorf("init runtime asset %s: %w", file.path, err)
		}

		recordRuntimePath(&result, rel, created)
	}

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
