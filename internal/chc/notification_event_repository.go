package chc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNotificationEventNotFound = errors.New("notification event not found")

type NotificationAuditRecord struct {
	Event   NotificationEvent
	Channel string
}

type NotificationEventRepository interface {
	CreateNotificationEvent(ctx context.Context, record NotificationAuditRecord) error
	UpdateNotificationEvent(ctx context.Context, record NotificationAuditRecord) error
	GetNotificationEvent(ctx context.Context, id string) (NotificationAuditRecord, error)
	ListProjectNotificationEvents(ctx context.Context, projectID string) ([]NotificationAuditRecord, error)
	Close() error
}

type sqliteNotificationEventRepository struct {
	db *sql.DB
}

func NewNotificationEventRepository(
	ctx context.Context,
	env UserEnvironment,
	assetsFS fs.FS,
) (NotificationEventRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open notification event database: %w", err)
	}

	err = configureSQLiteDatabase(ctx, db)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	repository, err := NewNotificationEventRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return repository, nil
}

func NewNotificationEventRepositoryFromDB(
	ctx context.Context,
	db *sql.DB,
	assetsFS fs.FS,
) (NotificationEventRepository, error) {
	migrator := NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate notification event database: %w", err)
	}

	return &sqliteNotificationEventRepository{db: db}, nil
}

func (r *sqliteNotificationEventRepository) CreateNotificationEvent(
	ctx context.Context,
	record NotificationAuditRecord,
) error {
	err := ValidateNotificationAuditRecord(record)
	if err != nil {
		return err
	}

	event := record.Event

	_, err = r.db.ExecContext(ctx, insertNotificationEventSQL(),
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
		formatProjectTime(event.CreatedAt),
		formatOptionalNotificationTime(event.DeliveredAt),
		formatProjectTime(event.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create notification event: %w", err)
	}

	return nil
}

func (r *sqliteNotificationEventRepository) UpdateNotificationEvent(
	ctx context.Context,
	record NotificationAuditRecord,
) error {
	err := ValidateNotificationAuditRecord(record)
	if err != nil {
		return err
	}

	event := record.Event

	result, err := r.db.ExecContext(ctx, updateNotificationEventSQL(),
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
		formatProjectTime(event.CreatedAt),
		formatOptionalNotificationTime(event.DeliveredAt),
		formatProjectTime(event.UpdatedAt),
		event.ID,
	)
	if err != nil {
		return fmt.Errorf("update notification event: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect updated notification event rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("%w: %s", ErrNotificationEventNotFound, event.ID)
	}

	return nil
}

func (r *sqliteNotificationEventRepository) GetNotificationEvent(
	ctx context.Context,
	id string,
) (NotificationAuditRecord, error) {
	row := r.db.QueryRowContext(ctx, selectNotificationEventSQL()+" WHERE id = ?", id)

	record, err := scanNotificationAuditRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return NotificationAuditRecord{}, fmt.Errorf("%w: %s", ErrNotificationEventNotFound, id)
	}

	if err != nil {
		return NotificationAuditRecord{}, err
	}

	return record, nil
}

func (r *sqliteNotificationEventRepository) ListProjectNotificationEvents(
	ctx context.Context,
	projectID string,
) ([]NotificationAuditRecord, error) {
	rows, err := r.db.QueryContext(
		ctx,
		selectNotificationEventSQL()+" WHERE project_id = ? ORDER BY created_at",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list project notification events: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked through rows.Err below.
		_ = rows.Close()
	}()

	records := []NotificationAuditRecord{}

	for rows.Next() {
		record, err := scanNotificationAuditRecord(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("list notification event rows: %w", err)
	}

	return records, nil
}

func (r *sqliteNotificationEventRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close notification event repository database: %w", err)
	}

	return nil
}

func ValidateNotificationAuditRecord(record NotificationAuditRecord) error {
	return ValidateNotificationEvent(record.Event)
}

type notificationAuditScanner interface {
	Scan(dest ...any) error
}

func scanNotificationAuditRecord(scanner notificationAuditScanner) (NotificationAuditRecord, error) {
	var record NotificationAuditRecord

	var eventType, severity, status, source string

	var createdAt, updatedAt string

	var deliveredAt sql.NullString

	err := scanner.Scan(
		&record.Event.ID,
		&record.Event.ProjectID,
		&eventType,
		&severity,
		&status,
		&source,
		&record.Channel,
		&record.Event.References.TicketKey,
		&record.Event.References.ExecutionID,
		&record.Event.References.LoopRunID,
		&record.Event.References.DaemonPID,
		&record.Event.Payload.Summary,
		&record.Event.Payload.Detail,
		&record.Event.FailureReason,
		&createdAt,
		&deliveredAt,
		&updatedAt,
	)
	if err != nil {
		return NotificationAuditRecord{}, fmt.Errorf("scan notification event: %w", err)
	}

	parsedCreatedAt, err := parseProjectTime(createdAt)
	if err != nil {
		return NotificationAuditRecord{}, err
	}

	parsedUpdatedAt, err := parseProjectTime(updatedAt)
	if err != nil {
		return NotificationAuditRecord{}, err
	}

	record.Event.Type = NotificationEventType(eventType)
	record.Event.Severity = NotificationSeverity(severity)
	record.Event.Status = NotificationStatus(status)
	record.Event.Source = NotificationSource(source)
	record.Event.CreatedAt = parsedCreatedAt
	record.Event.UpdatedAt = parsedUpdatedAt

	record.Event.DeliveredAt, err = parseOptionalNotificationTime(deliveredAt)
	if err != nil {
		return NotificationAuditRecord{}, err
	}

	return record, nil
}

func insertNotificationEventSQL() string {
	return `INSERT INTO notification_events (
    id, project_id, type, severity, status, source, channel, ticket_key,
    execution_id, loop_run_id, daemon_pid, summary, detail, failure_reason,
    created_at, delivered_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
}

func updateNotificationEventSQL() string {
	return `UPDATE notification_events
SET project_id = ?, type = ?, severity = ?, status = ?, source = ?,
    channel = ?, ticket_key = ?, execution_id = ?, loop_run_id = ?,
    daemon_pid = ?, summary = ?, detail = ?, failure_reason = ?,
    created_at = ?, delivered_at = ?, updated_at = ?
WHERE id = ?`
}

func selectNotificationEventSQL() string {
	return `SELECT id, project_id, type, severity, status, source, channel,
    ticket_key, execution_id, loop_run_id, daemon_pid, summary, detail,
    failure_reason, created_at, delivered_at, updated_at
FROM notification_events`
}

func formatOptionalNotificationTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return formatProjectTime(*value)
}

func parseOptionalNotificationTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := parseProjectTime(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
