package core

import (
	"database/sql"

	"forge.adrianpk.com/chr/internal/core/database"
)

// notificationEventRow keeps event conversions aligned with the retained audit columns.
type notificationEventRow struct {
	record                              NotificationAuditRecord
	eventType, severity, status, source string
	createdAt, updatedAt                string
	deliveredAt                         sql.NullString
}

const notificationEventColumns = `id, project_id, type, severity, status, source, channel, ticket_key,
    execution_id, loop_run_id, daemon_pid, summary, detail, failure_reason,
    created_at, delivered_at, updated_at`

func (row notificationEventRow) values() []any {
	record := row.record
	event := record.Event

	return []any{
		event.ID,
		event.ProjectID,
		string(event.Type),
		string(event.Severity),
		string(event.Status),
		string(event.Source),
		record.Channel,
		event.References.TicketKey,
		event.References.ExecutionID,
		event.References.LoopRunID,
		event.References.DaemonPID,
		event.Payload.Summary,
		event.Payload.Detail,
		event.FailureReason,
		database.FormatTime(event.CreatedAt),
		formatOptionalNotificationTime(event.DeliveredAt),
		database.FormatTime(event.UpdatedAt),
	}
}

func (row *notificationEventRow) scanTargets() []any {
	record := &row.record

	return []any{
		&record.Event.ID,
		&record.Event.ProjectID,
		&row.eventType,
		&row.severity,
		&row.status,
		&row.source,
		&record.Channel,
		&record.Event.References.TicketKey,
		&record.Event.References.ExecutionID,
		&record.Event.References.LoopRunID,
		&record.Event.References.DaemonPID,
		&record.Event.Payload.Summary,
		&record.Event.Payload.Detail,
		&record.Event.FailureReason,
		&row.createdAt,
		&row.deliveredAt,
		&row.updatedAt,
	}
}

func insertNotificationEventSQL() string {
	return `INSERT INTO notification_events (` + notificationEventColumns + `) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
}

func selectNotificationEventSQL() string {
	return `SELECT ` + notificationEventColumns + ` FROM notification_events`
}
