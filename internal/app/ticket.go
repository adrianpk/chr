package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/core"
)

var (
	reviewProviderDiagnosticRunner core.ReviewProviderDiagnosticRunner = core.OSReviewProviderDiagnosticRunner{}
	reviewRemoteCommandRunner      core.ReviewRemoteCommandRunner      = core.OSReviewRemoteCommandRunner{}
)

func newTicketCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "ticket",
		Usage: "Manage workspace tickets",
		Commands: []*cli.Command{
			newTicketApplyResultCommand(stdout),
			newTicketBoardCommand(stdout),
			newTicketCommentCommand(stdout),
			newTicketListCommand(stdout),
			newTicketLoopCommand(stdout),
			newTicketMergeCommand(stdout),
			newTicketNewCommand(stdout),
			newTicketReviewCommand(stdout),
			newTicketReviewModeCommand(stdout),
			newTicketReviewRefCommand(stdout),
			newTicketShowCommand(stdout),
			newTicketStartCommand(stdout),
			newTicketStatusCommand(stdout),
		},
	}
}

func newTicketBoardCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "board",
		Usage:  "Inspect Board columns for the resolved project",
		Flags:  ticketProjectFlags(),
		Action: ticketBoardAction(stdout),
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

func newTicketReviewRefCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "review-ref",
		Usage:  "Attach branch and pull request review metadata to one ticket",
		Flags:  ticketReviewRefFlags(),
		Action: ticketReviewRefAction(stdout),
	}
}

func newTicketReviewCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "review",
		Usage:  "Apply an inspection-only ticket review result",
		Flags:  ticketReviewFlags(),
		Action: ticketReviewAction(stdout),
	}
}

func newTicketMergeCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "merge",
		Usage:  "Merge a reviewed ticket pull request",
		Flags:  ticketMergeFlags(),
		Action: ticketMergeAction(stdout),
	}
}

func newTicketReviewModeCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "review-mode",
		Usage: "Inspect ticket review workflow mode",
		Commands: []*cli.Command{
			newTicketReviewModeDiagnosticsCommand(stdout),
			newTicketReviewModeStatusCommand(stdout),
		},
	}
}

func newTicketReviewModeDiagnosticsCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "diagnostics",
		Usage:  "Verify ticket review provider access",
		Flags:  ticketReviewModeDiagnosticsFlags(),
		Action: ticketReviewModeDiagnosticsAction(stdout),
	}
}

func newTicketReviewModeStatusCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Print active ticket review workflow mode",
		Flags:  ticketReviewModeStatusFlags(),
		Action: ticketReviewModeStatusAction(stdout),
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

func ticketReviewRefFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "branch",
			Usage: "ticket implementation branch",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "review reference type",
			Value: "pull_request",
		},
		&cli.StringFlag{
			Name:  "provider",
			Usage: "review provider",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "pull request URL",
		},
		&cli.StringFlag{
			Name:  "id",
			Usage: "provider pull request identifier",
		},
		&cli.StringFlag{
			Name:  "ref-status",
			Usage: "provider pull request status",
			Value: "open",
		},
	)
}

func ticketReviewFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "result",
			Usage: "review result: pass, needs_changes, insufficient_evidence, needs_human_decision",
			Value: core.TicketReviewCommandResultPass,
		},
		&cli.StringSliceFlag{
			Name:  "finding",
			Usage: "blocking review finding to write remotely",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print review plan without writing ticket or remote comments",
		},
	)
}

func ticketMergeFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "strategy",
			Usage: "merge strategy: merge, rebase, squash",
			Value: core.TicketMergeStrategyRebase,
		},
		&cli.StringFlag{
			Name:  "base",
			Usage: "expected pull request base branch; defaults to the configured integration branch",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print merge preflight without merging or writing ticket state",
		},
	)
}

func ticketReviewModeStatusFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the active mode without performing review work",
		},
	)
}

func ticketReviewModeDiagnosticsFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "url",
			Usage: "pull request URL to inspect",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "skip remote pull request readability checks",
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

func ticketBoardAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketBoard(ctx, cmd, stdout)
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

func ticketReviewRefAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketReviewRef(ctx, cmd, stdout)
	}
}

func ticketReviewAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketReview(ctx, cmd, stdout)
	}
}

func ticketMergeAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketMerge(ctx, cmd, stdout)
	}
}

func ticketReviewModeStatusAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketReviewModeStatus(ctx, cmd, stdout)
	}
}

func ticketReviewModeDiagnosticsAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTicketReviewModeDiagnostics(ctx, cmd, stdout)
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
	AssetsFS       fs.FS
	Env            core.UserEnvironment
	ProjectContext commandProjectContext
	Ticket         core.IndexedTicket
	Record         core.ExecutionRecord
	Runtime        core.RuntimeValidationResult
}

type ticketStartPreflightOptions struct {
	PrepareWorktree bool
}

type ticketStartSelection struct {
	ProjectContext commandProjectContext
	Ticket         core.IndexedTicket
}

type ticketStartWorkspace struct {
	Ticket                core.IndexedTicket
	TicketPath            string
	TicketBranch          string
	ExecutionWorkspaceDir string
	WorkspaceBaseRef      string
	WorkspaceBaseSHA      string
	PreparedWorktree      bool
}

type ticketApplyResultPlan struct {
	ProjectContext commandProjectContext
	Record         core.ExecutionRecord
	Result         core.ExecutionResult
	Application    core.ResultApplication
	ReportPath     string
}

