package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func (factory commandFactory) runtimeCommand() *cli.Command {
	return &cli.Command{
		Name:  "runtime",
		Usage: "Manage project runtime assets",
		Commands: []*cli.Command{
			factory.runtimeInitCommand(),
			factory.runtimeValidateCommand(),
		},
	}
}

func (factory commandFactory) provingCommand() *cli.Command {
	return &cli.Command{
		Name:  "proving",
		Usage: "Run temporary MVP proving-ground workflows",
		Commands: []*cli.Command{
			factory.provingRunCommand(),
			factory.provingRalphCommand(),
			factory.provingValidateCommand(),
		},
	}
}

func (factory commandFactory) provingRunCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run one direct proving objective in the agent sandbox",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  "objective",
				Usage: "project-relative proving objective path",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print the proving sandbox plan without running Docker",
			},
		},
		Action: factory.runProvingRun,
	}
}

func (factory commandFactory) provingRalphCommand() *cli.Command {
	return &cli.Command{
		Name:  "ralph",
		Usage: "Run a bounded proving-ground Ralph Loop for one objective",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  "objective",
				Usage: "project-relative proving objective path",
			},
			&cli.IntFlag{
				Name:  "max-attempts",
				Usage: "maximum Ralph Loop attempts",
			},
		},
		Action: factory.runProvingRalph,
	}
}

func (factory commandFactory) provingValidateCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Run validation commands for one proving objective",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  "objective",
				Usage: "project-relative proving objective path",
			},
		},
		Action: factory.runProvingValidate,
	}
}

func (factory commandFactory) runtimeInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Install missing runtime defaults in the resolved project",
		Flags: append(projectRegistryFlags(),
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print planned changes without writing files",
			},
		),
		Action: factory.runRuntimeInit,
	}
}

func (factory commandFactory) runtimeValidateCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Validate the resolved project's runtime boundary",
		Flags: append(projectRegistryFlags(),
			&cli.StringFlag{
				Name:  "dir",
				Usage: "project directory override",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print the validation plan without running Docker",
			},
		),
		Action: factory.runRuntimeValidate,
	}
}

func (factory commandFactory) runRuntimeInit(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("runtime init does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	result, err := core.NewRuntimeAssets(factory.deps.assetsFS).Init(ctx, core.RuntimeInitOptions{
		ProjectDir: projectContext.ProjectDir,
		DryRun:     cmd.Bool("dry-run"),
	})
	if err != nil {
		return err
	}

	return printRuntimeInitResult(factory.io.Stdout, result)
}

func (factory commandFactory) runRuntimeValidate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("runtime validate does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	result, err := validateCodexRuntime(core.RuntimeValidationOptions{
		ProjectDir: projectContext.ProjectDir,
		Config:     config.Runtime,
	})
	if err != nil {
		return err
	}

	err = printRuntimeValidateResult(factory.io.Stdout, result, cmd.Bool("dry-run"))
	if err != nil {
		return err
	}

	if !core.RuntimeValidationPassed(result.Checks) {
		return fmt.Errorf("runtime validation failed")
	}

	return nil
}

func (factory commandFactory) runProvingRun(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("proving run does not accept positional arguments")
	}

	objectivePath := cmd.String("objective")
	if objectivePath == "" {
		return fmt.Errorf("proving run requires --objective")
	}

	result, err := core.RunProvingSandbox(ctx, core.ProvingSandboxRunOptions{
		Worker:        newCodexWorker(),
		ProjectDir:    cmd.String("dir"),
		ObjectivePath: objectivePath,
		DryRun:        cmd.Bool("dry-run"),
	})
	if err != nil {
		return err
	}

	return printProvingRunResult(factory.io.Stdout, result)
}

