package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) ticketLoopCommand() *cli.Command {
	return &cli.Command{
		Name:   "loop",
		Usage:  "Preview a foreground loop over eligible tickets",
		Flags:  ticketLoopFlags(),
		Action: factory.runTicketLoop,
	}
}

func ticketLoopFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the planned loop selection without executing tickets",
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
		&cli.IntFlag{
			Name:  "max-iterations",
			Usage: "maximum tickets to execute in one loop run",
		},
	)
}

func (factory commandFactory) runTicketLoop(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket loop does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	project := operations.projectContext

	if err != nil {
		return err
	}

	request := ticketLoopRequest{Query: ticketLoopQueryOptionsFromCommand(cmd, project.ProjectID), Run: ticketLoopRunOptions{MaxIterations: cmd.Int("max-iterations")}, DryRun: cmd.Bool("dry-run")}

	outcome, err := operations.runLegacyTicketLoop(ctx, request)
	if request.DryRun {
		if err != nil {
			return err
		}

		return printTicketLoopPlan(factory.io.Stdout, project, request.Query, outcome.Selection.Tickets, true)
	}

	if outcome.Execution.Record.ID != "" {
		if printErr := printTicketLoopExecution(factory.io.Stdout, outcome.Execution); printErr != nil {
			return printErr
		}
	}

	return err
}

func ticketLoopQueryOptionsFromCommand(cmd *cli.Command, projectID string) core.TicketLoopQueryOptions {
	return core.TicketLoopQueryOptions{
		ProjectID: projectID,
		Statuses:  cmd.StringSlice("status"),
		Filters: core.TicketLoopFilters{
			Priorities:   cmd.StringSlice("priority"),
			Severities:   cmd.StringSlice("severity"),
			ProjectAreas: cmd.StringSlice("project-area"),
			Tags:         cmd.StringSlice("tag"),
			Types:        cmd.StringSlice("type"),
			UseCases:     cmd.StringSlice("use-case"),
		},
		Order: core.TicketLoopOrder(cmd.String("order")),
	}
}

func printTicketLoopPlan(stdout io.Writer, projectContext commandProjectContext, opts core.TicketLoopQueryOptions, tickets []core.IndexedTicket, dryRun bool) error {
	order := opts.Order
	if order == "" {
		order = core.TicketLoopOrderCreatedAt
	}

	lines := []string{
		"loop: planned",
		fmt.Sprintf("project: %s", projectContext.ProjectID),
		fmt.Sprintf("dry_run: %t", dryRun),
		fmt.Sprintf("order: %s", order),
		fmt.Sprintf("eligible: %d", len(tickets)),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket loop result: %w", err)
		}
	}

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
			return fmt.Errorf("write ticket loop result: %w", err)
		}
	}

	return nil
}

func printTicketLoopExecution(stdout io.Writer, result ticketLoopExecutionResult) error {
	skippedCount := result.Record.EligibleCount - result.Record.ProcessedCount - result.Record.FailedCount
	if skippedCount < 0 {
		skippedCount = 0
	}

	lines := []string{
		fmt.Sprintf("loop: %s", result.Record.ID),
		fmt.Sprintf("status: %s", result.Record.Status),
		fmt.Sprintf("eligible: %d", result.Record.EligibleCount),
		fmt.Sprintf("processed: %d", result.Record.ProcessedCount),
		fmt.Sprintf("skipped: %d", skippedCount),
		fmt.Sprintf("blocked: %d", result.Record.BlockedCount),
		fmt.Sprintf("failed: %d", result.Record.FailedCount),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket loop execution result: %w", err)
		}
	}

	for _, execution := range result.Executions {
		_, err := fmt.Fprintf(stdout, "execution: %s\t%s\t%s\n", execution.ID, execution.TicketKey, execution.Status)
		if err != nil {
			return fmt.Errorf("write ticket loop execution result: %w", err)
		}
	}

	return nil
}
