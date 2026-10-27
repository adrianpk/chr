package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func (s service) observeLegacyWork(ctx context.Context, config core.WorkspaceConfig, result *WorkObservation) error {
	// The same canonical/legacy document owner used by selection reads avoids a
	// write-through index refresh during observation.
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: s.workflowNamespace})
	if err != nil {
		return err
	}

	queue, err := core.ReadLegacyWorkQueue(ctx, layout, s.projectContext.ProjectID)
	if err != nil {
		return err
	}

	if err = s.observeLegacyQueue(ctx, config, queue, result); err != nil {
		return err
	}

	retained, err := executioncore.ReadLegacyWork(ctx, s.env, s.projectContext.ProjectID)
	if err != nil {
		return err
	}

	tickets := map[string]ticketcore.Indexed{}
	for _, ticket := range queue.Tickets {
		tickets[ticket.Key] = ticket
	}

	streams := 0

	for _, execution := range retained.Executions {
		current := WorkContext{ID: "execution:" + execution.ID, Kind: "execution", TicketKey: execution.TicketKey, Title: tickets[execution.TicketKey].Title, Scope: tickets[execution.TicketKey].UseCase, Status: string(execution.Status), UpdatedAt: execution.UpdatedAt, Message: execution.FailureReason}
		if err = s.observeLegacyContext(ctx, result, &current, executioncore.RecordEventStreamPath(execution), execution.ID, "", &streams); err != nil {
			return err
		}
	}

	for _, loop := range retained.Loops {
		current := WorkContext{ID: "loop:" + loop.ID, Kind: "loop", Status: string(loop.Status), UpdatedAt: loop.UpdatedAt, Message: loop.FailureReason, Eligible: loop.EligibleCount, Processed: loop.ProcessedCount, Blocked: loop.BlockedCount, Failed: loop.FailedCount}
		if err = s.observeLegacyContext(ctx, result, &current, executioncore.LoopRunEventStreamPath(loop), "", loop.ID, &streams); err != nil {
			return err
		}
	}

	return nil
}

func (s service) observeLegacyQueue(ctx context.Context, config core.WorkspaceConfig, queue core.LegacyWorkQueue, result *WorkObservation) error {
	preflight, err := core.RunWorkSelectionPreflight(ctx, core.WorkSelectionPreflightOptions{ProjectDir: s.projectContext.ProjectDir, Tickets: queue.Eligible, BranchPolicy: config.Work.BranchPolicy, Workspace: config.Work.Workspace, RequireBranchPolicy: true})
	if err != nil {
		return err
	}

	result.Excluded = map[string]int{"adopted_workflow": queue.Excluded, "selection_precedence": len(queue.Eligible) - len(preflight.SelectedTickets)}
	for _, ticket := range preflight.SelectedTickets {
		item := WorkQueueItem{Key: ticket.Key, Title: ticket.Title, Scope: ticket.UseCase, Readiness: "ready"}
		if !preflight.Allowed {
			item.Readiness = "blocked"
			item.Reason = preflight.BlockReasons[0].Detail
		}

		bindWorkQueueReadiness(result, &item)

		result.Queue = append(result.Queue, item)
	}

	return nil
}

func (s service) observeLegacyContext(ctx context.Context, result *WorkObservation, current *WorkContext, path, executionID, loopID string, streams *int) error {
	if err := s.observeLegacyActivity(ctx, result, current, path, executionID, loopID, streams); err != nil {
		return err
	}

	if current.Phase == "" {
		current.Phase = current.Status
	}

	if current.Status == "running" || current.Status == "pending" {
		result.Current = append(result.Current, *current)
	} else {
		result.Results = append(result.Results, *current)
	}

	return nil
}

func (s service) observeLegacyActivity(ctx context.Context, result *WorkObservation, current *WorkContext, path, executionID, loopID string, streams *int) error {
	if path == "" || *streams >= 32 {
		return nil
	}

	*streams++

	stream, err := executioncore.ReadRecentAgentEvents(ctx, path)
	if err != nil {
		return err
	}

	if len(stream.Malformed) > 0 {
		result.Warnings = append(result.Warnings, "Some retained activity entries are malformed.")
	}

	for _, event := range stream.Events {
		if !legacyWorkEventMatches(event, s.projectContext.ProjectID, executionID, loopID) {
			continue
		}

		if err = appendObservedLegacyEvent(result, current, event); err != nil {
			return err
		}
	}

	return nil
}

func legacyWorkEventMatches(event executioncore.AgentEvent, project, execution, loop string) bool {
	if event.ProjectID != project {
		return false
	}

	if execution != "" {
		return event.ExecutionID == execution
	}

	return event.LoopID == loop && event.ExecutionID == ""
}

func appendObservedLegacyEvent(result *WorkObservation, current *WorkContext, event executioncore.AgentEvent) error {
	if current.TicketKey != "" && event.TicketKey != "" && current.TicketKey != event.TicketKey {
		return nil
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(encoded)

	result.Events = append(result.Events, WorkEvent{ID: current.ID + ":" + hex.EncodeToString(hash[:]), ContextID: current.ID, TicketKey: event.TicketKey, Phase: string(event.Phase), Message: event.Message, Time: event.Timestamp})
	if event.Timestamp.Before(current.ActivityAt) {
		return nil
	}

	current.Phase, current.Message = string(event.Phase), event.Message
	current.ActivityAt = event.Timestamp

	return nil
}
