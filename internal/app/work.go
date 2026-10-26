package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) workCommand() *cli.Command {
	return &cli.Command{
		Name:  "work",
		Usage: "Run agent work for tickets",
		Commands: []*cli.Command{
			factory.workConfigCommand(),
			factory.workUpgradeCommand(),
			factory.workTrackerCommand(),
			factory.workDoctorCommand(),
			factory.workEventsCommand(),
			factory.workLogsCommand(),
			factory.workTicketCommand(),
			factory.workRunAllCommand(),
			factory.workExtentCommand("slice", "slice"),
			factory.workExtentCommand("slice-set", "slice_set"),
			factory.workExtentCommand("use-case", "use_case"),
			factory.workExtentCommand("loose", "loose"),
			factory.workControlCommand("status"),
			factory.workControlCommand("continue"),
			factory.workControlCommand("pause"),
			factory.workControlCommand("cancel"),
			factory.workControlCommand("resume"),
			factory.workControlCommand("approve"),
			factory.workControlCommand("grant"),
			factory.workControlCommand("answer-decision"),
			factory.workControlCommand("reconcile"),
		},
	}
}

func (factory commandFactory) workDoctorCommand() *cli.Command {
	return &cli.Command{
		Name:   "doctor",
		Usage:  "Inspect work readiness without starting agent work",
		Flags:  workDoctorFlags(),
		Action: factory.workDoctorAction,
	}
}

func (factory commandFactory) workEventsCommand() *cli.Command {
	return &cli.Command{
		Name:      "events",
		Usage:     "Summarize one agent event stream",
		ArgsUsage: "<event-stream-path>",
		Flags:     workProjectFlags(),
		Action:    factory.workEventsAction,
	}
}

func (factory commandFactory) workLogsCommand() *cli.Command {
	return &cli.Command{
		Name:      "logs",
		Usage:     "Locate runtime logs for one execution",
		ArgsUsage: "<execution-id>",
		Flags:     workProjectFlags(),
		Action:    factory.workLogsAction,
	}
}

func (factory commandFactory) workTicketCommand() *cli.Command {
	return &cli.Command{
		Name:      "ticket",
		Usage:     "Run agent work for one ticket",
		ArgsUsage: "<ticket-key>",
		Flags:     workWorkflowExtentFlags(false),
		Action:    factory.workTicketAction,
	}
}

func (factory commandFactory) workRunAllCommand() *cli.Command {
	return &cli.Command{
		Name:   "run-all",
		Usage:  "Run agent work for all ready tickets",
		Flags:  workRunAllFlags(),
		Action: factory.workRunAllAction,
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

func workDoctorFlags() []cli.Flag {
	return append(workProjectFlags(),
		&cli.BoolFlag{
			Name:  "fix-safe",
			Usage: "apply safe readiness repairs before printing the report",
		},
		&cli.StringFlag{
			Name:  "ticket",
			Usage: "ticket key to include ticket worktree readiness",
		},
		&cli.StringFlag{
			Name:  "ticket-branch",
			Usage: "ticket branch override for ticket worktree readiness",
		},
	)
}

func workRunAllFlags() []cli.Flag {
	return append(workProjectFlags(),
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"preview"},
			Usage:   "print the ready-ticket queue without executing agent work",
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
			Usage: "maximum integration cycles to run; 0 means until no eligible tickets or a cycle barrier is reached",
		},
	)
}

