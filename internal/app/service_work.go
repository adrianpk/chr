package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkOperationResult is the TUI-facing result for agent work operations.
type WorkOperationResult struct {
	Message string
	Tickets []core.IndexedTicket
}

// AgentWorkResult summarizes runnable agent work for the dashboard.
type AgentWorkResult struct {
	Message       string
	Tickets       []core.IndexedTicket
	Execution     *core.ExecutionRecord
	LoopRun       *core.LoopRunRecord
	Executions    []core.ExecutionRecord
	Merges        []core.TicketMergeCommandResult
	SelectedQueue []core.IndexedTicket
	EligibleCount int
}

// AgentWorkSelectionOptions filters agent work selection.
type AgentWorkSelectionOptions struct {
	Statuses      []string
	Filters       core.TicketLoopFilters
	Order         core.TicketLoopOrder
	MaxIterations int
}

type agentWorkLoopSelection struct {
	Tickets       []core.IndexedTicket
	EligibleCount int
	ReviewConfig  core.ReviewWorkflowConfig
	WorkConfig    core.WorkConfig
}

type agentWorkIntegrationCycleState struct {
	Result        ticketLoopExecutionResult
	Record        core.LoopRunRecord
	SelectedQueue []core.IndexedTicket
}

type agentWorkMergeCycleOptions struct {
	Ticket       core.IndexedTicket
	Execution    core.ExecutionRecord
	WorkConfig   core.WorkConfig
	ReviewConfig core.ReviewWorkflowConfig
}

type agentWorkPostDevelopmentOptions struct {
	Ticket       core.IndexedTicket
	Policy       core.IntegrationCyclePolicy
	Record       core.LoopRunRecord
	Execution    core.ExecutionRecord
	Application  core.ResultApplication
	Selection    agentWorkLoopSelection
	ReviewConfig core.ReviewWorkflowConfig
	LoopSink     core.AgentEventSink
}

var newAgenticReviewRunner = func() core.AgenticReviewRunner {
	return core.CodexAgenticReviewRunner{}
}

func (s service) StartSelectedTicket(ticketKey string) (WorkOperationResult, error) {
	result, err := s.StartAgentWorkTicket(ticketKey)
	if err != nil {
		return WorkOperationResult{}, err
	}

	return workOperationResultFromAgentWork(result), nil
}

func (s service) StartReadyTicketLoop() (WorkOperationResult, error) {
	result, err := s.StartReadyAgentWorkLoop()
	if err != nil {
		return WorkOperationResult{}, err
	}

	return workOperationResultFromAgentWork(result), nil
}

func (s service) StartAgentWorkTicket(ticketKey string) (AgentWorkResult, error) {
	ticket, preflight, release, err := s.prepareAgentWorkTicket(ticketKey)
	if err != nil {
		return AgentWorkResult{}, err
	}
	defer release()

	if !core.RuntimeValidationPassed(preflight.Runtime.Checks) {
		return AgentWorkResult{}, core.NewExecutionFailure(
			core.ExecutionFailureRuntime,
			errors.New("runtime validation failed"),
		)
	}

	events := ticketExecutionEvents{
		Sink: agentWorkExecutionEventSink(preflight.Record),
	}

	record, err := runTicketStartExecution(s.ctx, s.env, preflight.Record, preflight.Runtime, events, s.assetsFS)
	if err != nil {
		return AgentWorkResult{}, err
	}

	plan, err := buildTicketApplyResultPlanForRecord(s.ctx, s.projectContext, record)
	if err != nil {
		return AgentWorkResult{}, err
	}

	err = applyTicketResult(s.ctx, s.env, plan, events, s.assetsFS)
	if err != nil {
		return AgentWorkResult{}, err
	}

	result, err := s.reloadWorkOperationState()
	if err != nil {
		return AgentWorkResult{}, err
	}

	result.Message = "Execution " + record.ID + " for ticket " + ticket.Key + " finished with status " +
		string(record.Status) + ". Result applied to ticket. Result: " +
		displayPath(s.projectContext.ProjectDir, record.Artifacts.ResultPath)

	if record.FailureReason != "" {
		result.Message += " Failure: " + record.FailureReason
	}

	return AgentWorkResult{
		Message:   result.Message,
		Tickets:   result.Tickets,
		Execution: &record,
	}, nil
}

