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
	notifycore "forge.adrianpk.com/chr/internal/core/notify"
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
	event := notifycore.Event{
		ID:        workflowNotificationEventID(plan.ProjectContext.ProjectID, "ticket-blocked", now),
		ProjectID: plan.ProjectContext.ProjectID,
		Type:      notifycore.EventTicketBlocked,
		Severity:  notifycore.SeverityWarning,
		Status:    notifycore.StatusSkipped,
		Source:    notifycore.SourceTicketExecution,
		References: notifycore.References{
			TicketKey:   plan.Application.TicketKey,
			ExecutionID: plan.Application.ExecutionID,
		},
		Payload: notifycore.Payload{
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
) (notifycore.Event, bool) {
	if record.Status == core.LoopRunStatusSucceeded && record.ProcessedCount == 0 {
		return notifycore.Event{}, false
	}

	notificationKind := loopRunNotificationType(record.Status)

	eventType, severity := notificationKind.Type, notificationKind.Severity
	if eventType == "" {
		return notifycore.Event{}, false
	}

	return notifycore.Event{
		ID:        workflowNotificationEventID(projectContext.ProjectID, string(eventType), now),
		ProjectID: projectContext.ProjectID,
		Type:      eventType,
		Severity:  severity,
		Status:    notifycore.StatusSkipped,
		Source:    notifycore.SourceTicketLoop,
		References: notifycore.References{
			LoopRunID: record.ID,
		},
		Payload: notifycore.Payload{
			Summary: loopRunNotificationSummary(record),
			Detail:  loopRunNotificationDetail(record),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, true
}

func loopRunNotificationType(status core.LoopRunStatus) loopNotificationKind {
	switch status {
	case core.LoopRunStatusSucceeded:
		return loopNotificationKind{Type: notifycore.EventLoopFinished, Severity: notifycore.SeverityInfo}
	case core.LoopRunStatusBlocked:
		return loopNotificationKind{Type: notifycore.EventLoopBlocked, Severity: notifycore.SeverityWarning}
	case core.LoopRunStatusFailed:
		return loopNotificationKind{Type: notifycore.EventLoopFailed, Severity: notifycore.SeverityError}
	default:
		return loopNotificationKind{Type: "", Severity: ""}
	}
}

func emitWorkflowNotification(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	event notifycore.Event,
	assetsFS fs.FS,
) error {
	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	record := workflowNotificationAuditRecord(event, config.Notifications)

	repository, err := notifycore.NewRepository(ctx, env, assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The command result already comes from the main operation.
		_ = repository.Close()
	}()

	if record.Event.Status != notifycore.StatusPending {
		return repository.Create(ctx, record)
	}

	target, err := createTelegramNotification(ctx, repository, record, projectContext, config.Notifications.Telegram)
	if err != nil {
		return workflowNotificationFailure(ctx, repository, record.Event.ID, err)
	}

	_, sendErr := dispatchTelegramNotification(ctx, repository, record, projectContext, target)

	return workflowNotificationFailure(ctx, repository, record.Event.ID, sendErr)
}

func workflowNotificationFailure(ctx context.Context, repository notifycore.Repository, id string, failure error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if failure == nil {
		return nil
	}

	saved, err := repository.Get(ctx, id)
	if err != nil {
		return errors.Join(failure, err)
	}

	if saved.Event.Status == notifycore.StatusFailed {
		return nil
	}

	return failure
}

func workflowNotificationAuditRecord(
	event notifycore.Event,
	config notifycore.Config,
) notifycore.AuditRecord {
	if !config.Telegram.Enabled {
		event.Status = notifycore.StatusSkipped

		return notifycore.AuditRecord{Event: event}
	}

	event.Status = notifycore.StatusPending
	record := notifycore.AuditRecord{
		Event:   event,
		Channel: "telegram",
	}

	err := notifycore.ValidateConfig(config)
	if err != nil {
		record.Event.Status = notifycore.StatusFailed
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

type loopNotificationKind struct {
	Type     notifycore.EventType
	Severity notifycore.Severity
}
