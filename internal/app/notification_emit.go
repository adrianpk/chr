package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func emitTicketBlockedNotification(
	ctx context.Context,
	env core.UserEnvironment,
	plan ticketApplyResultPlan,
	assetsFS fs.FS,
) error {
	if plan.Application.TicketStatus != core.TicketResultStatusBlocked {
		return nil
	}

	now := time.Now().UTC()
	event := core.NotificationEvent{
		ID:        workflowNotificationEventID(plan.ProjectContext.ProjectID, "ticket-blocked", now),
		ProjectID: plan.ProjectContext.ProjectID,
		Type:      core.NotificationEventTicketBlocked,
		Severity:  core.NotificationSeverityWarning,
		Status:    core.NotificationStatusSkipped,
		Source:    core.NotificationSourceTicketExecution,
		References: core.NotificationReferences{
			TicketKey:   plan.Application.TicketKey,
			ExecutionID: plan.Application.ExecutionID,
		},
		Payload: core.NotificationPayload{
			Summary: fmt.Sprintf("Ticket %s is blocked", plan.Application.TicketKey),
			Detail:  ticketBlockedNotificationDetail(plan),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	return emitWorkflowNotification(ctx, env, plan.ProjectContext, event, assetsFS)
}

func emitLoopRunNotification(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	record core.LoopRunRecord,
	assetsFS fs.FS,
) error {
	event, ok := loopRunNotificationEvent(projectContext, record, time.Now().UTC())
	if !ok {
		return nil
	}

	return emitWorkflowNotification(ctx, env, projectContext, event, assetsFS)
}

func loopRunNotificationEvent(
	projectContext commandProjectContext,
	record core.LoopRunRecord,
	now time.Time,
) (core.NotificationEvent, bool) {
	if record.Status == core.LoopRunStatusSucceeded && record.ProcessedCount == 0 {
		return core.NotificationEvent{}, false
	}

	eventType, severity := loopRunNotificationType(record.Status)
	if eventType == "" {
		return core.NotificationEvent{}, false
	}

	return core.NotificationEvent{
		ID:        workflowNotificationEventID(projectContext.ProjectID, string(eventType), now),
		ProjectID: projectContext.ProjectID,
		Type:      eventType,
		Severity:  severity,
		Status:    core.NotificationStatusSkipped,
		Source:    core.NotificationSourceTicketLoop,
		References: core.NotificationReferences{
			LoopRunID: record.ID,
		},
		Payload: core.NotificationPayload{
			Summary: loopRunNotificationSummary(record),
			Detail:  loopRunNotificationDetail(record),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, true
}

func loopRunNotificationType(status core.LoopRunStatus) (core.NotificationEventType, core.NotificationSeverity) {
	switch status {
	case core.LoopRunStatusSucceeded:
		return core.NotificationEventLoopFinished, core.NotificationSeverityInfo
	case core.LoopRunStatusBlocked:
		return core.NotificationEventLoopBlocked, core.NotificationSeverityWarning
	case core.LoopRunStatusFailed:
		return core.NotificationEventLoopFailed, core.NotificationSeverityError
	default:
		return "", ""
	}
}

func emitWorkflowNotification(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	event core.NotificationEvent,
	assetsFS fs.FS,
) error {
	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	record := workflowNotificationAuditRecord(event, config.Notifications)

	repository, err := core.NewNotificationEventRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = repository.Close()
	}()

	if record.Event.Status != core.NotificationStatusPending {
		return repository.CreateNotificationEvent(ctx, record)
	}

	target, err := createTelegramNotification(ctx, repository, record, projectContext, config.Notifications.Telegram)
	if err != nil {
		return workflowNotificationFailure(ctx, repository, record.Event.ID, err)
	}

	_, sendErr := dispatchTelegramNotification(ctx, repository, record, projectContext, target)

	return workflowNotificationFailure(ctx, repository, record.Event.ID, sendErr)
}

func workflowNotificationFailure(ctx context.Context, repository core.NotificationEventRepository, id string, failure error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if failure == nil {
		return nil
	}

	saved, err := repository.GetNotificationEvent(ctx, id)
	if err != nil {
		return errors.Join(failure, err)
	}

	if saved.Event.Status == core.NotificationStatusFailed {
		return nil
	}

	return failure
}

func workflowNotificationAuditRecord(
	event core.NotificationEvent,
	config core.NotificationConfig,
) core.NotificationAuditRecord {
	if !config.Telegram.Enabled {
		event.Status = core.NotificationStatusSkipped

		return core.NotificationAuditRecord{Event: event}
	}

	event.Status = core.NotificationStatusPending
	record := core.NotificationAuditRecord{
		Event:   event,
		Channel: "telegram",
	}

	err := core.ValidateNotificationConfig(config)
	if err != nil {
		record.Event.Status = core.NotificationStatusFailed
		record.Event.FailureReason = err.Error()
	}

	return record
}

func ticketBlockedNotificationDetail(plan ticketApplyResultPlan) string {
	detail := fmt.Sprintf("Result: %s", plan.Result.Status)
	if plan.Result.Summary != "" {
		detail = detail + "\nWorker-reported summary: " + plan.Result.Summary
	}

	return detail
}

func loopRunNotificationSummary(record core.LoopRunRecord) string {
	switch record.Status {
	case core.LoopRunStatusSucceeded:
		return "Ticket loop finished"
	case core.LoopRunStatusBlocked:
		return "Ticket loop stopped on a blocker"
	case core.LoopRunStatusFailed:
		return "Ticket loop failed"
	default:
		return "Ticket loop changed state"
	}
}

func loopRunNotificationDetail(record core.LoopRunRecord) string {
	detail := fmt.Sprintf(
		"Eligible: %d\nProcessed: %d\nBlocked: %d\nFailed: %d",
		record.EligibleCount,
		record.ProcessedCount,
		record.BlockedCount,
		record.FailedCount,
	)
	if record.FailureReason != "" {
		detail = detail + "\nFailure: " + record.FailureReason
	}

	return detail
}

func workflowNotificationEventID(projectID, kind string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(projectID + "\x00notification\x00" + kind + "\x00" + createdAt.Format(time.RFC3339Nano)))

	return "notify-" + hex.EncodeToString(sum[:8])
}