func (s service) prepareAgentWorkTicket(ticketKey string) (core.IndexedTicket, ticketStartPreflight, func(), error) {
	err := s.ensureAgentWorkRuntime()
	if err != nil {
		return core.IndexedTicket{}, ticketStartPreflight{}, noopRelease, err
	}

	ticket, err := s.ticketByKey(ticketKey)
	if err != nil {
		return core.IndexedTicket{}, ticketStartPreflight{}, noopRelease, err
	}

	tickets, err := s.listProjectTickets()
	if err != nil {
		return core.IndexedTicket{}, ticketStartPreflight{}, noopRelease, err
	}

	err = enforceWorkSelectionPreflight(s.ctx, s.projectContext, tickets, ticket.Key)
	if err != nil {
		return core.IndexedTicket{}, ticketStartPreflight{}, noopRelease, err
	}

	lock, err := acquireAgentOperationLock(s.ctx, s.projectContext, core.AgentOperationDevelopment)
	if err != nil {
		return core.IndexedTicket{}, ticketStartPreflight{}, noopRelease, err
	}

	preflight, err := buildTicketStartPreflightForTicket(s.ctx, s.env, s.projectContext, ticket, s.assetsFS)
	if err != nil {
		releaseErr := lock.Release()
		if releaseErr != nil {
			return core.IndexedTicket{}, ticketStartPreflight{}, noopRelease, releaseErr
		}

		return core.IndexedTicket{}, ticketStartPreflight{}, noopRelease, core.NewExecutionFailure(core.ExecutionFailurePreflight, err)
	}

	return ticket, preflight, releaseAgentOperationLock(lock), nil
}

func (s service) StartReadyAgentWorkLoop() (AgentWorkResult, error) {
	err := s.ensureAgentWorkRuntime()
	if err != nil {
		return AgentWorkResult{}, err
	}

	return s.startAgentWorkLoop(defaultAgentWorkSelectionOptions(1))
}

func (s service) StartAllReadyAgentWork() (AgentWorkResult, error) {
	return s.StartAgentWorkLoop(defaultAgentWorkSelectionOptions(0))
}

func (s service) StartAgentWorkLoop(opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	err := s.ensureAgentWorkRuntime()
	if err != nil {
		return AgentWorkResult{}, err
	}

	if err := validateAgentWorkSelectionOptions(opts); err != nil {
		return AgentWorkResult{}, err
	}

	return s.startAgentWorkLoop(opts)
}

func (s service) PreviewAllReadyAgentWork() (AgentWorkResult, error) {
	return s.PreviewAgentWorkLoop(defaultAgentWorkSelectionOptions(0))
}

func (s service) PreviewAgentWorkLoop(opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	if err := validateAgentWorkSelectionOptions(opts); err != nil {
		return AgentWorkResult{}, err
	}

	selection, err := s.agentWorkLoopSelection(opts)
	if err != nil {
		return AgentWorkResult{}, err
	}

	reloaded, err := s.reloadWorkOperationState()
	if err != nil {
		return AgentWorkResult{}, err
	}

	return AgentWorkResult{
		Message:       fmt.Sprintf("Ready agent-work queue contains %d eligible tickets.", selection.EligibleCount),
		Tickets:       reloaded.Tickets,
		SelectedQueue: selection.Tickets,
		EligibleCount: selection.EligibleCount,
	}, nil
}

