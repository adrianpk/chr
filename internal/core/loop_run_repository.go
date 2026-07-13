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
)

// ErrLoopRunNotFound identifies a package boundary error.
var ErrLoopRunNotFound = errors.New("loop run not found")

// LoopRunRepository defines a package boundary interface.
type LoopRunRepository interface {
	CreateLoopRun(ctx context.Context, record LoopRunRecord) error
	UpdateLoopRun(ctx context.Context, record LoopRunRecord) error
	GetLoopRun(ctx context.Context, id string) (LoopRunRecord, error)
	ListProjectLoopRuns(ctx context.Context, projectID string) ([]LoopRunRecord, error)
	Close() error
}

type sqliteLoopRunRepository struct {
	db *sql.DB
}

// NewLoopRunRepository constructs the corresponding package service or value.
func NewLoopRunRepository(ctx context.Context, env UserEnvironment, assetsFS fs.FS) (LoopRunRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open loop run database: %w", err)
	}

	err = configureSQLiteDatabase(ctx, db)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	repository, err := NewLoopRunRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return repository, nil
}

// NewLoopRunRepositoryFromDB constructs the corresponding package service or value.
func NewLoopRunRepositoryFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (LoopRunRepository, error) {
	migrator := NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate loop run database: %w", err)
	}

	return &sqliteLoopRunRepository{db: db}, nil
}

// CreateLoopRun executes the corresponding package operation.
func (r *sqliteLoopRunRepository) CreateLoopRun(ctx context.Context, record LoopRunRecord) error {
	err := ValidateLoopRunRecord(record)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, insertLoopRunSQL(),
		record.ID,
		record.ArtifactID,
		record.ArtifactDir,
		record.ProjectID,
		string(record.Status),
		record.EligibleCount,
		record.ProcessedCount,
		record.BlockedCount,
		record.FailedCount,
		record.FailureReason,
		formatProjectTime(record.CreatedAt),
		formatOptionalLoopRunTime(record.StartedAt),
		formatOptionalLoopRunTime(record.FinishedAt),
		formatProjectTime(record.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create loop run: %w", err)
	}

	return nil
}

// UpdateLoopRun executes the corresponding package operation.
func (r *sqliteLoopRunRepository) UpdateLoopRun(ctx context.Context, record LoopRunRecord) error {
	err := ValidateLoopRunRecord(record)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, updateLoopRunSQL(),
		record.ArtifactID,
		record.ArtifactDir,
		record.ProjectID,
		string(record.Status),
		record.EligibleCount,
		record.ProcessedCount,
		record.BlockedCount,
		record.FailedCount,
		record.FailureReason,
		formatProjectTime(record.CreatedAt),
		formatOptionalLoopRunTime(record.StartedAt),
		formatOptionalLoopRunTime(record.FinishedAt),
		formatProjectTime(record.UpdatedAt),
		record.ID,
	)
	if err != nil {
		return fmt.Errorf("update loop run: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect updated loop run rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("%w: %s", ErrLoopRunNotFound, record.ID)
	}

	return nil
}

// GetLoopRun executes the corresponding package operation.
func (r *sqliteLoopRunRepository) GetLoopRun(ctx context.Context, id string) (LoopRunRecord, error) {
	row := r.db.QueryRowContext(ctx, selectLoopRunSQL()+" WHERE id = ?", id)

	record, err := scanLoopRun(row)
	if errors.Is(err, sql.ErrNoRows) {
		return LoopRunRecord{}, fmt.Errorf("%w: %s", ErrLoopRunNotFound, id)
	}

	if err != nil {
		return LoopRunRecord{}, err
	}

	return record, nil
}

// ListProjectLoopRuns lists the requested data.
func (r *sqliteLoopRunRepository) ListProjectLoopRuns(ctx context.Context, projectID string) ([]LoopRunRecord, error) {
	rows, err := r.db.QueryContext(
		ctx,
		selectLoopRunSQL()+" WHERE project_id = ? ORDER BY created_at",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list project loop runs: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked through rows.Err below.
		_ = rows.Close()
	}()

	records := []LoopRunRecord{}

	for rows.Next() {
		record, err := scanLoopRun(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("list loop run rows: %w", err)
	}

	return records, nil
}

// Close executes the corresponding package operation.
func (r *sqliteLoopRunRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close loop run repository database: %w", err)
	}

	return nil
}

type loopRunScanner interface {
	Scan(dest ...any) error
}

func scanLoopRun(scanner loopRunScanner) (LoopRunRecord, error) {
	var record LoopRunRecord

	var status string

	var createdAt, updatedAt string

	var startedAt, finishedAt sql.NullString

	err := scanner.Scan(
		&record.ID,
		&record.ArtifactID,
		&record.ArtifactDir,
		&record.ProjectID,
		&status,
		&record.EligibleCount,
		&record.ProcessedCount,
		&record.BlockedCount,
		&record.FailedCount,
		&record.FailureReason,
		&createdAt,
		&startedAt,
		&finishedAt,
		&updatedAt,
	)
	if err != nil {
		return LoopRunRecord{}, fmt.Errorf("scan loop run: %w", err)
	}

	parsedCreatedAt, err := parseProjectTime(createdAt)
	if err != nil {
		return LoopRunRecord{}, err
	}

	parsedUpdatedAt, err := parseProjectTime(updatedAt)
	if err != nil {
		return LoopRunRecord{}, err
	}

	record.Status = LoopRunStatus(status)
	record.CreatedAt = parsedCreatedAt
	record.UpdatedAt = parsedUpdatedAt

	record.StartedAt, err = parseOptionalLoopRunTime(startedAt)
	if err != nil {
		return LoopRunRecord{}, err
	}

	record.FinishedAt, err = parseOptionalLoopRunTime(finishedAt)
	if err != nil {
		return LoopRunRecord{}, err
	}

	return record, nil
}

func insertLoopRunSQL() string {
	return `INSERT INTO loop_runs (
    id, artifact_id, artifact_dir, project_id, status, eligible_count, processed_count, blocked_count,
    failed_count, failure_reason, created_at, started_at, finished_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
}

func updateLoopRunSQL() string {
	return `UPDATE loop_runs
SET artifact_id = ?, artifact_dir = ?, project_id = ?, status = ?, eligible_count = ?, processed_count = ?,
    blocked_count = ?, failed_count = ?, failure_reason = ?, created_at = ?,
    started_at = ?, finished_at = ?, updated_at = ?
WHERE id = ?`
}

func selectLoopRunSQL() string {
	return `SELECT id, artifact_id, artifact_dir, project_id, status, eligible_count, processed_count,
    blocked_count, failed_count, failure_reason, created_at, started_at,
    finished_at, updated_at
FROM loop_runs`
}

func formatOptionalLoopRunTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return formatProjectTime(*value)
}

func parseOptionalLoopRunTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := parseProjectTime(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
