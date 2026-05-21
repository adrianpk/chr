package chcapp

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

func newMaintenanceCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "maintenance",
		Usage: "Inspect and recover local chc state",
		Commands: []*cli.Command{
			newMaintenanceDBCommand(stdout),
			newMaintenanceExecutionCommand(stdout),
			newMaintenanceRegistryCommand(stdout),
		},
	}
}

func newMaintenanceDBCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "Inspect local database state",
		Commands: []*cli.Command{
			newMaintenanceDBStatusCommand(stdout),
		},
	}
}

func newMaintenanceDBStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report local database migration status",
		Flags:  projectRegistryFlags(),
		Action: maintenanceDBStatusAction(stdout),
	}
}

func newMaintenanceExecutionCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "execution",
		Usage: "Inspect execution records and artifacts",
		Commands: []*cli.Command{
			newMaintenanceExecutionArtifactsCommand(stdout),
		},
	}
}

func newMaintenanceExecutionArtifactsCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "artifacts",
		Usage:  "Report execution artifact drift for the resolved project",
		Flags:  maintenanceProjectFlags(),
		Action: maintenanceExecutionArtifactsAction(stdout),
	}
}

func newMaintenanceRegistryCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "registry",
		Usage: "Inspect project registry state",
		Commands: []*cli.Command{
			newMaintenanceRegistryDisableCommand(stdout),
			newMaintenanceRegistryStatusCommand(stdout),
			newMaintenanceRegistrySetPathCommand(stdout),
		},
	}
}

func newMaintenanceRegistryDisableCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "disable",
		Usage:     "Disable a registered project",
		ArgsUsage: "<project-name>",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print planned repair without updating the registry",
		}),
		Action: maintenanceRegistryDisableAction(stdout),
	}
}

func newMaintenanceRegistryStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report project registry diagnostics",
		Flags:  projectRegistryFlags(),
		Action: maintenanceRegistryStatusAction(stdout),
	}
}

func newMaintenanceRegistrySetPathCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "set-path",
		Usage:     "Update a registered project path",
		ArgsUsage: "<project-name> <project-dir>",
		Flags: append(projectRegistryFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print planned repair without updating the registry",
		}),
		Action: maintenanceRegistrySetPathAction(stdout),
	}
}

func maintenanceProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}

func maintenanceDBStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runMaintenanceDBStatus(ctx, cmd, stdout)
	}
}

func maintenanceExecutionArtifactsAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runMaintenanceExecutionArtifacts(ctx, cmd, stdout)
	}
}

func maintenanceRegistryStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runMaintenanceRegistryStatus(ctx, cmd, stdout)
	}
}

func maintenanceRegistryDisableAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runMaintenanceRegistryDisable(ctx, cmd, stdout)
	}
}

func maintenanceRegistrySetPathAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runMaintenanceRegistrySetPath(ctx, cmd, stdout)
	}
}

func runMaintenanceDBStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance db status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	status, err := chc.NewMigrationStatusChecker(assetsFS).Check(ctx, env)
	if err != nil {
		return err
	}

	return printMigrationStatus(stdout, status)
}

func runMaintenanceExecutionArtifacts(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance execution artifacts does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:    cmd.String("dir"),
		UseExplicitDir: true,
	})
	if err != nil {
		return err
	}

	repository, err := chc.NewExecutionRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	diagnostics, err := chc.NewExecutionArtifactDriftService(repository).ListDiagnostics(
		ctx,
		projectContext.ProjectID,
	)
	if err != nil {
		return err
	}

	return printExecutionArtifactDriftDiagnostics(stdout, diagnostics)
}

func runMaintenanceRegistryStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance registry status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	diagnostics, err := chc.NewProjectRegistryDiagnosticsService(
		registry,
		chc.NewActiveProjectStore(env),
	).ListDiagnostics(ctx)
	if err != nil {
		return err
	}

	return printProjectRegistryDiagnostics(stdout, diagnostics)
}

func runMaintenanceRegistryDisable(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("maintenance registry disable requires exactly one project name")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
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
		return printProjectRegistryDisable(stdout, project, true)
	}

	project, err = registry.DisableProject(ctx, projectName)
	if err != nil {
		return err
	}

	return printProjectRegistryDisable(stdout, project, false)
}

func runMaintenanceRegistrySetPath(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
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

	registry, err := chc.NewProjectRegistry(ctx, env, assetsFS)
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
		return printProjectRegistrySetPath(stdout, project, previousPath, true)
	}

	project, err = registry.UpdateProjectPath(ctx, projectName, projectDir)
	if err != nil {
		return err
	}

	return printProjectRegistrySetPath(stdout, project, previousPath, false)
}