func (factory commandFactory) workDoctorAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("work doctor does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	projectContext := operations.projectContext

	if err != nil {
		return err
	}

	workConfig, err := workSelectionWorkConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	doctorTicket, err := workDoctorTicketContext(cmd, operations)
	ticketKey, ticketBranch := doctorTicket.Key, doctorTicket.Branch

	if err != nil {
		return err
	}

	if cmd.Bool("fix-safe") {
		result, err := core.RepairWorkReadiness(ctx, core.WorkReadinessRepairOptions{
			ProjectDir:   projectContext.ProjectDir,
			Policy:       workConfig.BranchPolicy,
			Workspace:    workConfig.Workspace,
			TicketKey:    ticketKey,
			TicketBranch: ticketBranch,
		})
		if err != nil {
			return err
		}

		return printWorkRepairResult(factory.io.Stdout, projectContext, result)
	}

	report, err := core.CheckWorkReadiness(ctx, core.WorkReadinessOptions{
		ProjectDir:   projectContext.ProjectDir,
		Policy:       workConfig.BranchPolicy,
		Workspace:    workConfig.Workspace,
		TicketKey:    ticketKey,
		TicketBranch: ticketBranch,
	})
	if err != nil {
		return err
	}

	return printWorkReadinessReport(factory.io.Stdout, projectContext, report)
}

func workDoctorTicketContext(cmd *cli.Command, operations service) (workDoctorTicket, error) {
	ticketKey := strings.TrimSpace(cmd.String("ticket"))

	ticketBranch := strings.TrimSpace(cmd.String("ticket-branch"))
	if ticketKey == "" {
		return workDoctorTicket{Key: "", Branch: ticketBranch}, nil
	}

	if ticketBranch != "" {
		return workDoctorTicket{Key: ticketKey, Branch: ticketBranch}, nil
	}

	ticket, err := operations.ticketByKey(ticketKey)
	if err != nil {
		return workDoctorTicket{Key: "", Branch: ""}, err
	}

	return workDoctorTicket{Key: ticketKey, Branch: ticket.Branch}, nil
}

func (factory commandFactory) workTicketAction(ctx context.Context, cmd *cli.Command) error {
	if err := workWorkflowArguments(cmd, false); err != nil {
		return err
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	config, err := agentWorkLoopWorkspaceConfig(operations.projectContext.ProjectDir)
	if err != nil {
		return err
	}

	if core.WorkAuthority(config.Work) == core.WorkStateAuthorityTickets {
		return (workWorkflowCLI{stdout: factory.io.Stdout}).extent(ctx, cmd, operations, "ticket")
	}

	if usesWorkflowTicketFlags(cmd) {
		return fmt.Errorf("workflow ticket controls require tickets authority; current authority is tracker_compat")
	}

	result, err := operations.StartAgentWorkTicket(cmd.Args().Get(0))
	if err != nil {
		return err
	}

	return printAgentWorkTicket(factory.io.Stdout, result)
}

func usesWorkflowTicketFlags(cmd *cli.Command) bool {
	for _, name := range []string{"preview", "json", "max-steps", "offset", "limit", "snapshot"} {
		if cmd.IsSet(name) {
			return true
		}
	}

	return false
}

func (factory commandFactory) workRunAllAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("work run-all does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	projectContext := operations.projectContext

	if err != nil {
		return err
	}

	selection := agentWorkSelectionOptionsFromCommand(cmd)
	if cmd.Bool("dry-run") {
		result, err := operations.PreviewAgentWorkLoop(selection)
		if err != nil {
			return err
		}

		return printAgentWorkRunAllPreview(factory.io.Stdout, projectContext, selection, result)
	}

	result, err := operations.StartAgentWorkLoop(selection)
	if err != nil {
		return err
	}

	return printAgentWorkRunAllExecution(factory.io.Stdout, result)
}

func (factory commandFactory) workEventsAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("work events requires exactly one event stream path argument")
	}

	resolvedCommandService, err := factory.projectService(ctx, cmd)
	projectContext := resolvedCommandService.projectContext

	if err != nil {
		return err
	}

	path := resolveWorkEventStreamPath(projectContext.ProjectDir, cmd.Args().First())

	stream, err := core.ReadAgentEventStream(ctx, path)
	if err != nil {
		return err
	}

	summary := core.SummarizeAgentEventStream(stream)

	return printWorkEventStreamSummary(factory.io.Stdout, projectContext.ProjectDir, summary)
}

