package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func agentWorkLoopEventSink(projectDir string, record executioncore.LoopRunRecord) executioncore.AgentEventSink {
	if record.ArtifactID != "" {
		return executioncore.NewFileAgentEventSink(executioncore.LoopRunEventStreamPath(record))
	}

	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return nil
	}

	plan, err := harnessruntime.ResolvePlan(harnessruntime.PlanOptions{
		ExtraMounts: legacyCodexMounts(),
		ProjectDir:  projectDir,
		Config:      config.Runtime,
		RunID:       record.ID,
	})
	if err != nil {
		return nil
	}

	return executioncore.NewFileAgentEventSink(executioncore.AgentEventStreamPath(plan.RunArtifactDir))
}

func agentWorkExecutionEventSink(record executioncore.ExecutionRecord) executioncore.AgentEventSink {
	return executioncore.NewFileAgentEventSink(executioncore.ExecutionAgentEventStreamPath(record))
}

func emitAgentWorkEvent(ctx context.Context, sink executioncore.AgentEventSink, event executioncore.AgentEvent) {
	if sink == nil {
		return
	}

	event.SchemaVersion = executioncore.AgentEventSchemaVersion
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if event.Level == "" {
		event.Level = executioncore.AgentEventLevelInfo
	}

	err := sink.EmitAgentEvent(ctx, event)
	if err != nil {
		return
	}
}

func agentWorkTicketRefs(ticket ticketcore.Indexed) executioncore.AgentEventRefs {
	path := ticket.RelativePath
	if path == "" {
		path = ticket.Path
	}

	return compactAppAgentEventRefs(executioncore.AgentEventRefs{
		"ticket_path": path,
	})
}

func agentWorkApplicationRefs(plan ticketApplyResultPlan) executioncore.AgentEventRefs {
	return compactAppAgentEventRefs(executioncore.AgentEventRefs{
		"ticket_path":             plan.Application.TicketPath,
		"result_path":             plan.Application.Evidence.ResultPath,
		"runtime_log_path":        plan.Application.Evidence.LogPath,
		"validation_command":      plan.Application.Evidence.ValidationCommand,
		"validation_log_path":     plan.Application.Evidence.ValidationPath,
		"execution_report_path":   executioncore.ExecutionReportPath(plan.Record),
		"result_application_path": plan.ReportPath,
	})
}

func agentWorkReviewRefs(ticket ticketcore.Indexed) executioncore.AgentEventRefs {
	return compactAppAgentEventRefs(executioncore.AgentEventRefs{
		"ticket_path":         agentWorkTicketPath(ticket),
		"ticket_branch":       ticket.Branch,
		"review_ref":          firstAgentWorkPullRequestURL(ticket),
		"review_comments_url": ticket.ReviewCommentsURL,
	})
}

func agentWorkMergeRefs(result core.TicketMergeCommandResult) executioncore.AgentEventRefs {
	return compactAppAgentEventRefs(executioncore.AgentEventRefs{
		"review_ref":    result.Preflight.ReviewRef.URL,
		"merge_commit":  result.MergeCommit,
		"merge_blocker": result.MergeBlocker,
	})
}

func agentWorkTicketPath(ticket ticketcore.Indexed) string {
	path := strings.TrimSpace(ticket.RelativePath)
	if path != "" {
		return path
	}

	return ticket.Path
}

func firstAgentWorkPullRequestURL(ticket ticketcore.Indexed) string {
	for _, ref := range ticket.ReviewRefs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return ref.URL
		}
	}

	return ""
}

func compactAppAgentEventRefs(refs executioncore.AgentEventRefs) executioncore.AgentEventRefs {
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

func loopFinishedAgentEventLevel(record executioncore.LoopRunRecord) executioncore.AgentEventLevel {
	switch record.Status {
	case executioncore.LoopRunStatusSucceeded:
		return executioncore.AgentEventLevelInfo
	case executioncore.LoopRunStatusBlocked:
		return executioncore.AgentEventLevelWarn
	default:
		return executioncore.AgentEventLevelError
	}
}

func ticketStatusAgentEventLevel(status string) executioncore.AgentEventLevel {
	if status == core.TicketResultStatusBlocked {
		return executioncore.AgentEventLevelWarn
	}

	return executioncore.AgentEventLevelInfo
}

func loopFinishedAgentEventMessage(record executioncore.LoopRunRecord) string {
	return fmt.Sprintf(
		"loop finished: %s, %d processed, %d blocked, %d failed",
		record.Status,
		record.ProcessedCount,
		record.BlockedCount,
		record.FailedCount,
	)
}

func emitLoopStartedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		Phase:     executioncore.AgentEventPhaseLoopStarted,
		Message:   fmt.Sprintf("loop started: %d eligible tickets", record.EligibleCount),
	})
}

func emitTicketSelectedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord, ticket ticketcore.Indexed) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		TicketKey: ticket.Key,
		Phase:     executioncore.AgentEventPhaseTicketSelected,
		Message:   fmt.Sprintf("%s selected: %s", ticket.Key, ticket.Title),
		Refs:      agentWorkTicketRefs(ticket),
	})
}

func emitCycleRehydratedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord, execution executioncore.ExecutionRecord, application core.ResultApplication) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   application.TicketKey,
		Phase:       executioncore.AgentEventPhaseCycleRehydrated,
		Message:     fmt.Sprintf("%s cycle rehydrated: ticket status %s", application.TicketKey, application.TicketStatus),
		Refs: compactAppAgentEventRefs(executioncore.AgentEventRefs{
			"ticket_path":           application.TicketPath,
			"execution_report_path": executioncore.ExecutionReportPath(execution),
			"runtime_log_path":      application.Evidence.LogPath,
			"validation_log_path":   application.Evidence.ValidationPath,
		}),
	})
}

func emitReviewStartedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, review reviewStartedEvent) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID:   review.Record.ProjectID,
		LoopID:      review.Record.ID,
		ExecutionID: review.Execution.ID,
		TicketKey:   review.Ticket.Key,
		Phase:       executioncore.AgentEventPhaseReviewStarted,
		Message:     fmt.Sprintf("%s review started: %s", review.Ticket.Key, review.Mode),
		Refs:        agentWorkReviewRefs(review.Ticket),
	})
}

func emitReviewFinishedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord, execution executioncore.ExecutionRecord, ticket ticketcore.Indexed) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   ticket.Key,
		Phase:       executioncore.AgentEventPhaseReviewFinished,
		Level:       ticketReviewAgentEventLevel(ticket),
		Message:     fmt.Sprintf("%s review finished: %s", ticket.Key, nonEmptyAgentWorkEventValue(ticket.ReviewResult, ticket.Status)),
		Refs:        agentWorkReviewRefs(ticket),
	})
}

func emitRepairStartedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord, ticket ticketcore.Indexed) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		TicketKey: ticket.Key,
		Phase:     executioncore.AgentEventPhaseRepairStarted,
		Message:   fmt.Sprintf("%s repair started: review round %d", ticket.Key, ticket.ReviewRound),
		Refs:      agentWorkReviewRefs(ticket),
	})
}

func emitRepairFinishedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord, execution executioncore.ExecutionRecord, ticket ticketcore.Indexed) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   ticket.Key,
		Phase:       executioncore.AgentEventPhaseRepairFinished,
		Level:       ticketReviewAgentEventLevel(ticket),
		Message:     fmt.Sprintf("%s repair finished: %s", ticket.Key, nonEmptyAgentWorkEventValue(ticket.ReviewResult, ticket.Status)),
		Refs:        agentWorkReviewRefs(ticket),
	})
}

func emitMergeStartedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord, execution executioncore.ExecutionRecord, ticket ticketcore.Indexed) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   ticket.Key,
		Phase:       executioncore.AgentEventPhaseMergeStarted,
		Message:     fmt.Sprintf("%s merge started", ticket.Key),
		Refs:        agentWorkReviewRefs(ticket),
	})
}

func emitMergeFinishedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord, execution executioncore.ExecutionRecord, result core.TicketMergeCommandResult) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      record.ID,
		ExecutionID: execution.ID,
		TicketKey:   result.TicketKey,
		Phase:       executioncore.AgentEventPhaseMergeFinished,
		Level:       mergeAgentEventLevel(result),
		Message:     fmt.Sprintf("%s merge finished: %s", result.TicketKey, nonEmptyAgentWorkEventValue(result.MergeResult, result.MergeState)),
		Refs:        agentWorkMergeRefs(result),
	})
}

func emitBranchPreparedAgentEvent(ctx context.Context, events ticketExecutionEvents, record executioncore.ExecutionRecord) {
	if record.Workspace.Mode != workspacecore.ExecutionWorktree {
		return
	}

	emitAgentWorkEvent(ctx, events.Sink, executioncore.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Phase:       executioncore.AgentEventPhaseBranchPrepared,
		Message:     fmt.Sprintf("%s branch prepared: %s", record.TicketKey, record.Workspace.TicketBranch),
		Refs: compactAppAgentEventRefs(executioncore.AgentEventRefs{
			"ticket_path":             record.TicketPath,
			"ticket_branch":           record.Workspace.TicketBranch,
			"project_checkout_dir":    record.Workspace.ProjectCheckoutDir,
			"execution_workspace_dir": record.Workspace.ExecutionWorkspaceDir,
			"workspace_base_ref":      record.Workspace.BaseRef,
			"workspace_base_sha":      record.Workspace.BaseSHA,
		}),
	})
}

