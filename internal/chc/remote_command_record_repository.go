package chc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// ErrRemoteCommandRecordNotFound identifies a package boundary error.
var ErrRemoteCommandRecordNotFound = errors.New("remote command record not found")

// RemoteCommandRecordRepository defines a package boundary interface.
type RemoteCommandRecordRepository interface {
	CreateRemoteCommandRecord(ctx context.Context, record RemoteCommandRecord) error
	GetRemoteCommandRecord(ctx context.Context, id string) (RemoteCommandRecord, error)
	ListRemoteCommandRecords(ctx context.Context) ([]RemoteCommandRecord, error)
	Close() error
}

type sqliteRemoteCommandRecordRepository struct {
	db *sql.DB
}

// NewRemoteCommandRecordRepository constructs the corresponding package service or value.
func NewRemoteCommandRecordRepository(
	ctx context.Context,
	env UserEnvironment,
	assetsFS fs.FS,
) (RemoteCommandRecordRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open remote command database: %w", err)
	}

	err = configureSQLiteDatabase(ctx, db)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	repository, err := NewRemoteCommandRecordRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return repository, nil
}

// NewRemoteCommandRecordRepositoryFromDB constructs the corresponding package service or value.
func NewRemoteCommandRecordRepositoryFromDB(
	ctx context.Context,
	db *sql.DB,
	assetsFS fs.FS,
) (RemoteCommandRecordRepository, error) {
	migrator := NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate remote command database: %w", err)
	}

	return &sqliteRemoteCommandRecordRepository{db: db}, nil
}

// CreateRemoteCommandRecord executes the corresponding package operation.
func (r *sqliteRemoteCommandRecordRepository) CreateRemoteCommandRecord(
	ctx context.Context,
	record RemoteCommandRecord,
) error {
	err := ValidateRemoteCommandRecord(record)
	if err != nil {
		return err
	}

	request := record.Request

	_, err = r.db.ExecContext(ctx, insertRemoteCommandRecordSQL(),
		request.ID,
		string(record.Status),
		request.Sender.Channel,
		request.Sender.ExternalID,
		request.Sender.DisplayName,
		request.ProjectSelector,
		string(request.Intent),
		request.TicketKey,
		request.Comment,
		mustEncodeRemoteCommandValues(request.Loop.Statuses),
		mustEncodeRemoteCommandValues(request.Loop.Priorities),
		mustEncodeRemoteCommandValues(request.Loop.Severities),
		mustEncodeRemoteCommandValues(request.Loop.ProjectAreas),
		mustEncodeRemoteCommandValues(request.Loop.Tags),
		mustEncodeRemoteCommandValues(request.Loop.Types),
		mustEncodeRemoteCommandValues(request.Loop.UseCases),
		request.Loop.MaxIterations,
		request.RawText,
		record.HandoffCommand,
		record.RejectionReason,
		formatProjectTime(request.ReceivedAt),
		formatProjectTime(request.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("create remote command record: %w", err)
	}

	return nil
}

// GetRemoteCommandRecord executes the corresponding package operation.
func (r *sqliteRemoteCommandRecordRepository) GetRemoteCommandRecord(
	ctx context.Context,
	id string,
) (RemoteCommandRecord, error) {
	row := r.db.QueryRowContext(ctx, selectRemoteCommandRecordSQL()+" WHERE id = ?", id)

	record, err := scanRemoteCommandRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return RemoteCommandRecord{}, fmt.Errorf("%w: %s", ErrRemoteCommandRecordNotFound, id)
	}

	if err != nil {
		return RemoteCommandRecord{}, err
	}

	return record, nil
}

