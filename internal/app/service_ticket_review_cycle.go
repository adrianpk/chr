package app

import (
	"context"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

var newAgenticReviewRunner = func() core.AgenticReviewRunner {
	return core.HarnessAgenticReviewRunner{Runner: codex.Runner{}}
}

func (s service) repairAgentWorkCycleTicket(ctx context.Context, state *agentWorkIntegrationCycleState, opts agentWorkPostDevelopmentOptions, ticket ticketcore.Indexed) (agentWorkRepairResult, error) {
	if !agentWorkCycleRunsRepair(opts.Policy, ticket) {
		return agentWorkRepairResult{Ticket: ticket, Execution: opts.Execution, Stop: true}, nil
	}

	if !ticketcore.ReviewRepairAllowed(ticket, ticketcore.DefaultReviewRepairLimit) {
		return agentWorkRepairResult{Ticket: ticket, Execution: opts.Execution, Stop: true}, nil
	}

	emitRepairStartedAgentEvent(ctx, opts.LoopSink, state.Record, ticket)

	ticketOutcome, err := s.executeLoopTicket(ctx, ticket, ticketExecutionEvents{
		LoopID: state.Record.ID,
		Sink:   opts.LoopSink,
	})

	execution, application := ticketOutcome.Record, ticketOutcome.Application
	if err != nil {
		return agentWorkRepairResult{Ticket: ticket, Execution: execution, Stop: false}, err
	}

	state.Result.Executions = append(state.Result.Executions, execution)
	state.Result.Applications = append(state.Result.Applications, application)

	if application.TicketStatus == core.TicketResultStatusBlocked {
		emitRepairFinishedAgentEvent(ctx, opts.LoopSink, state.Record, execution, ticket)

		return agentWorkRepairResult{Ticket: ticket, Execution: execution, Stop: true}, nil
	}

	reviewResult, err := s.reviewAgentWorkCycleTicket(ctx, agentWorkPostDevelopmentOptions{
		Ticket:       ticket,
		Policy:       opts.Policy,
		Record:       state.Record,
		Execution:    execution,
		Application:  application,
		Selection:    opts.Selection,
		ReviewConfig: opts.ReviewConfig,
		LoopSink:     opts.LoopSink,
	})

	reviewedTicket, stop := reviewResult.Ticket, reviewResult.Stop
	if err != nil {
		return agentWorkRepairResult{Ticket: ticketcore.Indexed{}, Execution: execution, Stop: false}, err
	}

	emitRepairFinishedAgentEvent(ctx, opts.LoopSink, state.Record, execution, reviewedTicket)

	return agentWorkRepairResult{Ticket: reviewedTicket, Execution: execution, Stop: stop}, nil
}

func (s service) reviewAgentWorkCycleTicket(ctx context.Context, opts agentWorkPostDevelopmentOptions) (agentWorkReviewResult, error) {
	cycleTicket := opts.Ticket
	runsAgenticReview := agentWorkCycleRunsAgenticReview(opts.Policy, opts.Application)
	runsTrustedReview := agentWorkCycleRunsTrustedAutoMergeReview(opts.Policy, opts.Application)

	var err error

	if runsAgenticReview {
		emitReviewStartedAgentEvent(ctx, opts.LoopSink, reviewStartedEvent{Record: opts.Record, Execution: opts.Execution, Ticket: opts.Ticket, Mode: opts.Policy.Mode})

		cycleTicket, err = s.runAgenticReviewCycle(ctx, opts.Ticket, opts.Execution, opts.Application, opts.ReviewConfig)
	}

	if runsTrustedReview {
		emitReviewStartedAgentEvent(ctx, opts.LoopSink, reviewStartedEvent{Record: opts.Record, Execution: opts.Execution, Ticket: opts.Ticket, Mode: opts.Policy.Mode})

		cycleTicket, err = s.runTrustedAutoMergeReviewCycle(opts.Ticket, opts.Execution, opts.Application)
	}

	if err != nil {
		return agentWorkReviewResult{Ticket: ticketcore.Indexed{}, Stop: false}, err
	}

	if runsAgenticReview || runsTrustedReview {
		emitReviewFinishedAgentEvent(ctx, opts.LoopSink, opts.Record, opts.Execution, cycleTicket)
	}

	stop := agentWorkCycleStopsBeforeMerge(opts.Policy) || !agentWorkCycleTicketReadyForMerge(opts.Policy, cycleTicket)

	return agentWorkReviewResult{Ticket: cycleTicket, Stop: stop}, nil
}

func (s service) runAgenticReviewCycle(ctx context.Context, selected ticketcore.Indexed, execution executioncore.Record, application core.ResultApplication, config workspaceconfig.ReviewWorkflowConfig) (ticketcore.Indexed, error) {
	reviewedAt := s.now().UTC()
	reviewProjectDir := agenticReviewProjectDir(s.projectContext, execution)

	ticket, err := parseAgenticReviewTicket(ctx, s.projectContext.ProjectID, reviewProjectDir, application.TicketPath, reviewedAt)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	ticket = agentWorkTicketWithSelectedReviewMetadata(ticket, selected)
	s.generationExecutionDir = reviewProjectDir

	bound, err := s.bindGeneration(ctx, "review-"+execution.ID, "agentic_review")
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	ctx = bound.context(ctx)

	runner := newAgenticReviewRunner()
	if native, ok := runner.(core.HarnessAgenticReviewRunner); ok && bound.generationRunner != nil {
		native.Runner = bound.generationRunner
		runner = native
	}

	artifact, err := runner.Review(ctx, core.AgenticReviewContext{
		ProjectID:           s.projectContext.ProjectID,
		ProjectDir:          reviewProjectDir,
		Ticket:              ticket,
		Execution:           execution,
		Application:         application,
		ExecutionReportPath: executioncore.ReportPath(execution),
	})
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	result, err := core.ApplyAgenticReviewResult(ctx, core.ApplyAgenticReviewResultOptions{
		Ticket:     ticket,
		Artifact:   artifact,
		ReviewedAt: reviewedAt,
	})
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	err = writeAgenticReviewNonBlockingComments(ctx, ticket.Path, artifact.Result.NonBlockingNotes, config.NonBlockingCommentLimit, reviewedAt)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	if err := s.refreshAgenticReviewTicketIndex(execution, ticket.Path, reviewedAt); err != nil {
		return ticketcore.Indexed{}, err
	}

	ticket.Status = result.Status
	ticket.ReviewState = result.ReviewState
	ticket.ReviewResult = result.ReviewResult
	ticket.ReviewRound = result.ReviewRound

	return ticket, nil
}

func (s service) runTrustedAutoMergeReviewCycle(selected ticketcore.Indexed, execution executioncore.Record, application core.ResultApplication) (ticketcore.Indexed, error) {
	reviewedAt := s.now().UTC()
	reviewProjectDir := agenticReviewProjectDir(s.projectContext, execution)

	ticket, err := parseAgenticReviewTicket(s.ctx, s.projectContext.ProjectID, reviewProjectDir, application.TicketPath, reviewedAt)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	ticket = agentWorkTicketWithSelectedReviewMetadata(ticket, selected)

	result, err := core.ApplyTrustedAutoMergeReview(s.ctx, core.ApplyTrustedAutoMergeReviewOptions{
		Ticket:     ticket,
		ReviewedAt: reviewedAt,
	})
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	if err := s.refreshAgenticReviewTicketIndex(execution, ticket.Path, reviewedAt); err != nil {
		return ticketcore.Indexed{}, err
	}

	ticket.Status = result.Status
	ticket.ReviewState = result.ReviewState
	ticket.ReviewResult = result.ReviewResult
	ticket.ReviewRound = result.ReviewRound

	return ticket, nil
}

func agentWorkTicketWithSelectedReviewMetadata(ticket, selected ticketcore.Indexed) ticketcore.Indexed {
	if ticket.Branch == "" {
		ticket.Branch = selected.Branch
	}

	if len(ticket.ReviewRefs) == 0 {
		ticket.ReviewRefs = selected.ReviewRefs
	}

	return ticket
}

func parseAgenticReviewTicket(ctx context.Context, projectID, projectDir, ticketPath string, indexedAt time.Time) (ticketcore.Indexed, error) {
	return ticketcore.NewParser().Parse(ctx, ticketcore.ParseOptions{
		ProjectID:  projectID,
		ProjectDir: projectDir,
		Path:       ticketPath,
		IndexedAt:  indexedAt,
	})
}

func writeAgenticReviewNonBlockingComments(ctx context.Context, ticketPath string, notes []workflowstate.ReviewFinding, limit int, createdAt time.Time) error {
	for index, note := range limitedAgenticReviewNotes(notes, limit) {
		err := workflowcore.NewTicketCommenter().AddComment(ctx, workflowcore.AddTicketCommentOptions{
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

func limitedAgenticReviewNotes(notes []workflowstate.ReviewFinding, limit int) []workflowstate.ReviewFinding {
	if limit <= 0 || len(notes) == 0 {
		return nil
	}

	if len(notes) <= limit {
		return notes
	}

	return notes[:limit]
}

func renderAgenticReviewNonBlockingComment(note workflowstate.ReviewFinding) string {
	return fmt.Sprintf("Agentic review note (%s): %s\n\n%s", note.Category, note.Title, note.Detail)
}

func (s service) refreshAgenticReviewTicketIndex(execution executioncore.Record, ticketPath string, reviewedAt time.Time) error {
	if execution.Workspace.Mode == workspacecore.ExecutionWorktree {
		return nil
	}

	index, err := workflowcore.NewWorkflowTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = index.Close()
	}()

	_, err = ticketcore.NewReconciler(index).ReconcileTicket(s.ctx, ticketcore.ReconcileTicketOptions{
		ProjectID:    s.projectContext.ProjectID,
		ProjectDir:   s.projectContext.ProjectDir,
		TicketPath:   ticketPath,
		ReconciledAt: reviewedAt,
	})

	return err
}

func agenticReviewProjectDir(projectContext commandProjectContext, execution executioncore.Record) string {
	if execution.Workspace.Mode == workspacecore.ExecutionWorktree && execution.Workspace.ExecutionWorkspaceDir != "" {
		return execution.Workspace.ExecutionWorkspaceDir
	}

	return projectContext.ProjectDir
}

type agentWorkRepairResult struct {
	Ticket    ticketcore.Indexed
	Execution executioncore.Record
	Stop      bool
}

type agentWorkReviewResult struct {
	Ticket ticketcore.Indexed
	Stop   bool
}
