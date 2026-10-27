package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

type ticketLoopExecutionResult struct {
	AuthorityExcluded int
	Record            core.LoopRunRecord
	Executions        []core.ExecutionRecord
	Applications      []core.ResultApplication
	Merges            []core.TicketMergeCommandResult
}

type ticketLoopRunOptions struct {
	MaxIterations int
}

func selectTicketLoopTickets(ctx context.Context, index ticketcore.TicketIndex, projectContext commandProjectContext, queryOptions core.TicketLoopQueryOptions, runOptions ticketLoopRunOptions) (ticketLoopSelection, error) {
	tickets, err := core.NewTicketLoopQuery(index).EligibleTickets(ctx, queryOptions)
	if err != nil {
		return ticketLoopSelection{Tickets: nil, EligibleCount: 0}, err
	}

	tickets, err = applyTicketLoopPreflight(ctx, index, projectContext, tickets)
	if err != nil {
		return ticketLoopSelection{Tickets: nil, EligibleCount: 0}, err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, runOptions)

	return ticketLoopSelection{Tickets: tickets, EligibleCount: len(eligibleTickets)}, nil
}

func applyTicketLoopPreflight(ctx context.Context, index ticketcore.TicketIndex, projectContext commandProjectContext, tickets []ticketcore.Indexed) ([]ticketcore.Indexed, error) {
	allTickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return nil, err
	}

	workConfig, err := workSelectionWorkConfig(projectContext.ProjectDir)
	if err != nil {
		return nil, err
	}

	preflight, err := core.RunWorkSelectionPreflight(ctx, core.WorkSelectionPreflightOptions{
		ProjectDir:   projectContext.ProjectDir,
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

func ticketsForTicketLoopRun(tickets []ticketcore.Indexed, opts ticketLoopRunOptions) []ticketcore.Indexed {
	if opts.MaxIterations == 0 || len(tickets) <= opts.MaxIterations {
		return tickets
	}

	return tickets[:opts.MaxIterations]
}

func deduplicateTicketLoopTickets(tickets []ticketcore.Indexed) []ticketcore.Indexed {
	selected := make([]ticketcore.Indexed, 0, len(tickets))
	seen := map[string]struct{}{}

	for _, ticket := range tickets {
		_, ok := seen[ticket.Key]
		if ok {
			continue
		}

		seen[ticket.Key] = struct{}{}
		selected = append(selected, ticket)
	}

	return selected
}

func enforceWorkSelectionPreflight(ctx context.Context, projectContext commandProjectContext, tickets []ticketcore.Indexed, selectedTicketKey string) error {
	workConfig, err := workSelectionWorkConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	preflight, err := core.RunWorkSelectionPreflight(ctx, core.WorkSelectionPreflightOptions{
		ProjectDir:        projectContext.ProjectDir,
		Tickets:           tickets,
		SelectedTicketKey: selectedTicketKey,
		BranchPolicy:      workConfig.BranchPolicy,
		Workspace:         workConfig.Workspace,
	})
	if err != nil {
		return err
	}

	if !preflight.Allowed {
		return formatWorkSelectionPreflightError(preflight)
	}

	return nil
}

func formatWorkSelectionPreflightError(preflight core.WorkSelectionPreflightResult) error {
	if len(preflight.BlockReasons) == 0 {
		return fmt.Errorf("work selection preflight blocked")
	}

	reason := preflight.BlockReasons[0]
	if reason.TicketKey != "" {
		return fmt.Errorf("work selection preflight blocked: %s ticket=%s detail=%s", reason.Code, reason.TicketKey, reason.Detail)
	}

	return fmt.Errorf("work selection preflight blocked: %s detail=%s", reason.Code, reason.Detail)
}

func hasReviewRepairTicket(tickets []ticketcore.Indexed) bool {
	for _, ticket := range tickets {
		readyForDev := ticket.Status == ticketcore.StatusReadyForDev
		changesRequested := ticket.ReviewState == "changes_requested" || ticket.ReviewResult == "needs_changes"

		if readyForDev && changesRequested {
			return true
		}
	}

	return false
}

func (s service) runTicketLoopExecution(ctx context.Context, selection ticketLoopSelection) (ticketLoopExecutionResult, error) {
	repository, err := core.NewLoopRunRepository(ctx, s.env, s.assetsFS)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}
	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	loop := ticketLoop{Repository: repository, Record: newRunningLoopRunRecord(s.projectContext.ProjectID, selection.EligibleCount)}

	loop.Record, err = allocateLoopRunArtifact(s.projectContext.ProjectDir, loop.Record)
	if err != nil {
		return ticketLoopExecutionResult{}, err
	}

	if err = repository.CreateLoopRun(ctx, loop.Record); err != nil {
		return ticketLoopExecutionResult{}, err
	}

	loop.Events = agentWorkLoopEventSink(s.projectContext.ProjectDir, loop.Record)
	emitLoopStartedAgentEvent(ctx, loop.Events, loop.Record)

	loop.Result.Record = loop.Record
	for _, ticket := range selection.Tickets {
		emitTicketSelectedAgentEvent(ctx, loop.Events, loop.Record, ticket)

		ticketOutcome, err := s.executeLoopTicket(ctx, ticket, ticketExecutionEvents{LoopID: loop.Record.ID, Sink: loop.Events})
		execution, application := ticketOutcome.Record, ticketOutcome.Application

		if err != nil {
			return s.failTicketLoopExecution(ctx, loop, err)
		}

		loop.Result.Executions = append(loop.Result.Executions, execution)
		loop.Result.Applications = append(loop.Result.Applications, application)

		loop.Record.ProcessedCount++
		if application.TicketStatus == core.TicketResultStatusBlocked {
			loop.Record.BlockedCount++

			break
		}
	}

	return s.finishTicketLoopExecution(ctx, loop)
}

func (s service) failTicketLoopExecution(ctx context.Context, loop ticketLoop, cause error) (ticketLoopExecutionResult, error) {
	loop.Record.FailedCount++

	loop.Record = failLoopRunRecord(loop.Record, cause)
	if err := loop.Repository.UpdateLoopRun(ctx, loop.Record); err != nil {
		return ticketLoopExecutionResult{}, err
	}

	loop.Result.Record = loop.Record
	emitLoopFailedAgentEvents(ctx, loop.Events, loop.Record, cause)

	if loop.Record.ArtifactID != "" {
		if err := opscore.RefreshArtifactMetadata(loop.Record.ArtifactDir, string(loop.Record.Status), time.Now().UTC()); err != nil {
			return ticketLoopExecutionResult{}, err
		}
	}

	if err := emitLoopRunNotification(ctx, s.env, s.projectContext, loop.Record, s.assetsFS); err != nil {
		return ticketLoopExecutionResult{}, err
	}

	return loop.Result, cause
}

func (s service) finishTicketLoopExecution(ctx context.Context, loop ticketLoop) (ticketLoopExecutionResult, error) {
	loop.Record = finishLoopRunRecord(loop.Record)
	if err := loop.Repository.UpdateLoopRun(ctx, loop.Record); err != nil {
		return ticketLoopExecutionResult{}, err
	}

	loop.Result.Record = loop.Record
	emitLoopFinishedAgentEvent(ctx, loop.Events, loop.Record)

	if loop.Record.ArtifactID != "" {
		if err := opscore.RefreshArtifactMetadata(loop.Record.ArtifactDir, string(loop.Record.Status), time.Now().UTC()); err != nil {
			return ticketLoopExecutionResult{}, err
		}
	}

	if err := emitLoopRunNotification(ctx, s.env, s.projectContext, loop.Record, s.assetsFS); err != nil {
		return ticketLoopExecutionResult{}, err
	}

	return loop.Result, nil
}

func (s service) executeLoopTicket(ctx context.Context, ticket ticketcore.Indexed, events ticketExecutionEvents) (ticketExecutionOutcome, error) {
	preflight, err := s.prepareTicketStart(ctx, ticket, ticketStartPreflightOptions{PrepareWorktree: true})
	if err != nil {
		return ticketExecutionOutcome{Record: core.ExecutionRecord{}, Application: core.ResultApplication{}}, err
	}

	events.Sink = core.NewAgentEventFanoutSink(events.Sink, agentWorkExecutionEventSink(preflight.Record))

	record, err := s.runTicketStartExecution(ctx, ticketExecutionInput{Record: preflight.Record, Runtime: preflight.Runtime, Events: events})
	if err != nil {
		return ticketExecutionOutcome{Record: record, Application: core.ResultApplication{}}, err
	}

	plan, err := buildTicketApplyResultPlanForRecord(ctx, s.projectContext, record)
	if err != nil {
		return ticketExecutionOutcome{Record: record, Application: core.ResultApplication{}}, err
	}

	err = s.applyTicketResult(ctx, plan, events)
	if err != nil {
		return ticketExecutionOutcome{Record: record, Application: core.ResultApplication{}}, err
	}

	return ticketExecutionOutcome{Record: record, Application: plan.Application}, nil
}

func newRunningLoopRunRecord(projectID string, eligibleCount int) core.LoopRunRecord {
	now := time.Now().UTC()

	return core.LoopRunRecord{
		ID:            loopRunID(projectID, now),
		ProjectID:     projectID,
		Status:        core.LoopRunStatusRunning,
		EligibleCount: eligibleCount,
		CreatedAt:     now,
		StartedAt:     &now,
		UpdatedAt:     now,
	}
}

func allocateLoopRunArtifact(projectDir string, record core.LoopRunRecord) (core.LoopRunRecord, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return core.LoopRunRecord{}, err
	}

	artifact, err := opscore.CreateArtifact(opscore.ArtifactOptions{
		Layout:      layout,
		Type:        "run",
		Title:       "Agent work loop",
		Producer:    "chr",
		ExternalIDs: []string{record.ID},
	})
	if err != nil {
		return core.LoopRunRecord{}, err
	}

	record.ArtifactID = artifact.ID
	record.ArtifactDir = artifact.Dir

	return record, nil
}

