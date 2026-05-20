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

const daemonStateID = "daemon"

var ErrDaemonStateNotFound = errors.New("daemon state not found")

type DaemonStateRepository interface {
	GetDaemonState(ctx context.Context) (DaemonState, error)
	UpsertDaemonState(ctx context.Context, state DaemonState) error
	UpdateDaemonHeartbeat(ctx context.Context, heartbeatAt time.Time) error
	ClearDaemonState(ctx context.Context) error
	Close() error
}

type sqliteDaemonStateRepository struct {
	db *sql.DB
}

func NewDaemonStateRepository(ctx context.Context, env UserEnvironment, assetsFS fs.FS) (DaemonStateRepository, error) {
	if err := os.MkdirAll(env.DataDir(), 0o755); err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open daemon state database: %w", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := NewDaemonStateRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return repository, nil
}

func NewDaemonStateRepositoryFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (DaemonStateRepository, error) {
	migrator := NewMigrator(assetsFS)
	if err := migrator.Migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("migrate daemon state database: %w", err)
	}

	return &sqliteDaemonStateRepository{db: db}, nil
}

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

func (r *sqliteDaemonStateRepository) UpsertDaemonState(ctx context.Context, state DaemonState) error {
	if err := ValidateDaemonState(state); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx, upsertDaemonStateSQL(),
		daemonStateID,
		string(state.Status),
		state.PID,
		formatOptionalDaemonTime(state.StartedAt),
		formatOptionalDaemonTime(state.HeartbeatAt),
		formatOptionalDaemonTime(state.StopRequestedAt),
		state.FailureReason,
		formatProjectTime(state.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert daemon state: %w", err)
	}

	return nil
}

func (r *sqliteDaemonStateRepository) UpdateDaemonHeartbeat(ctx context.Context, heartbeatAt time.Time) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE daemon_state SET heartbeat_at = ?, updated_at = ? WHERE id = ?`,
		formatProjectTime(heartbeatAt),
		formatProjectTime(heartbeatAt),
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

func (r *sqliteDaemonStateRepository) ClearDaemonState(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM daemon_state WHERE id = ?", daemonStateID)
	if err != nil {
		return fmt.Errorf("clear daemon state: %w", err)
	}

	return nil
}

func (r *sqliteDaemonStateRepository) Close() error {
	if err := r.db.Close(); err != nil {
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

	parsedUpdatedAt, err := parseProjectTime(updatedAt)
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

	return formatProjectTime(*value)
}

func parseOptionalDaemonTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := parseProjectTime(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