func (s service) startAgentWorkLoop(opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	lock, err := acquireAgentOperationLock(s.ctx, s.projectContext, core.AgentOperationDevelopment)
	if err != nil {
		return AgentWorkResult{}, err
	}

	defer releaseAgentOperationLock(lock)()

	result, selectedQueue, eligibleCount, err := s.runAgentWorkIntegrationCycles(opts)
	if err != nil {
		return AgentWorkResult{}, err
	}

	reloaded, err := s.reloadWorkOperationState()
	if err != nil {
		return AgentWorkResult{}, err
	}

	reloaded.Message = "Loop " + result.Record.ID + " finished with status " + string(result.Record.Status) +
		fmt.Sprintf(": %d processed, %d blocked, %d failed.", result.Record.ProcessedCount, result.Record.BlockedCount, result.Record.FailedCount)

	if result.Record.FailureReason != "" {
		reloaded.Message += " Failure: " + result.Record.FailureReason
	}

	return AgentWorkResult{
		Message:       reloaded.Message,
		Tickets:       reloaded.Tickets,
		LoopRun:       &result.Record,
		Executions:    result.Executions,
		Merges:        result.Merges,
		SelectedQueue: selectedQueue,
		EligibleCount: eligibleCount,
	}, nil
}

func (s service) RequestDaemonStop() (WorkOperationResult, error) {
	requestedAt := s.now().UTC()

	stopResult, err := requestDaemonStopState(s.ctx, s.env, s.assetsFS, requestedAt)
	if err != nil {
		return WorkOperationResult{}, err
	}

	if stopResult.AlreadyStopped {
		result, reloadErr := s.reloadWorkOperationState()
		if reloadErr != nil {
			return WorkOperationResult{}, reloadErr
		}

		result.Message = "Daemon is already stopped."

		return result, nil
	}

	result, err := s.reloadWorkOperationState()
	if err != nil {
		return WorkOperationResult{}, err
	}

	result.Message = "Daemon stop requested at " + stopResult.RequestedAt.Format(time.RFC3339) + "."

	return result, nil
}

