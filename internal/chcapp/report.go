package chcapp

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

func newReportCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "report",
		Usage: "Inspect project reports",
		Commands: []*cli.Command{
			newReportListCommand(stdout),
		},
	}
}

func newReportListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List reports for the resolved project",
		Flags:  reportListFlags(),
		Action: reportListAction(stdout),
	}
}

func reportListFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "kind",
			Usage: "report kind: execution or loop",
		},
		&cli.StringFlag{
			Name:  "ticket",
			Usage: "ticket key filter",
		},
		&cli.StringFlag{
			Name:  "execution",
			Usage: "execution id filter",
		},
		&cli.StringFlag{
			Name:  "loop",
			Usage: "loop id filter",
		},
		&cli.StringFlag{
			Name:  "status",
			Usage: "report status filter",
		},
	)
}

func reportListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runReportList(ctx, cmd, stdout)
	}
}

func runReportList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("report list does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	filter, err := reportListFilterFromCommand(cmd)
	if err != nil {
		return err
	}

	reports, err := chc.NewReportLister().ListProjectReports(ctx, projectContext.ProjectDir, filter)
	if err != nil {
		return err
	}

	return printReportList(stdout, reports)
}

func reportListFilterFromCommand(cmd *cli.Command) (chc.ReportListFilter, error) {
	kind, err := reportKindFromString(cmd.String("kind"))
	if err != nil {
		return chc.ReportListFilter{}, err
	}

	return chc.ReportListFilter{
		Kind:        kind,
		TicketKey:   cmd.String("ticket"),
		ExecutionID: cmd.String("execution"),
		LoopID:      cmd.String("loop"),
		Status:      cmd.String("status"),
	}, nil
}

func reportKindFromString(value string) (chc.ReportKind, error) {
	switch value {
	case "":
		return "", nil
	case string(chc.ReportKindExecution):
		return chc.ReportKindExecution, nil
	case string(chc.ReportKindLoop):
		return chc.ReportKindLoop, nil
	default:
		return "", fmt.Errorf("unsupported report kind: %s", value)
	}
}

func printReportList(stdout io.Writer, reports []chc.ReportListEntry) error {
	for _, report := range reports {
		_, err := fmt.Fprintf(
			stdout,
			"kind: %s path: %s ticket: %s execution: %s loop: %s status: %s\n",
			report.Kind,
			report.RelativePath,
			emptyReportListField(report.TicketKey),
			emptyReportListField(report.ExecutionID),
			emptyReportListField(report.LoopID),
			emptyReportListField(report.Status),
		)
		if err != nil {
			return fmt.Errorf("write report list output: %w", err)
		}
	}

	return nil
}

func emptyReportListField(value string) string {
	if value == "" {
		return "none"
	}

	return value
}
