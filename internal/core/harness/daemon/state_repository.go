package daemon

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

const stateID = "daemon"

// ErrStateNotFound identifies a package boundary error.
var ErrStateNotFound = errors.New("daemon state not found")

// Repository defines a package boundary interface.
type Repository interface {
	Get(ctx context.Context) (State, error)
	Upsert(ctx context.Context, state State) error
	UpdateHeartbeat(ctx context.Context, heartbeatAt time.Time) error
	RequestStop(ctx context.Context, requestedAt time.Time) error
	Clear(ctx context.Context) error
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
		return nil, fmt.Errorf("open daemon state database: %w", err)
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
		return nil, fmt.Errorf("migrate daemon state database: %w", err)
	}

	return &sqliteRepository{db: db}, nil
}

// Get executes the corresponding package operation.
func (r *sqliteRepository) Get(ctx context.Context) (State, error) {
	row := r.db.QueryRowContext(ctx, selectStateSQL()+" WHERE id = ?", stateID)

	state, err := scanState(row)
	if errors.Is(err, sql.ErrNoRows) {
		return State{}, ErrStateNotFound
	}

	if err != nil {
		return State{}, err
	}

	return state, nil
}

// Upsert executes the corresponding package operation.
func (r *sqliteRepository) Upsert(ctx context.Context, state State) error {
	err := ValidateState(state)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, upsertStateSQL(),
		stateID,
		string(state.Status),
		state.PID,
		formatOptionalTime(state.StartedAt),
		formatOptionalTime(state.HeartbeatAt),
		formatOptionalTime(state.StopRequestedAt),
		state.FailureReason,
		database.FormatTime(state.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert daemon state: %w", err)
	}

	return nil
}

// UpdateHeartbeat executes the corresponding package operation.
func (r *sqliteRepository) UpdateHeartbeat(ctx context.Context, heartbeatAt time.Time) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE daemon_state SET heartbeat_at = ?, updated_at = ? WHERE id = ?`,
		database.FormatTime(heartbeatAt),
		database.FormatTime(heartbeatAt),
		stateID,
	)
	if err != nil {
		return fmt.Errorf("update daemon heartbeat: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect updated daemon heartbeat rows: %w", err)
	}

	if affected == 0 {
		return ErrStateNotFound
	}

	return nil
}

// RequestStop executes the corresponding package operation.
func (r *sqliteRepository) RequestStop(ctx context.Context, requestedAt time.Time) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE daemon_state
SET status = ?, stop_requested_at = ?, updated_at = ?
WHERE id = ? AND status IN (?, ?, ?)`,
		string(Stopping),
		database.FormatTime(requestedAt),
		database.FormatTime(requestedAt),
		stateID,
		string(Starting),
		string(Running),
		string(Stopping),
	)
	if err != nil {
		return fmt.Errorf("request daemon stop: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect daemon stop request rows: %w", err)
	}

	if affected == 0 {
		return ErrStateNotFound
	}

	return nil
}

// Clear executes the corresponding package operation.
func (r *sqliteRepository) Clear(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM daemon_state WHERE id = ?", stateID)
	if err != nil {
		return fmt.Errorf("clear daemon state: %w", err)
	}

	return nil
}

// Close executes the corresponding package operation.
func (r *sqliteRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close daemon state repository database: %w", err)
	}

	return nil
}

type stateScanner interface {
	Scan(dest ...any) error
}

func scanState(scanner stateScanner) (State, error) {
	var state State

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
		return State{}, fmt.Errorf("scan daemon state: %w", err)
	}

	if id != stateID {
		return State{}, fmt.Errorf("unexpected daemon state id: %s", id)
	}

	parsedUpdatedAt, err := database.ParseTime(updatedAt)
	if err != nil {
		return State{}, err
	}

	state.Status = Status(status)
	state.UpdatedAt = parsedUpdatedAt

	state.StartedAt, err = parseOptionalTime(startedAt)
	if err != nil {
		return State{}, err
	}

	state.HeartbeatAt, err = parseOptionalTime(heartbeatAt)
	if err != nil {
		return State{}, err
	}

	state.StopRequestedAt, err = parseOptionalTime(stopRequestedAt)
	if err != nil {
		return State{}, err
	}

	return state, nil
}

func upsertStateSQL() string {
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

func selectStateSQL() string {
	return `SELECT id, status, pid, started_at, heartbeat_at, stop_requested_at,
    failure_reason, updated_at
FROM daemon_state`
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
