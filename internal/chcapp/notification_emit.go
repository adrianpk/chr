package chcapp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

func emitTicketBlockedNotification(
	ctx context.Context,
	env chc.UserEnvironment,
	plan ticketApplyResultPlan,
	assetsFS fs.FS,
) error {
	if plan.Application.TicketStatus != chc.TicketResultStatusBlocked {
		return nil
	}

	now := time.Now().UTC()
	event := chc.NotificationEvent{
		ID:        workflowNotificationEventID(plan.ProjectContext.ProjectID, "ticket-blocked", now),
		ProjectID: plan.ProjectContext.ProjectID,
		Type:      chc.NotificationEventTicketBlocked,
		Severity:  chc.NotificationSeverityWarning,
		Status:    chc.NotificationStatusSkipped,
		Source:    chc.NotificationSourceTicketExecution,
		References: chc.NotificationReferences{
			TicketKey:   plan.Application.TicketKey,
			ExecutionID: plan.Application.ExecutionID,
		},
		Payload: chc.NotificationPayload{
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
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	record chc.LoopRunRecord,
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
	record chc.LoopRunRecord,
	now time.Time,
) (chc.NotificationEvent, bool) {
	if record.Status == chc.LoopRunStatusSucceeded && record.ProcessedCount == 0 {
		return chc.NotificationEvent{}, false
	}

	eventType, severity := loopRunNotificationType(record.Status)
	if eventType == "" {
		return chc.NotificationEvent{}, false
	}

	return chc.NotificationEvent{
		ID:        workflowNotificationEventID(projectContext.ProjectID, string(eventType), now),
		ProjectID: projectContext.ProjectID,
		Type:      eventType,
		Severity:  severity,
		Status:    chc.NotificationStatusSkipped,
		Source:    chc.NotificationSourceTicketLoop,
		References: chc.NotificationReferences{
			LoopRunID: record.ID,
		},
		Payload: chc.NotificationPayload{
			Summary: loopRunNotificationSummary(record),
			Detail:  loopRunNotificationDetail(record),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, true
}

func loopRunNotificationType(status chc.LoopRunStatus) (chc.NotificationEventType, chc.NotificationSeverity) {
	switch status {
	case chc.LoopRunStatusSucceeded:
		return chc.NotificationEventLoopFinished, chc.NotificationSeverityInfo
	case chc.LoopRunStatusBlocked:
		return chc.NotificationEventLoopBlocked, chc.NotificationSeverityWarning
	case chc.LoopRunStatusFailed:
		return chc.NotificationEventLoopFailed, chc.NotificationSeverityError
	default:
		return "", ""
	}
}

func emitWorkflowNotification(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	event chc.NotificationEvent,
	assetsFS fs.FS,
) error {
	config, err := chc.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	record := workflowNotificationAuditRecord(event, config.Notifications)

	repository, err := chc.NewNotificationEventRepository(ctx, env, assetsFS)
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

	if record.Event.Status != chc.NotificationStatusPending {
		return nil
	}

	return dispatchWorkflowTelegramNotification(ctx, repository, record, config.Notifications.Telegram)
}

func workflowNotificationAuditRecord(
	event chc.NotificationEvent,
	config chc.NotificationConfig,
) chc.NotificationAuditRecord {
	if !config.Telegram.Enabled {
		event.Status = chc.NotificationStatusSkipped

		return chc.NotificationAuditRecord{Event: event}
	}

	event.Status = chc.NotificationStatusPending
	record := chc.NotificationAuditRecord{
		Event:   event,
		Channel: "telegram",
	}

	err := chc.ValidateNotificationConfig(config)
	if err != nil {
		record.Event.Status = chc.NotificationStatusFailed
		record.Event.FailureReason = err.Error()
	}

	return record
}

func dispatchWorkflowTelegramNotification(
	ctx context.Context,
	repository chc.NotificationEventRepository,
	record chc.NotificationAuditRecord,
	telegram chc.TelegramNotificationConfig,
) error {
	message := chc.RenderTelegramNotificationMessage(record.Event)
	message.ChatID = telegram.ChatID

	token := os.Getenv(telegram.BotTokenEnv)
	_, sendErr := notifyTelegramSender.SendMessage(ctx, token, message)
	finishedAt := time.Now().UTC()

	if sendErr != nil {
		record.Event.Status = chc.NotificationStatusFailed
		record.Event.FailureReason = sendErr.Error()
		record.Event.UpdatedAt = finishedAt

		return repository.UpdateNotificationEvent(ctx, record)
	}

	record.Event.Status = chc.NotificationStatusDelivered
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

func loopRunNotificationSummary(record chc.LoopRunRecord) string {
	switch record.Status {
	case chc.LoopRunStatusSucceeded:
		return fmt.Sprintf("Ticket loop %s finished", record.ID)
	case chc.LoopRunStatusBlocked:
		return fmt.Sprintf("Ticket loop %s stopped on a blocker", record.ID)
	case chc.LoopRunStatusFailed:
		return fmt.Sprintf("Ticket loop %s failed", record.ID)
	default:
		return fmt.Sprintf("Ticket loop %s changed state", record.ID)
	}
}

func loopRunNotificationDetail(record chc.LoopRunRecord) string {
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
