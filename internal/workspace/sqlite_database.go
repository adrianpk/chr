package workspace

import (
	"context"
	"database/sql"
	"fmt"
)

const sqliteBusyTimeoutMilliseconds = 5000

func configureSQLiteDatabase(ctx context.Context, db *sql.DB) error {
	db.SetMaxOpenConns(1)

	_, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMilliseconds))
	if err != nil {
		return fmt.Errorf("configure sqlite busy timeout: %w", err)
	}

	return nil
}