func (factory commandFactory) runProvingRalph(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("proving ralph does not accept positional arguments")
	}

	objectivePath := cmd.String("objective")
	if objectivePath == "" {
		return fmt.Errorf("proving ralph requires --objective")
	}

	result, err := core.RunProvingRalphLoop(ctx, core.ProvingRalphLoopOptions{
		Worker:        newCodexWorker(),
		ProjectDir:    cmd.String("dir"),
		ObjectivePath: objectivePath,
		MaxAttempts:   cmd.Int("max-attempts"),
	})
	if err != nil {
		return err
	}

	err = printProvingRalphResult(factory.io.Stdout, result)
	if err != nil {
		return err
	}

	if result.Status != core.ProvingRalphLoopStatusPassed {
		return fmt.Errorf("proving ralph loop failed")
	}

	return nil
}

func (factory commandFactory) runProvingValidate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("proving validate does not accept positional arguments")
	}

	objectivePath := cmd.String("objective")
	if objectivePath == "" {
		return fmt.Errorf("proving validate requires --objective")
	}

	result, err := core.RunProvingValidation(ctx, cmd.String("dir"), objectivePath)
	if err != nil {
		return err
	}

	err = printProvingValidationResult(factory.io.Stdout, result)
	if err != nil {
		return err
	}

	if !result.Passed {
		return fmt.Errorf("proving validation failed")
	}

	return nil
}

func printProvingRalphResult(stdout io.Writer, result core.ProvingRalphLoopResult) error {
	lines := []string{
		fmt.Sprintf("proving_ralph: %s", result.Status),
		fmt.Sprintf("objective_id: %s", result.Objective.ID),
		fmt.Sprintf("ticket_key: %s", result.Objective.TicketKey),
		fmt.Sprintf("attempts: %d", len(result.Attempts)),
		fmt.Sprintf("stop_reason: %s", result.StopReason),
		fmt.Sprintf("summary_path: %s", result.SummaryPath),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write proving ralph result: %w", err)
		}
	}

	for _, attempt := range result.Attempts {
		_, err := fmt.Fprintf(
			stdout,
			"attempt: %d objective_id=%s validation_passed=%t validation_path=%s\n",
			attempt.Number,
			attempt.ObjectiveID,
			attempt.ValidationPassed,
			attempt.ValidationPath,
		)
		if err != nil {
			return fmt.Errorf("write proving ralph attempt result: %w", err)
		}
	}

	return nil
}

func printProvingRunResult(stdout io.Writer, result core.ProvingSandboxRunResult) error {
	status := "completed"
	if result.DryRun {
		status = "planned"
	}

	lines := []string{
		fmt.Sprintf("proving: %s", status),
		fmt.Sprintf("dry_run: %t", result.DryRun),
		fmt.Sprintf("objective_id: %s", result.Objective.ID),
		fmt.Sprintf("ticket_key: %s", result.Objective.TicketKey),
		fmt.Sprintf("image: %s", result.Plan.Image),
		fmt.Sprintf("dockerfile: %s", result.Plan.DockerfilePath),
		fmt.Sprintf("context_path: %s", result.Plan.WorkerContextPath),
		fmt.Sprintf("result_path: %s", result.Plan.ResultPath),
		fmt.Sprintf("log_path: %s", result.Plan.LogPath),
		fmt.Sprintf("validation_path: %s", result.Plan.ValidationPath),
		fmt.Sprintf("build_command: %s", formatHarnessCommand(result.Build)),
		fmt.Sprintf("run_command: %s", formatHarnessCommand(result.Run)),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write proving result: %w", err)
		}
	}

	return nil
}

func printProvingValidationResult(stdout io.Writer, result core.ProvingValidationResult) error {
	lines := []string{
		fmt.Sprintf("proving_validation: %t", result.Passed),
		fmt.Sprintf("objective_id: %s", result.Objective.ID),
		fmt.Sprintf("ticket_key: %s", result.Objective.TicketKey),
		fmt.Sprintf("validation_path: %s", result.LogPath),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write proving validation result: %w", err)
		}
	}

	for _, command := range result.Commands {
		_, err := fmt.Fprintf(stdout, "command: %s exit_code=%d\n", command.Command, command.ExitCode)
		if err != nil {
			return fmt.Errorf("write proving validation command result: %w", err)
		}
	}

	return nil
}

func formatHarnessCommand(command core.HarnessCommand) string {
	parts := append([]string{command.Program}, command.Args...)

	return strings.Join(parts, " ")
}

