package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
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

func agentWorkReviewRefs(ticket core.IndexedTicket) core.AgentEventRefs {
	return compactAppAgentEventRefs(core.AgentEventRefs{
		"ticket_path":         agentWorkTicketPath(ticket),
		"ticket_branch":       ticket.Branch,
		"review_ref":          firstAgentWorkPullRequestURL(ticket),
		"review_comments_url": ticket.ReviewCommentsURL,
	})
}

func agentWorkMergeRefs(result core.TicketMergeCommandResult) core.AgentEventRefs {
	return compactAppAgentEventRefs(core.AgentEventRefs{
		"review_ref":    result.Preflight.ReviewRef.URL,
		"merge_commit":  result.MergeCommit,
		"merge_blocker": result.MergeBlocker,
	})
}

func agentWorkTicketPath(ticket core.IndexedTicket) string {
	path := strings.TrimSpace(ticket.RelativePath)
	if path != "" {
		return path
	}

	return ticket.Path
}

func firstAgentWorkPullRequestURL(ticket core.IndexedTicket) string {
	for _, ref := range ticket.ReviewRefs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return ref.URL
		}
	}

	return ""
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

func emitCycleRehydratedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, execution core.ExecutionRecord, application core.ResultApplication) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   application.TicketKey,
		Phase:       core.AgentEventPhaseCycleRehydrated,
		Message:     fmt.Sprintf("%s cycle rehydrated: ticket status %s", application.TicketKey, application.TicketStatus),
		Refs: compactAppAgentEventRefs(core.AgentEventRefs{
			"ticket_path":           application.TicketPath,
			"execution_report_path": core.ExecutionReportPath(execution),
			"runtime_log_path":      application.Evidence.LogPath,
			"validation_log_path":   application.Evidence.ValidationPath,
		}),
	})
}

func emitReviewStartedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, execution core.ExecutionRecord, ticket core.IndexedTicket, mode core.IntegrationCycleMode) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   ticket.Key,
		Phase:       core.AgentEventPhaseReviewStarted,
		Message:     fmt.Sprintf("%s review started: %s", ticket.Key, mode),
		Refs:        agentWorkReviewRefs(ticket),
	})
}

func emitReviewFinishedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, execution core.ExecutionRecord, ticket core.IndexedTicket) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   ticket.Key,
		Phase:       core.AgentEventPhaseReviewFinished,
		Level:       ticketReviewAgentEventLevel(ticket),
		Message:     fmt.Sprintf("%s review finished: %s", ticket.Key, nonEmptyAgentWorkEventValue(ticket.ReviewResult, ticket.Status)),
		Refs:        agentWorkReviewRefs(ticket),
	})
}

func emitRepairStartedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, ticket core.IndexedTicket) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		TicketKey: ticket.Key,
		Phase:     core.AgentEventPhaseRepairStarted,
		Message:   fmt.Sprintf("%s repair started: review round %d", ticket.Key, ticket.ReviewRound),
		Refs:      agentWorkReviewRefs(ticket),
	})
}

func emitRepairFinishedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, execution core.ExecutionRecord, ticket core.IndexedTicket) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   ticket.Key,
		Phase:       core.AgentEventPhaseRepairFinished,
		Level:       ticketReviewAgentEventLevel(ticket),
		Message:     fmt.Sprintf("%s repair finished: %s", ticket.Key, nonEmptyAgentWorkEventValue(ticket.ReviewResult, ticket.Status)),
		Refs:        agentWorkReviewRefs(ticket),
	})
}

func emitMergeStartedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, execution core.ExecutionRecord, ticket core.IndexedTicket) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   ticket.Key,
		Phase:       core.AgentEventPhaseMergeStarted,
		Message:     fmt.Sprintf("%s merge started", ticket.Key),
		Refs:        agentWorkReviewRefs(ticket),
	})
}

func emitMergeFinishedAgentEvent(ctx context.Context, sink core.AgentEventSink, record core.LoopRunRecord, execution core.ExecutionRecord, result core.TicketMergeCommandResult) {
	emitAgentWorkEvent(ctx, sink, core.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   result.TicketKey,
		Phase:       core.AgentEventPhaseMergeFinished,
		Level:       mergeAgentEventLevel(result),
		Message:     fmt.Sprintf("%s merge finished: %s", result.TicketKey, nonEmptyAgentWorkEventValue(result.MergeResult, result.MergeState)),
		Refs:        agentWorkMergeRefs(result),
	})
}

func emitBranchPreparedAgentEvent(ctx context.Context, events ticketExecutionEvents, record core.ExecutionRecord) {
	if record.Workspace.Mode != core.WorkWorkspaceModeWorktree {
		return
	}

	emitAgentWorkEvent(ctx, events.Sink, core.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Phase:       core.AgentEventPhaseBranchPrepared,
		Message:     fmt.Sprintf("%s branch prepared: %s", record.TicketKey, record.Workspace.TicketBranch),
		Refs: compactAppAgentEventRefs(core.AgentEventRefs{
			"ticket_path":             record.TicketPath,
			"ticket_branch":           record.Workspace.TicketBranch,
			"project_checkout_dir":    record.Workspace.ProjectCheckoutDir,
			"execution_workspace_dir": record.Workspace.ExecutionWorkspaceDir,
			"workspace_base_ref":      record.Workspace.BaseRef,
			"workspace_base_sha":      record.Workspace.BaseSHA,
		}),
	})
}

func emitExecutionCreatedAgentEvent(ctx context.Context, events ticketExecutionEvents, record core.ExecutionRecord) {
	emitAgentWorkEvent(ctx, events.Sink, core.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Phase:       core.AgentEventPhaseExecutionCreated,
		Message:     fmt.Sprintf("%s execution created: %s", record.TicketKey, record.ID),
		Refs: core.AgentEventRefs{
			"ticket_path":         record.TicketPath,
			"result_path":         record.Artifacts.ResultPath,
			"runtime_log_path":    record.Artifacts.LogPath,
			"validation_log_path": record.Artifacts.ValidationPath,
		},
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

func ticketReviewAgentEventLevel(ticket core.IndexedTicket) core.AgentEventLevel {
	warns := core.IsTicketReviewRepair(ticket) ||
		ticket.ReviewResult == core.TicketReviewCommandResultInsufficientEvidence ||
		ticket.ReviewResult == core.TicketReviewCommandResultNeedsHumanDecision

	if warns {
		return core.AgentEventLevelWarn
	}

	return core.AgentEventLevelInfo
}

func mergeAgentEventLevel(result core.TicketMergeCommandResult) core.AgentEventLevel {
	if agentWorkCycleMergeBlocked(result) {
		return core.AgentEventLevelWarn
	}

	return core.AgentEventLevelInfo
}

func nonEmptyAgentWorkEventValue(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}

	return fallback
}
