package app

import (
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
	Scope         string
	Authority     core.WorkStateAuthority
	Excluded      map[string]int
	Workflow      *core.WorkflowInvocation
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

	if core.WorkAuthority(config.Work) == core.WorkStateAuthorityTickets {
		return s.startNativeAgentWork(config, core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: ticketKey}}, 0, "")
	}

	result, err := s.startLegacyAgentWorkTicket(ticketKey)

	return legacyAgentWorkResult(result), err
}

func (s service) startLegacyAgentWorkTicket(ticketKey string) (AgentWorkResult, error) {
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
	ticket, err := s.legacyAgentWorkTicket(ticketKey)
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
	return s.StartAgentWorkLoop(defaultAgentWorkSelectionOptions(1))
}

func (s service) StartAllReadyAgentWork() (AgentWorkResult, error) {
	return s.StartAgentWorkLoop(defaultAgentWorkSelectionOptions(0))
}

func (s service) StartAgentWorkLoop(opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	if err := validateAgentWorkSelectionOptions(opts); err != nil {
		return AgentWorkResult{}, err
	}

	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return AgentWorkResult{}, err
	}

	if core.WorkAuthority(config.Work) == core.WorkStateAuthorityTickets {
		return s.startNativeAgentWork(config, nativeAgentWorkSelection(opts), opts.MaxIterations, "")
	}

	if err := s.ensureAgentWorkRuntime(); err != nil {
		return AgentWorkResult{}, err
	}

	result, err := s.startAgentWorkLoop(opts)

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

	if core.WorkAuthority(config.Work) == core.WorkStateAuthorityTickets {
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
		Authority:     core.WorkStateAuthorityTrackerCompat,
		Excluded:      map[string]int{"ticket_native": selection.ExcludedAdopted},
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
		Excluded:      map[string]int{"ticket_native": result.AuthorityExcluded},
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

func (s service) legacyAgentWorkTicket(ticketKey string) (core.IndexedTicket, error) {
	if err := s.checkLegacyAgentWorkKey(ticketKey); err != nil {
		return core.IndexedTicket{}, err
	}

	err := s.ensureAgentWorkRuntime()
	if err != nil {
		return core.IndexedTicket{}, err
	}

	ticket, err := s.ticketByKey(ticketKey)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	if err := core.CheckLegacyWorkTicket(ticket.Path); err != nil {
		return core.IndexedTicket{}, err
	}

	tickets, err := s.listProjectTickets()
	if err != nil {
		return core.IndexedTicket{}, err
	}

	tickets, _, err = core.FilterLegacyWorkTickets(s.ctx, tickets)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	err = enforceWorkSelectionPreflight(s.ctx, s.projectContext, tickets, ticket.Key)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	return ticket, nil
}