// ListRemoteCommandRecords lists the requested data.
func (r *sqliteRemoteCommandRecordRepository) ListRemoteCommandRecords(
	ctx context.Context,
) ([]RemoteCommandRecord, error) {
	rows, err := r.db.QueryContext(ctx, selectRemoteCommandRecordSQL()+" ORDER BY created_at")
	if err != nil {
		return nil, fmt.Errorf("list remote command records: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked through rows.Err below.
		_ = rows.Close()
	}()

	records := []RemoteCommandRecord{}

	for rows.Next() {
		record, err := scanRemoteCommandRecord(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("list remote command rows: %w", err)
	}

	return records, nil
}

// Close executes the corresponding package operation.
func (r *sqliteRemoteCommandRecordRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close remote command repository database: %w", err)
	}

	return nil
}

type remoteCommandRecordScanner interface {
	Scan(dest ...any) error
}

func scanRemoteCommandRecord(scanner remoteCommandRecordScanner) (RemoteCommandRecord, error) {
	var record RemoteCommandRecord

	var status, intent, receivedAt, createdAt string

	var loop remoteCommandLoopColumns

	err := scanner.Scan(
		&record.Request.ID,
		&status,
		&record.Request.Sender.Channel,
		&record.Request.Sender.ExternalID,
		&record.Request.Sender.DisplayName,
		&record.Request.ProjectSelector,
		&intent,
		&record.Request.TicketKey,
		&record.Request.Comment,
		&loop.Statuses,
		&loop.Priorities,
		&loop.Severities,
		&loop.ProjectAreas,
		&loop.Tags,
		&loop.Types,
		&loop.UseCases,
		&record.Request.Loop.MaxIterations,
		&record.Request.RawText,
		&record.HandoffCommand,
		&record.RejectionReason,
		&receivedAt,
		&createdAt,
	)
	if err != nil {
		return RemoteCommandRecord{}, fmt.Errorf("scan remote command record: %w", err)
	}

	return finishScannedRemoteCommandRecord(record, status, intent, receivedAt, createdAt, loop)
}

type remoteCommandLoopColumns struct {
	Statuses     string
	Priorities   string
	Severities   string
	ProjectAreas string
	Tags         string
	Types        string
	UseCases     string
}

func finishScannedRemoteCommandRecord(
	record RemoteCommandRecord,
	status string,
	intent string,
	receivedAt string,
	createdAt string,
	loop remoteCommandLoopColumns,
) (RemoteCommandRecord, error) {
	var err error

	record.Status = RemoteCommandRecordStatus(status)
	record.Request.Intent = RemoteCommandIntent(intent)

	record.Request.ReceivedAt, err = parseProjectTime(receivedAt)
	if err != nil {
		return RemoteCommandRecord{}, err
	}

	record.Request.CreatedAt, err = parseProjectTime(createdAt)
	if err != nil {
		return RemoteCommandRecord{}, err
	}

	record.Request.Loop, err = decodeRemoteCommandLoop(loop, record.Request.Loop.MaxIterations)
	if err != nil {
		return RemoteCommandRecord{}, err
	}

	return record, nil
}

func decodeRemoteCommandLoop(
	columns remoteCommandLoopColumns,
	maxIterations int,
) (RemoteCommandLoopOptions, error) {
	loop := RemoteCommandLoopOptions{MaxIterations: maxIterations}

	var err error

	loop.Statuses, err = decodeRemoteCommandValues(columns.Statuses)
	if err != nil {
		return RemoteCommandLoopOptions{}, err
	}

	loop.Priorities, err = decodeRemoteCommandValues(columns.Priorities)
	if err != nil {
		return RemoteCommandLoopOptions{}, err
	}

	loop.Severities, err = decodeRemoteCommandValues(columns.Severities)
	if err != nil {
		return RemoteCommandLoopOptions{}, err
	}

	loop.ProjectAreas, err = decodeRemoteCommandValues(columns.ProjectAreas)
	if err != nil {
		return RemoteCommandLoopOptions{}, err
	}

	loop.Tags, err = decodeRemoteCommandValues(columns.Tags)
	if err != nil {
		return RemoteCommandLoopOptions{}, err
	}

	loop.Types, err = decodeRemoteCommandValues(columns.Types)
	if err != nil {
		return RemoteCommandLoopOptions{}, err
	}

	loop.UseCases, err = decodeRemoteCommandValues(columns.UseCases)
	if err != nil {
		return RemoteCommandLoopOptions{}, err
	}

	return loop, nil
}

func insertRemoteCommandRecordSQL() string {
	return `INSERT INTO remote_command_records (
    id, status, channel, sender_external_id, sender_display_name,
    project_selector, intent, ticket_key, comment, loop_statuses,
    loop_priorities, loop_severities, loop_project_areas, loop_tags,
    loop_types, loop_use_cases, loop_max_iterations, raw_text,
    handoff_command, rejection_reason, received_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
}

func selectRemoteCommandRecordSQL() string {
	return `SELECT
    id, status, channel, sender_external_id, sender_display_name,
    project_selector, intent, ticket_key, comment, loop_statuses,
    loop_priorities, loop_severities, loop_project_areas, loop_tags,
    loop_types, loop_use_cases, loop_max_iterations, raw_text,
    handoff_command, rejection_reason, received_at, created_at
FROM remote_command_records`
}

func mustEncodeRemoteCommandValues(values []string) string {
	encoded, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func decodeRemoteCommandValues(encoded string) ([]string, error) {
	if strings.TrimSpace(encoded) == "" {
		return nil, nil
	}

	var values []string

	err := json.Unmarshal([]byte(encoded), &values)
	if err != nil {
		return nil, fmt.Errorf("decode remote command values: %w", err)
	}

	return values, nil
}
