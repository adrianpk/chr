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
			newReportShowCommand(stdout),
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

func newReportShowCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show one project report",
		ArgsUsage: "[report-id]",
		Flags:     reportShowFlags(),
		Action:    reportShowAction(stdout),
	}
}

func reportShowFlags() []cli.Flag {
	return append(reportListFlags(),
		&cli.BoolFlag{
			Name:  "metadata",
			Usage: "print report metadata instead of body",
		},
	)
}

func reportListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runReportList(ctx, cmd, stdout)
	}
}

func reportShowAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runReportShow(ctx, cmd, stdout)
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

func runReportShow(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() > 1 {
		return fmt.Errorf("report show accepts at most one positional argument")
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

	request, err := reportLookupRequestFromCommand(cmd)
	if err != nil {
		return err
	}

	result, err := chc.NewReportLookupService().LookupReport(ctx, projectContext.ProjectDir, request)
	if err != nil {
		return err
	}

	if cmd.Bool("metadata") {
		return printReportList(stdout, []chc.ReportListEntry{result.Entry})
	}

	_, err = fmt.Fprint(stdout, result.Body)
	if err != nil {
		return fmt.Errorf("write report body: %w", err)
	}

	return nil
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

func reportLookupRequestFromCommand(cmd *cli.Command) (chc.ReportLookupRequest, error) {
	kind, err := reportKindFromString(cmd.String("kind"))
	if err != nil {
		return chc.ReportLookupRequest{}, err
	}

	reportID := ""
	if cmd.NArg() == 1 {
		reportID = cmd.Args().First()
	}

	request := chc.ReportLookupRequest{
		ReportID:    reportID,
		Kind:        kind,
		TicketKey:   cmd.String("ticket"),
		ExecutionID: cmd.String("execution"),
		LoopID:      cmd.String("loop"),
		Status:      cmd.String("status"),
	}

	if reportLookupRequestIsEmpty(request) {
		return chc.ReportLookupRequest{}, fmt.Errorf("report show requires a report id or lookup filter")
	}

	return request, nil
}

func reportLookupRequestIsEmpty(request chc.ReportLookupRequest) bool {
	return request.ReportID == "" &&
		request.Kind == "" &&
		request.TicketKey == "" &&
		request.ExecutionID == "" &&
		request.LoopID == "" &&
		request.Status == ""
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
