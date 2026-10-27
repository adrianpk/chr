package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

func (factory commandFactory) ticketApplyResultCommand() *cli.Command {
	return &cli.Command{
		Name:   "apply-result",
		Usage:  "Apply one completed execution result to its ticket",
		Flags:  ticketApplyResultFlags(),
		Action: factory.runTicketApplyResult,
	}
}

func ticketApplyResultFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the planned result application without writing ticket state",
		},
	)
}

func (factory commandFactory) runTicketApplyResult(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket apply-result requires exactly one execution id argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	plan, err := factory.buildTicketApplyResultPlan(ctx, cmd, env)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		err = newService(ctx, env, plan.ProjectContext, factory.deps.assetsFS).applyTicketResult(ctx, plan, ticketExecutionEvents{})
		if err != nil {
			return err
		}
	}

	return printTicketApplyResult(factory.io.Stdout, plan, cmd.Bool("dry-run"))
}

func (factory commandFactory) buildTicketApplyResultPlan(ctx context.Context, cmd *cli.Command, env core.UserEnvironment) (ticketApplyResultPlan, error) {
	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	repository, err := executioncore.NewRepository(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = repository.Close()
	}()

	record, err := repository.GetExecution(ctx, cmd.Args().First())
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	return buildTicketApplyResultPlanForRecord(ctx, projectContext, record)
}

func printTicketApplyResult(stdout io.Writer, plan ticketApplyResultPlan, dryRun bool) error {
	lines := []string{
		fmt.Sprintf("execution: %s", plan.Application.ExecutionID),
		fmt.Sprintf("ticket: %s", plan.Application.TicketKey),
		fmt.Sprintf("ticket_path: %s", plan.Application.TicketPath),
		fmt.Sprintf("status: %s", plan.Application.TicketStatus),
		fmt.Sprintf("result_status: %s", plan.Result.Status),
		fmt.Sprintf("execution_report_path: %s", executioncore.ReportPath(plan.Record)),
		fmt.Sprintf("report_path: %s", plan.ReportPath),
		fmt.Sprintf("dry_run: %t", dryRun),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket apply-result output: %w", err)
		}
	}

	return nil
}
