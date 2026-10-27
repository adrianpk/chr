package database

import (
	"context"
	"database/sql"
	"fmt"
)

const sqliteBusyTimeoutMilliseconds = 5000

// Configure applies the existing connection limit and busy timeout.
func Configure(ctx context.Context, db *sql.DB) error {
	db.SetMaxOpenConns(1)

	_, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMilliseconds))
	if err != nil {
		return fmt.Errorf("configure sqlite busy timeout: %w", err)
	}

	return nil
}