func printRuntimeValidateResult(
	stdout io.Writer,
	result core.RuntimeValidationResult,
	dryRun bool,
) error {
	plan := result.Plan

	_, err := fmt.Fprintf(stdout, "runtime: %s\n", displayPath(plan.ProjectDir, workspacecore.UserPath(plan.ProjectDir, "runtime")))
	if err != nil {
		return fmt.Errorf("write runtime result: %w", err)
	}

	lines := []string{
		fmt.Sprintf("dry_run: %t", dryRun),
		fmt.Sprintf("workspace_mode: %s", plan.WorkspaceMode),
		fmt.Sprintf("project_checkout_dir: %s", displayPath(plan.ProjectDir, plan.ProjectCheckoutDir)),
		fmt.Sprintf("execution_workspace_dir: %s", displayPath(plan.ProjectDir, plan.ExecutionWorkspaceDir)),
		fmt.Sprintf("image: %s", plan.Image),
		fmt.Sprintf("dockerfile: %s", displayPath(plan.ProjectDir, plan.DockerfilePath)),
		fmt.Sprintf("workdir: %s", plan.ContainerWorkdir),
		fmt.Sprintf("artifact_dir: %s", displayPath(plan.ProjectDir, plan.ArtifactDir)),
		fmt.Sprintf("run_artifact_dir: %s", displayPath(plan.ProjectDir, plan.RunArtifactDir)),
		fmt.Sprintf("result_path: %s", displayPath(plan.ProjectDir, plan.ResultPath)),
		fmt.Sprintf("log_path: %s", displayPath(plan.ProjectDir, plan.LogPath)),
		fmt.Sprintf("validation_command: %s", plan.ValidationCommand),
	}

	for _, line := range lines {
		_, err = fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write runtime plan: %w", err)
		}
	}

	for _, mount := range plan.Mounts {
		_, err = fmt.Fprintf(stdout, "mount: %s -> %s read_only=%t\n", displayPath(plan.ProjectDir, mount.Source), mount.Target, mount.ReadOnly)
		if err != nil {
			return fmt.Errorf("write runtime mount: %w", err)
		}
	}

	return printRuntimeValidationChecks(stdout, plan.ProjectDir, result.Checks)
}

func printRuntimeValidationChecks(
	stdout io.Writer,
	projectDir string,
	checks []core.RuntimeValidationCheck,
) error {
	_, err := fmt.Fprintln(stdout, "checks:")
	if err != nil {
		return fmt.Errorf("write runtime checks header: %w", err)
	}

	for _, check := range checks {
		status := "fail"
		if check.OK {
			status = "ok"
		}

		message := check.Message
		if check.OK && message == "ok" {
			message = ""
		}

		if message == "" {
			_, err = fmt.Fprintf(stdout, "  %-4s  %-12s %s\n", status, check.Name, displayPath(projectDir, check.Path))
		} else {
			_, err = fmt.Fprintf(stdout, "  %-4s  %-12s %s  %s\n", status, check.Name, displayPath(projectDir, check.Path), message)
		}

		if err != nil {
			return fmt.Errorf("write runtime check: %w", err)
		}
	}

	return nil
}

func printRuntimeInitResult(stdout io.Writer, result core.RuntimeInitResult) error {
	_, err := fmt.Fprintf(stdout, "runtime: %s\n", result.Root)
	if err != nil {
		return fmt.Errorf("write runtime result: %w", err)
	}

	for _, created := range result.Created {
		_, err = fmt.Fprintf(stdout, "create %s\n", created)
		if err != nil {
			return fmt.Errorf("write created path: %w", err)
		}
	}

	for _, skipped := range result.Skipped {
		_, err = fmt.Fprintf(stdout, "skip existing %s\n", skipped)
		if err != nil {
			return fmt.Errorf("write skipped path: %w", err)
		}
	}

	_, err = fmt.Fprintln(stdout, "done")
	if err != nil {
		return fmt.Errorf("write completion result: %w", err)
	}

	return nil
}