type ticketLoopExecutionResult struct {
	Record       core.LoopRunRecord
	Executions   []core.ExecutionRecord
	Applications []core.ResultApplication
	Merges       []core.TicketMergeCommandResult
}

type ticketExecutionEvents struct {
	LoopID string
	Sink   core.AgentEventSink
}

type ticketLoopRunOptions struct {
	MaxIterations int
}

var newTicketStartHarnessExecutor = func() core.HarnessExecutor {
	return core.OSHarnessExecutor{}
}

type ticketStartWorktreePreparer interface {
	Prepare(context.Context, core.WorktreePlanOptions) (core.WorktreePrepareResult, error)
}

var newTicketStartWorktreePreparer = func() ticketStartWorktreePreparer {
	service := core.NewWorktreeService()

	return service
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

func ticketOperationResultTicket(result TicketOperationResult, key string) (core.IndexedTicket, error) {
	for _, ticket := range result.Tickets {
		if ticket.Key == key {
			return ticket, nil
		}
	}

	return core.IndexedTicket{}, fmt.Errorf("ticket not found after operation: %s", key)
}

func runTicketStart(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket start requires exactly one ticket key argument")
	}

	preflight, err := buildTicketStartPreflight(ctx, cmd)
	if err != nil {
		return core.NewExecutionFailure(core.ExecutionFailurePreflight, err)
	}

	if !core.RuntimeValidationPassed(preflight.Runtime.Checks) {
		preflight.Record.Status = core.ExecutionStatusBlocked
		preflight.Record.FailureReason = core.NewExecutionFailure(
			core.ExecutionFailureRuntime,
			fmt.Errorf("runtime validation failed"),
		).Error()

		err = printTicketStartResult(stdout, preflight.Ticket, preflight.Record, preflight.Runtime, cmd.Bool("dry-run"))
		if err != nil {
			return err
		}

		return core.NewExecutionFailure(core.ExecutionFailureRuntime, fmt.Errorf("runtime validation failed"))
	}

	if !cmd.Bool("dry-run") {
		lock, err := acquireAgentOperationLock(ctx, preflight.ProjectContext, core.AgentOperationDevelopment)
		if err != nil {
			return err
		}

		defer releaseAgentOperationLock(lock)()

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

	tickets, eligibleCount, err := selectTicketLoopTickets(ctx, index, projectContext, queryOptions, runOptions)
	if err != nil {
		return err
	}

	if !cmd.Bool("dry-run") {
		deps, err := appDependenciesFromCommand(cmd)
		if err != nil {
			return err
		}

		lock, err := acquireAgentOperationLock(ctx, projectContext, core.AgentOperationDevelopment)
		if err != nil {
			return err
		}

		defer releaseAgentOperationLock(lock)()

		return runAndPrintTicketLoopExecution(ctx, stdout, env, projectContext, tickets, eligibleCount, deps.assetsFS)
	}

	return printTicketLoopPlan(stdout, projectContext, queryOptions, tickets, cmd.Bool("dry-run"))
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

func acquireAgentOperationLock(ctx context.Context, projectContext commandProjectContext, operation string) (*core.AgentOperationLock, error) {
	return core.AcquireAgentOperationLock(ctx, core.AgentOperationLockOptions{
		ProjectDir: projectContext.ProjectDir,
		Operation:  operation,
		Owner:      "chc",
	})
}

func releaseAgentOperationLock(lock *core.AgentOperationLock) func() {
	return func() {
		//nolint:errcheck // Operation result already comes from the main workflow.
		_ = lock.Release()
	}
}

func noopRelease() {}

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

		err = applyTicketResult(ctx, env, plan, ticketExecutionEvents{}, deps.assetsFS)
		if err != nil {
			return err
		}
	}

	return printTicketApplyResult(stdout, plan, cmd.Bool("dry-run"))
}

