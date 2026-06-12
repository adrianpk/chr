package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/core"
)

func newProjectCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage registered projects",
		Commands: []*cli.Command{
			newProjectAddCommand(stdout, deps),
			newProjectCurrentCommand(stdout, deps),
			newProjectListCommand(stdout, deps),
			newProjectSelectCommand(stdout, deps),
		},
	}
}

func newProjectAddCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Register a local project",
		Flags: append(projectRegistryFlags(), &cli.StringFlag{
			Name:  "name",
			Usage: "registered project name",
		}),
		Action: projectAddAction(stdout, deps),
	}
}

func projectAddAction(stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectAdd(ctx, cmd, stdout, deps.assetsFS)
	}
}

func runProjectAdd(ctx context.Context, cmd *cli.Command, stdout io.Writer, assetsFS fs.FS) error {
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

	registry, err := core.NewProjectRegistry(ctx, env, assetsFS)
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

	project := core.Project{
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

	_, err = fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path)
	if err != nil {
		return fmt.Errorf("write project add result: %w", err)
	}

	return nil
}

func newProjectListCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List registered projects",
		Flags:  projectRegistryFlags(),
		Action: projectListAction(stdout, deps),
	}
}

func projectListAction(stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectList(ctx, cmd, stdout, deps.assetsFS)
	}
}

func runProjectList(ctx context.Context, cmd *cli.Command, stdout io.Writer, assetsFS fs.FS) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("project list does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := core.NewProjectRegistry(ctx, env, assetsFS)
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
		_, err = fmt.Fprintf(stdout, "%s\t%s\t%t\t%s\n", project.Name, project.Path, project.Enabled, project.ID)
		if err != nil {
			return fmt.Errorf("write project list result: %w", err)
		}
	}

	return nil
}

func newProjectCurrentCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:   "current",
		Usage:  "Print the active project",
		Flags:  projectRegistryFlags(),
		Action: projectCurrentAction(stdout, deps),
	}
}

func projectCurrentAction(stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectCurrent(ctx, cmd, stdout, deps.assetsFS)
	}
}

func runProjectCurrent(ctx context.Context, cmd *cli.Command, stdout io.Writer, assetsFS fs.FS) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("project current does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	activeProject, err := core.NewActiveProjectStore(env).LoadActiveProject()
	if err != nil {
		return err
	}

	registry, err := core.NewProjectRegistry(ctx, env, assetsFS)
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

	_, err = fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path)
	if err != nil {
		return fmt.Errorf("write project current result: %w", err)
	}

	return nil
}

func newProjectSelectCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:   "select",
		Usage:  "Select the active project",
		Flags:  projectRegistryFlags(),
		Action: projectSelectAction(stdout, deps),
	}
}

func projectSelectAction(stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runProjectSelect(ctx, cmd, stdout, deps.assetsFS)
	}
}

func runProjectSelect(ctx context.Context, cmd *cli.Command, stdout io.Writer, assetsFS fs.FS) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("project select requires exactly one project name or path argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := core.NewProjectRegistry(ctx, env, assetsFS)
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

	activeProject, err := core.NewActiveProject(project.ID)
	if err != nil {
		return err
	}

	store := core.NewActiveProjectStore(env)

	err = store.SetActiveProject(activeProject)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "project: %s %s\n", project.Name, project.Path)
	if err != nil {
		return fmt.Errorf("write project select result: %w", err)
	}

	return nil
}

func resolveProjectSelection(ctx context.Context, registry core.ProjectRegistry, selector string) (core.Project, error) {
	project, err := registry.GetProjectByName(ctx, selector)
	if err == nil {
		return project, nil
	}

	if !errors.Is(err, core.ErrProjectNotFound) {
		return core.Project{}, err
	}

	projectPath, pathErr := filepath.Abs(selector)
	if pathErr != nil {
		return core.Project{}, fmt.Errorf("resolve project selector path: %w", pathErr)
	}

	project, err = registry.GetProjectByPath(ctx, projectPath)
	if err != nil {
		return core.Project{}, fmt.Errorf("resolve project selection: %w", err)
	}

	return project, nil
}
