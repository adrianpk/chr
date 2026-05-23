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

func (s service) StartSelectedTicket(ticketKey string) (WorkOperationResult, error) {
	ticket, err := s.ticketByKey(ticketKey)
	if err != nil {
		return WorkOperationResult{}, err
	}

	preflight, err := buildTicketStartPreflightForTicket(s.ctx, s.env, s.projectContext, ticket)
	if err != nil {
		return WorkOperationResult{}, chc.NewExecutionFailure(chc.ExecutionFailurePreflight, err)
	}

	if !chc.RuntimeValidationPassed(preflight.Runtime.Checks) {
		return WorkOperationResult{}, chc.NewExecutionFailure(
			chc.ExecutionFailureRuntime,
			errors.New("runtime validation failed"),
		)
	}

	record, err := runTicketStartExecution(s.ctx, s.env, preflight.Record, preflight.Runtime)
	if err != nil {
		return WorkOperationResult{}, err
	}

	result, err := s.reloadWorkOperationState()
	if err != nil {
		return WorkOperationResult{}, err
	}

	result.Message = "Execution " + record.ID + " for ticket " + ticket.Key + " finished with status " +
		string(record.Status) + ". Result: " + displayPath(s.projectContext.ProjectDir, record.Artifacts.ResultPath)

	if record.FailureReason != "" {
		result.Message += " Failure: " + record.FailureReason
	}

	return result, nil
}

func (s service) StartReadyTicketLoop() (WorkOperationResult, error) {
	tickets, eligibleCount, err := s.readyLoopTickets()
	if err != nil {
		return WorkOperationResult{}, err
	}

	result, err := runTicketLoopExecution(s.ctx, s.env, s.projectContext, tickets, eligibleCount)
	if err != nil {
		return WorkOperationResult{}, err
	}

	reloaded, err := s.reloadWorkOperationState()
	if err != nil {
		return WorkOperationResult{}, err
	}

	reloaded.Message = "Loop " + result.Record.ID + " finished with status " + string(result.Record.Status) +
		fmt.Sprintf(": %d processed, %d blocked, %d failed.", result.Record.ProcessedCount, result.Record.BlockedCount, result.Record.FailedCount)

	if result.Record.FailureReason != "" {
		reloaded.Message += " Failure: " + result.Record.FailureReason
	}

	return reloaded, nil
}

func (s service) RequestDaemonStop() (WorkOperationResult, error) {
	repository, err := chc.NewDaemonStateRepository(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return WorkOperationResult{}, err
	}

	defer func() {
		//nolint:errcheck // Stop result already comes from the main operation.
		_ = repository.Close()
	}()

	requestedAt := time.Now().UTC()

	err = repository.RequestDaemonStop(s.ctx, requestedAt)
	if errors.Is(err, chc.ErrDaemonStateNotFound) {
		result, reloadErr := s.reloadWorkOperationState()
		if reloadErr != nil {
			return WorkOperationResult{}, reloadErr
		}

		result.Message = "Daemon is already stopped."

		return result, nil
	}

	if err != nil {
		return WorkOperationResult{}, err
	}

	result, err := s.reloadWorkOperationState()
	if err != nil {
		return WorkOperationResult{}, err
	}

	result.Message = "Daemon stop requested at " + requestedAt.Format(time.RFC3339) + "."

	return result, nil
}

func (s service) readyLoopTickets() ([]chc.IndexedTicket, int, error) {
	index, err := chc.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		//nolint:errcheck // Query result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		return nil, 0, err
	}

	tickets, err := chc.NewTicketLoopQuery(index).EligibleTickets(s.ctx, chc.TicketLoopQueryOptions{
		ProjectID: s.projectContext.ProjectID,
		Statuses:  []string{"ready"},
		Order:     chc.TicketLoopOrderKey,
	})
	if err != nil {
		return nil, 0, err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, ticketLoopRunOptions{MaxIterations: 1})

	return tickets, len(eligibleTickets), nil
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
