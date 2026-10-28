package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// WorkOperationResult is the TUI-facing result for agent work operations.
type WorkOperationResult struct {
	Message string
	Tickets []ticketcore.Indexed
}

// AgentWorkResult summarizes runnable agent work for the dashboard.
type AgentWorkResult struct {
	Scope         string
	Authority     workspaceconfig.WorkStateAuthority
	Excluded      map[string]int
	Workflow      *workflowcore.Invocation
	Message       string
	Tickets       []ticketcore.Indexed
	Execution     *executioncore.Record
	LoopRun       *executioncore.LoopRunRecord
	Executions    []executioncore.Record
	Merges        []core.TicketMergeCommandResult
	SelectedQueue []ticketcore.Indexed
	EligibleCount int
}

// AgentWorkSelectionOptions filters agent work selection.
type AgentWorkSelectionOptions struct {
	Statuses      []string
	Filters       workflowstate.TicketFilters
	Order         workflowstate.TicketOrder
	MaxIterations int
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
	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return AgentWorkResult{}, err
	}

	if workspaceconfig.WorkAuthority(config.Work) == workspaceconfig.WorkStateAuthorityTickets {
		return s.startNativeAgentWork(config, workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: ticketKey}}, 0, "")
	}

	result, err := s.startLegacyAgentWorkTicket(ticketKey)

	return legacyAgentWorkResult(result), err
}

func (s service) startLegacyAgentWorkTicket(ticketKey string) (AgentWorkResult, error) {
	preparedTicket, err := s.prepareAgentWorkTicket(ticketKey)
	ticket, preflight, release := preparedTicket.Ticket, preparedTicket.Preflight, preparedTicket.Release

	if err != nil {
		return AgentWorkResult{}, err
	}

	defer release()

	if !harnessruntime.ValidationPassed(preflight.Runtime.Checks) {
		return AgentWorkResult{}, executioncore.NewFailure(
			executioncore.FailureRuntime,
			errors.New("runtime validation failed"),
		)
	}

	events := ticketExecutionEvents{
		Sink: agentWorkExecutionEventSink(preflight.Record),
	}

	record, err := s.runTicketStartExecution(s.ctx, ticketExecutionInput{Record: preflight.Record, Runtime: preflight.Runtime, Events: events})
	if err != nil {
		return AgentWorkResult{}, err
	}

	plan, err := buildTicketApplyResultPlanForRecord(s.ctx, s.projectContext, record)
	if err != nil {
		return AgentWorkResult{}, err
	}

	err = s.applyTicketResult(s.ctx, plan, events)
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

func (s service) prepareAgentWorkTicket(ticketKey string) (preparedAgentWorkTicket, error) {
	ticket, err := s.legacyAgentWorkTicket(ticketKey)
	if err != nil {
		return preparedAgentWorkTicket{Ticket: ticketcore.Indexed{}, Preflight: ticketStartPreflight{}, Release: noopRelease}, err
	}

	lock, err := acquireAgentOperationLock(s.ctx, s.projectContext, workflowcore.AgentOperationDevelopment)
	if err != nil {
		return preparedAgentWorkTicket{Ticket: ticketcore.Indexed{}, Preflight: ticketStartPreflight{}, Release: noopRelease}, err
	}

	preflight, err := s.prepareTicketStart(s.ctx, ticket, ticketStartPreflightOptions{PrepareWorktree: true})
	if err != nil {
		releaseErr := lock.Release()
		if releaseErr != nil {
			return preparedAgentWorkTicket{Ticket: ticketcore.Indexed{}, Preflight: ticketStartPreflight{}, Release: noopRelease}, releaseErr
		}

		return preparedAgentWorkTicket{Ticket: ticketcore.Indexed{}, Preflight: ticketStartPreflight{}, Release: noopRelease}, executioncore.NewFailure(executioncore.FailurePreflight, err)
	}

	return preparedAgentWorkTicket{Ticket: ticket, Preflight: preflight, Release: releaseAgentOperationLock(lock)}, nil
}

func (s service) StartReadyAgentWorkLoop() (AgentWorkResult, error) {
	return s.StartAgentWorkLoop(s.ctx, defaultAgentWorkSelectionOptions(1))
}

func (s service) StartAllReadyAgentWork() (AgentWorkResult, error) {
	return s.StartAgentWorkLoop(s.ctx, defaultAgentWorkSelectionOptions(0))
}

func (s service) StartAgentWorkLoop(ctx context.Context, opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	if err := validateAgentWorkSelectionOptions(opts); err != nil {
		return AgentWorkResult{}, err
	}

	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return AgentWorkResult{}, err
	}

	if workspaceconfig.WorkAuthority(config.Work) == workspaceconfig.WorkStateAuthorityTickets {
		return s.startNativeAgentWork(config, nativeAgentWorkSelection(opts), opts.MaxIterations, "")
	}

	if err := s.ensureAgentWorkRuntime(); err != nil {
		return AgentWorkResult{}, err
	}

	result, err := s.startAgentWorkLoop(ctx, opts)

	return legacyAgentWorkResult(result), err
}

