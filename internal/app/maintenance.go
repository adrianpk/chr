package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) maintenanceCommand() *cli.Command {
	return &cli.Command{
		Name:  "maintenance",
		Usage: "Inspect and recover local chr state",
		Commands: []*cli.Command{
			factory.maintenanceBundleCommand(),
			factory.maintenanceDBCommand(),
			factory.maintenanceExecutionCommand(),
			factory.maintenanceRegistryCommand(),
		},
	}
}

func (factory commandFactory) maintenanceBundleCommand() *cli.Command {
	return &cli.Command{
		Name:   "bundle",
		Usage:  "Write a local diagnostic bundle for the resolved project",
		Flags:  maintenanceProjectFlags(),
		Action: factory.runMaintenanceBundle,
	}
}

func (factory commandFactory) maintenanceDBCommand() *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "Inspect local database state",
		Commands: []*cli.Command{
			factory.maintenanceDBStatusCommand(),
		},
	}
}

func (factory commandFactory) maintenanceDBStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Report local database migration status",
		Flags:  projectRegistryFlags(),
		Action: factory.runMaintenanceDBStatus,
	}
}

func (factory commandFactory) maintenanceExecutionCommand() *cli.Command {
	return &cli.Command{
		Name:  "execution",
		Usage: "Inspect execution records and artifacts",
		Commands: []*cli.Command{
			factory.maintenanceExecutionArtifactsCommand(),
			factory.maintenanceExecutionCleanupPreviewCommand(),
		},
	}
}

func (factory commandFactory) maintenanceExecutionArtifactsCommand() *cli.Command {
	return &cli.Command{
		Name:   "artifacts",
		Usage:  "Report execution artifact drift for the resolved project",
		Flags:  maintenanceProjectFlags(),
		Action: factory.runMaintenanceExecutionArtifacts,
	}
}

func (factory commandFactory) maintenanceExecutionCleanupPreviewCommand() *cli.Command {
	return &cli.Command{
		Name:   "cleanup-preview",
		Usage:  "Preview unreferenced runtime artifact directories",
		Flags:  maintenanceProjectFlags(),
		Action: factory.runMaintenanceExecutionCleanupPreview,
	}
}

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

func maintenanceProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}

func (factory commandFactory) runMaintenanceBundle(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance bundle does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:    cmd.String("dir"),
		UseExplicitDir: true,
	})
	if err != nil {
		return err
	}

	bundle, err := collectDiagnosticBundle(ctx, env, projectContext, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	path, err := core.WriteDiagnosticBundle(ctx, projectContext.ProjectDir, bundle)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"bundle_path: %s\n", path)
	if err != nil {
		return fmt.Errorf("write diagnostic bundle result: %w", err)
	}

	return nil
}

func (factory commandFactory) runMaintenanceDBStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance db status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	status, err := core.NewMigrationStatusChecker(factory.deps.assetsFS).Check(ctx, env)
	if err != nil {
		return err
	}

	return printMigrationStatus(factory.io.Stdout, status)
}

func collectDiagnosticBundle(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) (core.DiagnosticBundle, error) {
	migrationStatus, err := core.NewMigrationStatusChecker(assetsFS).Check(ctx, env)
	if err != nil {
		return core.DiagnosticBundle{}, err
	}

	registryDiagnostics, err := collectDiagnosticBundleRegistry(ctx, env, assetsFS)
	if err != nil {
		return core.DiagnosticBundle{}, err
	}

	repository, err := core.NewExecutionRepository(ctx, env, assetsFS)
	if err != nil {
		return core.DiagnosticBundle{}, err
	}

	defer func() {
		//nolint:errcheck // Bundle result already comes from the main operation.
		_ = repository.Close()
	}()

	executionArtifacts, err := core.NewExecutionArtifactDriftService(repository).ListDiagnostics(
		ctx,
		projectContext.ProjectID,
	)
	if err != nil {
		return core.DiagnosticBundle{}, err
	}

	cleanupPreview, err := core.NewRuntimeArtifactCleanupPreviewService(repository).Preview(
		ctx,
		projectContext.ProjectDir,
		projectContext.ProjectID,
	)
	if err != nil {
		return core.DiagnosticBundle{}, err
	}

	reports, err := core.NewReportLister().ListProjectReports(ctx, projectContext.ProjectDir, core.ReportListFilter{})
	if err != nil {
		return core.DiagnosticBundle{}, err
	}

	return core.DiagnosticBundle{
		ProjectDir:             projectContext.ProjectDir,
		ProjectID:              projectContext.ProjectID,
		MigrationStatus:        migrationStatus,
		RegistryDiagnostics:    registryDiagnostics,
		ExecutionArtifactDrift: executionArtifacts,
		RuntimeCleanupPreview:  cleanupPreview,
		Reports:                reports,
	}, nil
}

