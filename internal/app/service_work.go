package app

import (
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chc/internal/core"
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

	plan, err := buildTicketApplyResultPlanForRecord(s.projectContext, record)
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

	tickets, eligibleCount, err := s.agentWorkLoopTickets(defaultAgentWorkSelectionOptions(1))
	if err != nil {
		return AgentWorkResult{}, err
	}

	return s.startAgentWorkLoop(tickets, eligibleCount)
}

func (s service) StartAllReadyAgentWork() (AgentWorkResult, error) {
	return s.StartAgentWorkLoop(defaultAgentWorkSelectionOptions(0))
}

func (s service) StartAgentWorkLoop(opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	err := s.ensureAgentWorkRuntime()
	if err != nil {
		return AgentWorkResult{}, err
	}

	tickets, eligibleCount, err := s.agentWorkLoopTickets(opts)
	if err != nil {
		return AgentWorkResult{}, err
	}

	return s.startAgentWorkLoop(tickets, eligibleCount)
}

func (s service) PreviewAllReadyAgentWork() (AgentWorkResult, error) {
	return s.PreviewAgentWorkLoop(defaultAgentWorkSelectionOptions(0))
}

func (s service) PreviewAgentWorkLoop(opts AgentWorkSelectionOptions) (AgentWorkResult, error) {
	tickets, eligibleCount, err := s.agentWorkLoopTickets(opts)
	if err != nil {
		return AgentWorkResult{}, err
	}

	reloaded, err := s.reloadWorkOperationState()
	if err != nil {
		return AgentWorkResult{}, err
	}

	return AgentWorkResult{
		Message:       fmt.Sprintf("Ready agent-work queue contains %d eligible tickets.", eligibleCount),
		Tickets:       reloaded.Tickets,
		SelectedQueue: tickets,
		EligibleCount: eligibleCount,
	}, nil
}

func (s service) startAgentWorkLoop(tickets []core.IndexedTicket, eligibleCount int) (AgentWorkResult, error) {
	lock, err := acquireAgentOperationLock(s.ctx, s.projectContext, core.AgentOperationDevelopment)
	if err != nil {
		return AgentWorkResult{}, err
	}

	defer releaseAgentOperationLock(lock)()

	result, err := runTicketLoopExecution(s.ctx, s.env, s.projectContext, tickets, eligibleCount, s.assetsFS)
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
		SelectedQueue: tickets,
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

func (s service) agentWorkLoopTickets(opts AgentWorkSelectionOptions) ([]core.IndexedTicket, int, error) {
	index, err := core.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		//nolint:errcheck // Query result already comes from the main workflow.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		return nil, 0, err
	}

	tickets, err := core.NewTicketLoopQuery(index).EligibleTickets(s.ctx, core.TicketLoopQueryOptions{
		ProjectID: s.projectContext.ProjectID,
		Statuses:  agentWorkStatuses(opts.Statuses),
		Filters:   opts.Filters,
		Order:     agentWorkOrder(opts.Order),
	})
	if err != nil {
		return nil, 0, err
	}

	tickets, err = s.applyAgentWorkSelectionPreflight(index, tickets)
	if err != nil {
		return nil, 0, err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, ticketLoopRunOptions{MaxIterations: opts.MaxIterations})

	return tickets, len(eligibleTickets), nil
}

func (s service) applyAgentWorkSelectionPreflight(index core.TicketIndex, tickets []core.IndexedTicket) ([]core.IndexedTicket, error) {
	allTickets, err := index.ListProjectTickets(s.ctx, s.projectContext.ProjectID)
	if err != nil {
		return nil, err
	}

	branchPolicy, err := workSelectionBranchPolicy(s.projectContext.ProjectDir)
	if err != nil {
		return nil, err
	}

	preflight, err := core.RunWorkSelectionPreflight(s.ctx, core.WorkSelectionPreflightOptions{
		ProjectDir:   s.projectContext.ProjectDir,
		Tickets:      allTickets,
		BranchPolicy: branchPolicy,
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
