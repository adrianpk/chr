package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// UserFileName is the existing user database filename.
const UserFileName = "core.db"

// FormatTime retains UTC RFC3339 storage spelling.
func FormatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

// OptionalTime stores an absent timestamp as SQL NULL.
func OptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return FormatTime(*value)
}

// ParseTime reads the existing timestamp format and error context.
func ParseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse project timestamp: %w", err)
	}

	return parsed, nil
}

// BoolInt encodes boolean SQL values as zero or one.
func BoolInt(value bool) int {
	if value {
		return 1
	}

	return 0
}

// CloseOnError closes a failed connection and preserves both errors.
func CloseOnError(db *sql.DB, err error) error {
	closeErr := db.Close()
	if closeErr != nil {
		return fmt.Errorf("%w; close database failed: %w", err, closeErr)
	}

	return err
}

// TransactionFailure retains cancellation when SQLite reports a completed transaction.
// This owner alone commits/rolls back the transaction. database/sql can win
// that race after cancellation and return the bare ErrTxDone instead of the
// context error. Wrapped/joined checkpoint failures must remain failures.
func TransactionFailure(ctx context.Context, err error) error {
	if ctx.Err() == nil {
		return err
	}

	if _, wrapped := err.(interface{ Unwrap() error }); wrapped {
		return err
	}

	if _, joined := err.(interface{ Unwrap() []error }); joined {
		return err
	}

	if errors.Is(err, sql.ErrTxDone) {
		return ctx.Err()
	}

	return err
}

// DataDirectory supplies the existing database location.
type DataDirectory interface{ DataDir() string }
