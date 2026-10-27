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
)

const daemonStateID = "daemon"

// ErrDaemonStateNotFound identifies a package boundary error.
var ErrDaemonStateNotFound = errors.New("daemon state not found")

// DaemonStateRepository defines a package boundary interface.
type DaemonStateRepository interface {
	GetDaemonState(ctx context.Context) (DaemonState, error)
	UpsertDaemonState(ctx context.Context, state DaemonState) error
	UpdateDaemonHeartbeat(ctx context.Context, heartbeatAt time.Time) error
	RequestDaemonStop(ctx context.Context, requestedAt time.Time) error
	ClearDaemonState(ctx context.Context) error
	Close() error
}

type sqliteDaemonStateRepository struct {
	db *sql.DB
}

// NewDaemonStateRepository constructs the corresponding package service or value.
func NewDaemonStateRepository(ctx context.Context, env UserEnvironment, assetsFS fs.FS) (DaemonStateRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), database.UserFileName))
	if err != nil {
		return nil, fmt.Errorf("open daemon state database: %w", err)
	}

	err = database.Configure(ctx, db)
	if err != nil {
		return nil, database.CloseOnError(db, err)
	}

	repository, err := NewDaemonStateRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, database.CloseOnError(db, err)
	}

	return repository, nil
}

// NewDaemonStateRepositoryFromDB constructs the corresponding package service or value.
func NewDaemonStateRepositoryFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (DaemonStateRepository, error) {
	migrator := database.NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate daemon state database: %w", err)
	}

	return &sqliteDaemonStateRepository{db: db}, nil
}

// GetDaemonState executes the corresponding package operation.
func (r *sqliteDaemonStateRepository) GetDaemonState(ctx context.Context) (DaemonState, error) {
	row := r.db.QueryRowContext(ctx, selectDaemonStateSQL()+" WHERE id = ?", daemonStateID)

	state, err := scanDaemonState(row)
	if errors.Is(err, sql.ErrNoRows) {
		return DaemonState{}, ErrDaemonStateNotFound
	}

	if err != nil {
		return DaemonState{}, err
	}

	return state, nil
}

// UpsertDaemonState executes the corresponding package operation.
func (r *sqliteDaemonStateRepository) UpsertDaemonState(ctx context.Context, state DaemonState) error {
	err := ValidateDaemonState(state)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, upsertDaemonStateSQL(),
		daemonStateID,
		string(state.Status),
		state.PID,
		formatOptionalDaemonTime(state.StartedAt),
		formatOptionalDaemonTime(state.HeartbeatAt),
		formatOptionalDaemonTime(state.StopRequestedAt),
		state.FailureReason,
		database.FormatTime(state.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert daemon state: %w", err)
	}

	return nil
}

// UpdateDaemonHeartbeat executes the corresponding package operation.
func (r *sqliteDaemonStateRepository) UpdateDaemonHeartbeat(ctx context.Context, heartbeatAt time.Time) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE daemon_state SET heartbeat_at = ?, updated_at = ? WHERE id = ?`,
		database.FormatTime(heartbeatAt),
		database.FormatTime(heartbeatAt),
		daemonStateID,
	)
	if err != nil {
		return fmt.Errorf("update daemon heartbeat: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect updated daemon heartbeat rows: %w", err)
	}

	if affected == 0 {
		return ErrDaemonStateNotFound
	}

	return nil
}

// RequestDaemonStop executes the corresponding package operation.
func (r *sqliteDaemonStateRepository) RequestDaemonStop(ctx context.Context, requestedAt time.Time) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE daemon_state
SET status = ?, stop_requested_at = ?, updated_at = ?
WHERE id = ? AND status IN (?, ?, ?)`,
		string(DaemonStatusStopping),
		database.FormatTime(requestedAt),
		database.FormatTime(requestedAt),
		daemonStateID,
		string(DaemonStatusStarting),
		string(DaemonStatusRunning),
		string(DaemonStatusStopping),
	)
	if err != nil {
		return fmt.Errorf("request daemon stop: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect daemon stop request rows: %w", err)
	}

	if affected == 0 {
		return ErrDaemonStateNotFound
	}

	return nil
}

// ClearDaemonState executes the corresponding package operation.
func (r *sqliteDaemonStateRepository) ClearDaemonState(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM daemon_state WHERE id = ?", daemonStateID)
	if err != nil {
		return fmt.Errorf("clear daemon state: %w", err)
	}

	return nil
}

// Close executes the corresponding package operation.
func (r *sqliteDaemonStateRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close daemon state repository database: %w", err)
	}

	return nil
}

type daemonStateScanner interface {
	Scan(dest ...any) error
}

func scanDaemonState(scanner daemonStateScanner) (DaemonState, error) {
	var state DaemonState

	var id, status, updatedAt string

	var startedAt, heartbeatAt, stopRequestedAt sql.NullString

	err := scanner.Scan(
		&id,
		&status,
		&state.PID,
		&startedAt,
		&heartbeatAt,
		&stopRequestedAt,
		&state.FailureReason,
		&updatedAt,
	)
	if err != nil {
		return DaemonState{}, fmt.Errorf("scan daemon state: %w", err)
	}

	if id != daemonStateID {
		return DaemonState{}, fmt.Errorf("unexpected daemon state id: %s", id)
	}

	parsedUpdatedAt, err := database.ParseTime(updatedAt)
	if err != nil {
		return DaemonState{}, err
	}

	state.Status = DaemonStatus(status)
	state.UpdatedAt = parsedUpdatedAt

	state.StartedAt, err = parseOptionalDaemonTime(startedAt)
	if err != nil {
		return DaemonState{}, err
	}

	state.HeartbeatAt, err = parseOptionalDaemonTime(heartbeatAt)
	if err != nil {
		return DaemonState{}, err
	}

	state.StopRequestedAt, err = parseOptionalDaemonTime(stopRequestedAt)
	if err != nil {
		return DaemonState{}, err
	}

	return state, nil
}

func upsertDaemonStateSQL() string {
	return `INSERT INTO daemon_state (
    id, status, pid, started_at, heartbeat_at, stop_requested_at,
    failure_reason, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    status = excluded.status,
    pid = excluded.pid,
    started_at = excluded.started_at,
    heartbeat_at = excluded.heartbeat_at,
    stop_requested_at = excluded.stop_requested_at,
    failure_reason = excluded.failure_reason,
    updated_at = excluded.updated_at`
}

func selectDaemonStateSQL() string {
	return `SELECT id, status, pid, started_at, heartbeat_at, stop_requested_at,
    failure_reason, updated_at
FROM daemon_state`
}

func formatOptionalDaemonTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return database.FormatTime(*value)
}

func parseOptionalDaemonTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := database.ParseTime(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