func (factory commandFactory) workLogsAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("work logs requires exactly one execution id argument")
	}

	operations, err := factory.projectService(ctx, cmd)
	projectContext := operations.projectContext

	if err != nil {
		return err
	}

	repository, err := core.NewExecutionRepository(ctx, operations.env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Lookup result is already complete at this point.
		_ = repository.Close()
	}()

	record, err := repository.GetExecution(ctx, cmd.Args().First())
	if err != nil {
		return err
	}

	if record.ProjectID != projectContext.ProjectID {
		return fmt.Errorf("%w: %s", core.ErrExecutionNotFound, cmd.Args().First())
	}

	return printWorkExecutionLogs(factory.io.Stdout, projectContext.ProjectDir, record)
}

func resolveWorkEventStreamPath(projectDir, path string) string {
	path = strings.TrimSpace(path)
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	return filepath.Join(projectDir, filepath.FromSlash(path))
}

func agentWorkSelectionOptionsFromCommand(cmd *cli.Command) AgentWorkSelectionOptions {
	return AgentWorkSelectionOptions{
		Statuses: cmd.StringSlice("status"),
		Filters: core.TicketLoopFilters{
			Priorities:   cmd.StringSlice("priority"),
			Severities:   cmd.StringSlice("severity"),
			ProjectAreas: cmd.StringSlice("project-area"),
			Tags:         cmd.StringSlice("tag"),
			Types:        cmd.StringSlice("type"),
			UseCases:     cmd.StringSlice("use-case"),
		},
		Order:         core.TicketLoopOrder(cmd.String("order")),
		MaxIterations: cmd.Int("max-iterations"),
	}
}

func printAgentWorkTicket(stdout io.Writer, result AgentWorkResult) error {
	lines := []string{
		"work: ticket",
		fmt.Sprintf("message: %s", result.Message),
	}

	lines = append(lines, agentWorkAuthorityLines(result)...)
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
		fmt.Sprintf("cycle_ticket: %s", agentWorkPreviewCycleTicket(result.SelectedQueue)),
		fmt.Sprintf("remaining_eligible: %d", agentWorkPreviewRemainingEligible(result)),
	}

	lines = append(lines, agentWorkAuthorityLines(result)...)
	if err := printAgentWorkLines(stdout, lines); err != nil {
		return err
	}

	return printAgentWorkTicketQueue(stdout, result.SelectedQueue)
}

func agentWorkPreviewCycleTicket(queue []core.IndexedTicket) string {
	if len(queue) == 0 {
		return "(none)"
	}

	return queue[0].Key
}

func agentWorkPreviewRemainingEligible(result AgentWorkResult) int {
	remaining := result.EligibleCount - len(result.SelectedQueue)
	if remaining < 0 {
		return 0
	}

	return remaining
}

func printAgentWorkRunAllExecution(stdout io.Writer, result AgentWorkResult) error {
	lines := []string{
		"work: run-all",
		fmt.Sprintf("message: %s", result.Message),
	}

	lines = append(lines, agentWorkAuthorityLines(result)...)
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

	for _, merge := range result.Merges {
		_, err := fmt.Fprintf(stdout, "merge: %s\t%s\t%s\t%s\n", merge.TicketKey, merge.MergeState, merge.MergeResult, merge.MergeCommit)
		if err != nil {
			return fmt.Errorf("write agent work result: %w", err)
		}
	}

	return nil
}

func printWorkExecutionLogs(stdout io.Writer, projectDir string, record core.ExecutionRecord) error {
	entries := []struct {
		label string
		path  string
	}{
		{label: "run_dir", path: record.Artifacts.RunDir},
		{label: "runtime_log", path: record.Artifacts.LogPath},
		{label: "validation_log", path: record.Artifacts.ValidationPath},
		{label: "result", path: record.Artifacts.ResultPath},
		{label: "execution_report", path: core.ExecutionReportPath(record)},
		{label: "events", path: core.ExecutionAgentEventStreamPath(record)},
	}

	lines := []string{
		"work: logs",
		fmt.Sprintf("execution: %s", record.ID),
		fmt.Sprintf("ticket: %s", record.TicketKey),
		fmt.Sprintf("status: %s", record.Status),
		fmt.Sprintf("validation_command: %s", record.Runtime.ValidationCommand),
	}

	for _, entry := range entries {
		exists, err := workLogPathExists(entry.path)
		if err != nil {
			return err
		}

		lines = append(lines,
			fmt.Sprintf("%s_path: %s", entry.label, relativeProjectPath(projectDir, entry.path)),
			fmt.Sprintf("%s_exists: %t", entry.label, exists),
		)
	}

	return printAgentWorkLines(stdout, lines)
}

