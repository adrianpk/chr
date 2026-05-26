package chcapp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/chc"
)

func newTicketCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "ticket",
		Usage: "Manage workspace tickets",
		Commands: []*cli.Command{
			newTicketApplyResultCommand(stdout),
			newTicketCommentCommand(stdout),
			newTicketListCommand(stdout),
			newTicketLoopCommand(stdout),
			newTicketNewCommand(stdout),
			newTicketShowCommand(stdout),
			newTicketStartCommand(stdout),
			newTicketStatusCommand(stdout),
		},
	}
}

func newTicketNewCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create a ticket in the current workspace",
		Flags:  ticketNewFlags(),
		Action: ticketNewAction(stdout),
	}
}

func newTicketListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List tickets for the resolved project",
		Flags:  ticketProjectFlags(),
		Action: ticketListAction(stdout),
	}
}

func newTicketLoopCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "loop",
		Usage:  "Preview a foreground loop over eligible tickets",
		Flags:  ticketLoopFlags(),
		Action: ticketLoopAction(stdout),
	}
}

func newTicketShowCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show one ticket by key",
		Flags:  ticketProjectFlags(),
		Action: ticketShowAction(stdout),
	}
}

func newTicketCommentCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "comment",
		Usage:  "Append a comment to one ticket",
		Flags:  ticketCommentFlags(),
		Action: ticketCommentAction(stdout),
	}
}

func newTicketApplyResultCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "apply-result",
		Usage:  "Apply one completed execution result to its ticket",
		Flags:  ticketApplyResultFlags(),
		Action: ticketApplyResultAction(stdout),
	}
}

func newTicketStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Update one ticket status",
		Flags:  ticketProjectFlags(),
		Action: ticketStatusAction(stdout),
	}
}

func newTicketStartCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Prepare one ticket execution",
		Flags:  ticketStartFlags(),
		Action: ticketStartAction(stdout),
	}
}

func ticketProjectFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
	)
}

func ticketStartFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the execution preflight without creating an execution record",
		},
	)
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

func ticketApplyResultFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the planned result application without writing ticket state",
		},
	)
}

func ticketCommentFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "author",
			Usage: "comment author",
			Value: "user",
		},
	)
}

func ticketNewFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "priority",
			Usage: "ticket priority",
		},
		&cli.StringFlag{
			Name:  "synopsis",
			Usage: "ticket synopsis",
		},
		&cli.StringFlag{
			Name:  "severity",
			Usage: "ticket severity",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "ticket type",
		},
		&cli.StringSliceFlag{
			Name:  "tag",
			Usage: "ticket tag",
		},
		&cli.StringFlag{
			Name:  "use-case",
			Usage: "ticket use case",
		},
		&cli.StringFlag{
			Name:  "project-area",
			Usage: "ticket project area",
		},
	)
}

func ticketListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketList(ctx, cmd, stdout)
	}
}

func ticketShowAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketShow(ctx, cmd, stdout)
	}
}

func ticketCommentAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketComment(ctx, cmd, stdout)
	}
}

func ticketStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketStatus(ctx, cmd, stdout)
	}
}

func ticketStartAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketStart(ctx, cmd, stdout)
	}
}

func ticketLoopAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketLoop(ctx, cmd, stdout)
	}
}

func ticketApplyResultAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketApplyResult(ctx, cmd, stdout)
	}
}

type ticketStartPreflight struct {
	AssetsFS fs.FS
	Env      chc.UserEnvironment
	Ticket   chc.IndexedTicket
	Record   chc.ExecutionRecord
	Runtime  chc.RuntimeValidationResult
}

type ticketStartSelection struct {
	ProjectContext commandProjectContext
	Ticket         chc.IndexedTicket
}

type ticketApplyResultPlan struct {
	ProjectContext commandProjectContext
	Record         chc.ExecutionRecord
	Result         chc.ExecutionResult
	Application    chc.ResultApplication
	ReportPath     string
}

type ticketLoopExecutionResult struct {
	Record       chc.LoopRunRecord
	Executions   []chc.ExecutionRecord
	Applications []chc.ResultApplication
}

type ticketLoopRunOptions struct {
	MaxIterations int
}

