package notify

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
)

// ErrEventNotFound identifies a package boundary error.
var ErrEventNotFound = errors.New("notification event not found")

// AuditRecord describes a package boundary value.
type AuditRecord struct {
	Event   Event
	Channel string
}

// Repository defines a package boundary interface.
type Repository interface {
	CreateTelegram(context.Context, TelegramInput) error
	ReadTelegram(context.Context, string, TelegramTarget) (TelegramLookup, error)
	ReserveTelegram(context.Context, string, TelegramTarget, bool) (bool, error)
	FinishTelegram(context.Context, TelegramCompletion) error
	Create(ctx context.Context, record AuditRecord) error
	Update(ctx context.Context, record AuditRecord) error
	Get(ctx context.Context, id string) (AuditRecord, error)
	ListProject(ctx context.Context, projectID string) ([]AuditRecord, error)
	Close() error
}

type sqliteRepository struct {
	db *sql.DB
}

// NewRepository constructs the corresponding package service or value.
func NewRepository(ctx context.Context, env database.DataDirectory, assetsFS fs.FS) (Repository, error) {
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

	repository, err := NewRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, database.CloseOnError(db, err)
	}

	return repository, nil
}

// NewRepositoryFromDB constructs the corresponding package service or value.
func NewRepositoryFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (Repository, error) {
	migrator := database.NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate notification event database: %w", err)
	}

	return &sqliteRepository{db: db}, nil
}

// Create executes the corresponding package operation.
func (r *sqliteRepository) Create(ctx context.Context, record AuditRecord) error {
	err := ValidateAuditRecord(record)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, insertEventSQL(), (eventRow{record: record}).values()...)
	if err != nil {
		return fmt.Errorf("create notification event: %w", err)
	}

	return nil
}

// Update executes the corresponding package operation.
func (r *sqliteRepository) Update(ctx context.Context, record AuditRecord) error {
	err := ValidateAuditRecord(record)
	if err != nil {
		return err
	}

	event := record.Event

	values := (eventRow{record: record}).values()

	result, err := r.db.ExecContext(ctx, updateEventSQL(), append(values[1:], values[0])...)
	if err != nil {
		return fmt.Errorf("update notification event: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect updated notification event rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("%w: %s", ErrEventNotFound, event.ID)
	}

	return nil
}

// Get executes the corresponding package operation.
func (r *sqliteRepository) Get(ctx context.Context, id string) (AuditRecord, error) {
	row := r.db.QueryRowContext(ctx, selectEventSQL()+" WHERE id = ?", id)

	record, err := scanAuditRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return AuditRecord{}, fmt.Errorf("%w: %s", ErrEventNotFound, id)
	}

	if err != nil {
		return AuditRecord{}, err
	}

	return record, nil
}

// ListProject lists the requested data.
func (r *sqliteRepository) ListProject(ctx context.Context, projectID string) ([]AuditRecord, error) {
	rows, err := r.db.QueryContext(
		ctx,
		selectEventSQL()+" WHERE project_id = ? ORDER BY created_at",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list project notification events: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked through rows.Err below.
		_ = rows.Close()
	}()

	records := []AuditRecord{}

	for rows.Next() {
		record, err := scanAuditRecord(rows)
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
func (r *sqliteRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close notification event repository database: %w", err)
	}

	return nil
}

// ValidateAuditRecord validates the corresponding package boundary value.
func ValidateAuditRecord(record AuditRecord) error {
	return ValidateEvent(record.Event)
}

type auditScanner interface {
	Scan(dest ...any) error
}

func scanAuditRecord(scanner auditScanner) (AuditRecord, error) {
	var row eventRow
	if err := scanner.Scan(row.scanTargets()...); err != nil {
		return AuditRecord{}, fmt.Errorf("scan notification event: %w", err)
	}

	record := row.record

	parsedCreatedAt, err := database.ParseTime(row.createdAt)
	if err != nil {
		return AuditRecord{}, err
	}

	parsedUpdatedAt, err := database.ParseTime(row.updatedAt)
	if err != nil {
		return AuditRecord{}, err
	}

	record.Event.Type = EventType(row.eventType)
	record.Event.Severity = Severity(row.severity)
	record.Event.Status = Status(row.status)
	record.Event.Source = Source(row.source)
	record.Event.CreatedAt = parsedCreatedAt
	record.Event.UpdatedAt = parsedUpdatedAt

	record.Event.DeliveredAt, err = parseOptionalTime(row.deliveredAt)
	if err != nil {
		return AuditRecord{}, err
	}

	return record, nil
}

func updateEventSQL() string {
	return `UPDATE notification_events
SET project_id = ?, type = ?, severity = ?, status = ?, source = ?,
    channel = ?, ticket_key = ?, execution_id = ?, loop_run_id = ?,
    daemon_pid = ?, summary = ?, detail = ?, failure_reason = ?,
    created_at = ?, delivered_at = ?, updated_at = ?
WHERE id = ?`
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return database.FormatTime(*value)
}

func parseOptionalTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := database.ParseTime(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
