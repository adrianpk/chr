package app

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) maintenanceBundleCommand() *cli.Command {
	return &cli.Command{
		Name:   "bundle",
		Usage:  "Write a local diagnostic bundle for the resolved project",
		Flags:  maintenanceProjectFlags(),
		Action: factory.runMaintenanceBundle,
	}
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