func (s service) runAgentWorkIntegrationCycles(opts AgentWorkSelectionOptions) (ticketLoopExecutionResult, []core.IndexedTicket, int, error) {
	repository, err := core.NewLoopRunRepository(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, nil, 0, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	selection, err := s.agentWorkLoopSelection(opts)
	if err != nil {
		return ticketLoopExecutionResult{}, nil, 0, err
	}

	record := newRunningLoopRunRecord(s.projectContext.ProjectID, selection.EligibleCount)
	if err := repository.CreateLoopRun(s.ctx, record); err != nil {
		return ticketLoopExecutionResult{}, nil, 0, err
	}

	loopSink := agentWorkLoopEventSink(s.projectContext.ProjectDir, record)
	emitLoopStartedAgentEvent(s.ctx, loopSink, record)

	state := agentWorkIntegrationCycleState{
		Result:        ticketLoopExecutionResult{Record: record},
		Record:        record,
		SelectedQueue: make([]core.IndexedTicket, 0),
	}
	initialEligibleCount := selection.EligibleCount
	processedCycles := 0

	for len(selection.Tickets) > 0 && agentWorkCycleLimitAllowsNext(opts, processedCycles) {
		var stop bool

		state, stop, err = s.runAgentWorkIntegrationCycle(selection, state, loopSink)
		if err != nil {
			return s.failAgentWorkIntegrationCycles(repository, state, loopSink, initialEligibleCount, err)
		}

		processedCycles++

		if stop {
			break
		}

		selection, err = s.agentWorkLoopSelection(opts)
		if err != nil {
			return s.failAgentWorkIntegrationCycles(repository, state, loopSink, initialEligibleCount, err)
		}
	}

	finished, err := finishTicketLoopExecution(s.ctx, s.env, repository, s.projectContext, state.Result, state.Record, loopSink, s.assetsFS)

	return finished, state.SelectedQueue, initialEligibleCount, err
}

func (s service) runAgentWorkIntegrationCycle(selection agentWorkLoopSelection, state agentWorkIntegrationCycleState, loopSink core.AgentEventSink) (agentWorkIntegrationCycleState, bool, error) {
	ticket := selection.Tickets[0]

	cyclePlan, err := core.PlanIntegrationCycle(core.IntegrationCyclePlanOptions{
		ProjectID:    s.projectContext.ProjectID,
		Ticket:       ticket,
		ReviewConfig: selection.ReviewConfig,
	})
	if err != nil {
		return state, false, err
	}

	state.SelectedQueue = append(state.SelectedQueue, ticket)
	emitTicketSelectedAgentEvent(s.ctx, loopSink, state.Record, ticket)

	execution, application, err := executeLoopTicket(s.ctx, s.env, s.projectContext, ticket, ticketExecutionEvents{
		LoopID: state.Record.ID,
		Sink:   loopSink,
	}, s.assetsFS)
	if err != nil {
		return state, false, err
	}

	state.Result.Executions = append(state.Result.Executions, execution)
	state.Result.Applications = append(state.Result.Applications, application)
	state.Record.ProcessedCount++

	stop, err := s.finishAgentWorkIntegrationCycle(&state, agentWorkPostDevelopmentOptions{
		Ticket:       ticket,
		Policy:       cyclePlan.Policy,
		Record:       state.Record,
		Execution:    execution,
		Application:  application,
		Selection:    selection,
		ReviewConfig: selection.ReviewConfig,
		LoopSink:     loopSink,
	})
	if err != nil {
		return state, false, err
	}

	return state, stop, nil
}

func (s service) finishAgentWorkIntegrationCycle(state *agentWorkIntegrationCycleState, opts agentWorkPostDevelopmentOptions) (bool, error) {
	emitCycleRehydratedAgentEvent(s.ctx, opts.LoopSink, state.Record, opts.Execution, opts.Application)

	if opts.Application.TicketStatus == core.TicketResultStatusBlocked {
		state.Record.BlockedCount++

		return true, nil
	}

	cycleTicket, stop, err := s.reviewAgentWorkCycleTicket(opts)
	if err != nil {
		return false, err
	}

	finalExecution := opts.Execution
	if stop {
		cycleTicket, finalExecution, stop, err = s.repairAgentWorkCycleTicket(state, opts, cycleTicket)
		if err != nil {
			return false, err
		}
	}

	if stop {
		state.Record.BlockedCount++

		return true, nil
	}

	if !opts.Policy.RequiresMerge {
		return false, nil
	}

	emitMergeStartedAgentEvent(s.ctx, opts.LoopSink, state.Record, finalExecution, cycleTicket)

	merge, err := s.runAgentWorkMergeCycle(agentWorkMergeCycleOptions{
		Ticket:       cycleTicket,
		Execution:    finalExecution,
		WorkConfig:   opts.Selection.WorkConfig,
		ReviewConfig: opts.ReviewConfig,
	})
	if err != nil {
		return false, err
	}

	state.Result.Merges = append(state.Result.Merges, merge)
	emitMergeFinishedAgentEvent(s.ctx, opts.LoopSink, state.Record, finalExecution, merge)

	if agentWorkCycleMergeBlocked(merge) {
		state.Record.BlockedCount++

		return true, nil
	}

	return false, nil
}

func (s service) repairAgentWorkCycleTicket(state *agentWorkIntegrationCycleState, opts agentWorkPostDevelopmentOptions, ticket core.IndexedTicket) (core.IndexedTicket, core.ExecutionRecord, bool, error) {
	if !agentWorkCycleRunsRepair(opts.Policy, ticket) {
		return ticket, opts.Execution, true, nil
	}

	if !core.TicketReviewRepairAllowed(ticket, core.DefaultTicketReviewRepairLimit) {
		return ticket, opts.Execution, true, nil
	}

	emitRepairStartedAgentEvent(s.ctx, opts.LoopSink, state.Record, ticket)

	execution, application, err := executeLoopTicket(s.ctx, s.env, s.projectContext, ticket, ticketExecutionEvents{
		LoopID: state.Record.ID,
		Sink:   opts.LoopSink,
	}, s.assetsFS)
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

	artifact, err := newAgenticReviewRunner().Review(s.ctx, core.AgenticReviewContext{
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

	result, err := core.ApplyAgenticReviewResult(core.ApplyAgenticReviewResultOptions{
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

	result, err := core.ApplyTrustedAutoMergeReview(core.ApplyTrustedAutoMergeReviewOptions{
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

func (s service) runAgentWorkMergeCycle(opts agentWorkMergeCycleOptions) (core.TicketMergeCommandResult, error) {
	mergedAt := s.now().UTC()

	tickets, err := s.agentWorkMergeTickets(opts.Execution, mergedAt)
	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	result, err := core.RunTicketMergeCommand(s.ctx, core.TicketMergeCommandOptions{
		Ticket:        opts.Ticket,
		Tickets:       tickets,
		Remote:        core.NewRemoteReviewAdapter(opts.ReviewConfig.Providers, reviewRemoteCommandRunner),
		ExpectedBase:  opts.WorkConfig.BranchPolicy.IntegrationBranch,
		MergeStrategy: core.TicketMergeStrategyRebase,
		ProjectDir:    s.projectContext.ProjectDir,
		BranchPolicy:  opts.WorkConfig.BranchPolicy,
		Workspace:     opts.WorkConfig.Workspace,
		MergedAt:      mergedAt,
	})
	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	if err := s.refreshAgenticReviewTicketIndex(opts.Execution, opts.Ticket.Path, mergedAt); err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	return result, nil
}

func (s service) agentWorkMergeTickets(execution core.ExecutionRecord, indexedAt time.Time) ([]core.IndexedTicket, error) {
	if execution.Workspace.Mode == core.WorkWorkspaceModeWorktree {
		return parseProjectTicketsFromDir(s.ctx, s.projectContext.ProjectID, agenticReviewProjectDir(s.projectContext, execution), indexedAt)
	}

	return s.listProjectTickets()
}

func parseProjectTicketsFromDir(ctx context.Context, projectID, projectDir string, indexedAt time.Time) ([]core.IndexedTicket, error) {
	files, err := core.NewTicketFileDiscoverer().DiscoverTicketFiles(ctx, core.DiscoverTicketFilesOptions{ProjectDir: projectDir})
	if err != nil {
		return nil, err
	}

	parser := core.NewTicketParser()

	tickets := make([]core.IndexedTicket, 0, len(files))
	for _, file := range files {
		ticket, err := parser.ParseTicket(ctx, core.ParseTicketOptions{
			ProjectID:  projectID,
			ProjectDir: projectDir,
			Path:       file.Path,
			IndexedAt:  indexedAt,
		})
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
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

func (s service) failAgentWorkIntegrationCycles(repository core.LoopRunRepository, state agentWorkIntegrationCycleState, eventSink core.AgentEventSink, initialEligibleCount int, cause error) (ticketLoopExecutionResult, []core.IndexedTicket, int, error) {
	failed, err := failTicketLoopExecution(s.ctx, s.env, repository, s.projectContext, state.Result, state.Record, eventSink, cause, s.assetsFS)

	return failed, state.SelectedQueue, initialEligibleCount, err
}

func (s service) agentWorkLoopSelection(opts AgentWorkSelectionOptions) (agentWorkLoopSelection, error) {
	index, err := core.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	defer func() {
		//nolint:errcheck // Query result already comes from the main workflow.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	tickets, err := core.NewTicketLoopQuery(index).EligibleTickets(s.ctx, core.TicketLoopQueryOptions{
		ProjectID: s.projectContext.ProjectID,
		Statuses:  agentWorkStatuses(opts.Statuses),
		Filters:   opts.Filters,
		Order:     agentWorkOrder(opts.Order),
	})
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	tickets, err = s.applyAgentWorkSelectionPreflight(index, tickets, config.Work)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, ticketLoopRunOptions{MaxIterations: 1})

	return agentWorkLoopSelection{
		Tickets:       tickets,
		EligibleCount: len(eligibleTickets),
		ReviewConfig:  config.Review,
		WorkConfig:    config.Work,
	}, nil
}

func (s service) applyAgentWorkSelectionPreflight(index core.TicketIndex, tickets []core.IndexedTicket, workConfig core.WorkConfig) ([]core.IndexedTicket, error) {
	allTickets, err := index.ListProjectTickets(s.ctx, s.projectContext.ProjectID)
	if err != nil {
		return nil, err
	}

	preflight, err := core.RunWorkSelectionPreflight(s.ctx, core.WorkSelectionPreflightOptions{
		ProjectDir:   s.projectContext.ProjectDir,
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

func agentWorkLoopWorkspaceConfig(projectDir string) (core.WorkspaceConfig, error) {
	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return core.WorkspaceConfig{}, err
	}

	if err := core.ValidateWorkConfig(config.Work); err != nil {
		return core.WorkspaceConfig{}, err
	}

	if err := core.ValidateReviewWorkflowConfig(config.Review); err != nil {
		return core.WorkspaceConfig{}, err
	}

	return config, nil
}

func agentWorkCycleLimitAllowsNext(opts AgentWorkSelectionOptions, processedCycles int) bool {
	return opts.MaxIterations == 0 || processedCycles < opts.MaxIterations
}

func validateAgentWorkSelectionOptions(opts AgentWorkSelectionOptions) error {
	if opts.MaxIterations < 0 {
		return fmt.Errorf("max-iterations must be greater than or equal to zero")
	}

	return nil
}

func agentWorkCycleRunsAgenticReview(policy core.IntegrationCyclePolicy, application core.ResultApplication) bool {
	return policy.Mode == core.IntegrationCycleModeAgenticReview && application.TicketStatus != core.TicketResultStatusBlocked
}

func agentWorkCycleRunsTrustedAutoMergeReview(policy core.IntegrationCyclePolicy, application core.ResultApplication) bool {
	return policy.Mode == core.IntegrationCycleModeTrustedAutoMerge && application.TicketStatus != core.TicketResultStatusBlocked
}

func agentWorkCycleRunsRepair(policy core.IntegrationCyclePolicy, ticket core.IndexedTicket) bool {
	return policy.Mode == core.IntegrationCycleModeAgenticReview && core.IsTicketReviewRepair(ticket)
}

func agentWorkCycleStopsBeforeMerge(policy core.IntegrationCyclePolicy) bool {
	return policy.RequiresHumanReview
}

func agentWorkCycleTicketReadyForMerge(policy core.IntegrationCyclePolicy, ticket core.IndexedTicket) bool {
	if !policy.RequiresMerge {
		return true
	}

	return ticket.Status == core.TicketStatusReadyForMerge
}

func agentWorkCycleMergeBlocked(result core.TicketMergeCommandResult) bool {
	return result.MergeState != core.TicketMergeStateMerged || result.Status != core.TicketStatusDone
}

func defaultAgentWorkSelectionOptions(maxIterations int) AgentWorkSelectionOptions {
	return AgentWorkSelectionOptions{
		Statuses:      []string{core.TicketStatusReadyForDev},
		Order:         core.TicketLoopOrderCreatedAt,
		MaxIterations: maxIterations,
	}
}

func agentWorkStatuses(statuses []string) []string {
	if len(statuses) == 0 {
		return []string{core.TicketStatusReadyForDev}
	}

	return statuses
}

func agentWorkOrder(order core.TicketLoopOrder) core.TicketLoopOrder {
	if order == "" {
		return core.TicketLoopOrderCreatedAt
	}

	return order
}

func agenticReviewProjectDir(projectContext commandProjectContext, execution core.ExecutionRecord) string {
	if execution.Workspace.Mode == core.WorkWorkspaceModeWorktree && execution.Workspace.ExecutionWorkspaceDir != "" {
		return execution.Workspace.ExecutionWorkspaceDir
	}

	return projectContext.ProjectDir
}

func workOperationResultFromAgentWork(result AgentWorkResult) WorkOperationResult {
	return WorkOperationResult{
		Message: result.Message,
		Tickets: result.Tickets,
	}
}

func (s service) reloadWorkOperationState() (WorkOperationResult, error) {
	tickets, err := s.listProjectTickets()
	if err != nil {
		return WorkOperationResult{}, err
	}

	return WorkOperationResult{
		Tickets: tickets,
	}, nil
}

func (s service) ensureAgentWorkRuntime() error {
	_, err := core.NewRuntimeAssets(s.assetsFS).Init(s.ctx, core.RuntimeInitOptions{
		ProjectDir: s.projectContext.ProjectDir,
	})
	if err != nil {
		return fmt.Errorf("initialize agent runtime: %w", err)
	}

	return nil
}
