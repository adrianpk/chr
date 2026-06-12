package workspace

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
	ListTicketExecutions(ctx context.Context, projectID, ticketKey string) ([]ExecutionRecord, error)
	Close() error
}

type sqliteExecutionRepository struct {
	db *sql.DB
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

	_, err = r.db.ExecContext(ctx, insertExecutionSQL(),
		record.ID,
		record.ProjectID,
		record.TicketKey,
		record.TicketPath,
		string(record.Status),
		record.Runtime.Image,
		record.Runtime.DockerfilePath,
		record.Runtime.ContainerWorkdir,
		record.Runtime.ValidationCommand,
		record.Artifacts.RunDir,
		record.Artifacts.ResultPath,
		record.Artifacts.LogPath,
		record.Artifacts.ValidationPath,
		record.FailureReason,
		formatProjectTime(record.CreatedAt),
		formatOptionalExecutionTime(record.StartedAt),
		formatOptionalExecutionTime(record.FinishedAt),
		formatProjectTime(record.UpdatedAt),
	)
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

	result, err := r.db.ExecContext(ctx, updateExecutionSQL(),
		record.ProjectID,
		record.TicketKey,
		record.TicketPath,
		string(record.Status),
		record.Runtime.Image,
		record.Runtime.DockerfilePath,
		record.Runtime.ContainerWorkdir,
		record.Runtime.ValidationCommand,
		record.Artifacts.RunDir,
		record.Artifacts.ResultPath,
		record.Artifacts.LogPath,
		record.Artifacts.ValidationPath,
		record.FailureReason,
		formatProjectTime(record.CreatedAt),
		formatOptionalExecutionTime(record.StartedAt),
		formatOptionalExecutionTime(record.FinishedAt),
		formatProjectTime(record.UpdatedAt),
		record.ID,
	)
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

func scanExecution(scanner executionScanner) (ExecutionRecord, error) {
	var record ExecutionRecord

	var status string

	var createdAt, updatedAt string

	var startedAt, finishedAt sql.NullString

	err := scanner.Scan(
		&record.ID,
		&record.ProjectID,
		&record.TicketKey,
		&record.TicketPath,
		&status,
		&record.Runtime.Image,
		&record.Runtime.DockerfilePath,
		&record.Runtime.ContainerWorkdir,
		&record.Runtime.ValidationCommand,
		&record.Artifacts.RunDir,
		&record.Artifacts.ResultPath,
		&record.Artifacts.LogPath,
		&record.Artifacts.ValidationPath,
		&record.FailureReason,
		&createdAt,
		&startedAt,
		&finishedAt,
		&updatedAt,
	)
	if err != nil {
		return ExecutionRecord{}, fmt.Errorf("scan execution: %w", err)
	}

	parsedCreatedAt, err := parseProjectTime(createdAt)
	if err != nil {
		return ExecutionRecord{}, err
	}

	parsedUpdatedAt, err := parseProjectTime(updatedAt)
	if err != nil {
		return ExecutionRecord{}, err
	}

	record.Status = ExecutionStatus(status)
	record.CreatedAt = parsedCreatedAt
	record.UpdatedAt = parsedUpdatedAt

	record.StartedAt, err = parseOptionalExecutionTime(startedAt)
	if err != nil {
		return ExecutionRecord{}, err
	}

	record.FinishedAt, err = parseOptionalExecutionTime(finishedAt)
	if err != nil {
		return ExecutionRecord{}, err
	}

	return record, nil
}

func insertExecutionSQL() string {
	return `INSERT INTO execution_records (
    id, project_id, ticket_key, ticket_path, status,
    runtime_image, runtime_dockerfile_path, runtime_container_workdir,
    runtime_validation_command, artifact_run_dir, artifact_result_path,
    artifact_log_path, artifact_validation_path, failure_reason,
    created_at, started_at, finished_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
}

func updateExecutionSQL() string {
	return `UPDATE execution_records
SET project_id = ?, ticket_key = ?, ticket_path = ?, status = ?,
    runtime_image = ?, runtime_dockerfile_path = ?, runtime_container_workdir = ?,
    runtime_validation_command = ?, artifact_run_dir = ?, artifact_result_path = ?,
    artifact_log_path = ?, artifact_validation_path = ?, failure_reason = ?,
    created_at = ?, started_at = ?, finished_at = ?, updated_at = ?
WHERE id = ?`
}

func selectExecutionSQL() string {
	return `SELECT id, project_id, ticket_key, ticket_path, status,
    runtime_image, runtime_dockerfile_path, runtime_container_workdir,
    runtime_validation_command, artifact_run_dir, artifact_result_path,
    artifact_log_path, artifact_validation_path, failure_reason,
    created_at, started_at, finished_at, updated_at
FROM execution_records`
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
