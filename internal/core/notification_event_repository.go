package core

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

	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// ErrNotificationEventNotFound identifies a package boundary error.
var ErrNotificationEventNotFound = errors.New("notification event not found")

// NotificationAuditRecord describes a package boundary value.
type NotificationAuditRecord struct {
	Event   NotificationEvent
	Channel string
}

// NotificationEventRepository defines a package boundary interface.
type NotificationEventRepository interface {
	CreateTelegramNotification(context.Context, TelegramNotificationInput) error
	ReadTelegramNotification(context.Context, string, TelegramNotificationTarget) (TelegramNotificationDelivery, bool, error)
	ReserveTelegramNotification(context.Context, string, TelegramNotificationTarget, bool) (bool, error)
	FinishTelegramNotification(context.Context, string, TelegramNotificationTarget, tgmessage.Acknowledgement, string, bool, time.Time) error
	CreateNotificationEvent(ctx context.Context, record NotificationAuditRecord) error
	UpdateNotificationEvent(ctx context.Context, record NotificationAuditRecord) error
	GetNotificationEvent(ctx context.Context, id string) (NotificationAuditRecord, error)
	ListProjectNotificationEvents(ctx context.Context, projectID string) ([]NotificationAuditRecord, error)
	Close() error
}

type sqliteNotificationEventRepository struct {
	db *sql.DB
}

// NewNotificationEventRepository constructs the corresponding package service or value.
func NewNotificationEventRepository(
	ctx context.Context,
	env UserEnvironment,
	assetsFS fs.FS,
) (NotificationEventRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), database.UserFileName))
	if err != nil {
		return nil, fmt.Errorf("open notification event database: %w", err)
	}

	err = database.Configure(ctx, db)
	if err != nil {
		return nil, database.CloseOnError(db, err)
	}

	repository, err := NewNotificationEventRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, database.CloseOnError(db, err)
	}

	return repository, nil
}

// NewNotificationEventRepositoryFromDB constructs the corresponding package service or value.
func NewNotificationEventRepositoryFromDB(
	ctx context.Context,
	db *sql.DB,
	assetsFS fs.FS,
) (NotificationEventRepository, error) {
	migrator := database.NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate notification event database: %w", err)
	}

	return &sqliteNotificationEventRepository{db: db}, nil
}

// CreateNotificationEvent executes the corresponding package operation.
func (r *sqliteNotificationEventRepository) CreateNotificationEvent(
	ctx context.Context,
	record NotificationAuditRecord,
) error {
	err := ValidateNotificationAuditRecord(record)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, insertNotificationEventSQL(), (notificationEventRow{record: record}).values()...)
	if err != nil {
		return fmt.Errorf("create notification event: %w", err)
	}

	return nil
}

// UpdateNotificationEvent executes the corresponding package operation.
func (r *sqliteNotificationEventRepository) UpdateNotificationEvent(
	ctx context.Context,
	record NotificationAuditRecord,
) error {
	err := ValidateNotificationAuditRecord(record)
	if err != nil {
		return err
	}

	event := record.Event

	values := (notificationEventRow{record: record}).values()

	result, err := r.db.ExecContext(ctx, updateNotificationEventSQL(), append(values[1:], values[0])...)
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

// GetNotificationEvent executes the corresponding package operation.
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

// ListProjectNotificationEvents lists the requested data.
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

// Close executes the corresponding package operation.
func (r *sqliteNotificationEventRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close notification event repository database: %w", err)
	}

	return nil
}

// ValidateNotificationAuditRecord validates the corresponding package boundary value.
func ValidateNotificationAuditRecord(record NotificationAuditRecord) error {
	return ValidateNotificationEvent(record.Event)
}

type notificationAuditScanner interface {
	Scan(dest ...any) error
}

func scanNotificationAuditRecord(scanner notificationAuditScanner) (NotificationAuditRecord, error) {
	var row notificationEventRow
	if err := scanner.Scan(row.scanTargets()...); err != nil {
		return NotificationAuditRecord{}, fmt.Errorf("scan notification event: %w", err)
	}

	record := row.record

	parsedCreatedAt, err := database.ParseTime(row.createdAt)
	if err != nil {
		return NotificationAuditRecord{}, err
	}

	parsedUpdatedAt, err := database.ParseTime(row.updatedAt)
	if err != nil {
		return NotificationAuditRecord{}, err
	}

	record.Event.Type = NotificationEventType(row.eventType)
	record.Event.Severity = NotificationSeverity(row.severity)
	record.Event.Status = NotificationStatus(row.status)
	record.Event.Source = NotificationSource(row.source)
	record.Event.CreatedAt = parsedCreatedAt
	record.Event.UpdatedAt = parsedUpdatedAt

	record.Event.DeliveredAt, err = parseOptionalNotificationTime(row.deliveredAt)
	if err != nil {
		return NotificationAuditRecord{}, err
	}

	return record, nil
}

func updateNotificationEventSQL() string {
	return `UPDATE notification_events
SET project_id = ?, type = ?, severity = ?, status = ?, source = ?,
    channel = ?, ticket_key = ?, execution_id = ?, loop_run_id = ?,
    daemon_pid = ?, summary = ?, detail = ?, failure_reason = ?,
    created_at = ?, delivered_at = ?, updated_at = ?
WHERE id = ?`
}

func formatOptionalNotificationTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return database.FormatTime(*value)
}

func parseOptionalNotificationTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := database.ParseTime(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
