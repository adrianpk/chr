package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func (factory commandFactory) projectCommand() *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage registered projects",
		Commands: []*cli.Command{
			factory.projectAddCommand(),
			factory.projectCurrentCommand(),
			factory.projectListCommand(),
			factory.projectSelectCommand(),
		},
	}
}

func (factory commandFactory) projectAddCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Register a local project",
		Flags: append(projectRegistryFlags(), &cli.StringFlag{
			Name:  "name",
			Usage: "registered project name",
		}),
		Action: factory.runProjectAdd,
	}
}

func (factory commandFactory) runProjectAdd(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("project add requires exactly one project directory argument")
	}

	projectDir, err := filepath.Abs(cmd.Args().First())
	if err != nil {
		return fmt.Errorf("resolve project directory: %w", err)
	}

	err = requireWorkspace(projectDir)
	if err != nil {
		return err
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := projectcore.NewRegistry(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	projectName := cmd.String("name")
	if projectName == "" {
		projectName = filepath.Base(projectDir)
	}

	project := projectcore.Project{
		ID:           projectID(projectDir),
		Name:         projectName,
		Path:         projectDir,
		Enabled:      true,
		RegisteredAt: time.Now().UTC(),
	}

	err = registry.AddProject(ctx, project)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"project: %s %s\n", project.Name, project.Path)
	if err != nil {
		return fmt.Errorf("write project add result: %w", err)
	}

	return nil
}

func (factory commandFactory) projectListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List registered projects",
		Flags:  projectRegistryFlags(),
		Action: factory.runProjectList,
	}
}

func (factory commandFactory) runProjectList(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("project list does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := projectcore.NewRegistry(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	projects, err := registry.ListProjects(ctx)
	if err != nil {
		return err
	}

	for _, project := range projects {
		_, err = fmt.Fprintf(factory.io.Stdout,
			"%s\t%s\t%t\t%s\n", project.Name, project.Path, project.Enabled, project.ID)
		if err != nil {
			return fmt.Errorf("write project list result: %w", err)
		}
	}

	return nil
}

func (factory commandFactory) projectCurrentCommand() *cli.Command {
	return &cli.Command{
		Name:   "current",
		Usage:  "Print the active project",
		Flags:  projectRegistryFlags(),
		Action: factory.runProjectCurrent,
	}
}

func (factory commandFactory) runProjectCurrent(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("project current does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	activeProject, err := projectcore.NewActiveStore(env).LoadActiveProject()
	if err != nil {
		return err
	}

	registry, err := projectcore.NewRegistry(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	project, err := registry.GetProjectByID(ctx, activeProject.ProjectID)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"project: %s %s\n", project.Name, project.Path)
	if err != nil {
		return fmt.Errorf("write project current result: %w", err)
	}

	return nil
}

func (factory commandFactory) projectSelectCommand() *cli.Command {
	return &cli.Command{
		Name:   "select",
		Usage:  "Select the active project",
		Flags:  projectRegistryFlags(),
		Action: factory.runProjectSelect,
	}
}

func (factory commandFactory) runProjectSelect(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("project select requires exactly one project name or path argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := projectcore.NewRegistry(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	project, err := resolveProjectSelection(ctx, registry, cmd.Args().First())
	if err != nil {
		return err
	}

	if !project.Enabled {
		return fmt.Errorf("project is disabled: %s", project.Name)
	}

	_, err = resolveRegisteredProjectDirectory(project)
	if err != nil {
		return err
	}

	activeProject, err := projectcore.NewActive(project.ID)
	if err != nil {
		return err
	}

	store := projectcore.NewActiveStore(env)

	err = store.SetActiveProject(activeProject)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"project: %s %s\n", project.Name, project.Path)
	if err != nil {
		return fmt.Errorf("write project select result: %w", err)
	}

	return nil
}

func resolveProjectSelection(ctx context.Context, registry projectcore.Registry, selector string) (projectcore.Project, error) {
	project, err := registry.GetProjectByName(ctx, selector)
	if err == nil {
		return project, nil
	}

	if !errors.Is(err, projectcore.ErrNotFound) {
		return projectcore.Project{}, err
	}

	projectPath, pathErr := filepath.Abs(selector)
	if pathErr != nil {
		return projectcore.Project{}, fmt.Errorf("resolve project selector path: %w", pathErr)
	}

	project, err = registry.GetProjectByPath(ctx, projectPath)
	if err != nil {
		return projectcore.Project{}, fmt.Errorf("resolve project selection: %w", err)
	}

	return project, nil
}