func printWorkEventStreamSummary(stdout io.Writer, projectDir string, summary core.AgentEventStreamSummary) error {
	lines := []string{
		"work: events",
		fmt.Sprintf("path: %s", relativeProjectPath(projectDir, summary.Path)),
		fmt.Sprintf("events: %d", summary.EventCount),
		fmt.Sprintf("malformed: %d", summary.MalformedCount),
		fmt.Sprintf("degraded: %t", summary.Degraded),
		fmt.Sprintf("status: %s", summary.Status),
		fmt.Sprintf("latest_time: %s", formatWorkEventSummaryTime(summary.LatestTimestamp)),
		fmt.Sprintf("latest_loop: %s", emptyWorkEventSummaryField(summary.LatestLoopID)),
		fmt.Sprintf("latest_ticket: %s", emptyWorkEventSummaryField(summary.LatestTicketKey)),
		fmt.Sprintf("latest_execution: %s", emptyWorkEventSummaryField(summary.LatestExecutionID)),
		fmt.Sprintf("latest_phase: %s", emptyWorkEventSummaryField(string(summary.LatestPhase))),
		fmt.Sprintf("latest_level: %s", emptyWorkEventSummaryField(string(summary.LatestLevel))),
		fmt.Sprintf("latest_message: %s", emptyWorkEventSummaryField(summary.LatestMessage)),
	}

	return printAgentWorkLines(stdout, lines)
}

func formatWorkEventSummaryTime(value time.Time) string {
	if value.IsZero() {
		return "none"
	}

	return value.Format(time.RFC3339)
}

func emptyWorkEventSummaryField(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}

	return value
}

func workLogPathExists(path string) (bool, error) {
	if path == "" {
		return false, nil
	}

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("inspect work log path %s: %w", path, err)
}