func emitExecutionCreatedAgentEvent(ctx context.Context, events ticketExecutionEvents, record executioncore.ExecutionRecord) {
	emitAgentWorkEvent(ctx, events.Sink, executioncore.AgentEvent{
		ProjectID:   record.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Phase:       executioncore.AgentEventPhaseExecutionCreated,
		Message:     fmt.Sprintf("%s execution created: %s", record.TicketKey, record.ID),
		Refs: executioncore.AgentEventRefs{
			"ticket_path":         record.TicketPath,
			"result_path":         record.Artifacts.ResultPath,
			"runtime_log_path":    record.Artifacts.LogPath,
			"validation_log_path": record.Artifacts.ValidationPath,
		},
	})
}

func emitLoopFailedAgentEvents(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord, cause error) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		Phase:     executioncore.AgentEventPhaseWorkFailed,
		Level:     executioncore.AgentEventLevelError,
		Message:   "loop failed: " + cause.Error(),
	})

	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		Phase:     executioncore.AgentEventPhaseLoopFinished,
		Level:     executioncore.AgentEventLevelError,
		Message:   loopFinishedAgentEventMessage(record),
	})
}

func emitLoopFinishedAgentEvent(ctx context.Context, sink executioncore.AgentEventSink, record executioncore.LoopRunRecord) {
	emitAgentWorkEvent(ctx, sink, executioncore.AgentEvent{
		ProjectID: record.ProjectID,
		LoopID:    record.ID,
		Phase:     executioncore.AgentEventPhaseLoopFinished,
		Level:     loopFinishedAgentEventLevel(record),
		Message:   loopFinishedAgentEventMessage(record),
	})
}

func emitTicketApplicationAgentEvents(ctx context.Context, events ticketExecutionEvents, plan ticketApplyResultPlan) {
	emitAgentWorkEvent(ctx, events.Sink, executioncore.AgentEvent{
		ProjectID:   plan.Application.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: plan.Application.ExecutionID,
		TicketKey:   plan.Application.TicketKey,
		Phase:       executioncore.AgentEventPhaseResultApplied,
		Message:     fmt.Sprintf("%s result applied", plan.Application.TicketKey),
		Refs:        agentWorkApplicationRefs(plan),
	})

	emitTicketStatusAgentEvent(ctx, events, plan)
	emitTicketBlockedAgentEvent(ctx, events, plan)
}

func emitTicketStatusAgentEvent(ctx context.Context, events ticketExecutionEvents, plan ticketApplyResultPlan) {
	emitAgentWorkEvent(ctx, events.Sink, executioncore.AgentEvent{
		ProjectID:   plan.Application.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: plan.Application.ExecutionID,
		TicketKey:   plan.Application.TicketKey,
		Phase:       executioncore.AgentEventPhaseTicketStatusUpdated,
		Level:       ticketStatusAgentEventLevel(plan.Application.TicketStatus),
		Message:     fmt.Sprintf("%s moved to %s", plan.Application.TicketKey, plan.Application.TicketStatus),
		Refs:        agentWorkApplicationRefs(plan),
	})
}

func emitTicketBlockedAgentEvent(ctx context.Context, events ticketExecutionEvents, plan ticketApplyResultPlan) {
	if plan.Application.TicketStatus != core.TicketResultStatusBlocked {
		return
	}

	emitAgentWorkEvent(ctx, events.Sink, executioncore.AgentEvent{
		ProjectID:   plan.Application.ProjectID,
		LoopID:      events.LoopID,
		ExecutionID: plan.Application.ExecutionID,
		TicketKey:   plan.Application.TicketKey,
		Phase:       executioncore.AgentEventPhaseWorkBlocked,
		Level:       executioncore.AgentEventLevelWarn,
		Message:     fmt.Sprintf("%s blocked: %s", plan.Application.TicketKey, plan.Application.BlockedReason.Reason),
		Refs:        agentWorkApplicationRefs(plan),
	})
}

func ticketReviewAgentEventLevel(ticket ticketcore.Indexed) executioncore.AgentEventLevel {
	warns := core.IsTicketReviewRepair(ticket) ||
		ticket.ReviewResult == core.TicketReviewCommandResultInsufficientEvidence ||
		ticket.ReviewResult == core.TicketReviewCommandResultNeedsHumanDecision

	if warns {
		return executioncore.AgentEventLevelWarn
	}

	return executioncore.AgentEventLevelInfo
}

func mergeAgentEventLevel(result core.TicketMergeCommandResult) executioncore.AgentEventLevel {
	if agentWorkCycleMergeBlocked(result) {
		return executioncore.AgentEventLevelWarn
	}

	return executioncore.AgentEventLevelInfo
}

func nonEmptyAgentWorkEventValue(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}

	return fallback
}

type reviewStartedEvent struct {
	Record    executioncore.LoopRunRecord
	Execution executioncore.ExecutionRecord
	Ticket    ticketcore.Indexed
	Mode      core.IntegrationCycleMode
}
