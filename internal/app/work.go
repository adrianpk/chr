package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/workspace"
)

func newWorkCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "work",
		Usage: "Run agent work for tickets",
		Commands: []*cli.Command{
			newWorkTicketCommand(stdout, deps),
			newWorkRunAllCommand(stdout, deps),
		},
	}
}

func newWorkTicketCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:      "ticket",
		Usage:     "Run agent work for one ticket",
		ArgsUsage: "<ticket-key>",
		Flags:     workProjectFlags(),
		Action:    workTicketAction(stdout, deps),
	}
}

func newWorkRunAllCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:   "run-all",
		Usage:  "Run agent work for all ready tickets",
		Flags:  workRunAllFlags(),
		Action: workRunAllAction(stdout, deps),
	}
}

func workProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}

func workRunAllFlags() []cli.Flag {
	return append(workProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the ready-ticket queue without executing agent work",
		},
		&cli.StringSliceFlag{
			Name:  "status",
			Usage: "eligible ticket status",
		},
		&cli.StringSliceFlag{
			Name:  "priority",
			Usage: "ticket priority filter",
		},
		&cli.StringSliceFlag{
			Name:  "severity",
			Usage: "ticket severity filter",
		},
		&cli.StringSliceFlag{
			Name:  "project-area",
			Usage: "ticket project area filter",
		},
		&cli.StringSliceFlag{
			Name:  "tag",
			Usage: "ticket tag filter",
		},
		&cli.StringSliceFlag{
			Name:  "type",
			Usage: "ticket type filter",
		},
		&cli.StringSliceFlag{
			Name:  "use-case",
			Usage: "ticket use case filter",
		},
		&cli.StringFlag{
			Name:  "order",
			Usage: "ticket order: key, created_at, priority, severity, priority_severity",
		},
	)
}

func workTicketAction(stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() != 1 {
			return fmt.Errorf("work ticket requires exactly one ticket key argument")
		}

		operations, _, err := workServiceFromCommand(ctx, cmd, deps)
		if err != nil {
			return err
		}

		result, err := operations.StartAgentWorkTicket(cmd.Args().Get(0))
		if err != nil {
			return err
		}

		return printAgentWorkTicket(stdout, result)
	}
}

func workRunAllAction(stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		operations, projectContext, err := workServiceFromCommand(ctx, cmd, deps)
		if err != nil {
			return err
		}

		selection := agentWorkSelectionOptionsFromCommand(cmd)
		if cmd.Bool("dry-run") {
			result, err := operations.PreviewAgentWorkLoop(selection)
			if err != nil {
				return err
			}

			return printAgentWorkRunAllPreview(stdout, projectContext, selection, result)
		}

		result, err := operations.StartAgentWorkLoop(selection)
		if err != nil {
			return err
		}

		return printAgentWorkRunAllExecution(stdout, result)
	}
}

func workServiceFromCommand(
	ctx context.Context,
	cmd *cli.Command,
	deps appDependencies,
) (service, commandProjectContext, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return service{}, commandProjectContext{}, err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return service{}, commandProjectContext{}, err
	}

	return newService(ctx, env, projectContext, deps.assetsFS), projectContext, nil
}

func agentWorkSelectionOptionsFromCommand(cmd *cli.Command) AgentWorkSelectionOptions {
	return AgentWorkSelectionOptions{
		Statuses: cmd.StringSlice("status"),
		Filters: workspace.TicketLoopFilters{
			Priorities:   cmd.StringSlice("priority"),
			Severities:   cmd.StringSlice("severity"),
			ProjectAreas: cmd.StringSlice("project-area"),
			Tags:         cmd.StringSlice("tag"),
			Types:        cmd.StringSlice("type"),
			UseCases:     cmd.StringSlice("use-case"),
		},
		Order: workspace.TicketLoopOrder(cmd.String("order")),
	}
}

func printAgentWorkTicket(stdout io.Writer, result AgentWorkResult) error {
	lines := []string{
		"work: ticket",
		fmt.Sprintf("message: %s", result.Message),
	}

	if result.Execution != nil {
		lines = append(lines,
			fmt.Sprintf("ticket: %s", result.Execution.TicketKey),
			fmt.Sprintf("execution: %s", result.Execution.ID),
			fmt.Sprintf("status: %s", result.Execution.Status),
		)
	}

	return printAgentWorkLines(stdout, lines)
}

func printAgentWorkRunAllPreview(
	stdout io.Writer,
	projectContext commandProjectContext,
	selection AgentWorkSelectionOptions,
	result AgentWorkResult,
) error {
	lines := []string{
		"work: run-all",
		fmt.Sprintf("project: %s", projectContext.ProjectID),
		"dry_run: true",
		fmt.Sprintf("order: %s", agentWorkOrder(selection.Order)),
		fmt.Sprintf("eligible: %d", result.EligibleCount),
	}

	if err := printAgentWorkLines(stdout, lines); err != nil {
		return err
	}

	return printAgentWorkTicketQueue(stdout, result.SelectedQueue)
}

func printAgentWorkRunAllExecution(stdout io.Writer, result AgentWorkResult) error {
	lines := []string{
		"work: run-all",
		fmt.Sprintf("message: %s", result.Message),
	}

	if result.LoopRun != nil {
		skippedCount := result.LoopRun.EligibleCount - result.LoopRun.ProcessedCount - result.LoopRun.FailedCount
		if skippedCount < 0 {
			skippedCount = 0
		}

		lines = append(lines,
			fmt.Sprintf("loop: %s", result.LoopRun.ID),
			fmt.Sprintf("status: %s", result.LoopRun.Status),
			fmt.Sprintf("eligible: %d", result.LoopRun.EligibleCount),
			fmt.Sprintf("processed: %d", result.LoopRun.ProcessedCount),
			fmt.Sprintf("skipped: %d", skippedCount),
			fmt.Sprintf("blocked: %d", result.LoopRun.BlockedCount),
			fmt.Sprintf("failed: %d", result.LoopRun.FailedCount),
		)
	}

	if err := printAgentWorkLines(stdout, lines); err != nil {
		return err
	}

	for _, execution := range result.Executions {
		_, err := fmt.Fprintf(stdout, "execution: %s\t%s\t%s\n", execution.ID, execution.TicketKey, execution.Status)
		if err != nil {
			return fmt.Errorf("write agent work result: %w", err)
		}
	}

	return nil
}

func printAgentWorkTicketQueue(stdout io.Writer, tickets []workspace.IndexedTicket) error {
	for _, ticket := range tickets {
		_, err := fmt.Fprintf(
			stdout,
			"ticket: %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.Key,
			ticket.Title,
			ticket.Status,
			ticket.Priority,
			ticket.Severity,
			ticket.Type,
			ticket.UseCase,
			ticket.RelativePath,
		)
		if err != nil {
			return fmt.Errorf("write agent work result: %w", err)
		}
	}

	return nil
}

func printAgentWorkLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write agent work result: %w", err)
		}
	}

	return nil
}
