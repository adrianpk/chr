package database

import (
	"context"
	"database/sql"
	"net/url"
	"path/filepath"
)

// OpenReadOnly inspects existing user data without migration or file creation.
func OpenReadOnly(ctx context.Context, env DataDirectory) (*sql.DB, error) {
	path, err := filepath.Abs(UserDatabasePath(env))
	if err != nil {
		return nil, err
	}

	location := url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}

	db, err := sql.Open("sqlite", location.String())
	if err != nil {
		return nil, err
	}

	if err := Configure(ctx, db); err != nil {
		return nil, CloseOnError(db, err)
	}

	return db, nil
}