func printMigrationStatus(stdout io.Writer, status chc.MigrationStatus) error {
	lines := []string{
		fmt.Sprintf("database: %s", status.DatabasePath),
		fmt.Sprintf("database_exists: %t", status.DatabaseExists),
		fmt.Sprintf("schema_migrations_exists: %t", status.SchemaTableExists),
		fmt.Sprintf("available: %d", len(status.Available)),
		fmt.Sprintf("applied: %d", len(status.Applied)),
		fmt.Sprintf("pending: %d", len(status.Pending)),
		fmt.Sprintf("unknown: %d", len(status.Unknown)),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write migration status: %w", err)
		}
	}

	err := printAppliedMigrations(stdout, status.Applied)
	if err != nil {
		return err
	}

	err = printPendingMigrations(stdout, status.Pending)
	if err != nil {
		return err
	}

	return printUnknownMigrations(stdout, status.Unknown)
}

func printAppliedMigrations(stdout io.Writer, migrations []chc.AppliedMigration) error {
	for _, migration := range migrations {
		_, err := fmt.Fprintf(
			stdout,
			"applied_migration: %s %s %s\n",
			migration.Version,
			migration.Name,
			migration.AppliedAt,
		)
		if err != nil {
			return fmt.Errorf("write applied migration: %w", err)
		}
	}

	return nil
}

func printPendingMigrations(stdout io.Writer, migrations []chc.Migration) error {
	for _, migration := range migrations {
		_, err := fmt.Fprintf(stdout, "pending_migration: %s %s\n", migration.Version, migration.Name)
		if err != nil {
			return fmt.Errorf("write pending migration: %w", err)
		}
	}

	return nil
}

func printUnknownMigrations(stdout io.Writer, migrations []chc.AppliedMigration) error {
	for _, migration := range migrations {
		_, err := fmt.Fprintf(
			stdout,
			"unknown_migration: %s %s %s\n",
			migration.Version,
			migration.Name,
			migration.AppliedAt,
		)
		if err != nil {
			return fmt.Errorf("write unknown migration: %w", err)
		}
	}

	return nil
}

func printProjectRegistryDiagnostics(
	stdout io.Writer,
	diagnostics []chc.ProjectRegistryDiagnostic,
) error {
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

func printExecutionArtifactDriftDiagnostics(
	stdout io.Writer,
	diagnostics []chc.ExecutionArtifactDriftDiagnostic,
) error {
	drifted := 0
	missingArtifacts := 0

	for _, diagnostic := range diagnostics {
		if diagnostic.Status != chc.ExecutionArtifactDriftStatusOK {
			drifted++
		}

		for _, artifact := range diagnostic.Artifacts {
			if !artifact.Exists {
				missingArtifacts++
			}
		}
	}

	lines := []string{
		fmt.Sprintf("executions: %d", len(diagnostics)),
		fmt.Sprintf("drifted: %d", drifted),
		fmt.Sprintf("missing_artifacts: %d", missingArtifacts),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write execution artifact drift summary: %w", err)
		}
	}

	for _, diagnostic := range diagnostics {
		err := printExecutionArtifactDriftDiagnostic(stdout, diagnostic)
		if err != nil {
			return err
		}
	}

	return nil
}

func printExecutionArtifactDriftDiagnostic(
	stdout io.Writer,
	diagnostic chc.ExecutionArtifactDriftDiagnostic,
) error {
	_, err := fmt.Fprintf(
		stdout,
		"execution: %s ticket: %s status: %s artifact_status: %s\n",
		diagnostic.Execution.ID,
		diagnostic.Execution.TicketKey,
		diagnostic.Execution.Status,
		diagnostic.Status,
	)
	if err != nil {
		return fmt.Errorf("write execution artifact drift diagnostic: %w", err)
	}

	for _, artifact := range diagnostic.Artifacts {
		_, err = fmt.Fprintf(
			stdout,
			"artifact: %s name: %s exists: %t path: %s\n",
			diagnostic.Execution.ID,
			artifact.Name,
			artifact.Exists,
			artifact.Path,
		)
		if err != nil {
			return fmt.Errorf("write execution artifact path diagnostic: %w", err)
		}
	}

	return nil
}

func printProjectRegistryDisable(stdout io.Writer, project chc.Project, dryRun bool) error {
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

func printProjectRegistrySetPath(
	stdout io.Writer,
	project chc.Project,
	previousPath string,
	dryRun bool,
) error {
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
