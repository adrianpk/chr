package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
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

type ticketApplyResultPlan struct {
	ProjectContext commandProjectContext
	Record         core.ExecutionRecord
	Result         core.ExecutionResult
	Application    core.ResultApplication
	ReportPath     string
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
		err = applyTicketResult(ctx, env, plan, ticketExecutionEvents{}, factory.deps.assetsFS)
		if err != nil {
			return err
		}
	}

	return printTicketApplyResult(factory.io.Stdout, plan, cmd.Bool("dry-run"))
}

func (factory commandFactory) buildTicketApplyResultPlan(
	ctx context.Context,
	cmd *cli.Command,
	env core.UserEnvironment,
) (ticketApplyResultPlan, error) {
	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	repository, err := core.NewExecutionRepository(ctx, env, factory.deps.assetsFS)
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
		CommentAuthor:     "chr",
	})
	if err != nil {
		return ticketApplyResultPlan{}, err
	}

	return ticketApplyResultPlan{
		ProjectContext: projectContext,
		Record:         record,
		Result:         result,
		Application:    application,
		ReportPath:     ticketResultApplicationPath(record),
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

func ticketResultApplicationPath(record core.ExecutionRecord) string {
	if record.Artifacts.ArtifactID != "" {
		return filepath.Join(record.Artifacts.RunDir, "evidence", "result-application.md")
	}

	return filepath.Join(record.Artifacts.RunDir, "result-application.md")
}

func applyTicketResult(ctx context.Context, env core.UserEnvironment, plan ticketApplyResultPlan, events ticketExecutionEvents, assetsFS fs.FS) error {
	appliedAt := time.Now().UTC()

	err := writeTicketResultState(ctx, plan, appliedAt)
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

	if err := refreshTicketArtifactMetadata(plan.Record); err != nil {
		return err
	}

	return emitTicketBlockedNotification(ctx, env, plan, assetsFS)
}

func writeTicketResultState(ctx context.Context, plan ticketApplyResultPlan, appliedAt time.Time) error {
	if err := core.NewTicketResultStateWriter().WriteResultState(ctx, core.WriteResultStateOptions{
		Application: plan.Application,
		UpdatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	if err := core.NewTicketResultCommentWriter().WriteResultComment(ctx, core.WriteResultCommentOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	if err := core.NewTicketResultValidationWriter().WriteResultValidation(ctx, core.WriteResultValidationOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	if err := core.NewTicketResultChangedFilesWriter().WriteChangedFiles(ctx, core.WriteChangedFilesOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	}); err != nil {
		return err
	}

	return core.NewTicketResultBlockedReasonWriter().WriteBlockedReason(ctx, core.WriteBlockedReasonOptions{
		Application: plan.Application,
		CreatedAt:   appliedAt,
	})
}

func refreshTicketArtifactMetadata(record core.ExecutionRecord) error {
	if record.Artifacts.ArtifactID == "" {
		return nil
	}

	return core.RefreshOpsArtifactMetadata(record.Artifacts.RunDir, string(record.Status), time.Now().UTC())
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
