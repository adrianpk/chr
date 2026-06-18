package app

import (
	"context"
	"fmt"
	"time"

	"forge.adrianpk.com/chc/internal/core"
)

func agentWorkLoopEventSink(projectDir string, record core.LoopRunRecord) core.AgentEventSink {
	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return nil
	}

	plan, err := core.PlanRuntime(core.RuntimePlanOptions{
		ProjectDir: projectDir,
		Config:     config.Runtime,
		RunID:      record.ID,
	})
	if err != nil {
		return nil
	}

	return core.NewFileAgentEventSink(core.AgentEventStreamPath(plan.RunArtifactDir))
}

func agentWorkExecutionEventSink(record core.ExecutionRecord) core.AgentEventSink {
	return core.NewFileAgentEventSink(core.AgentEventStreamPath(record.Artifacts.RunDir))
}

func emitAgentWorkEvent(ctx context.Context, sink core.AgentEventSink, event core.AgentEvent) {
	if sink == nil {
		return
	}

	event.SchemaVersion = core.AgentEventSchemaVersion
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if event.Level == "" {
		event.Level = core.AgentEventLevelInfo
	}

	err := sink.EmitAgentEvent(ctx, event)
	if err != nil {
		return
	}
}

func agentWorkTicketRefs(ticket core.IndexedTicket) core.AgentEventRefs {
	path := ticket.RelativePath
	if path == "" {
		path = ticket.Path
	}

	return compactAppAgentEventRefs(core.AgentEventRefs{
		"ticket_path": path,
	})
}

func agentWorkApplicationRefs(plan ticketApplyResultPlan) core.AgentEventRefs {
	return compactAppAgentEventRefs(core.AgentEventRefs{
		"ticket_path":             plan.Application.TicketPath,
		"result_path":             plan.Application.Evidence.ResultPath,
		"runtime_log_path":        plan.Application.Evidence.LogPath,
		"validation_command":      plan.Application.Evidence.ValidationCommand,
		"validation_log_path":     plan.Application.Evidence.ValidationPath,
		"execution_report_path":   core.ExecutionReportPath(plan.Record),
		"result_application_path": plan.ReportPath,
	})
}

func compactAppAgentEventRefs(refs core.AgentEventRefs) core.AgentEventRefs {
	for key, value := range refs {
		if value == "" {
			delete(refs, key)
		}
	}

	if len(refs) == 0 {
		return nil
	}

	return refs
}

func loopFinishedAgentEventLevel(record core.LoopRunRecord) core.AgentEventLevel {
	switch record.Status {
	case core.LoopRunStatusSucceeded:
		return core.AgentEventLevelInfo
	case core.LoopRunStatusBlocked:
		return core.AgentEventLevelWarn
	default:
		return core.AgentEventLevelError
	}
}

func ticketStatusAgentEventLevel(status string) core.AgentEventLevel {
	if status == core.TicketResultStatusBlocked {
		return core.AgentEventLevelWarn
	}

	return core.AgentEventLevelInfo
}

func loopFinishedAgentEventMessage(record core.LoopRunRecord) string {
	return fmt.Sprintf(
		"loop finished: %s, %d processed, %d blocked, %d failed",
		record.Status,
		record.ProcessedCount,
		record.BlockedCount,
		record.FailedCount,
	)
}

func emitLoopStartedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		Phase:     core.AgentEventPhaseLoopStarted,
		Message:   fmt.Sprintf("loop started: %d eligible tickets", record.EligibleCount),
	})
}

func emitTicketSelectedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, ticket core.IndexedTicket) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		TicketKey: ticket.Key,
		Phase:     core.AgentEventPhaseTicketSelected,
		Message:   fmt.Sprintf("%s selected: %s", ticket.Key, ticket.Title),
		Refs:      agentWorkTicketRefs(ticket),
	})
}

func emitLoopFailedAgentEvents(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, cause error) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		Phase:     core.AgentEventPhaseWorkFailed,
		Level:     core.AgentEventLevelError,
		Message:   "loop failed: " + cause.Error(),
	})

	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		Phase:     core.AgentEventPhaseLoopFinished,
		Level:     core.AgentEventLevelError,
		Message:   loopFinishedAgentEventMessage(record),
	})
}

func emitLoopFinishedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		Phase:     core.AgentEventPhaseLoopFinished,
		Level:     loopFinishedAgentEventLevel(record),
		Message:   loopFinishedAgentEventMessage(record),
	})
}

func emitTicketApplicationAgentEvents(ctx context.Context, events ticketExecutionEvents, plan ticketApplyResultPlan) {
	emitAgentWorkEvent(ctx, events.Sink, core.AgentEvent{
		ProjectID:   plan.Application.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: plan.Application.ExecutionID,
		TicketKey:   plan.Application.TicketKey,
		Phase:       core.AgentEventPhaseResultApplied,
		Message:     fmt.Sprintf("%s result applied", plan.Application.TicketKey),
		Refs:        agentWorkApplicationRefs(plan),
	})

	emitTicketStatusAgentEvent(ctx, events, plan)
	emitTicketBlockedAgentEvent(ctx, events, plan)
}

func emitTicketStatusAgentEvent(ctx context.Context, events ticketExecutionEvents, plan ticketApplyResultPlan) {
	emitAgentWorkEvent(ctx, events.Sink, core.AgentEvent{
		ProjectID:   plan.Application.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: plan.Application.ExecutionID,
		TicketKey:   plan.Application.TicketKey,
		Phase:       core.AgentEventPhaseTicketStatusUpdated,
		Level:       ticketStatusAgentEventLevel(plan.Application.TicketStatus),
		Message:     fmt.Sprintf("%s moved to %s", plan.Application.TicketKey, plan.Application.TicketStatus),
		Refs:        agentWorkApplicationRefs(plan),
	})
}

func emitTicketBlockedAgentEvent(ctx context.Context, events ticketExecutionEvents, plan ticketApplyResultPlan) {
	if plan.Application.TicketStatus != core.TicketResultStatusBlocked {
		return
	}

	emitAgentWorkEvent(ctx, events.Sink, core.AgentEvent{
		ProjectID:   plan.Application.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: plan.Application.ExecutionID,
		TicketKey:   plan.Application.TicketKey,
		Phase:       core.AgentEventPhaseWorkBlocked,
		Level:       core.AgentEventLevelWarn,
		Message:     fmt.Sprintf("%s blocked: %s", plan.Application.TicketKey, plan.Application.BlockedReason.Reason),
		Refs:        agentWorkApplicationRefs(plan),
	})
}
