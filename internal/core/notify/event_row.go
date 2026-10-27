package notify

import (
	"database/sql"

	"forge.adrianpk.com/chr/internal/core/database"
)

// eventRow keeps event conversions aligned with the retained audit columns.
type eventRow struct {
	record                              AuditRecord
	eventType, severity, status, source string
	createdAt, updatedAt                string
	deliveredAt                         sql.NullString
}

const eventColumns = `id, project_id, type, severity, status, source, channel, ticket_key,
    execution_id, loop_run_id, daemon_pid, summary, detail, failure_reason,
    created_at, delivered_at, updated_at`

func (row eventRow) values() []any {
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
		formatOptionalTime(event.DeliveredAt),
		database.FormatTime(event.UpdatedAt),
	}
}

func (row *eventRow) scanTargets() []any {
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

func insertEventSQL() string {
	return `INSERT INTO notification_events (` + eventColumns + `) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
}

func selectEventSQL() string {
	return `SELECT ` + eventColumns + ` FROM notification_events`
}
