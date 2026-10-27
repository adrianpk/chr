package core

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// RuntimeAssets defines a package boundary interface.
type RuntimeAssets interface {
	Init(ctx context.Context, opts RuntimeInitOptions) (RuntimeInitResult, error)
}

type runtimeAssets struct {
	workspace *workspace
}

// RuntimeInitOptions describes a package boundary value.
type RuntimeInitOptions struct {
	ProjectDir string
	DryRun     bool
}

// RuntimeInitResult describes a package boundary value.
type RuntimeInitResult struct {
	Root    string
	Created []string
	Skipped []string
}

// NewRuntimeAssets constructs the corresponding package service or value.
func NewRuntimeAssets(assetsFS fs.FS) RuntimeAssets {
	return &runtimeAssets{workspace: &workspace{assetsFS: assetsFS}}
}

// Init executes the corresponding package operation.
func (r *runtimeAssets) Init(ctx context.Context, opts RuntimeInitOptions) (RuntimeInitResult, error) {
	err := ctx.Err()
	if err != nil {
		return RuntimeInitResult{}, err
	}

	absProjectDir, err := projectcore.ResolveDirectory(opts.ProjectDir)
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

		rel := filepath.Join(workspacecore.DirectoryName, UserDirectoryName, filepath.FromSlash(file.path))
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
