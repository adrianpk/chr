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

type TicketKeySequenceRepository interface {
	ReserveNextTicketKey(ctx context.Context, opts ReserveNextTicketKeyOptions) (int, error)
	Close() error
}

type ReserveNextTicketKeyOptions struct {
	ProjectID string
	Prefix    string
	Seed      int
	UpdatedAt time.Time
}

type sqliteTicketKeySequenceRepository struct {
	db *sql.DB
}

func NewTicketKeySequenceRepository(
	ctx context.Context,
	env UserEnvironment,
	assetsFS fs.FS,
) (TicketKeySequenceRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open ticket key sequence database: %w", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := NewTicketKeySequenceRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return repository, nil
}

func NewTicketKeySequenceRepositoryFromDB(
	ctx context.Context,
	db *sql.DB,
	assetsFS fs.FS,
) (TicketKeySequenceRepository, error) {
	migrator := NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate ticket key sequence database: %w", err)
	}

	return &sqliteTicketKeySequenceRepository{db: db}, nil
}

func (r *sqliteTicketKeySequenceRepository) ReserveNextTicketKey(
	ctx context.Context,
	opts ReserveNextTicketKeyOptions,
) (int, error) {
	if opts.ProjectID == "" {
		return 0, fmt.Errorf("project id is required")
	}

	if opts.Prefix == "" {
		return 0, fmt.Errorf("ticket key prefix is required")
	}

	if opts.Seed < 1 {
		return 0, fmt.Errorf("ticket key seed must be positive")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin ticket key sequence transaction: %w", err)
	}

	nextNumber, err := reserveNextTicketKey(ctx, tx, opts)
	if err != nil {
		return 0, rollbackWithError(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("commit ticket key sequence transaction: %w", err)
	}

	return nextNumber, nil
}

func (r *sqliteTicketKeySequenceRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close ticket key sequence database: %w", err)
	}

	return nil
}

func reserveNextTicketKey(ctx context.Context, tx *sql.Tx, opts ReserveNextTicketKeyOptions) (int, error) {
	var nextNumber int

	err := tx.QueryRowContext(
		ctx,
		"SELECT next_number FROM ticket_key_sequences WHERE project_id = ? AND prefix = ?",
		opts.ProjectID,
		opts.Prefix,
	).Scan(&nextNumber)
	if err == nil {
		return updateExistingTicketKeySequence(ctx, tx, opts, nextNumber)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("read ticket key sequence: %w", err)
	}

	return createTicketKeySequence(ctx, tx, opts)
}

func updateExistingTicketKeySequence(
	ctx context.Context,
	tx *sql.Tx,
	opts ReserveNextTicketKeyOptions,
	nextNumber int,
) (int, error) {
	reservedNumber := max(nextNumber, opts.Seed)

	_, err := tx.ExecContext(
		ctx,
		`UPDATE ticket_key_sequences
SET next_number = ?, updated_at = ?
WHERE project_id = ? AND prefix = ?`,
		reservedNumber+1,
		formatProjectTime(opts.UpdatedAt),
		opts.ProjectID,
		opts.Prefix,
	)
	if err != nil {
		return 0, fmt.Errorf("update ticket key sequence: %w", err)
	}

	return reservedNumber, nil
}

func createTicketKeySequence(ctx context.Context, tx *sql.Tx, opts ReserveNextTicketKeyOptions) (int, error) {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO ticket_key_sequences (project_id, prefix, next_number, updated_at)
VALUES (?, ?, ?, ?)`,
		opts.ProjectID,
		opts.Prefix,
		opts.Seed+1,
		formatProjectTime(opts.UpdatedAt),
	)
	if err != nil {
		return 0, fmt.Errorf("create ticket key sequence: %w", err)
	}

	return opts.Seed, nil
}