func collectDiagnosticBundleRegistry(
	ctx context.Context,
	env core.UserEnvironment,
	assetsFS fs.FS,
) ([]core.ProjectRegistryDiagnostic, error) {
	registry, err := core.NewProjectRegistry(ctx, env, assetsFS)
	if err != nil {
		return nil, err
	}

	defer func() {
		//nolint:errcheck // Bundle result already comes from the main operation.
		_ = registry.Close()
	}()

	diagnostics, err := core.NewProjectRegistryDiagnosticsService(
		registry,
		core.NewActiveProjectStore(env),
	).ListDiagnostics(ctx)
	if err != nil {
		return nil, err
	}

	return diagnostics, nil
}

func (factory commandFactory) runMaintenanceExecutionArtifacts(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance execution artifacts does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:    cmd.String("dir"),
		UseExplicitDir: true,
	})
	if err != nil {
		return err
	}

	repository, err := core.NewExecutionRepository(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	diagnostics, err := core.NewExecutionArtifactDriftService(repository).ListDiagnostics(
		ctx,
		projectContext.ProjectID,
	)
	if err != nil {
		return err
	}

	return printExecutionArtifactDriftDiagnostics(factory.io.Stdout, diagnostics)
}

func (factory commandFactory) runMaintenanceExecutionCleanupPreview(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance execution cleanup-preview does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:    cmd.String("dir"),
		UseExplicitDir: true,
	})
	if err != nil {
		return err
	}

	repository, err := core.NewExecutionRepository(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	preview, err := core.NewRuntimeArtifactCleanupPreviewService(repository).Preview(
		ctx,
		projectContext.ProjectDir,
		projectContext.ProjectID,
	)
	if err != nil {
		return err
	}

	return printRuntimeArtifactCleanupPreview(factory.io.Stdout, preview)
}

func (factory commandFactory) runMaintenanceRegistryStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("maintenance registry status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	registry, err := core.NewProjectRegistry(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = registry.Close()
	}()

	diagnostics, err := core.NewProjectRegistryDiagnosticsService(
		registry,
		core.NewActiveProjectStore(env),
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

	registry, err := core.NewProjectRegistry(ctx, env, factory.deps.assetsFS)
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

	registry, err := core.NewProjectRegistry(ctx, env, factory.deps.assetsFS)
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

func printMigrationStatus(stdout io.Writer, status core.MigrationStatus) error {
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

func printAppliedMigrations(stdout io.Writer, migrations []core.AppliedMigration) error {
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

func printPendingMigrations(stdout io.Writer, migrations []core.Migration) error {
	for _, migration := range migrations {
		_, err := fmt.Fprintf(stdout, "pending_migration: %s %s\n", migration.Version, migration.Name)
		if err != nil {
			return fmt.Errorf("write pending migration: %w", err)
		}
	}

	return nil
}

func printUnknownMigrations(stdout io.Writer, migrations []core.AppliedMigration) error {
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
	diagnostics []core.ProjectRegistryDiagnostic,
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
	diagnostics []core.ExecutionArtifactDriftDiagnostic,
) error {
	drifted := 0
	missingArtifacts := 0

	for _, diagnostic := range diagnostics {
		if diagnostic.Status != core.ExecutionArtifactDriftStatusOK {
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

func printRuntimeArtifactCleanupPreview(
	stdout io.Writer,
	preview core.RuntimeArtifactCleanupPreview,
) error {
	lines := []string{
		fmt.Sprintf("artifact_dir: %s", preview.ArtifactDir),
		fmt.Sprintf("referenced_run_dirs: %d", len(preview.ReferencedRunDirs)),
		fmt.Sprintf("cleanup_candidates: %d", len(preview.Candidates)),
		"mode: preview",
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write runtime artifact cleanup preview summary: %w", err)
		}
	}

	for _, candidate := range preview.Candidates {
		_, err := fmt.Fprintf(stdout, "candidate: %s\n", candidate.Path)
		if err != nil {
			return fmt.Errorf("write runtime artifact cleanup candidate: %w", err)
		}
	}

	return nil
}

func printExecutionArtifactDriftDiagnostic(
	stdout io.Writer,
	diagnostic core.ExecutionArtifactDriftDiagnostic,
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

func printProjectRegistryDisable(stdout io.Writer, project core.Project, dryRun bool) error {
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
	project core.Project,
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
