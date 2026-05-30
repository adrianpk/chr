package chcapp

import (
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

type WorkOperationResult struct {
	Message string
	Tickets []chc.IndexedTicket
}

type AgentWorkResult struct {
	Message       string
	Tickets       []chc.IndexedTicket
	Execution     *chc.ExecutionRecord
	LoopRun       *chc.LoopRunRecord
	Executions    []chc.ExecutionRecord
	SelectedQueue []chc.IndexedTicket
	EligibleCount int
}

type AgentWorkSelectionOptions struct {
	Statuses      []string
	Filters       chc.TicketLoopFilters
	Order         chc.TicketLoopOrder
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
	err := s.ensureAgentWorkRuntime()
	if err != nil {
		return AgentWorkResult{}, err
	}

	ticket, err := s.ticketByKey(ticketKey)
	if err != nil {
		return AgentWorkResult{}, err
	}

	preflight, err := buildTicketStartPreflightForTicket(s.ctx, s.env, s.projectContext, ticket, s.assetsFS)
	if err != nil {
		return AgentWorkResult{}, chc.NewExecutionFailure(chc.ExecutionFailurePreflight, err)
	}

	if !chc.RuntimeValidationPassed(preflight.Runtime.Checks) {
		return AgentWorkResult{}, chc.NewExecutionFailure(
			chc.ExecutionFailureRuntime,
			errors.New("runtime validation failed"),
		)
	}

	record, err := runTicketStartExecution(s.ctx, s.env, preflight.Record, preflight.Runtime, s.assetsFS)
	if err != nil {
		return AgentWorkResult{}, err
	}

	result, err := s.reloadWorkOperationState()
	if err != nil {
		return AgentWorkResult{}, err
	}

	result.Message = "Execution " + record.ID + " for ticket " + ticket.Key + " finished with status " +
		string(record.Status) + ". Result: " + displayPath(s.projectContext.ProjectDir, record.Artifacts.ResultPath)

	if record.FailureReason != "" {
		result.Message += " Failure: " + record.FailureReason
	}

	return AgentWorkResult{
		Message:   result.Message,
		Tickets:   result.Tickets,
		Execution: &record,
	}, nil
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

func (s service) startAgentWorkLoop(tickets []chc.IndexedTicket, eligibleCount int) (AgentWorkResult, error) {
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

func (s service) agentWorkLoopTickets(opts AgentWorkSelectionOptions) ([]chc.IndexedTicket, int, error) {
	index, err := chc.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return nil, 0, err
	}

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		closeErr := index.Close()
		if closeErr != nil {
			return nil, 0, closeErr
		}

		return nil, 0, err
	}

	tickets, err := chc.NewTicketLoopQuery(index).EligibleTickets(s.ctx, chc.TicketLoopQueryOptions{
		ProjectID: s.projectContext.ProjectID,
		Statuses:  agentWorkStatuses(opts.Statuses),
		Filters:   opts.Filters,
		Order:     agentWorkOrder(opts.Order),
	})
	if err != nil {
		closeErr := index.Close()
		if closeErr != nil {
			return nil, 0, closeErr
		}

		return nil, 0, err
	}

	err = index.Close()
	if err != nil {
		return nil, 0, err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, ticketLoopRunOptions{MaxIterations: opts.MaxIterations})

	return tickets, len(eligibleTickets), nil
}

func defaultAgentWorkSelectionOptions(maxIterations int) AgentWorkSelectionOptions {
	return AgentWorkSelectionOptions{
		Statuses:      []string{chc.TicketStatusReadyForDev},
		Order:         chc.TicketLoopOrderCreatedAt,
		MaxIterations: maxIterations,
	}
}

func agentWorkStatuses(statuses []string) []string {
	if len(statuses) == 0 {
		return []string{chc.TicketStatusReadyForDev}
	}

	return statuses
}

func agentWorkOrder(order chc.TicketLoopOrder) chc.TicketLoopOrder {
	if order == "" {
		return chc.TicketLoopOrderCreatedAt
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
	_, err := chc.NewRuntimeAssets(s.assetsFS).Init(s.ctx, chc.RuntimeInitOptions{
		ProjectDir: s.projectContext.ProjectDir,
	})
	if err != nil {
		return fmt.Errorf("initialize agent runtime: %w", err)
	}

	return nil
}
