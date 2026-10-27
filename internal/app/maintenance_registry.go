package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func (factory commandFactory) maintenanceRegistryCommand() *cli.Command {
	return &cli.Command{
		Name:  "registry",
		Usage: "Inspect project registry state",
		Commands: []*cli.Command{
			factory.maintenanceRegistryDisableCommand(),
			factory.maintenanceRegistryStatusCommand(),
			factory.maintenanceRegistrySetPathCommand(),
		},
	}
}

func (factory commandFactory) maintenanceRegistryDisableCommand() *cli.Command {
	return &cli.Command{
		Name:      "disable",
		Usage:     "Disable a registered project",
		ArgsUsage: "<project-name>",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print planned repair without updating the registry",
		}),
		Action: factory.runMaintenanceRegistryDisable,
	}
}

func (factory commandFactory) maintenanceRegistryStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report project registry diagnostics",
		Flags:  projectRegistryFlags(),
		Action: factory.runMaintenanceRegistryStatus,
	}
}

func (factory commandFactory) maintenanceRegistrySetPathCommand() *cli.Command {
	return &cli.Command{
		Name:      "set-path",
		Usage:     "Update a registered project path",
		ArgsUsage: "<project-name> <project-dir>",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print planned repair without updating the registry",
		}),
		Action: factory.runMaintenanceRegistrySetPath,
	}
}

func (factory commandFactory) runMaintenanceRegistryStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance registry status does not accept positional arguments")
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

	diagnostics, err := projectcore.NewDiagnosticsService(
		registry,
		projectcore.NewActiveStore(env),
		core.DirectoryName,
	).ListDiagnostics(ctx)
	if err != nil {
		return err
	}

	return printProjectRegistryDiagnostics(factory.io.Stdout, diagnostics)
}

func (factory commandFactory) runMaintenanceRegistryDisable(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("maintenance registry disable requires exactly one project name")
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

	projectName := cmd.Args().First()

	project, err := registry.GetProjectByName(ctx, projectName)
	if err != nil {
		return err
	}

	if cmd.Bool("dry-run") {
		return printProjectRegistryDisable(factory.io.Stdout, project, true)
	}

	project, err = registry.DisableProject(ctx, projectName)
	if err != nil {
		return err
	}

	return printProjectRegistryDisable(factory.io.Stdout, project, false)
}

func (factory commandFactory) runMaintenanceRegistrySetPath(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("maintenance registry set-path requires project name and project directory")
	}

	projectDir, err := filepath.Abs(cmd.Args().Get(1))
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

	projectName := cmd.Args().First()

	project, err := registry.GetProjectByName(ctx, projectName)
	if err != nil {
		return err
	}

	previousPath := project.Path
	project.Path = projectDir

	if cmd.Bool("dry-run") {
		return printProjectRegistrySetPath(factory.io.Stdout, project, previousPath, true)
	}

	project, err = registry.UpdateProjectPath(ctx, projectName, projectDir)
	if err != nil {
		return err
	}

	return printProjectRegistrySetPath(factory.io.Stdout, project, previousPath, false)
}

func printProjectRegistryDiagnostics(stdout io.Writer, diagnostics []projectcore.Diagnostic) error {
	_, err := fmt.Fprintf(stdout, "projects: %d\n", len(diagnostics))
	if err != nil {
		return fmt.Errorf("write project registry diagnostics: %w", err)
	}

	for _, diagnostic := range diagnostics {
		_, err = fmt.Fprintf(
			stdout,
			"project: %s path: %s enabled: %t active: %t path_exists: %t workspace_exists: %t status: %s id: %s\n",
			diagnostic.Project.Name,
			diagnostic.Project.Path,
			diagnostic.Project.Enabled,
			diagnostic.Active,
			diagnostic.PathExists,
			diagnostic.WorkspaceExists,
			diagnostic.Status,
			diagnostic.Project.ID,
		)
		if err != nil {
			return fmt.Errorf("write project registry diagnostic: %w", err)
		}
	}

	return nil
}

func printProjectRegistryDisable(stdout io.Writer, project projectcore.Project, dryRun bool) error {
	action := "disabled"
	if dryRun {
		action = "would_disable"
	}

	_, err := fmt.Fprintf(stdout, "%s: %s path: %s id: %s\n", action, project.Name, project.Path, project.ID)
	if err != nil {
		return fmt.Errorf("write project registry disable result: %w", err)
	}

	return nil
}

func printProjectRegistrySetPath(stdout io.Writer, project projectcore.Project, previousPath string, dryRun bool) error {
	action := "updated_path"
	if dryRun {
		action = "would_update_path"
	}

	_, err := fmt.Fprintf(
		stdout,
		"%s: %s from: %s to: %s id: %s\n",
		action,
		project.Name,
		previousPath,
		project.Path,
		project.ID,
	)
	if err != nil {
		return fmt.Errorf("write project registry set-path result: %w", err)
	}

	return nil
}
