package app

import (
	"forge.adrianpk.com/chr/internal/core"
)

type agentWorkIntegrationCycleState struct {
	Result        ticketLoopExecutionResult
	Record        core.LoopRunRecord
	SelectedQueue []core.IndexedTicket
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

func (s service) runAgentWorkIntegrationCycles(opts AgentWorkSelectionOptions) (agentWorkIntegrationResult, error) {
	repository, err := core.NewLoopRunRepository(s.ctx, s.env, s.assetsFS)
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

	if err := repository.CreateLoopRun(s.ctx, record); err != nil {
		return agentWorkIntegrationResult{Execution: ticketLoopExecutionResult{}, SelectedQueue: nil, EligibleCount: 0}, err
	}

	loopSink := agentWorkLoopEventSink(s.projectContext.ProjectDir, record)
	emitLoopStartedAgentEvent(s.ctx, loopSink, record)

	state := agentWorkIntegrationCycleState{
		Result:        ticketLoopExecutionResult{Record: record, AuthorityExcluded: selection.ExcludedAdopted},
		Record:        record,
		SelectedQueue: make([]core.IndexedTicket, 0),
	}
	initialEligibleCount := selection.EligibleCount
	processedCycles := 0

	for len(selection.Tickets) > 0 && agentWorkCycleLimitAllowsNext(opts, processedCycles) {
		var stop bool

		cycleResult, err := s.runAgentWorkIntegrationCycle(selection, state, loopSink)

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

func (s service) runAgentWorkIntegrationCycle(selection agentWorkLoopSelection, state agentWorkIntegrationCycleState, loopSink core.AgentEventSink) (agentWorkCycleResult, error) {
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
	emitTicketSelectedAgentEvent(s.ctx, loopSink, state.Record, ticket)

	ticketOutcome, err := s.executeLoopTicket(s.ctx, ticket, ticketExecutionEvents{
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
		return agentWorkCycleResult{State: state, Stop: false}, err
	}

	return agentWorkCycleResult{State: state, Stop: stop}, nil
}

func (s service) finishAgentWorkIntegrationCycle(state *agentWorkIntegrationCycleState, opts agentWorkPostDevelopmentOptions) (bool, error) {
	emitCycleRehydratedAgentEvent(s.ctx, opts.LoopSink, state.Record, opts.Execution, opts.Application)

	if opts.Application.TicketStatus == core.TicketResultStatusBlocked {
		state.Record.BlockedCount++

		return true, nil
	}

	reviewResult, err := s.reviewAgentWorkCycleTicket(opts)
	cycleTicket, stop := reviewResult.Ticket, reviewResult.Stop

	if err != nil {
		return false, err
	}

	finalExecution := opts.Execution
	if stop {
		repairResult, err := s.repairAgentWorkCycleTicket(state, opts, cycleTicket)
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

func (s service) failAgentWorkIntegrationCycles(repository core.LoopRunRepository, state agentWorkIntegrationCycleState, eventSink core.AgentEventSink, initialEligibleCount int, cause error) (agentWorkIntegrationResult, error) {
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

func (s service) finishAgentWorkIntegrationCycles(repository core.LoopRunRepository, state agentWorkIntegrationCycleState, loopSink core.AgentEventSink, initialEligibleCount int) (agentWorkIntegrationResult, error) {
	finished, err := s.finishTicketLoopExecution(s.ctx, ticketLoop{Repository: repository, Result: state.Result, Record: state.Record, Events: loopSink})

	finished.AuthorityExcluded = state.Result.AuthorityExcluded

	return agentWorkIntegrationResult{Execution: finished, SelectedQueue: state.SelectedQueue, EligibleCount: initialEligibleCount}, err
}

type agentWorkIntegrationResult struct {
	Execution     ticketLoopExecutionResult
	SelectedQueue []core.IndexedTicket
	EligibleCount int
}

type agentWorkCycleResult struct {
	State agentWorkIntegrationCycleState
	Stop  bool
}