func buildTicketApplyResultPlan(
	ctx context.Context,
	cmd *cli.Command,
	env core.UserEnvironment,
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

	repository, err := core.NewExecutionRepository(ctx, env, deps.assetsFS)
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

func buildTicketApplyResultPlanForRecord(
	ctx context.Context,
	projectContext commandProjectContext,
	record core.ExecutionRecord,
) (ticketApplyResultPlan, error) {
	if record.ProjectID != projectContext.ProjectID {
		return ticketApplyResultPlan{}, fmt.Errorf("execution project mismatch: %s", record.ProjectID)
	}

	result, err := core.LoadExecutionResult(record.Artifacts.ResultPath)
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	readiness, err := ticketEvidenceReadiness(ctx, ticketEvidenceRequest{
		ProjectDir: projectContext.ProjectDir,
		TicketPath: record.TicketPath,
		Execution:  &record,
		Result:     &result,
	})
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	changedFiles, err := discoverTicketApplyResultChangedFiles(ctx, record)
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	application, err := core.NewResultApplication(core.ResultApplicationOptions{
		Execution:         record,
		Result:            result,
		EvidenceReadiness: &readiness,
		ChangedFiles:      changedFiles,
		CommentAuthor:     "chc",
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

func discoverTicketApplyResultChangedFiles(ctx context.Context, record core.ExecutionRecord) ([]core.ResultApplicationChangedFile, error) {
	if record.Workspace.Mode != core.WorkWorkspaceModeWorktree || record.Workspace.ExecutionWorkspaceDir == "" {
		return nil, nil
	}

	changedFiles, err := core.NewChangedFileDiscoverer().DiscoverChangedFiles(ctx, core.DiscoverChangedFilesOptions{
		ProjectDir:   record.Workspace.ProjectCheckoutDir,
		WorkspaceDir: record.Workspace.ExecutionWorkspaceDir,
	})
	if err != nil {
		return nil, err
	}

	result := make([]core.ResultApplicationChangedFile, 0, len(changedFiles))
	for _, changedFile := range changedFiles {
		result = append(result, core.ResultApplicationChangedFile(changedFile))
	}

	return result, nil
}

func applyTicketResult(ctx context.Context, env core.UserEnvironment, plan ticketApplyResultPlan, events ticketExecutionEvents, assetsFS fs.FS) error {
	appliedAt := time.Now().UTC()

	err := core.NewTicketResultStateWriter().WriteResultState(ctx, core.WriteResultStateOptions{
		Application: plan.Application,
		UpdatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = core.NewTicketResultCommentWriter().WriteResultComment(ctx, core.WriteResultCommentOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = core.NewTicketResultValidationWriter().WriteResultValidation(ctx, core.WriteResultValidationOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = core.NewTicketResultChangedFilesWriter().WriteChangedFiles(ctx, core.WriteChangedFilesOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	})
	if err != nil {
		return err
	}

	err = core.NewTicketResultBlockedReasonWriter().WriteBlockedReason(ctx, core.WriteBlockedReasonOptions{
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

	emitTicketApplicationAgentEvents(ctx, events, plan)

	return emitTicketBlockedNotification(ctx, env, plan, assetsFS)
}

func refreshAppliedTicketIndex(
	ctx context.Context,
	env core.UserEnvironment,
	plan ticketApplyResultPlan,
	appliedAt time.Time,
	assetsFS fs.FS,
) error {
	if plan.Record.Workspace.Mode == core.WorkWorkspaceModeWorktree {
		return nil
	}

	index, err := core.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = index.Close()
	}()

	result, err := core.NewTicketReconciler(index).ReconcileProjectTicket(ctx, core.ReconcileProjectTicketOptions{
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
	opts := core.ExecutionReportOptions{
		Execution: plan.Record,
		Result:    &plan.Result,
	}

	err := addTicketExecutionReportEvidenceOptions(&opts, plan)
	if err != nil {
		return err
	}

	report, err := core.NewExecutionReport(opts)
	if err != nil {
		return err
	}

	_, err = core.NewExecutionReportWriter().WriteExecutionReport(ctx, report)
	if err != nil {
		return err
	}

	return nil
}

func addTicketExecutionReportEvidenceOptions(opts *core.ExecutionReportOptions, plan ticketApplyResultPlan) error {
	config, err := core.LoadWorkspaceConfig(plan.ProjectContext.ProjectDir)
	if err != nil {
		return fmt.Errorf("load evidence strategy config: %w", err)
	}

	if !config.Evidence.Strategy.Enabled || !config.Evidence.Strategy.ExecutionReportCopy {
		return nil
	}

	checklist, checklistError, err := ticketExecutionReportEvidenceChecklist(plan.Application.TicketPath)
	if err != nil {
		return err
	}

	opts.IncludeEvidence = true
	opts.EvidenceChecklist = checklist
	opts.EvidenceChecklistError = checklistError

	return nil
}

func ticketExecutionReportEvidenceChecklist(path string) (*core.EvidenceStrategyChecklist, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read evidence strategy ticket: %w", err)
	}

	checklist, checklistError := ticketExecutionReportEvidenceChecklistContent(string(content))

	return checklist, checklistError, nil
}

func ticketExecutionReportEvidenceChecklistContent(content string) (*core.EvidenceStrategyChecklist, string) {
	checklist, err := core.ParseEvidenceStrategyChecklist(content)
	if err != nil {
		return nil, err.Error()
	}

	return &checklist, ""
}

func printTicketApplyResult(stdout io.Writer, plan ticketApplyResultPlan, dryRun bool) error {
	lines := []string{
		fmt.Sprintf("execution: %s", plan.Application.ExecutionID),
		fmt.Sprintf("ticket: %s", plan.Application.TicketKey),
		fmt.Sprintf("ticket_path: %s", plan.Application.TicketPath),
		fmt.Sprintf("status: %s", plan.Application.TicketStatus),
		fmt.Sprintf("result_status: %s", plan.Result.Status),
		fmt.Sprintf("execution_report_path: %s", core.ExecutionReportPath(plan.Record)),
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
) (core.ExecutionRecord, error) {
	record, err := runTicketStartExecution(ctx, preflight.Env, preflight.Record, preflight.Runtime, ticketExecutionEvents{
		Sink: agentWorkExecutionEventSink(preflight.Record),
	}, preflight.AssetsFS)
	if err == nil {
		return record, nil
	}

	printErr := printTicketStartResult(stdout, preflight.Ticket, record, preflight.Runtime, false)
	if printErr != nil {
		return core.ExecutionRecord{}, printErr
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

	return buildTicketStartPreflightForTicketWithOptions(ctx, env, selection.ProjectContext, selection.Ticket, deps.assetsFS, ticketStartPreflightOptions{
		PrepareWorktree: !cmd.Bool("dry-run"),
	})
}

func buildTicketStartPreflightForTicket(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	ticket core.IndexedTicket,
	assetsFS fs.FS,
) (ticketStartPreflight, error) {
	return buildTicketStartPreflightForTicketWithOptions(ctx, env, projectContext, ticket, assetsFS, ticketStartPreflightOptions{
		PrepareWorktree: true,
	})
}

func buildTicketStartPreflightForTicketWithOptions(ctx context.Context, env core.UserEnvironment, projectContext commandProjectContext, ticket core.IndexedTicket, assetsFS fs.FS, opts ticketStartPreflightOptions) (ticketStartPreflight, error) {
	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	workspace, err := prepareTicketStartWorkspace(ctx, projectContext, ticket, config, opts)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record := newPendingExecutionRecord(projectContext.ProjectID, workspace.Ticket)
	if workspace.TicketPath != "" {
		record.TicketPath = workspace.TicketPath
	}

	result, err := core.ValidateRuntime(core.RuntimeValidationOptions{
		ProjectDir:            projectContext.ProjectDir,
		Config:                config.Runtime,
		Workspace:             config.Work.Workspace,
		ExecutionWorkspaceDir: workspace.ExecutionWorkspaceDir,
		TicketKey:             workspace.Ticket.Key,
		TicketBranch:          workspace.TicketBranch,
		WorkspaceBaseRef:      workspace.WorkspaceBaseRef,
		WorkspaceBaseSHA:      workspace.WorkspaceBaseSHA,
		RunID:                 record.ID,
	})
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record = core.ExecutionRecordFromRuntimePlan(record, result.Plan)

	return ticketStartPreflight{
		AssetsFS:       assetsFS,
		Env:            env,
		ProjectContext: projectContext,
		Ticket:         workspace.Ticket,
		Record:         record,
		Runtime:        result,
	}, nil
}

func prepareTicketStartWorkspace(ctx context.Context, projectContext commandProjectContext, ticket core.IndexedTicket, config core.WorkspaceConfig, opts ticketStartPreflightOptions) (ticketStartWorkspace, error) {
	workspace := ticketStartWorkspace{
		Ticket:     ticket,
		TicketPath: ticket.Path,
	}
	if config.Work.Workspace.Mode != core.WorkWorkspaceModeWorktree {
		return workspace, nil
	}

	selection, err := core.SelectTicketBranch(ticket)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	planOptions := core.WorktreePlanOptions{
		ProjectDir:   projectContext.ProjectDir,
		TicketKey:    ticket.Key,
		Branch:       selection.Branch,
		BranchPolicy: config.Work.BranchPolicy,
		Workspace:    config.Work.Workspace,
		WorktreePath: workspace.ExecutionWorkspaceDir,
	}

	if !opts.PrepareWorktree {
		return previewTicketStartWorktree(ticket, planOptions)
	}

	return prepareTicketStartWorktree(ctx, projectContext.ProjectDir, ticket, config, selection, planOptions)
}

func previewTicketStartWorktree(ticket core.IndexedTicket, planOptions core.WorktreePlanOptions) (ticketStartWorkspace, error) {
	plan, err := core.NewWorktreePlan(planOptions)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	return ticketStartWorkspaceFromPlan(ticket, plan)
}

func prepareTicketStartWorktree(ctx context.Context, projectDir string, ticket core.IndexedTicket, config core.WorkspaceConfig, selection core.TicketBranchSelection, planOptions core.WorktreePlanOptions) (ticketStartWorkspace, error) {
	result, err := newTicketStartWorktreePreparer().Prepare(ctx, planOptions)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	workspace, err := ticketStartWorkspaceFromPlan(ticket, result.Plan)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	workspace.PreparedWorktree = true

	if selection.Derived {
		workspace, err = writeDerivedTicketBranch(ctx, workspace, selection.Branch)
		if err != nil {
			return ticketStartWorkspace{}, err
		}
	}

	if err := writeTicketStartWorktreeMetadata(result.Plan); err != nil {
		return ticketStartWorkspace{}, err
	}

	if err := ensureTicketStartRuntimeArtifactDir(projectDir, config, workspace); err != nil {
		return ticketStartWorkspace{}, err
	}

	return workspace, nil
}

func ticketStartWorkspaceFromPlan(ticket core.IndexedTicket, plan core.WorktreePlan) (ticketStartWorkspace, error) {
	ticketPath, err := equivalentProjectPath(plan.ProjectDir, plan.WorktreePath, ticket.Path)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	return ticketStartWorkspace{
		Ticket:                ticket,
		TicketPath:            ticketPath,
		TicketBranch:          plan.Branch,
		ExecutionWorkspaceDir: plan.WorktreePath,
		WorkspaceBaseRef:      plan.BaseRef,
		WorkspaceBaseSHA:      plan.BaseSHA,
	}, nil
}

func writeDerivedTicketBranch(ctx context.Context, workspace ticketStartWorkspace, branch string) (ticketStartWorkspace, error) {
	err := core.NewTicketReviewMetadataUpdater().UpdateReviewMetadata(ctx, core.UpdateTicketReviewMetadataOptions{
		Path:      workspace.TicketPath,
		Branch:    branch,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	workspace.Ticket.Branch = branch

	return workspace, nil
}

func writeTicketStartWorktreeMetadata(plan core.WorktreePlan) error {
	path, err := core.WorktreeMetadataPath(plan)
	if err != nil {
		return err
	}

	return core.WriteWorktreeMetadata(path, core.NewWorktreeMetadata(plan, time.Now().UTC()))
}

func ensureTicketStartRuntimeArtifactDir(projectDir string, config core.WorkspaceConfig, workspace ticketStartWorkspace) error {
	plan, err := core.PlanRuntime(core.RuntimePlanOptions{
		ProjectDir:            projectDir,
		Config:                config.Runtime,
		Workspace:             config.Work.Workspace,
		ExecutionWorkspaceDir: workspace.ExecutionWorkspaceDir,
		TicketKey:             workspace.Ticket.Key,
		TicketBranch:          workspace.TicketBranch,
		WorkspaceBaseRef:      workspace.WorkspaceBaseRef,
		WorkspaceBaseSHA:      workspace.WorkspaceBaseSHA,
	})
	if err != nil {
		return err
	}

	err = os.MkdirAll(plan.ArtifactDir, 0o755)
	if err != nil {
		return fmt.Errorf("create execution workspace artifact directory: %w", err)
	}

	return nil
}

func equivalentProjectPath(sourceRoot, targetRoot, path string) (string, error) {
	relative, err := filepath.Rel(sourceRoot, path)
	if err != nil || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return "", fmt.Errorf("path %q is not under source root %q", path, sourceRoot)
	}

	return filepath.Join(targetRoot, relative), nil
}

func selectTicketStartTicket(ctx context.Context, cmd *cli.Command, env core.UserEnvironment, assetsFS fs.FS) (ticketStartSelection, error) {
	projectContext, err := newCommandProjectContextResolver(env, assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return ticketStartSelection{}, err
	}

	index, err := core.NewTicketIndex(ctx, env, assetsFS)
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

	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return ticketStartSelection{}, err
	}

	err = enforceWorkSelectionPreflight(ctx, projectContext, tickets, ticket.Key)
	if err != nil {
		return ticketStartSelection{}, err
	}

	return ticketStartSelection{
		ProjectContext: projectContext,
		Ticket:         ticket,
	}, nil
}

func newPendingExecutionRecord(projectID string, ticket core.IndexedTicket) core.ExecutionRecord {
	now := time.Now().UTC()

	return core.ExecutionRecord{
		ID:         executionID(projectID, ticket.Key, now),
		ProjectID:  projectID,
		TicketKey:  ticket.Key,
		TicketPath: ticket.Path,
		Status:     core.ExecutionStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func runTicketStartExecution(
	ctx context.Context,
	env core.UserEnvironment,
	record core.ExecutionRecord,
	result core.RuntimeValidationResult,
	events ticketExecutionEvents,
	assetsFS fs.FS,
) (core.ExecutionRecord, error) {
	repository, err := core.NewExecutionRepository(ctx, env, assetsFS)
	if err != nil {
		return core.ExecutionRecord{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	err = repository.CreateExecution(ctx, record)
	if err != nil {
		return core.ExecutionRecord{}, err
	}

	emitBranchPreparedAgentEvent(ctx, events, record)
	emitExecutionCreatedAgentEvent(ctx, events, record)

	err = core.MaterializeCodexWorkerContextForExecution(ctx, record, result.Plan)
	if err != nil {
		return core.ExecutionRecord{}, core.NewExecutionFailure(core.ExecutionFailurePreflight, err)
	}

	invocation, err := core.NewHarnessInvocation(core.HarnessInvocationOptions{
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		TicketPath:  record.TicketPath,
		Plan:        result.Plan,
	})
	if err != nil {
		return core.ExecutionRecord{}, err
	}

	runner := core.NewExecutionRunner(repository, newTicketStartHarnessExecutor())
	runner.EventSink = events.Sink

	return runner.Run(ctx, core.RunExecutionOptions{
		Record:     record,
		Invocation: invocation,
		LoopID:     events.LoopID,
	})
}

func executionID(projectID, ticketKey string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(projectID + "\x00" + ticketKey + "\x00" + createdAt.Format(time.RFC3339Nano)))

	return "exec-" + hex.EncodeToString(sum[:8])
}

func printTicketStartResult(
	stdout io.Writer,
	ticket core.IndexedTicket,
	record core.ExecutionRecord,
	result core.RuntimeValidationResult,
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

func printTicketStartFailure(stdout io.Writer, record core.ExecutionRecord) error {
	if record.FailureReason == "" {
		return nil
	}

	kind := core.ExecutionFailureKindFromReason(record.FailureReason)
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

func printTicketStartDiagnostics(stdout io.Writer, record core.ExecutionRecord) error {
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

func runTicketReviewRef(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket review-ref requires exactly one ticket key argument")
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

	request, err := ticketReviewReferenceRequestFromCommand(cmd)
	if err != nil {
		return err
	}

	result, err := newService(ctx, env, projectContext, deps.assetsFS).AttachTicketReviewReference(request)
	if err != nil {
		return err
	}

	updated, err := ticketOperationResultTicket(result, request.Key)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "ticket: %s review_ref\n", updated.Key)
	if err != nil {
		return fmt.Errorf("write ticket review-ref result: %w", err)
	}

	return nil
}

func runTicketReview(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket review requires exactly one ticket key argument")
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

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	adapter := core.NewRemoteReviewAdapter(config.Review.Providers, reviewRemoteCommandRunner)

	release, err := acquireTicketReviewLock(ctx, projectContext, cmd.Bool("dry-run"))
	if err != nil {
		return err
	}
	defer release()

	result, err := core.RunTicketReviewCommand(ctx, core.TicketReviewCommandOptions{
		Ticket:     ticket,
		Config:     config,
		Result:     cmd.String("result"),
		Findings:   cmd.StringSlice("finding"),
		Adapter:    adapter,
		DryRun:     cmd.Bool("dry-run"),
		ReviewedAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	err = refreshTicketIndexAfterReview(ctx, index, projectContext, cmd.Bool("dry-run"))
	if err != nil {
		return err
	}

	return printTicketReviewResult(stdout, projectContext, result)
}

func runTicketMerge(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket merge requires exactly one ticket key argument")
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

	if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
		return err
	}

	ticket, tickets, err := ticketMergeCommandTickets(ctx, index, projectContext, cmd.Args().First())
	if err != nil {
		return err
	}

	config, expectedBase, err := ticketMergeCommandConfig(projectContext.ProjectDir, cmd)
	if err != nil {
		return err
	}

	release, err := acquireTicketMergeLock(ctx, projectContext, cmd.Bool("dry-run"))
	if err != nil {
		return err
	}
	defer release()

	result, err := core.RunTicketMergeCommand(ctx, core.TicketMergeCommandOptions{
		Ticket:        ticket,
		Tickets:       tickets,
		Remote:        core.NewRemoteReviewAdapter(config.Review.Providers, reviewRemoteCommandRunner),
		ExpectedBase:  expectedBase,
		MergeStrategy: cmd.String("strategy"),
		ProjectDir:    projectContext.ProjectDir,
		BranchPolicy:  config.Work.BranchPolicy,
		Workspace:     config.Work.Workspace,
		DryRun:        cmd.Bool("dry-run"),
		MergedAt:      time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	if err := refreshTicketIndexAfterMerge(ctx, index, projectContext, cmd.Bool("dry-run")); err != nil {
		return err
	}

	return printTicketMergeResult(stdout, projectContext, result)
}

func ticketMergeCommandConfig(projectDir string, cmd *cli.Command) (core.WorkspaceConfig, string, error) {
	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return core.WorkspaceConfig{}, "", err
	}

	if err := core.ValidateWorkConfig(config.Work); err != nil {
		return core.WorkspaceConfig{}, "", err
	}

	return config, ticketMergeExpectedBase(cmd, config.Work.BranchPolicy), nil
}

func ticketMergeExpectedBase(cmd *cli.Command, policy core.WorkBranchPolicy) string {
	if base := strings.TrimSpace(cmd.String("base")); cmd.IsSet("base") && base != "" {
		return base
	}

	return policy.IntegrationBranch
}

func ticketMergeCommandTickets(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext, key string) (core.IndexedTicket, []core.IndexedTicket, error) {
	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return core.IndexedTicket{}, nil, err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, key)
	if err != nil {
		return core.IndexedTicket{}, nil, err
	}

	return ticket, tickets, nil
}

func refreshTicketIndexAfterReview(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext, dryRun bool) error {
	if dryRun {
		return nil
	}

	return refreshProjectTicketIndex(ctx, index, projectContext)
}

func refreshTicketIndexAfterMerge(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext, dryRun bool) error {
	if dryRun {
		return nil
	}

	return refreshProjectTicketIndex(ctx, index, projectContext)
}

func acquireTicketReviewLock(ctx context.Context, projectContext commandProjectContext, dryRun bool) (func(), error) {
	if dryRun {
		return noopRelease, nil
	}

	lock, err := acquireAgentOperationLock(ctx, projectContext, core.AgentOperationReview)
	if err != nil {
		return noopRelease, err
	}

	return releaseAgentOperationLock(lock), nil
}

func acquireTicketMergeLock(ctx context.Context, projectContext commandProjectContext, dryRun bool) (func(), error) {
	if dryRun {
		return noopRelease, nil
	}

	lock, err := acquireAgentOperationLock(ctx, projectContext, core.AgentOperationMerge)
	if err != nil {
		return noopRelease, err
	}

	return releaseAgentOperationLock(lock), nil
}

func printTicketReviewResult(
	stdout io.Writer,
	projectContext commandProjectContext,
	result core.TicketReviewCommandResult,
) error {
	lines := []string{
		"ticket_review: result",
		"project_id: " + projectContext.ProjectID,
		"ticket: " + result.TicketKey,
		"result: " + result.ReviewResult,
		"review_state: " + result.ReviewState,
		fmt.Sprintf("review_round: %d", result.ReviewRound),
		"status: " + result.Status,
		"review_ref: " + result.ReviewRef.URL,
		"provider: " + string(result.RemoteSnapshot.Reference.Provider),
		fmt.Sprintf("remote_comments: %d", len(result.RemoteComments)),
		fmt.Sprintf("dry_run: %t", result.DryRun),
		"policy: " + core.TicketReviewCommandInspectionOnlyPolicySummary,
	}

	for _, finding := range result.BlockingFindings {
		lines = append(lines, "finding: "+finding)
	}

	for _, comment := range result.RemoteComments {
		lines = append(lines, "remote_comment: "+comment.Placement+" "+comment.Kind)
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket review result: %w", err)
		}
	}

	return nil
}

func printTicketMergeResult(stdout io.Writer, projectContext commandProjectContext, result core.TicketMergeCommandResult) error {
	lines := []string{
		"ticket_merge: result",
		"project_id: " + projectContext.ProjectID,
		"ticket: " + result.TicketKey,
		fmt.Sprintf("allowed: %t", result.Preflight.Allowed),
		"status: " + result.Status,
		"merge_state: " + result.MergeState,
		"merge_result: " + result.MergeResult,
		"merge_strategy: " + result.MergeStrategy,
		"merge_commit: " + result.MergeCommit,
		"review_ref: " + result.Preflight.ReviewRef.URL,
		"provider: " + string(result.Preflight.Remote.Reference.Provider),
		fmt.Sprintf("dry_run: %t", result.DryRun),
	}

	for _, blocker := range result.Preflight.Blockers {
		lines = append(lines, "blocker: "+blocker.Code+" kind="+blocker.Kind+" detail="+blocker.Detail)
	}

	if result.WorktreeCleanup.Status != "" {
		lines = append(lines,
			"worktree_cleanup: "+result.WorktreeCleanup.Status,
			"worktree_cleanup_detail: "+result.WorktreeCleanup.Detail,
		)
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket merge result: %w", err)
		}
	}

	return nil
}

func runTicketReviewModeStatus(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket review-mode status does not accept positional arguments")
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

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	return printTicketReviewModeStatus(stdout, projectContext, config.Review, cmd.Bool("dry-run"))
}

func runTicketReviewModeDiagnostics(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket review-mode diagnostics does not accept positional arguments")
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

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	diagnostic := core.RunReviewProviderDiagnostics(ctx, core.ReviewProviderDiagnosticOptions{
		Config: config.Review.Providers,
		URL:    cmd.String("url"),
		DryRun: cmd.Bool("dry-run"),
		Runner: reviewProviderDiagnosticRunner,
	})

	return printTicketReviewModeDiagnostics(stdout, projectContext, diagnostic, cmd.Bool("dry-run"))
}

func printTicketReviewModeDiagnostics(
	stdout io.Writer,
	projectContext commandProjectContext,
	diagnostic core.ReviewProviderDiagnostic,
	dryRun bool,
) error {
	reference := diagnostic.Reference
	lines := []string{
		"review_provider_diagnostics: status",
		"project_id: " + projectContext.ProjectID,
		"project_dir: " + projectContext.ProjectDir,
		"url: " + reference.URL,
		"provider: " + string(diagnostic.Provider),
		"provider_status: " + diagnostic.Status,
		"host: " + reference.Host,
		"owner: " + reference.Owner,
		"repo: " + reference.Repo,
		fmt.Sprintf("number: %d", reference.Number),
		fmt.Sprintf("dry_run: %t", dryRun),
	}

	for _, check := range diagnostic.Checks {
		line := fmt.Sprintf("check.%s: %s", check.Name, check.Status)
		if check.Reason != "" {
			line += " reason=" + check.Reason
		}

		if check.Detail != "" {
			line += " detail=" + check.Detail
		}

		lines = append(lines, line)
	}

	return printTicketReviewModeStatusLines(stdout, lines)
}

func printTicketReviewModeStatus(
	stdout io.Writer,
	projectContext commandProjectContext,
	config core.ReviewWorkflowConfig,
	dryRun bool,
) error {
	status := "ready"

	err := core.ValidateReviewWorkflowConfig(config)
	if err != nil {
		status = "invalid"
	}

	lines := []string{
		"review_mode: status",
		"project_id: " + projectContext.ProjectID,
		"project_dir: " + projectContext.ProjectDir,
		"mode: " + string(config.Mode),
		"mode_status: " + status,
		"description: " + core.ReviewWorkflowModeDescription(config.Mode),
		fmt.Sprintf("dry_run: %t", dryRun),
		fmt.Sprintf("requires_remote_review: %t", config.Mode == core.ReviewWorkflowModeAgenticReview),
		fmt.Sprintf("requires_human_review: %t", config.Mode == core.ReviewWorkflowModeHumanReview),
		fmt.Sprintf("trusted_auto_merge: %t", config.Mode == core.ReviewWorkflowModeTrustedAutoMerge),
		"clean_pass_remote_comment: false",
		fmt.Sprintf("non_blocking_comment_limit: %d", config.NonBlockingCommentLimit),
	}

	if err != nil {
		lines = append(lines, "error: "+err.Error())
	}

	return printTicketReviewModeStatusLines(stdout, lines)
}

func printTicketReviewModeStatusLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket review mode status: %w", err)
		}
	}

	return nil
}

func ticketReviewReferenceRequestFromCommand(cmd *cli.Command) (TicketReviewReferenceRequest, error) {
	request := TicketReviewReferenceRequest{
		Key:    cmd.Args().First(),
		Branch: cmd.String("branch"),
	}

	if strings.TrimSpace(cmd.String("url")) == "" {
		return request, nil
	}

	if strings.TrimSpace(cmd.String("provider")) == "" {
		return TicketReviewReferenceRequest{}, fmt.Errorf("ticket review-ref provider is required when url is set")
	}

	request.Ref = &core.TicketReviewRef{
		Type:     cmd.String("type"),
		Provider: cmd.String("provider"),
		URL:      cmd.String("url"),
		ID:       cmd.String("id"),
		Status:   cmd.String("ref-status"),
	}

	return request, nil
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

	return printTicketShow(stdout, ticket, ticketEvidenceSummary(ticket))
}

func printTicketShow(stdout io.Writer, ticket core.IndexedTicket, evidence core.EvidenceStrategySummary) error {
	lines := []string{
		"key: " + ticket.Key,
		"title: " + ticket.Title,
		"status: " + ticket.Status,
		"priority: " + ticket.Priority,
		"severity: " + ticket.Severity,
		"type: " + ticket.Type,
		"use_case: " + ticket.UseCase,
		"project_area: " + ticket.ProjectArea,
		"branch: " + ticket.Branch,
		"review_state: " + ticket.ReviewState,
		fmt.Sprintf("review_round: %d", ticket.ReviewRound),
		"review_result: " + ticket.ReviewResult,
		"review_comments_url: " + ticket.ReviewCommentsURL,
		"path: " + ticket.RelativePath,
	}

	lines = append(lines, ticketEvidenceShowLines(evidence)...)

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

	index, err := core.NewTicketIndex(ctx, env, deps.assetsFS)
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

func runTicketBoard(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket board does not accept positional arguments")
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

	board, err := newService(ctx, env, projectContext, deps.assetsFS).Board()
	if err != nil {
		return err
	}

	return printTicketBoard(stdout, board)
}

func printTicketBoard(stdout io.Writer, board core.Board) error {
	lines := []string{
		fmt.Sprintf("total: %d", board.TotalCount),
		fmt.Sprintf("normal: %d", board.NormalCount),
		fmt.Sprintf("attention: %d", board.AttentionCount),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket board result: %w", err)
		}
	}

	for _, column := range board.Columns {
		if err := printTicketBoardColumn(stdout, column); err != nil {
			return err
		}
	}

	return nil
}

func printTicketBoardColumn(stdout io.Writer, column core.BoardColumn) error {
	_, err := fmt.Fprintf(stdout, "column: %s\t%s\t%d\n", column.ID, column.Title, column.Count)
	if err != nil {
		return fmt.Errorf("write ticket board column: %w", err)
	}

	for _, ticket := range column.Tickets {
		_, err = fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.Key,
			ticket.Status,
			ticket.Priority,
			ticket.Type,
			ticket.Title,
			ticketEvidenceStatusCell(ticketEvidenceSummary(ticket)),
		)
		if err != nil {
			return fmt.Errorf("write ticket board row: %w", err)
		}
	}

	return nil
}

func printTicketList(stdout io.Writer, tickets []core.IndexedTicket) error {
	for _, ticket := range tickets {
		_, err := fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.Key,
			ticket.Title,
			ticket.Status,
			ticket.Priority,
			ticket.Severity,
			ticket.Type,
			ticket.UseCase,
			ticket.RelativePath,
			ticketEvidenceStatusCell(ticketEvidenceSummary(ticket)),
		)
		if err != nil {
			return fmt.Errorf("write ticket list result: %w", err)
		}
	}

	return nil
}

func ticketEvidenceSummary(ticket core.IndexedTicket) core.EvidenceStrategySummary {
	content, err := os.ReadFile(ticket.Path)
	if err != nil {
		return core.EvidenceStrategySummary{}
	}

	return core.SummarizeEvidenceStrategyChecklist(string(content))
}

func ticketEvidenceShowLines(evidence core.EvidenceStrategySummary) []string {
	if !ticketEvidenceVisible(evidence) {
		return nil
	}

	lines := []string{"evidence: " + string(evidence.Status)}
	if evidence.Error != "" {
		return append(lines, "evidence_error: "+evidence.Error)
	}

	for _, item := range evidence.Items {
		lines = append(lines, ticketEvidenceItemLine(item))
	}

	return lines
}

func ticketEvidenceItemLine(item core.EvidenceStrategySummaryItem) string {
	return ticketEvidenceItemKey(item.Label) + ": " + item.Status + " " + ticketEvidenceItemState(item.Checked)
}

func ticketEvidenceItemKey(label string) string {
	switch label {
	case core.EvidenceStrategyChecklistLabelStrategy:
		return "evidence_strategy"
	case core.EvidenceStrategyChecklistLabelGherkin:
		return "evidence_gherkin"
	case core.EvidenceStrategyChecklistLabelTests:
		return "evidence_tests"
	case core.EvidenceStrategyChecklistLabelValidation:
		return "evidence_validation"
	case core.EvidenceStrategyChecklistLabelSummary:
		return "evidence_summary"
	default:
		return "evidence_" + strings.ToLower(strings.ReplaceAll(label, " ", "_"))
	}
}

func ticketEvidenceItemState(checked bool) string {
	if checked {
		return "done"
	}

	return "pending"
}

func ticketEvidenceStatusCell(evidence core.EvidenceStrategySummary) string {
	if !ticketEvidenceVisible(evidence) {
		return "-"
	}

	return string(evidence.Status)
}

func ticketEvidenceVisible(evidence core.EvidenceStrategySummary) bool {
	return evidence.Status != "" && evidence.Status != core.EvidenceStrategySummaryStatusMissing
}

func openTicketIndexForCommand(
	ctx context.Context,
	cmd *cli.Command,
	env core.UserEnvironment,
) (commandProjectContext, core.TicketIndex, error) {
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

	index, err := core.NewTicketIndex(ctx, env, deps.assetsFS)
	if err != nil {
		return commandProjectContext{}, nil, err
	}

	return projectContext, index, nil
}

func refreshProjectTicketIndex(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext) error {
	result, err := core.NewTicketReconciler(index).ReconcileProjectTickets(ctx, core.ReconcileProjectTicketsOptions{
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