func (s service) PreviewAllReadyAgentWork() (AgentWorkResult, error) {
	return s.PreviewAgentWorkLoop(defaultAgentWorkSelectionOptions(0))
}

func (s service) PreviewAgentWorkLoop(opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return AgentWorkResult{}, err
	}

	if workspaceconfig.WorkAuthority(config.Work) == workspaceconfig.WorkStateAuthorityTickets {
		return s.previewNativeAgentWork(config, opts)
	}

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
		Message:       fmt.Sprintf("Authority tracker_compat; scope unadopted legacy tickets; %d adopted tickets excluded. Ready queue contains %d eligible tickets.", selection.ExcludedAdopted, selection.EligibleCount),
		Authority:     workspaceconfig.WorkStateAuthorityTrackerCompat,
		Excluded:      map[string]int{"ticket_native": selection.ExcludedAdopted},
		Tickets:       reloaded.Tickets,
		SelectedQueue: selection.Tickets,
		EligibleCount: selection.EligibleCount,
	}, nil
}

func (s service) startAgentWorkLoop(ctx context.Context, opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	lock, err := acquireAgentOperationLock(ctx, s.projectContext, workflowcore.AgentOperationDevelopment)
	if err != nil {
		return AgentWorkResult{}, err
	}

	defer releaseAgentOperationLock(lock)()

	integrationResult, err := s.runAgentWorkIntegrationCycles(ctx, opts)
	result, selectedQueue, eligibleCount := integrationResult.Execution, integrationResult.SelectedQueue, integrationResult.EligibleCount

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
		Excluded:      map[string]int{"ticket_native": result.AuthorityExcluded},
		Executions:    result.Executions,
		Merges:        result.Merges,
		SelectedQueue: selectedQueue,
		EligibleCount: eligibleCount,
	}, nil
}

func (s service) RequestStop() (WorkOperationResult, error) {
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

func (s service) legacyAgentWorkTicket(ticketKey string) (ticketcore.Indexed, error) {
	if err := s.checkLegacyAgentWorkKey(ticketKey); err != nil {
		return ticketcore.Indexed{}, err
	}

	err := s.ensureAgentWorkRuntime()
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	ticket, err := s.ticketByKey(ticketKey)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	if err := workflowcore.CheckLegacyWorkTicket(ticket.Path); err != nil {
		return ticketcore.Indexed{}, err
	}

	tickets, err := s.listProjectTickets()
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	legacySelection, legacySelectionErr := workflowcore.FilterLegacyWorkTickets(s.ctx, tickets)

	tickets, _, err = legacySelection.Tickets, legacySelection.Excluded, legacySelectionErr
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	err = enforceWorkSelectionPreflight(s.ctx, s.projectContext, tickets, ticket.Key)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	return ticket, nil
}

type preparedAgentWorkTicket struct {
	Ticket    ticketcore.Indexed
	Preflight ticketStartPreflight
	Release   func()
}