var newTicketStartHarnessExecutor = func() chc.HarnessExecutor {
	return chc.OSHarnessExecutor{}
}

func runTicketStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("ticket status requires exactly one ticket key and one status argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	status := cmd.Args().Get(1)

	result, err := newService(ctx, env, projectContext, deps.assetsFS).MoveTicketStatus(cmd.Args().Get(0), status)
	if err != nil {
		return err
	}

	updated, err := ticketOperationResultTicket(result, cmd.Args().Get(0))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "ticket: %s %s\n", updated.Key, status)
	if err != nil {
		return fmt.Errorf("write ticket status result: %w", err)
	}

	return nil
}

func ticketOperationResultTicket(result TicketOperationResult, key string) (chc.IndexedTicket, error) {
	for _, ticket := range result.Tickets {
		if ticket.Key == key {
			return ticket, nil
		}
	}

	return chc.IndexedTicket{}, fmt.Errorf("ticket not found after operation: %s", key)
}

func runTicketStart(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket start requires exactly one ticket key argument")
	}

	preflight, err := buildTicketStartPreflight(ctx, cmd)
	if err != nil {
		return chc.NewExecutionFailure(chc.ExecutionFailurePreflight, err)
	}

	if !chc.RuntimeValidationPassed(preflight.Runtime.Checks) {
		preflight.Record.Status = chc.ExecutionStatusBlocked
		preflight.Record.FailureReason = chc.NewExecutionFailure(
			chc.ExecutionFailureRuntime,
			fmt.Errorf("runtime validation failed"),
		).Error()

		err = printTicketStartResult(stdout, preflight.Ticket, preflight.Record, preflight.Runtime, cmd.Bool("dry-run"))
		if err != nil {
			return err
		}

		return chc.NewExecutionFailure(chc.ExecutionFailureRuntime, fmt.Errorf("runtime validation failed"))
	}

	if !cmd.Bool("dry-run") {
		record, err := runAndPrintTicketStartExecution(ctx, stdout, preflight)
		if err != nil {
			return err
		}

		preflight.Record = record
	}

	return printTicketStartResult(stdout, preflight.Ticket, preflight.Record, preflight.Runtime, cmd.Bool("dry-run"))
}

func runTicketLoop(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket loop does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, index, err := openTicketIndexForCommand(ctx, cmd, env)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return err
	}

	queryOptions := ticketLoopQueryOptionsFromCommand(cmd, projectContext.ProjectID)

	runOptions, err := ticketLoopRunOptionsFromCommand(cmd)
	if err != nil {
		return err
	}

	tickets, err := chc.NewTicketLoopQuery(index).EligibleTickets(ctx, queryOptions)
	if err != nil {
		return err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, runOptions)

	if !cmd.Bool("dry-run") {
		deps, err := appDependenciesFromCommand(cmd)
		if err != nil {
			return err
		}

		return runAndPrintTicketLoopExecution(ctx, stdout, env, projectContext, tickets, len(eligibleTickets), deps.assetsFS)
	}

	return printTicketLoopPlan(stdout, projectContext, queryOptions, tickets, cmd.Bool("dry-run"))
}

func ticketLoopQueryOptionsFromCommand(cmd *cli.Command, projectID string) chc.TicketLoopQueryOptions {
	return chc.TicketLoopQueryOptions{
		ProjectID: projectID,
		Statuses:  cmd.StringSlice("status"),
		Filters: chc.TicketLoopFilters{
			Priorities:   cmd.StringSlice("priority"),
			Severities:   cmd.StringSlice("severity"),
			ProjectAreas: cmd.StringSlice("project-area"),
			Tags:         cmd.StringSlice("tag"),
			Types:        cmd.StringSlice("type"),
			UseCases:     cmd.StringSlice("use-case"),
		},
		Order: chc.TicketLoopOrder(cmd.String("order")),
	}
}

func ticketLoopRunOptionsFromCommand(cmd *cli.Command) (ticketLoopRunOptions, error) {
	opts := ticketLoopRunOptions{
		MaxIterations: cmd.Int("max-iterations"),
	}
	if opts.MaxIterations < 0 {
		return ticketLoopRunOptions{}, fmt.Errorf("max-iterations must be greater than or equal to zero")
	}

	return opts, nil
}

