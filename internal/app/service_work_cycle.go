package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

type agentWorkIntegrationCycleState struct {
	Result        ticketLoopExecutionResult
	Record        executioncore.LoopRunRecord
	SelectedQueue []ticketcore.Indexed
}

type agentWorkPostDevelopmentOptions struct {
	Ticket       ticketcore.Indexed
	Policy       core.IntegrationCyclePolicy
	Record       executioncore.LoopRunRecord
	Execution    executioncore.Record
	Application  core.ResultApplication
	Selection    agentWorkLoopSelection
	ReviewConfig core.ReviewWorkflowConfig
	LoopSink     executioncore.AgentEventSink
}

func (s service) runAgentWorkIntegrationCycles(ctx context.Context, opts AgentWorkSelectionOptions) (agentWorkIntegrationResult, error) {
	repository, err := executioncore.NewLoopRunRepository(ctx, s.env, s.assetsFS)
	if err != nil {
		return agentWorkIntegrationResult{Execution: ticketLoopExecutionResult{}, SelectedQueue: nil, EligibleCount: 0}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	selection, err := s.agentWorkLoopSelection(opts)
	if err != nil {
		return agentWorkIntegrationResult{Execution: ticketLoopExecutionResult{}, SelectedQueue: nil, EligibleCount: 0}, err
	}

	record := newRunningLoopRunRecord(s.projectContext.ProjectID, selection.EligibleCount)

	record, err = allocateLoopRunArtifact(s.projectContext.ProjectDir, record)
	if err != nil {
		return agentWorkIntegrationResult{Execution: ticketLoopExecutionResult{}, SelectedQueue: nil, EligibleCount: 0}, err
	}

	if err := repository.CreateLoopRun(ctx, record); err != nil {
		return agentWorkIntegrationResult{Execution: ticketLoopExecutionResult{}, SelectedQueue: nil, EligibleCount: 0}, err
	}

	loopSink := agentWorkLoopEventSink(s.projectContext.ProjectDir, record)
	emitLoopStartedAgentEvent(ctx, loopSink, record)

	state := agentWorkIntegrationCycleState{
		Result:        ticketLoopExecutionResult{Record: record, AuthorityExcluded: selection.ExcludedAdopted},
		Record:        record,
		SelectedQueue: make([]ticketcore.Indexed, 0),
	}
	initialEligibleCount := selection.EligibleCount
	processedCycles := 0

	for len(selection.Tickets) > 0 && agentWorkCycleLimitAllowsNext(opts, processedCycles) {
		var stop bool

		cycleResult, err := s.runAgentWorkIntegrationCycle(ctx, selection, state, loopSink)

		state, stop = cycleResult.State, cycleResult.Stop
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

	return s.finishAgentWorkIntegrationCycles(repository, state, loopSink, initialEligibleCount)
}

func (s service) runAgentWorkIntegrationCycle(ctx context.Context, selection agentWorkLoopSelection, state agentWorkIntegrationCycleState, loopSink executioncore.AgentEventSink) (agentWorkCycleResult, error) {
	ticket := selection.Tickets[0]

	cyclePlan, err := core.PlanIntegrationCycle(core.IntegrationCyclePlanOptions{
		ProjectID:    s.projectContext.ProjectID,
		Ticket:       ticket,
		ReviewConfig: selection.ReviewConfig,
	})
	if err != nil {
		return agentWorkCycleResult{State: state, Stop: false}, err
	}

	state.SelectedQueue = append(state.SelectedQueue, ticket)
	emitTicketSelectedAgentEvent(ctx, loopSink, state.Record, ticket)

	ticketOutcome, err := s.executeLoopTicket(ctx, ticket, ticketExecutionEvents{
		LoopID: state.Record.ID,
		Sink:   loopSink,
	})
	execution, application := ticketOutcome.Record, ticketOutcome.Application

	if err != nil {
		return agentWorkCycleResult{State: state, Stop: false}, err
	}

	state.Result.Executions = append(state.Result.Executions, execution)
	state.Result.Applications = append(state.Result.Applications, application)
	state.Record.ProcessedCount++

	stop, err := s.finishAgentWorkIntegrationCycle(ctx, &state, agentWorkPostDevelopmentOptions{
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
		return agentWorkCycleResult{State: state, Stop: false}, err
	}

	return agentWorkCycleResult{State: state, Stop: stop}, nil
}

func (s service) finishAgentWorkIntegrationCycle(ctx context.Context, state *agentWorkIntegrationCycleState, opts agentWorkPostDevelopmentOptions) (bool, error) {
	emitCycleRehydratedAgentEvent(ctx, opts.LoopSink, state.Record, opts.Execution, opts.Application)

	if opts.Application.TicketStatus == core.TicketResultStatusBlocked {
		state.Record.BlockedCount++

		return true, nil
	}

	reviewResult, err := s.reviewAgentWorkCycleTicket(ctx, opts)
	cycleTicket, stop := reviewResult.Ticket, reviewResult.Stop

	if err != nil {
		return false, err
	}

	finalExecution := opts.Execution
	if stop {
		repairResult, err := s.repairAgentWorkCycleTicket(ctx, state, opts, cycleTicket)
		cycleTicket, finalExecution, stop = repairResult.Ticket, repairResult.Execution, repairResult.Stop

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

	emitMergeStartedAgentEvent(ctx, opts.LoopSink, state.Record, finalExecution, cycleTicket)

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
	emitMergeFinishedAgentEvent(ctx, opts.LoopSink, state.Record, finalExecution, merge)

	if agentWorkCycleMergeBlocked(merge) {
		state.Record.BlockedCount++

		return true, nil
	}

	return false, nil
}

func (s service) failAgentWorkIntegrationCycles(repository executioncore.LoopRunRepository, state agentWorkIntegrationCycleState, eventSink executioncore.AgentEventSink, initialEligibleCount int, cause error) (agentWorkIntegrationResult, error) {
	failed, err := s.failTicketLoopExecution(s.ctx, ticketLoop{Repository: repository, Result: state.Result, Record: state.Record, Events: eventSink}, cause)

	return agentWorkIntegrationResult{Execution: failed, SelectedQueue: state.SelectedQueue, EligibleCount: initialEligibleCount}, err
}

func agentWorkCycleLimitAllowsNext(opts AgentWorkSelectionOptions, processedCycles int) bool {
	return opts.MaxIterations == 0 || processedCycles < opts.MaxIterations
}

func agentWorkCycleRunsAgenticReview(policy core.IntegrationCyclePolicy, application core.ResultApplication) bool {
	return policy.Mode == core.IntegrationCycleModeAgenticReview && application.TicketStatus != core.TicketResultStatusBlocked
}

func agentWorkCycleRunsTrustedAutoMergeReview(policy core.IntegrationCyclePolicy, application core.ResultApplication) bool {
	return policy.Mode == core.IntegrationCycleModeTrustedAutoMerge && application.TicketStatus != core.TicketResultStatusBlocked
}

func agentWorkCycleRunsRepair(policy core.IntegrationCyclePolicy, ticket ticketcore.Indexed) bool {
	return policy.Mode == core.IntegrationCycleModeAgenticReview && core.IsTicketReviewRepair(ticket)
}

func agentWorkCycleStopsBeforeMerge(policy core.IntegrationCyclePolicy) bool {
	return policy.RequiresHumanReview
}

func agentWorkCycleTicketReadyForMerge(policy core.IntegrationCyclePolicy, ticket ticketcore.Indexed) bool {
	if !policy.RequiresMerge {
		return true
	}

	return ticket.Status == ticketcore.StatusReadyForMerge
}

func agentWorkCycleMergeBlocked(result core.TicketMergeCommandResult) bool {
	return result.MergeState != core.TicketMergeStateMerged || result.Status != ticketcore.StatusDone
}

func (s service) finishAgentWorkIntegrationCycles(repository executioncore.LoopRunRepository, state agentWorkIntegrationCycleState, loopSink executioncore.AgentEventSink, initialEligibleCount int) (agentWorkIntegrationResult, error) {
	finished, err := s.finishTicketLoopExecution(s.ctx, ticketLoop{Repository: repository, Result: state.Result, Record: state.Record, Events: loopSink})

	finished.AuthorityExcluded = state.Result.AuthorityExcluded

	return agentWorkIntegrationResult{Execution: finished, SelectedQueue: state.SelectedQueue, EligibleCount: initialEligibleCount}, err
}

type agentWorkIntegrationResult struct {
	Execution     ticketLoopExecutionResult
	SelectedQueue []ticketcore.Indexed
	EligibleCount int
}

type agentWorkCycleResult struct {
	State agentWorkIntegrationCycleState
	Stop  bool
}
