package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"time"

	"forge.adrianpk.com/chc/internal/workspace"
)

func emitTicketBlockedNotification(
	ctx context.Context,
	env workspace.UserEnvironment,
	plan ticketApplyResultPlan,
	assetsFS fs.FS,
) error {
	if plan.Application.TicketStatus != workspace.TicketResultStatusBlocked {
		return nil
	}

	now := time.Now().UTC()
	event := workspace.NotificationEvent{
		ID:        workflowNotificationEventID(plan.ProjectContext.ProjectID, "ticket-blocked", now),
		ProjectID: plan.ProjectContext.ProjectID,
		Type:      workspace.NotificationEventTicketBlocked,
		Severity:  workspace.NotificationSeverityWarning,
		Status:    workspace.NotificationStatusSkipped,
		Source:    workspace.NotificationSourceTicketExecution,
		References: workspace.NotificationReferences{
			TicketKey:   plan.Application.TicketKey,
			ExecutionID: plan.Application.ExecutionID,
		},
		Payload: workspace.NotificationPayload{
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
	env workspace.UserEnvironment,
	projectContext commandProjectContext,
	record workspace.LoopRunRecord,
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
	record workspace.LoopRunRecord,
	now time.Time,
) (workspace.NotificationEvent, bool) {
	if record.Status == workspace.LoopRunStatusSucceeded && record.ProcessedCount == 0 {
		return workspace.NotificationEvent{}, false
	}

	eventType, severity := loopRunNotificationType(record.Status)
	if eventType == "" {
		return workspace.NotificationEvent{}, false
	}

	return workspace.NotificationEvent{
		ID:        workflowNotificationEventID(projectContext.ProjectID, string(eventType), now),
		ProjectID: projectContext.ProjectID,
		Type:      eventType,
		Severity:  severity,
		Status:    workspace.NotificationStatusSkipped,
		Source:    workspace.NotificationSourceTicketLoop,
		References: workspace.NotificationReferences{
			LoopRunID: record.ID,
		},
		Payload: workspace.NotificationPayload{
			Summary: loopRunNotificationSummary(record),
			Detail:  loopRunNotificationDetail(record),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, true
}

func loopRunNotificationType(status workspace.LoopRunStatus) (workspace.NotificationEventType, workspace.NotificationSeverity) {
	switch status {
	case workspace.LoopRunStatusSucceeded:
		return workspace.NotificationEventLoopFinished, workspace.NotificationSeverityInfo
	case workspace.LoopRunStatusBlocked:
		return workspace.NotificationEventLoopBlocked, workspace.NotificationSeverityWarning
	case workspace.LoopRunStatusFailed:
		return workspace.NotificationEventLoopFailed, workspace.NotificationSeverityError
	default:
		return "", ""
	}
}

func emitWorkflowNotification(
	ctx context.Context,
	env workspace.UserEnvironment,
	projectContext commandProjectContext,
	event workspace.NotificationEvent,
	assetsFS fs.FS,
) error {
	config, err := workspace.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	record := workflowNotificationAuditRecord(event, config.Notifications)

	repository, err := workspace.NewNotificationEventRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = repository.Close()
	}()

	err = repository.CreateNotificationEvent(ctx, record)
	if err != nil {
		return err
	}

	if record.Event.Status != workspace.NotificationStatusPending {
		return nil
	}

	return dispatchWorkflowTelegramNotification(ctx, repository, record, config.Notifications.Telegram)
}

func workflowNotificationAuditRecord(
	event workspace.NotificationEvent,
	config workspace.NotificationConfig,
) workspace.NotificationAuditRecord {
	if !config.Telegram.Enabled {
		event.Status = workspace.NotificationStatusSkipped

		return workspace.NotificationAuditRecord{Event: event}
	}

	event.Status = workspace.NotificationStatusPending
	record := workspace.NotificationAuditRecord{
		Event:   event,
		Channel: "telegram",
	}

	err := workspace.ValidateNotificationConfig(config)
	if err != nil {
		record.Event.Status = workspace.NotificationStatusFailed
		record.Event.FailureReason = err.Error()
	}

	return record
}

func dispatchWorkflowTelegramNotification(
	ctx context.Context,
	repository workspace.NotificationEventRepository,
	record workspace.NotificationAuditRecord,
	telegram workspace.TelegramNotificationConfig,
) error {
	message := workspace.RenderTelegramNotificationMessage(record.Event)
	message.ChatID = telegram.ChatID

	token := os.Getenv(telegram.BotTokenEnv)
	_, sendErr := notifyTelegramSender.SendMessage(ctx, token, message)
	finishedAt := time.Now().UTC()

	if sendErr != nil {
		record.Event.Status = workspace.NotificationStatusFailed
		record.Event.FailureReason = sendErr.Error()
		record.Event.UpdatedAt = finishedAt

		return repository.UpdateNotificationEvent(ctx, record)
	}

	record.Event.Status = workspace.NotificationStatusDelivered
	record.Event.DeliveredAt = &finishedAt
	record.Event.UpdatedAt = finishedAt

	return repository.UpdateNotificationEvent(ctx, record)
}

func ticketBlockedNotificationDetail(plan ticketApplyResultPlan) string {
	detail := fmt.Sprintf("Execution: %s\nResult: %s", plan.Application.ExecutionID, plan.Result.Status)
	if plan.Result.Summary != "" {
		detail = detail + "\nSummary: " + plan.Result.Summary
	}

	return detail
}

func loopRunNotificationSummary(record workspace.LoopRunRecord) string {
	switch record.Status {
	case workspace.LoopRunStatusSucceeded:
		return fmt.Sprintf("Ticket loop %s finished", record.ID)
	case workspace.LoopRunStatusBlocked:
		return fmt.Sprintf("Ticket loop %s stopped on a blocker", record.ID)
	case workspace.LoopRunStatusFailed:
		return fmt.Sprintf("Ticket loop %s failed", record.ID)
	default:
		return fmt.Sprintf("Ticket loop %s changed state", record.ID)
	}
}

func loopRunNotificationDetail(record workspace.LoopRunRecord) string {
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