func printAgentWorkTicketQueue(stdout io.Writer, tickets []core.IndexedTicket) error {
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

func printWorkReadinessReport(stdout io.Writer, projectContext commandProjectContext, report core.WorkReadinessReport) error {
	if err := printAgentWorkLines(stdout, workReadinessReportLines(projectContext, report)); err != nil {
		return err
	}

	return printWorkReadinessDiagnostics(stdout, report)
}

func printWorkRepairResult(stdout io.Writer, projectContext commandProjectContext, result core.WorkReadinessRepairResult) error {
	lines := workReadinessReportLines(projectContext, result.After)
	lines = append(lines[:2], append([]string{
		"fix_safe: true",
		fmt.Sprintf("repairs_applied: %d", len(result.Applied)),
	}, lines[2:]...)...)

	if err := printAgentWorkLines(stdout, lines); err != nil {
		return err
	}

	for _, repair := range result.Applied {
		_, err := fmt.Fprintf(stdout, "repair_applied: %s\t%s\n", repair.Code, repair.Detail)
		if err != nil {
			return fmt.Errorf("write agent work result: %w", err)
		}
	}

	return printWorkReadinessDiagnostics(stdout, result.After)
}

func workReadinessReportLines(projectContext commandProjectContext, report core.WorkReadinessReport) []string {
	lines := []string{
		"work: doctor",
		fmt.Sprintf("project: %s", projectContext.ProjectID),
		fmt.Sprintf("status: %s", report.Status),
		fmt.Sprintf("mode: %s", report.Policy.Mode),
		fmt.Sprintf("workspace_mode: %s", report.Workspace.Mode),
		fmt.Sprintf("integration_branch: %s", emptyWorkDoctorField(report.Policy.IntegrationBranch)),
		fmt.Sprintf("canonical_remote: %s", emptyWorkDoctorField(report.Policy.CanonicalRemote)),
		fmt.Sprintf("require_clean_worktree: %t", report.Policy.RequireCleanWorktree),
		fmt.Sprintf("require_remote_tracking: %t", report.Policy.RequireRemoteTracking),
		fmt.Sprintf("current_branch: %s", emptyWorkDoctorField(report.CurrentBranch)),
		fmt.Sprintf("user_checkout_state: %s", emptyWorkDoctorState(report.UserCheckout.State)),
		fmt.Sprintf("user_checkout_path: %s", workDoctorPathField(projectContext.ProjectDir, report.UserCheckout.Path)),
		fmt.Sprintf("ticket_key: %s", emptyWorkDoctorField(report.TicketKey)),
		fmt.Sprintf("ticket_branch: %s", emptyWorkDoctorField(report.TicketBranch)),
		fmt.Sprintf("ticket_worktree_state: %s", emptyWorkDoctorState(report.TicketWorktree.State)),
		fmt.Sprintf("ticket_worktree_path: %s", workDoctorPathField(projectContext.ProjectDir, report.TicketWorktree.Path)),
		fmt.Sprintf("ticket_worktree_base_ref: %s", emptyWorkDoctorField(report.TicketWorktree.BaseRef)),
		fmt.Sprintf("ticket_worktree_base_sha: %s", emptyWorkDoctorField(report.TicketWorktree.BaseSHA)),
		fmt.Sprintf("diagnostics: %d", len(report.Diagnostics)),
	}

	return lines
}

func printWorkReadinessDiagnostics(stdout io.Writer, report core.WorkReadinessReport) error {
	for _, diagnostic := range report.Diagnostics {
		_, err := fmt.Fprintf(
			stdout,
			"diagnostic: %s\t%s\t%s\n",
			diagnostic.Code,
			diagnostic.Severity,
			diagnostic.Detail,
		)
		if err != nil {
			return fmt.Errorf("write agent work result: %w", err)
		}

		if diagnostic.Repair.Code == "" {
			continue
		}

		_, err = fmt.Fprintf(
			stdout,
			"repair: %s\t%s\t%t\t%t\t%s\n",
			diagnostic.Code,
			diagnostic.Repair.Code,
			diagnostic.Repair.Available,
			diagnostic.Repair.Safe,
			diagnostic.Repair.Detail,
		)
		if err != nil {
			return fmt.Errorf("write agent work result: %w", err)
		}
	}

	return nil
}

func emptyWorkDoctorField(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}

	return value
}

func emptyWorkDoctorState(value core.WorktreeState) string {
	if value == "" {
		return "unknown"
	}

	return string(value)
}

func workDoctorPathField(projectDir, path string) string {
	if strings.TrimSpace(path) == "" {
		return "none"
	}

	return emptyWorkDoctorField(relativeProjectPath(projectDir, path))
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

func agentWorkAuthorityLines(result AgentWorkResult) []string {
	if result.Authority == "" {
		return nil
	}

	scope := "unadopted_legacy"
	if result.Authority == core.WorkStateAuthorityTickets {
		scope = "loose"
	}

	if result.Scope != "" {
		scope = result.Scope
	}

	lines := []string{fmt.Sprintf("authority: %s", result.Authority)}
	if result.Workflow != nil {
		scope = result.Workflow.Selection.Selector.Extent
		lines = append(lines, fmt.Sprintf("invocation: %s", result.Workflow.ID), fmt.Sprintf("status: %s", result.Workflow.Status), fmt.Sprintf("confirmed: %d", result.Workflow.ConfirmedCursor))
	}

	lines = append(lines, "scope: "+scope)

	keys := make([]string, 0, len(result.Excluded))
	for key := range result.Excluded {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("excluded_%s: %d", key, result.Excluded[key]))
	}

	return lines
}

type workDoctorTicket struct {
	Key    string
	Branch string
}
