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

// ErrExecutionNotFound identifies a package boundary error.
var ErrExecutionNotFound = errors.New("execution not found")

// ExecutionRepository defines a package boundary interface.
type ExecutionRepository interface {
	CreateExecution(ctx context.Context, record ExecutionRecord) error
	UpdateExecution(ctx context.Context, record ExecutionRecord) error
	GetExecution(ctx context.Context, id string) (ExecutionRecord, error)
	ListProjectExecutions(ctx context.Context, projectID string) ([]ExecutionRecord, error)
	RequireProjectQuiescent(ctx context.Context, projectID string) error
	ListTicketExecutions(ctx context.Context, projectID, ticketKey string) ([]ExecutionRecord, error)
	Close() error
}

type sqliteExecutionRepository struct {
	db *sql.DB
}

// RequireProjectQuiescent observes at most one unfinished legacy owner without
// allocating a project's historical executions during a policy transition.
func (r *sqliteExecutionRepository) RequireProjectQuiescent(ctx context.Context, projectID string) error {
	var id string

	err := r.db.QueryRowContext(ctx, "SELECT id FROM execution_records WHERE project_id = ? AND status IN (?, ?) LIMIT 1", projectID, ExecutionStatusPending, ExecutionStatusRunning).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("observe legacy project ownership: %w", err)
	}

	return fmt.Errorf("unfinished legacy execution %s blocks project upgrade", id)
}

// NewExecutionRepository constructs the corresponding package service or value.
func NewExecutionRepository(ctx context.Context, env UserEnvironment, assetsFS fs.FS) (ExecutionRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open execution database: %w", err)
	}

	err = configureSQLiteDatabase(ctx, db)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	repository, err := NewExecutionRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return repository, nil
}

// NewExecutionRepositoryFromDB constructs the corresponding package service or value.
func NewExecutionRepositoryFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (ExecutionRepository, error) {
	migrator := NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate execution database: %w", err)
	}

	return &sqliteExecutionRepository{db: db}, nil
}

// CreateExecution executes the corresponding package operation.
func (r *sqliteExecutionRepository) CreateExecution(ctx context.Context, record ExecutionRecord) error {
	err := ValidateExecutionRecord(record)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, insertExecutionSQL(), (executionRow{record: record}).values()...)
	if err != nil {
		return fmt.Errorf("create execution: %w", err)
	}

	return nil
}

// UpdateExecution executes the corresponding package operation.
func (r *sqliteExecutionRepository) UpdateExecution(ctx context.Context, record ExecutionRecord) error {
	err := ValidateExecutionRecord(record)
	if err != nil {
		return err
	}

	values := (executionRow{record: record}).values()

	result, err := r.db.ExecContext(ctx, updateExecutionSQL(), append(values[1:], values[0])...)
	if err != nil {
		return fmt.Errorf("update execution: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect updated execution rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("%w: %s", ErrExecutionNotFound, record.ID)
	}

	return nil
}

// GetExecution executes the corresponding package operation.
func (r *sqliteExecutionRepository) GetExecution(ctx context.Context, id string) (ExecutionRecord, error) {
	row := r.db.QueryRowContext(ctx, selectExecutionSQL()+" WHERE id = ?", id)

	record, err := scanExecution(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ExecutionRecord{}, fmt.Errorf("%w: %s", ErrExecutionNotFound, id)
	}

	if err != nil {
		return ExecutionRecord{}, err
	}

	return record, nil
}

// ListProjectExecutions lists the requested data.
func (r *sqliteExecutionRepository) ListProjectExecutions(
	ctx context.Context,
	projectID string,
) ([]ExecutionRecord, error) {
	rows, err := r.db.QueryContext(
		ctx,
		selectExecutionSQL()+" WHERE project_id = ? ORDER BY created_at",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list project executions: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked through rows.Err below.
		_ = rows.Close()
	}()

	records, err := scanExecutionRows(rows)
	if err != nil {
		return nil, fmt.Errorf("list project execution rows: %w", err)
	}

	return records, nil
}

// ListTicketExecutions lists the requested data.
func (r *sqliteExecutionRepository) ListTicketExecutions(
	ctx context.Context,
	projectID, ticketKey string,
) ([]ExecutionRecord, error) {
	rows, err := r.db.QueryContext(
		ctx,
		selectExecutionSQL()+" WHERE project_id = ? AND ticket_key = ? ORDER BY created_at",
		projectID,
		ticketKey,
	)
	if err != nil {
		return nil, fmt.Errorf("list ticket executions: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked through rows.Err below.
		_ = rows.Close()
	}()

	records, err := scanExecutionRows(rows)
	if err != nil {
		return nil, fmt.Errorf("list ticket execution rows: %w", err)
	}

	return records, nil
}

func scanExecutionRows(rows *sql.Rows) ([]ExecutionRecord, error) {
	records := []ExecutionRecord{}

	for rows.Next() {
		record, err := scanExecution(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// Close executes the corresponding package operation.
func (r *sqliteExecutionRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close execution repository database: %w", err)
	}

	return nil
}

type executionScanner interface {
	Scan(dest ...any) error
}

type executionRecordTimes struct {
	CreatedAt  time.Time
	UpdatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
}

func scanExecution(scanner executionScanner) (ExecutionRecord, error) {
	var row executionRow
	if err := scanner.Scan(row.scanTargets()...); err != nil {
		return ExecutionRecord{}, fmt.Errorf("scan execution: %w", err)
	}

	return row.decoded()
}

func parseExecutionRecordTimes(createdAt, updatedAt string, startedAt, finishedAt sql.NullString) (executionRecordTimes, error) {
	parsedCreatedAt, err := parseProjectTime(createdAt)
	if err != nil {
		return executionRecordTimes{}, err
	}

	parsedUpdatedAt, err := parseProjectTime(updatedAt)
	if err != nil {
		return executionRecordTimes{}, err
	}

	parsedStartedAt, err := parseOptionalExecutionTime(startedAt)
	if err != nil {
		return executionRecordTimes{}, err
	}

	parsedFinishedAt, err := parseOptionalExecutionTime(finishedAt)
	if err != nil {
		return executionRecordTimes{}, err
	}

	return executionRecordTimes{
		CreatedAt:  parsedCreatedAt,
		UpdatedAt:  parsedUpdatedAt,
		StartedAt:  parsedStartedAt,
		FinishedAt: parsedFinishedAt,
	}, nil
}

func updateExecutionSQL() string {
	return `UPDATE execution_records
SET project_id = ?, ticket_key = ?, ticket_path = ?, status = ?,
    runtime_image = ?, runtime_dockerfile_path = ?, runtime_container_workdir = ?,
    runtime_validation_command = ?, workspace_mode = ?, workspace_project_checkout_dir = ?,
    workspace_execution_dir = ?, workspace_base_ref = ?, workspace_base_sha = ?,
    workspace_ticket_branch = ?, artifact_id = ?, artifact_run_dir = ?, artifact_result_path = ?,
    artifact_log_path = ?, artifact_validation_path = ?, failure_reason = ?,
    created_at = ?, started_at = ?, finished_at = ?, updated_at = ?
WHERE id = ?`
}

func formatOptionalExecutionTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return formatProjectTime(*value)
}

func parseOptionalExecutionTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := parseProjectTime(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