func ticketsForTicketLoopRun(tickets []chc.IndexedTicket, opts ticketLoopRunOptions) []chc.IndexedTicket {
	if opts.MaxIterations == 0 || len(tickets) <= opts.MaxIterations {
		return tickets
	}

	return tickets[:opts.MaxIterations]
}

func deduplicateTicketLoopTickets(tickets []chc.IndexedTicket) []chc.IndexedTicket {
	selected := make([]chc.IndexedTicket, 0, len(tickets))
	seen := map[string]struct{}{}

	for _, ticket := range tickets {
		_, ok := seen[ticket.Key]
		if ok {
			continue
		}

		seen[ticket.Key] = struct{}{}
		selected = append(selected, ticket)
	}

	return selected
}

func runAndPrintTicketLoopExecution(
	ctx context.Context,
	stdout io.Writer,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	tickets []chc.IndexedTicket,
	eligibleCount int,
	assetsFS fs.FS,
) error {
	result, err := runTicketLoopExecution(ctx, env, projectContext, tickets, eligibleCount, assetsFS)
	if result.Record.ID == "" {
		return err
	}

	printErr := printTicketLoopExecution(stdout, result)
	if printErr != nil {
		return printErr
	}

	return err
}

func printTicketLoopPlan(
	stdout io.Writer,
	projectContext commandProjectContext,
	opts chc.TicketLoopQueryOptions,
	tickets []chc.IndexedTicket,
	dryRun bool,
) error {
	order := opts.Order
	if order == "" {
		order = chc.TicketLoopOrderKey
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

func runTicketLoopExecution(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	tickets []chc.IndexedTicket,
	eligibleCount int,
	assetsFS fs.FS,
) (ticketLoopExecutionResult, error) {
	repository, err := chc.NewLoopRunRepository(ctx, env, assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	record := newRunningLoopRunRecord(projectContext.ProjectID, eligibleCount)

	err = repository.CreateLoopRun(ctx, record)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	result := ticketLoopExecutionResult{Record: record}

	for _, ticket := range tickets {
		execution, application, err := executeLoopTicket(ctx, env, projectContext, ticket, assetsFS)
		if err != nil {
			return failTicketLoopExecution(ctx, env, repository, projectContext, result, record, err, assetsFS)
		}

		result.Executions = append(result.Executions, execution)
		result.Applications = append(result.Applications, application)

		record.ProcessedCount++
		if application.TicketStatus == chc.TicketResultStatusBlocked {
			record.BlockedCount++

			break
		}
	}

	return finishTicketLoopExecution(ctx, env, repository, projectContext, result, record, assetsFS)
}

func failTicketLoopExecution(
	ctx context.Context,
	env chc.UserEnvironment,
	repository chc.LoopRunRepository,
	projectContext commandProjectContext,
	result ticketLoopExecutionResult,
	record chc.LoopRunRecord,
	cause error,
	assetsFS fs.FS,
) (ticketLoopExecutionResult, error) {
	record.FailedCount++
	record = failLoopRunRecord(record, cause)

	err := repository.UpdateLoopRun(ctx, record)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	result.Record = record

	err = emitLoopRunNotification(ctx, env, projectContext, record, assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	return result, cause
}

func finishTicketLoopExecution(
	ctx context.Context,
	env chc.UserEnvironment,
	repository chc.LoopRunRepository,
	projectContext commandProjectContext,
	result ticketLoopExecutionResult,
	record chc.LoopRunRecord,
	assetsFS fs.FS,
) (ticketLoopExecutionResult, error) {
	record = finishLoopRunRecord(record)

	err := repository.UpdateLoopRun(ctx, record)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	result.Record = record

	err = emitLoopRunNotification(ctx, env, projectContext, record, assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	return result, nil
}

func executeLoopTicket(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	ticket chc.IndexedTicket,
	assetsFS fs.FS,
) (chc.ExecutionRecord, chc.ResultApplication, error) {
	preflight, err := buildTicketStartPreflightForTicket(ctx, env, projectContext, ticket, assetsFS)
	if err != nil {
		return chc.ExecutionRecord{}, chc.ResultApplication{}, err
	}

	record, err := runTicketStartExecution(ctx, env, preflight.Record, preflight.Runtime, assetsFS)
	if err != nil {
		return record, chc.ResultApplication{}, err
	}

	plan, err := buildTicketApplyResultPlanForRecord(projectContext, record)
	if err != nil {
		return record, chc.ResultApplication{}, err
	}

	err = applyTicketResult(ctx, env, plan, assetsFS)
	if err != nil {
		return record, chc.ResultApplication{}, err
	}

	return record, plan.Application, nil
}

func newRunningLoopRunRecord(projectID string, eligibleCount int) chc.LoopRunRecord {
	now := time.Now().UTC()

	return chc.LoopRunRecord{
		ID:            loopRunID(projectID, now),
		ProjectID:     projectID,
		Status:        chc.LoopRunStatusRunning,
		EligibleCount: eligibleCount,
		CreatedAt:     now,
		StartedAt:     &now,
		UpdatedAt:     now,
	}
}

func finishLoopRunRecord(record chc.LoopRunRecord) chc.LoopRunRecord {
	finishedAt := time.Now().UTC()

	record.Status = chc.LoopRunStatusSucceeded
	if record.BlockedCount > 0 {
		record.Status = chc.LoopRunStatusBlocked
	}

	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt

	return record
}

func failLoopRunRecord(record chc.LoopRunRecord, cause error) chc.LoopRunRecord {
	finishedAt := time.Now().UTC()
	record.Status = chc.LoopRunStatusFailed
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	record.FailureReason = cause.Error()

	return record
}

func loopRunID(projectID string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(projectID + "\x00loop\x00" + createdAt.Format(time.RFC3339Nano)))

	return "loop-" + hex.EncodeToString(sum[:8])
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

func runTicketApplyResult(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket apply-result requires exactly one execution id argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	plan, err := buildTicketApplyResultPlan(ctx, cmd, env)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		deps, err := appDependenciesFromCommand(cmd)
		if err != nil {
			return err
		}

		err = applyTicketResult(ctx, env, plan, deps.assetsFS)
		if err != nil {
			return err
		}
	}

	return printTicketApplyResult(stdout, plan, cmd.Bool("dry-run"))
}

func buildTicketApplyResultPlan(
	ctx context.Context,
	cmd *cli.Command,
	env chc.UserEnvironment,
) (ticketApplyResultPlan, error) {
	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	repository, err := chc.NewExecutionRepository(ctx, env, deps.assetsFS)
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

	return buildTicketApplyResultPlanForRecord(projectContext, record)
}

func buildTicketApplyResultPlanForRecord(
	projectContext commandProjectContext,
	record chc.ExecutionRecord,
) (ticketApplyResultPlan, error) {
	if record.ProjectID != projectContext.ProjectID {
		return ticketApplyResultPlan{}, fmt.Errorf("execution project mismatch: %s", record.ProjectID)
	}

	result, err := chc.LoadExecutionResult(record.Artifacts.ResultPath)
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	application, err := chc.NewResultApplication(chc.ResultApplicationOptions{
		Execution:     record,
		Result:        result,
		CommentAuthor: "chc",
	})
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	return ticketApplyResultPlan{
		ProjectContext: projectContext,
		Record:         record,
		Result:         result,
		Application:    application,
		ReportPath:     filepath.Join(record.Artifacts.RunDir, "result-application.md"),
	}, nil
}

func applyTicketResult(ctx context.Context, env chc.UserEnvironment, plan ticketApplyResultPlan, assetsFS fs.FS) error {
	appliedAt := time.Now().UTC()

	err := chc.NewTicketResultStateWriter().WriteResultState(ctx, chc.WriteResultStateOptions{
		Application: plan.Application,
		UpdatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = chc.NewTicketResultCommentWriter().WriteResultComment(ctx, chc.WriteResultCommentOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = chc.NewTicketResultValidationWriter().WriteResultValidation(ctx, chc.WriteResultValidationOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = chc.NewTicketResultChangedFilesWriter().WriteChangedFiles(ctx, chc.WriteChangedFilesOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = chc.NewTicketResultBlockedReasonWriter().WriteBlockedReason(ctx, chc.WriteBlockedReasonOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = refreshAppliedTicketIndex(ctx, env, plan, appliedAt, assetsFS)
	if err != nil {
		return err
	}

	err = writeTicketApplyResultReport(plan, appliedAt)
	if err != nil {
		return err
	}

	err = writeTicketExecutionReport(ctx, plan)
	if err != nil {
		return err
	}

	return emitTicketBlockedNotification(ctx, env, plan, assetsFS)
}

func refreshAppliedTicketIndex(
	ctx context.Context,
	env chc.UserEnvironment,
	plan ticketApplyResultPlan,
	appliedAt time.Time,
	assetsFS fs.FS,
) error {
	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = index.Close()
	}()

	result, err := chc.NewTicketReconciler(index).ReconcileProjectTicket(ctx, chc.ReconcileProjectTicketOptions{
		ProjectID:    plan.ProjectContext.ProjectID,
		ProjectDir:   plan.ProjectContext.ProjectDir,
		TicketPath:   plan.Application.TicketPath,
		ReconciledAt: appliedAt,
	})
	if err != nil {
		return err
	}

	if result.HasBlockers() {
		firstBlocker, _ := result.FirstBlocker()

		return fmt.Errorf("refresh applied ticket index: %d blockers: %s %w", len(result.Blockers), firstBlocker.RelativePath, firstBlocker.Err)
	}

	return nil
}

func writeTicketApplyResultReport(plan ticketApplyResultPlan, appliedAt time.Time) error {
	report := fmt.Sprintf(
		"# Result Application\n\nexecution: %s\nticket: %s\nstatus: %s\napplied_at: %s\nresult: %s\nlog: %s\nvalidation: %s\n",
		plan.Application.ExecutionID,
		plan.Application.TicketKey,
		plan.Application.TicketStatus,
		appliedAt.Format(time.RFC3339),
		plan.Application.Evidence.ResultPath,
		plan.Application.Evidence.LogPath,
		plan.Application.Evidence.ValidationPath,
	)

	err := os.WriteFile(plan.ReportPath, []byte(report), 0o644)
	if err != nil {
		return fmt.Errorf("write result application report: %w", err)
	}

	return nil
}

func writeTicketExecutionReport(ctx context.Context, plan ticketApplyResultPlan) error {
	report, err := chc.NewExecutionReport(chc.ExecutionReportOptions{
		Execution: plan.Record,
		Result:    &plan.Result,
	})
	if err != nil {
		return err
	}

	_, err = chc.NewExecutionReportWriter().WriteExecutionReport(ctx, report)
	if err != nil {
		return err
	}

	return nil
}

func printTicketApplyResult(stdout io.Writer, plan ticketApplyResultPlan, dryRun bool) error {
	lines := []string{
		fmt.Sprintf("execution: %s", plan.Application.ExecutionID),
		fmt.Sprintf("ticket: %s", plan.Application.TicketKey),
		fmt.Sprintf("ticket_path: %s", plan.Application.TicketPath),
		fmt.Sprintf("status: %s", plan.Application.TicketStatus),
		fmt.Sprintf("result_status: %s", plan.Result.Status),
		fmt.Sprintf("execution_report_path: %s", chc.ExecutionReportPath(plan.Record)),
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

func runAndPrintTicketStartExecution(
	ctx context.Context,
	stdout io.Writer,
	preflight ticketStartPreflight,
) (chc.ExecutionRecord, error) {
	record, err := runTicketStartExecution(ctx, preflight.Env, preflight.Record, preflight.Runtime, preflight.AssetsFS)
	if err == nil {
		return record, nil
	}

	printErr := printTicketStartResult(stdout, preflight.Ticket, record, preflight.Runtime, false)
	if printErr != nil {
		return chc.ExecutionRecord{}, printErr
	}

	return record, err
}

func buildTicketStartPreflight(ctx context.Context, cmd *cli.Command) (ticketStartPreflight, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	selection, err := selectTicketStartTicket(ctx, cmd, env, deps.assetsFS)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	return buildTicketStartPreflightForTicket(ctx, env, selection.ProjectContext, selection.Ticket, deps.assetsFS)
}

func buildTicketStartPreflightForTicket(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	ticket chc.IndexedTicket,
	assetsFS fs.FS,
) (ticketStartPreflight, error) {
	config, err := chc.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record := newPendingExecutionRecord(projectContext.ProjectID, ticket)

	result, err := chc.ValidateRuntime(chc.RuntimeValidationOptions{
		ProjectDir: projectContext.ProjectDir,
		Config:     config.Runtime,
		TicketKey:  ticket.Key,
		RunID:      record.ID,
	})
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record = chc.ExecutionRecordFromRuntimePlan(record, result.Plan)

	return ticketStartPreflight{
		AssetsFS: assetsFS,
		Env:      env,
		Ticket:   ticket,
		Record:   record,
		Runtime:  result,
	}, nil
}

func selectTicketStartTicket(ctx context.Context, cmd *cli.Command, env chc.UserEnvironment, assetsFS fs.FS) (ticketStartSelection, error) {
	projectContext, err := newCommandProjectContextResolver(env, assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return ticketStartSelection{}, err
	}

	index, err := chc.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return ticketStartSelection{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return ticketStartSelection{}, err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().First())
	if err != nil {
		return ticketStartSelection{}, err
	}

	return ticketStartSelection{
		ProjectContext: projectContext,
		Ticket:         ticket,
	}, nil
}

func newPendingExecutionRecord(projectID string, ticket chc.IndexedTicket) chc.ExecutionRecord {
	now := time.Now().UTC()

	return chc.ExecutionRecord{
		ID:         executionID(projectID, ticket.Key, now),
		ProjectID:  projectID,
		TicketKey:  ticket.Key,
		TicketPath: ticket.Path,
		Status:     chc.ExecutionStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func runTicketStartExecution(
	ctx context.Context,
	env chc.UserEnvironment,
	record chc.ExecutionRecord,
	result chc.RuntimeValidationResult,
	assetsFS fs.FS,
) (chc.ExecutionRecord, error) {
	repository, err := chc.NewExecutionRepository(ctx, env, assetsFS)
	if err != nil {
		return chc.ExecutionRecord{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	err = repository.CreateExecution(ctx, record)
	if err != nil {
		return chc.ExecutionRecord{}, err
	}

	invocation, err := chc.NewHarnessInvocation(chc.HarnessInvocationOptions{
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		TicketPath:  record.TicketPath,
		Plan:        result.Plan,
	})
	if err != nil {
		return chc.ExecutionRecord{}, err
	}

	runner := chc.NewExecutionRunner(repository, newTicketStartHarnessExecutor())

	return runner.Run(ctx, chc.RunExecutionOptions{
		Record:     record,
		Invocation: invocation,
	})
}

func executionID(projectID, ticketKey string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(projectID + "\x00" + ticketKey + "\x00" + createdAt.Format(time.RFC3339Nano)))

	return "exec-" + hex.EncodeToString(sum[:8])
}

func printTicketStartResult(
	stdout io.Writer,
	ticket chc.IndexedTicket,
	record chc.ExecutionRecord,
	result chc.RuntimeValidationResult,
	dryRun bool,
) error {
	lines := []string{
		fmt.Sprintf("execution: %s", record.ID),
		fmt.Sprintf("ticket: %s", ticket.Key),
		fmt.Sprintf("ticket_path: %s", ticket.Path),
		fmt.Sprintf("ticket_relative_path: %s", ticket.RelativePath),
		fmt.Sprintf("status: %s", record.Status),
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket start result: %w", err)
		}
	}

	err := printTicketStartFailure(stdout, record)
	if err != nil {
		return err
	}

	return printRuntimeValidateResult(stdout, result, dryRun)
}

func printTicketStartFailure(stdout io.Writer, record chc.ExecutionRecord) error {
	if record.FailureReason == "" {
		return nil
	}

	kind := chc.ExecutionFailureKindFromReason(record.FailureReason)
	if kind != "" {
		_, err := fmt.Fprintf(stdout, "failure_kind: %s\n", kind)
		if err != nil {
			return fmt.Errorf("write ticket start failure kind: %w", err)
		}
	}

	_, err := fmt.Fprintf(stdout, "failure_reason: %s\n", record.FailureReason)
	if err != nil {
		return fmt.Errorf("write ticket start failure reason: %w", err)
	}

	return printTicketStartDiagnostics(stdout, record)
}

func printTicketStartDiagnostics(stdout io.Writer, record chc.ExecutionRecord) error {
	diagnostics := []struct {
		name string
		path string
	}{
		{name: "diagnostic_log_path", path: record.Artifacts.LogPath},
		{name: "diagnostic_result_path", path: record.Artifacts.ResultPath},
		{name: "diagnostic_validation_path", path: record.Artifacts.ValidationPath},
	}

	for _, diagnostic := range diagnostics {
		if diagnostic.path == "" {
			continue
		}

		_, err := fmt.Fprintf(stdout, "%s: %s\n", diagnostic.name, diagnostic.path)
		if err != nil {
			return fmt.Errorf("write ticket start diagnostic path: %w", err)
		}
	}

	return nil
}

func runTicketComment(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("ticket comment requires exactly one ticket key and one comment argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	result, err := newService(ctx, env, projectContext, deps.assetsFS).AddTicketComment(TicketCommentRequest{
		Key:    cmd.Args().Get(0),
		Author: cmd.String("author"),
		Body:   cmd.Args().Get(1),
	})
	if err != nil {
		return err
	}

	updated, err := ticketOperationResultTicket(result, cmd.Args().Get(0))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "ticket: %s\n", updated.Path)
	if err != nil {
		return fmt.Errorf("write ticket comment result: %w", err)
	}

	return nil
}

func runTicketShow(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket show requires exactly one ticket key argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, index, err := openTicketIndexForCommand(ctx, cmd, env)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().First())
	if err != nil {
		return err
	}

	return printTicketShow(stdout, ticket)
}

func printTicketShow(stdout io.Writer, ticket chc.IndexedTicket) error {
	lines := []string{
		"key: " + ticket.Key,
		"title: " + ticket.Title,
		"status: " + ticket.Status,
		"priority: " + ticket.Priority,
		"severity: " + ticket.Severity,
		"type: " + ticket.Type,
		"use_case: " + ticket.UseCase,
		"project_area: " + ticket.ProjectArea,
		"path: " + ticket.RelativePath,
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket show result: %w", err)
		}
	}

	return nil
}

func runTicketList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket list does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := chc.NewTicketIndex(ctx, env, deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return err
	}

	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return err
	}

	return printTicketList(stdout, tickets)
}

func printTicketList(stdout io.Writer, tickets []chc.IndexedTicket) error {
	for _, ticket := range tickets {
		_, err := fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
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
			return fmt.Errorf("write ticket list result: %w", err)
		}
	}

	return nil
}

func openTicketIndexForCommand(
	ctx context.Context,
	cmd *cli.Command,
	env chc.UserEnvironment,
) (commandProjectContext, chc.TicketIndex, error) {
	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return commandProjectContext{}, nil, err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return commandProjectContext{}, nil, err
	}

	index, err := chc.NewTicketIndex(ctx, env, deps.assetsFS)
	if err != nil {
		return commandProjectContext{}, nil, err
	}

	return projectContext, index, nil
}

func refreshProjectTicketIndex(ctx context.Context, index chc.TicketIndex, projectContext commandProjectContext) error {
	result, err := chc.NewTicketReconciler(index).ReconcileProjectTickets(ctx, chc.ReconcileProjectTicketsOptions{
		ProjectID:    projectContext.ProjectID,
		ProjectDir:   projectContext.ProjectDir,
		ReconciledAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	if result.HasBlockers() {
		firstBlocker, _ := result.FirstBlocker()

		return fmt.Errorf("refresh ticket index: %d blockers: %s %w", len(result.Blockers), firstBlocker.RelativePath, firstBlocker.Err)
	}

	return nil
}

func ticketNewAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketNew(ctx, cmd, stdout)
	}
}

func runTicketNew(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket new requires exactly one title argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	result, err := newService(ctx, env, projectContext, deps.assetsFS).CreateTicket(CreateTicketRequest{
		Title:       cmd.Args().First(),
		Synopsis:    cmd.String("synopsis"),
		Priority:    cmd.String("priority"),
		Severity:    cmd.String("severity"),
		Type:        cmd.String("type"),
		Tags:        cmd.StringSlice("tag"),
		UseCase:     cmd.String("use-case"),
		ProjectArea: cmd.String("project-area"),
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "ticket: %s\n", result.Path)
	if err != nil {
		return fmt.Errorf("write ticket result: %w", err)
	}

	return nil
}
