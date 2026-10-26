package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"time"

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

type ticketLoopExecutionResult struct {
	AuthorityExcluded int
	Record            core.LoopRunRecord
	Executions        []core.ExecutionRecord
	Applications      []core.ResultApplication
	Merges            []core.TicketMergeCommandResult
}

type ticketLoopRunOptions struct {
	MaxIterations int
}

func (factory commandFactory) runTicketLoop(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket loop does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, index, err := factory.openTicketIndexForCommand(ctx, cmd, env)
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

	tickets, eligibleCount, err := selectTicketLoopTickets(ctx, index, projectContext, queryOptions, runOptions)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		lock, err := acquireAgentOperationLock(ctx, projectContext, core.AgentOperationDevelopment)
		if err != nil {
			return err
		}

		defer releaseAgentOperationLock(lock)()

		return runAndPrintTicketLoopExecution(ctx, factory.io.Stdout, env, projectContext, tickets, eligibleCount, factory.deps.assetsFS)
	}

	return printTicketLoopPlan(factory.io.Stdout, projectContext, queryOptions, tickets, cmd.Bool("dry-run"))
}

func selectTicketLoopTickets(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext, queryOptions core.TicketLoopQueryOptions, runOptions ticketLoopRunOptions) ([]core.IndexedTicket, int, error) {
	tickets, err := core.NewTicketLoopQuery(index).EligibleTickets(ctx, queryOptions)
	if err != nil {
		return nil, 0, err
	}

	tickets, err = applyTicketLoopPreflight(ctx, index, projectContext, tickets)
	if err != nil {
		return nil, 0, err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, runOptions)

	return tickets, len(eligibleTickets), nil
}

func applyTicketLoopPreflight(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext, tickets []core.IndexedTicket) ([]core.IndexedTicket, error) {
	allTickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return nil, err
	}

	workConfig, err := workSelectionWorkConfig(projectContext.ProjectDir)
	if err != nil {
		return nil, err
	}

	preflight, err := core.RunWorkSelectionPreflight(ctx, core.WorkSelectionPreflightOptions{
		ProjectDir:   projectContext.ProjectDir,
		Tickets:      allTickets,
		BranchPolicy: workConfig.BranchPolicy,
		Workspace:    workConfig.Workspace,
	})
	if err != nil {
		return nil, err
	}

	if !preflight.Allowed {
		return nil, formatWorkSelectionPreflightError(preflight)
	}

	if len(preflight.SelectedTickets) > 0 && hasReviewRepairTicket(preflight.SelectedTickets) {
		return preflight.SelectedTickets, nil
	}

	return tickets, nil
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

func ticketLoopRunOptionsFromCommand(cmd *cli.Command) (ticketLoopRunOptions, error) {
	opts := ticketLoopRunOptions{
		MaxIterations: cmd.Int("max-iterations"),
	}
	if opts.MaxIterations < 0 {
		return ticketLoopRunOptions{}, fmt.Errorf("max-iterations must be greater than or equal to zero")
	}

	return opts, nil
}

func ticketsForTicketLoopRun(tickets []core.IndexedTicket, opts ticketLoopRunOptions) []core.IndexedTicket {
	if opts.MaxIterations == 0 || len(tickets) <= opts.MaxIterations {
		return tickets
	}

	return tickets[:opts.MaxIterations]
}

