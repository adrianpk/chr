package app

import (
	"context"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

var newAgenticReviewRunner = func() core.AgenticReviewRunner {
	return core.HarnessAgenticReviewRunner{Runner: codex.Runner{}}
}

func (s service) repairAgentWorkCycleTicket(state *agentWorkIntegrationCycleState, opts agentWorkPostDevelopmentOptions, ticket core.IndexedTicket) (core.IndexedTicket, core.ExecutionRecord, bool, error) {
	if !agentWorkCycleRunsRepair(opts.Policy, ticket) {
		return ticket, opts.Execution, true, nil
	}

	if !core.TicketReviewRepairAllowed(ticket, core.DefaultTicketReviewRepairLimit) {
		return ticket, opts.Execution, true, nil
	}

	emitRepairStartedAgentEvent(s.ctx, opts.LoopSink, state.Record, ticket)

	execution, application, err := s.executeLoopTicket(s.ctx, ticket, ticketExecutionEvents{
		LoopID: state.Record.ID,
		Sink:   opts.LoopSink,
	})
	if err != nil {
		return ticket, execution, false, err
	}

	state.Result.Executions = append(state.Result.Executions, execution)
	state.Result.Applications = append(state.Result.Applications, application)

	if application.TicketStatus == core.TicketResultStatusBlocked {
		emitRepairFinishedAgentEvent(s.ctx, opts.LoopSink, state.Record, execution, ticket)

		return ticket, execution, true, nil
	}

	reviewedTicket, stop, err := s.reviewAgentWorkCycleTicket(agentWorkPostDevelopmentOptions{
		Ticket:       ticket,
		Policy:       opts.Policy,
		Record:       state.Record,
		Execution:    execution,
		Application:  application,
		Selection:    opts.Selection,
		ReviewConfig: opts.ReviewConfig,
		LoopSink:     opts.LoopSink,
	})
	if err != nil {
		return core.IndexedTicket{}, execution, false, err
	}

	emitRepairFinishedAgentEvent(s.ctx, opts.LoopSink, state.Record, execution, reviewedTicket)

	return reviewedTicket, execution, stop, nil
}

func (s service) reviewAgentWorkCycleTicket(opts agentWorkPostDevelopmentOptions) (core.IndexedTicket, bool, error) {
	cycleTicket := opts.Ticket
	runsAgenticReview := agentWorkCycleRunsAgenticReview(opts.Policy, opts.Application)
	runsTrustedReview := agentWorkCycleRunsTrustedAutoMergeReview(opts.Policy, opts.Application)

	var err error

	if runsAgenticReview {
		emitReviewStartedAgentEvent(s.ctx, opts.LoopSink, opts.Record, opts.Execution, opts.Ticket, opts.Policy.Mode)

		cycleTicket, err = s.runAgenticReviewCycle(opts.Ticket, opts.Execution, opts.Application, opts.ReviewConfig)
	}

	if runsTrustedReview {
		emitReviewStartedAgentEvent(s.ctx, opts.LoopSink, opts.Record, opts.Execution, opts.Ticket, opts.Policy.Mode)

		cycleTicket, err = s.runTrustedAutoMergeReviewCycle(opts.Ticket, opts.Execution, opts.Application)
	}

	if err != nil {
		return core.IndexedTicket{}, false, err
	}

	if runsAgenticReview || runsTrustedReview {
		emitReviewFinishedAgentEvent(s.ctx, opts.LoopSink, opts.Record, opts.Execution, cycleTicket)
	}

	stop := agentWorkCycleStopsBeforeMerge(opts.Policy) || !agentWorkCycleTicketReadyForMerge(opts.Policy, cycleTicket)

	return cycleTicket, stop, nil
}

func (s service) runAgenticReviewCycle(selected core.IndexedTicket, execution core.ExecutionRecord, application core.ResultApplication, config core.ReviewWorkflowConfig) (core.IndexedTicket, error) {
	reviewedAt := s.now().UTC()
	reviewProjectDir := agenticReviewProjectDir(s.projectContext, execution)

	ticket, err := parseAgenticReviewTicket(s.ctx, s.projectContext.ProjectID, reviewProjectDir, application.TicketPath, reviewedAt)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	ticket = agentWorkTicketWithSelectedReviewMetadata(ticket, selected)
	s.generationExecutionDir = reviewProjectDir

	bound, _, err := s.bindGeneration(s.ctx, "review-"+execution.ID, "agentic_review")
	if err != nil {
		return core.IndexedTicket{}, err
	}

	s = bound

	runner := newAgenticReviewRunner()
	if native, ok := runner.(core.HarnessAgenticReviewRunner); ok && s.generationRunner != nil {
		native.Runner = s.generationRunner
		runner = native
	}

	artifact, err := runner.Review(s.ctx, core.AgenticReviewContext{
		ProjectID:           s.projectContext.ProjectID,
		ProjectDir:          reviewProjectDir,
		Ticket:              ticket,
		Execution:           execution,
		Application:         application,
		ExecutionReportPath: core.ExecutionReportPath(execution),
	})
	if err != nil {
		return core.IndexedTicket{}, err
	}

	result, err := core.ApplyAgenticReviewResult(s.ctx, core.ApplyAgenticReviewResultOptions{
		Ticket:     ticket,
		Artifact:   artifact,
		ReviewedAt: reviewedAt,
	})
	if err != nil {
		return core.IndexedTicket{}, err
	}

	err = writeAgenticReviewNonBlockingComments(s.ctx, ticket.Path, artifact.Result.NonBlockingNotes, config.NonBlockingCommentLimit, reviewedAt)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	if err := s.refreshAgenticReviewTicketIndex(execution, ticket.Path, reviewedAt); err != nil {
		return core.IndexedTicket{}, err
	}

	ticket.Status = result.Status
	ticket.ReviewState = result.ReviewState
	ticket.ReviewResult = result.ReviewResult
	ticket.ReviewRound = result.ReviewRound

	return ticket, nil
}

func (s service) runTrustedAutoMergeReviewCycle(selected core.IndexedTicket, execution core.ExecutionRecord, application core.ResultApplication) (core.IndexedTicket, error) {
	reviewedAt := s.now().UTC()
	reviewProjectDir := agenticReviewProjectDir(s.projectContext, execution)

	ticket, err := parseAgenticReviewTicket(s.ctx, s.projectContext.ProjectID, reviewProjectDir, application.TicketPath, reviewedAt)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	ticket = agentWorkTicketWithSelectedReviewMetadata(ticket, selected)

	result, err := core.ApplyTrustedAutoMergeReview(s.ctx, core.ApplyTrustedAutoMergeReviewOptions{
		Ticket:     ticket,
		ReviewedAt: reviewedAt,
	})
	if err != nil {
		return core.IndexedTicket{}, err
	}

	if err := s.refreshAgenticReviewTicketIndex(execution, ticket.Path, reviewedAt); err != nil {
		return core.IndexedTicket{}, err
	}

	ticket.Status = result.Status
	ticket.ReviewState = result.ReviewState
	ticket.ReviewResult = result.ReviewResult
	ticket.ReviewRound = result.ReviewRound

	return ticket, nil
}

func agentWorkTicketWithSelectedReviewMetadata(ticket, selected core.IndexedTicket) core.IndexedTicket {
	if ticket.Branch == "" {
		ticket.Branch = selected.Branch
	}

	if len(ticket.ReviewRefs) == 0 {
		ticket.ReviewRefs = selected.ReviewRefs
	}

	return ticket
}

func parseAgenticReviewTicket(ctx context.Context, projectID, projectDir, ticketPath string, indexedAt time.Time) (core.IndexedTicket, error) {
	return core.NewTicketParser().ParseTicket(ctx, core.ParseTicketOptions{
		ProjectID:  projectID,
		ProjectDir: projectDir,
		Path:       ticketPath,
		IndexedAt:  indexedAt,
	})
}

func writeAgenticReviewNonBlockingComments(ctx context.Context, ticketPath string, notes []core.AgenticReviewFinding, limit int, createdAt time.Time) error {
	for index, note := range limitedAgenticReviewNotes(notes, limit) {
		err := core.NewTicketCommenter().AddComment(ctx, core.AddTicketCommentOptions{
			Path:      ticketPath,
			Author:    "chr-reviewer",
			Body:      renderAgenticReviewNonBlockingComment(note),
			CreatedAt: createdAt.Add(time.Duration(index) * time.Nanosecond),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func limitedAgenticReviewNotes(notes []core.AgenticReviewFinding, limit int) []core.AgenticReviewFinding {
	if limit <= 0 || len(notes) == 0 {
		return nil
	}

	if len(notes) <= limit {
		return notes
	}

	return notes[:limit]
}

func renderAgenticReviewNonBlockingComment(note core.AgenticReviewFinding) string {
	return fmt.Sprintf("Agentic review note (%s): %s\n\n%s", note.Category, note.Title, note.Detail)
}

func (s service) refreshAgenticReviewTicketIndex(execution core.ExecutionRecord, ticketPath string, reviewedAt time.Time) error {
	if execution.Workspace.Mode == core.WorkWorkspaceModeWorktree {
		return nil
	}

	index, err := core.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = index.Close()
	}()

	_, err = core.NewTicketReconciler(index).ReconcileProjectTicket(s.ctx, core.ReconcileProjectTicketOptions{
		ProjectID:    s.projectContext.ProjectID,
		ProjectDir:   s.projectContext.ProjectDir,
		TicketPath:   ticketPath,
		ReconciledAt: reviewedAt,
	})

	return err
}

func agenticReviewProjectDir(projectContext commandProjectContext, execution core.ExecutionRecord) string {
	if execution.Workspace.Mode == core.WorkWorkspaceModeWorktree && execution.Workspace.ExecutionWorkspaceDir != "" {
		return execution.Workspace.ExecutionWorkspaceDir
	}

	return projectContext.ProjectDir
}
