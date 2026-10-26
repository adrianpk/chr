package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

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