func deduplicateTicketLoopTickets(tickets []core.IndexedTicket) []core.IndexedTicket {
	selected := make([]core.IndexedTicket, 0, len(tickets))
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

func enforceWorkSelectionPreflight(ctx context.Context, projectContext commandProjectContext, tickets []core.IndexedTicket, selectedTicketKey string) error {
	workConfig, err := workSelectionWorkConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	preflight, err := core.RunWorkSelectionPreflight(ctx, core.WorkSelectionPreflightOptions{
		ProjectDir:        projectContext.ProjectDir,
		Tickets:           tickets,
		SelectedTicketKey: selectedTicketKey,
		BranchPolicy:      workConfig.BranchPolicy,
		Workspace:         workConfig.Workspace,
	})
	if err != nil {
		return err
	}

	if !preflight.Allowed {
		return formatWorkSelectionPreflightError(preflight)
	}

	return nil
}

func formatWorkSelectionPreflightError(preflight core.WorkSelectionPreflightResult) error {
	if len(preflight.BlockReasons) == 0 {
		return fmt.Errorf("work selection preflight blocked")
	}

	reason := preflight.BlockReasons[0]
	if reason.TicketKey != "" {
		return fmt.Errorf("work selection preflight blocked: %s ticket=%s detail=%s", reason.Code, reason.TicketKey, reason.Detail)
	}

	return fmt.Errorf("work selection preflight blocked: %s detail=%s", reason.Code, reason.Detail)
}

func hasReviewRepairTicket(tickets []core.IndexedTicket) bool {
	for _, ticket := range tickets {
		readyForDev := ticket.Status == core.TicketStatusReadyForDev
		changesRequested := ticket.ReviewState == "changes_requested" || ticket.ReviewResult == "needs_changes"

		if readyForDev && changesRequested {
			return true
		}
	}

	return false
}

func runAndPrintTicketLoopExecution(
	ctx context.Context,
	stdout io.Writer,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	tickets []core.IndexedTicket,
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
	opts core.TicketLoopQueryOptions,
	tickets []core.IndexedTicket,
	dryRun bool,
) error {
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

func runTicketLoopExecution(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	tickets []core.IndexedTicket,
	eligibleCount int,
	assetsFS fs.FS,
) (ticketLoopExecutionResult, error) {
	repository, err := core.NewLoopRunRepository(ctx, env, assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	record := newRunningLoopRunRecord(projectContext.ProjectID, eligibleCount)

	record, err = allocateLoopRunArtifact(projectContext.ProjectDir, record)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	err = repository.CreateLoopRun(ctx, record)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	loopSink := agentWorkLoopEventSink(projectContext.ProjectDir, record)
	emitLoopStartedAgentEvent(ctx, loopSink, record)

	result := ticketLoopExecutionResult{Record: record}

	for _, ticket := range tickets {
		emitTicketSelectedAgentEvent(ctx, loopSink, record, ticket)

		execution, application, err := executeLoopTicket(ctx, env, projectContext, ticket, ticketExecutionEvents{
			LoopID: record.ID,
			Sink:   loopSink,
		}, assetsFS)
		if err != nil {
			return failTicketLoopExecution(ctx, env, repository, projectContext, result, record, loopSink, err, assetsFS)
		}

		result.Executions = append(result.Executions, execution)
		result.Applications = append(result.Applications, application)

		record.ProcessedCount++
		if application.TicketStatus == core.TicketResultStatusBlocked {
			record.BlockedCount++

			break
		}
	}

	return finishTicketLoopExecution(ctx, env, repository, projectContext, result, record, loopSink, assetsFS)
}

func failTicketLoopExecution(
	ctx context.Context,
	env core.UserEnvironment,
	repository core.LoopRunRepository,
	projectContext commandProjectContext,
	result ticketLoopExecutionResult,
	record core.LoopRunRecord,
	eventSink core.AgentEventSink,
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
	emitLoopFailedAgentEvents(ctx, eventSink, record, cause)

	if record.ArtifactID != "" {
		if err := core.RefreshOpsArtifactMetadata(record.ArtifactDir, string(record.Status), time.Now().UTC()); err != nil {
			return ticketLoopExecutionResult{}, err
		}
	}

	err = emitLoopRunNotification(ctx, env, projectContext, record, assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	return result, cause
}

func finishTicketLoopExecution(
	ctx context.Context,
	env core.UserEnvironment,
	repository core.LoopRunRepository,
	projectContext commandProjectContext,
	result ticketLoopExecutionResult,
	record core.LoopRunRecord,
	eventSink core.AgentEventSink,
	assetsFS fs.FS,
) (ticketLoopExecutionResult, error) {
	record = finishLoopRunRecord(record)

	err := repository.UpdateLoopRun(ctx, record)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	result.Record = record
	emitLoopFinishedAgentEvent(ctx, eventSink, record)

	if record.ArtifactID != "" {
		if err := core.RefreshOpsArtifactMetadata(record.ArtifactDir, string(record.Status), time.Now().UTC()); err != nil {
			return ticketLoopExecutionResult{}, err
		}
	}

	err = emitLoopRunNotification(ctx, env, projectContext, record, assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	return result, nil
}

func executeLoopTicket(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	ticket core.IndexedTicket,
	events ticketExecutionEvents,
	assetsFS fs.FS,
) (core.ExecutionRecord, core.ResultApplication, error) {
	preflight, err := buildTicketStartPreflightForTicket(ctx, env, projectContext, ticket, assetsFS)
	if err != nil {
		return core.ExecutionRecord{}, core.ResultApplication{}, err
	}

	events.Sink = core.NewAgentEventFanoutSink(events.Sink, agentWorkExecutionEventSink(preflight.Record))

	record, err := runTicketStartExecution(ctx, env, preflight.Record, preflight.Runtime, events, assetsFS)
	if err != nil {
		return record, core.ResultApplication{}, err
	}

	plan, err := buildTicketApplyResultPlanForRecord(ctx, projectContext, record)
	if err != nil {
		return record, core.ResultApplication{}, err
	}

	err = applyTicketResult(ctx, env, plan, events, assetsFS)
	if err != nil {
		return record, core.ResultApplication{}, err
	}

	return record, plan.Application, nil
}

func newRunningLoopRunRecord(projectID string, eligibleCount int) core.LoopRunRecord {
	now := time.Now().UTC()

	return core.LoopRunRecord{
		ID:            loopRunID(projectID, now),
		ProjectID:     projectID,
		Status:        core.LoopRunStatusRunning,
		EligibleCount: eligibleCount,
		CreatedAt:     now,
		StartedAt:     &now,
		UpdatedAt:     now,
	}
}

func allocateLoopRunArtifact(projectDir string, record core.LoopRunRecord) (core.LoopRunRecord, error) {
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return core.LoopRunRecord{}, err
	}

	artifact, err := core.CreateOpsArtifact(core.CreateOpsArtifactOptions{
		Layout:      layout,
		Type:        "run",
		Title:       "Agent work loop",
		Producer:    "chr",
		ExternalIDs: []string{record.ID},
	})
	if err != nil {
		return core.LoopRunRecord{}, err
	}

	record.ArtifactID = artifact.ID
	record.ArtifactDir = artifact.Dir

	return record, nil
}

func finishLoopRunRecord(record core.LoopRunRecord) core.LoopRunRecord {
	finishedAt := time.Now().UTC()

	record.Status = core.LoopRunStatusSucceeded
	if record.BlockedCount > 0 {
		record.Status = core.LoopRunStatusBlocked
	}

	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt

	return record
}

func failLoopRunRecord(record core.LoopRunRecord, cause error) core.LoopRunRecord {
	finishedAt := time.Now().UTC()
	record.Status = core.LoopRunStatusFailed
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
