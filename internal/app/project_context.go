package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

type commandProjectContext struct {
	ProjectDir  string
	ProjectID   string
	ProjectName string
}

type commandProjectContextOptions struct {
	ExplicitDir           string
	UseExplicitDir        bool
	AllowCurrentDirectory bool
}

type commandProjectContextResolver struct {
	env      core.UserEnvironment
	assetsFS fs.FS
}

func newCommandProjectContextResolver(env core.UserEnvironment, assetsFS fs.FS) commandProjectContextResolver {
	return commandProjectContextResolver{env: env, assetsFS: assetsFS}
}

func (r commandProjectContextResolver) Resolve(
	ctx context.Context,
	opts commandProjectContextOptions,
) (commandProjectContext, error) {
	if opts.UseExplicitDir {
		return resolveProjectDirectory(opts.ExplicitDir)
	}

	activeProject, err := projectcore.NewActiveStore(r.env).LoadActiveProject()
	if err == nil {
		return r.resolveActiveProject(ctx, activeProject)
	}

	if !errors.Is(err, projectcore.ErrActiveNotSelected) {
		return commandProjectContext{}, err
	}

	if opts.AllowCurrentDirectory {
		return resolveProjectDirectory(".")
	}

	return commandProjectContext{}, fmt.Errorf("project context is required")
}

func (r commandProjectContextResolver) resolveActiveProject(
	ctx context.Context,
	activeProject projectcore.Active,
) (commandProjectContext, error) {
	registry, err := projectcore.NewRegistry(ctx, r.env, r.assetsFS)
	if err != nil {
		return commandProjectContext{}, err
	}

	defer func() {
		//nolint:errcheck // Resolution result already comes from the main operation.
		_ = registry.Close()
	}()

	project, err := registry.GetProjectByID(ctx, activeProject.ProjectID)
	if err != nil {
		return commandProjectContext{}, err
	}

	if !project.Enabled {
		return commandProjectContext{}, fmt.Errorf("project is disabled: %s", project.Name)
	}

	return resolveRegisteredProjectDirectory(project)
}

func resolveProjectDirectory(projectDir string) (commandProjectContext, error) {
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return commandProjectContext{}, fmt.Errorf("resolve project directory: %w", err)
	}

	err = requireWorkspace(absProjectDir)
	if err != nil {
		return commandProjectContext{}, err
	}

	return commandProjectContext{
		ProjectDir:  absProjectDir,
		ProjectID:   projectID(absProjectDir),
		ProjectName: filepath.Base(absProjectDir),
	}, nil
}

func resolveRegisteredProjectDirectory(project projectcore.Project) (commandProjectContext, error) {
	if project.Path == "" {
		return commandProjectContext{}, fmt.Errorf("registered project path is required: %s", project.Name)
	}

	info, err := os.Stat(project.Path)
	if err != nil {
		return commandProjectContext{}, fmt.Errorf("inspect registered project path %q: %w", project.Path, err)
	}

	if !info.IsDir() {
		return commandProjectContext{}, fmt.Errorf("registered project path is not a directory: %s", project.Path)
	}

	workspacePath := filepath.Join(project.Path, core.DirectoryName)

	info, err = os.Stat(workspacePath)
	if err != nil {
		return commandProjectContext{}, fmt.Errorf("inspect registered project workspace %q: %w", workspacePath, err)
	}

	if !info.IsDir() {
		return commandProjectContext{}, fmt.Errorf("registered project workspace is not a directory: %s", workspacePath)
	}

	return commandProjectContext{
		ProjectDir:  project.Path,
		ProjectID:   project.ID,
		ProjectName: project.Name,
	}, nil
}