func finishLoopRunRecord(record core.LoopRunRecord) core.LoopRunRecord {
	finishedAt := time.Now().UTC()

	record.Status = core.LoopRunStatusSucceeded
	if record.BlockedCount > 0 {
		record.Status = core.LoopRunStatusBlocked
	}

	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt

	return record
}

func failLoopRunRecord(record core.LoopRunRecord, cause error) core.LoopRunRecord {
	finishedAt := time.Now().UTC()
	record.Status = core.LoopRunStatusFailed
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	record.FailureReason = cause.Error()

	return record
}

func loopRunID(projectID string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(projectID + "\x00loop\x00" + createdAt.Format(time.RFC3339Nano)))

	return "loop-" + hex.EncodeToString(sum[:8])
}

type ticketLoopRequest struct {
	Query  core.TicketLoopQueryOptions
	Run    ticketLoopRunOptions
	DryRun bool
}

type ticketLoopOperationResult struct {
	Selection ticketLoopSelection
	Execution ticketLoopExecutionResult
}

func (s service) runLegacyTicketLoop(ctx context.Context, request ticketLoopRequest) (ticketLoopOperationResult, error) {
	index, err := core.NewWorkflowTicketIndex(ctx, s.env, s.assetsFS)
	if err != nil {
		return ticketLoopOperationResult{}, err
	}
	defer func() {
		//nolint:errcheck // The operation result already owns its failure; the index is a rebuildable projection.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, s.projectContext); err != nil {
		return ticketLoopOperationResult{}, err
	}

	if request.Run.MaxIterations < 0 {
		return ticketLoopOperationResult{}, fmt.Errorf("max-iterations must be greater than or equal to zero")
	}

	loopSelection, err := selectTicketLoopTickets(ctx, index, s.projectContext, request.Query, request.Run)
	tickets, count := loopSelection.Tickets, loopSelection.EligibleCount

	if err != nil {
		return ticketLoopOperationResult{}, err
	}

	outcome := ticketLoopOperationResult{Selection: ticketLoopSelection{Tickets: tickets, EligibleCount: count}}
	if request.DryRun {
		return outcome, nil
	}

	lock, err := acquireAgentOperationLock(ctx, s.projectContext, core.AgentOperationDevelopment)
	if err != nil {
		return ticketLoopOperationResult{}, err
	}
	defer releaseAgentOperationLock(lock)()

	outcome.Execution, err = s.runTicketLoopExecution(ctx, outcome.Selection)

	return outcome, err
}

type ticketExecutionOutcome struct {
	Record      core.ExecutionRecord
	Application core.ResultApplication
}
