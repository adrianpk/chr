package core

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"path/filepath"
)

func openUserDatabaseReadOnly(env UserEnvironment) (*sql.DB, error) {
	path, err := filepath.Abs(UserDatabasePath(env))
	if err != nil {
		return nil, err
	}

	location := url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}

	return sql.Open("sqlite", location.String())
}

// ReadRegisteredProjects reads explicit identities without migrating or creating a registry.
func ReadRegisteredProjects(ctx context.Context, env UserEnvironment, ids []string) (projects []Project, err error) {
	if len(ids) == 0 || len(ids) > 50 {
		return nil, errors.New("one to 50 explicit project identities are required")
	}

	db, err := openUserDatabaseReadOnly(env)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	registry := sqliteProjectRegistry{db: db}
	for _, id := range ids {
		project, readErr := registry.GetProjectByID(ctx, id)
		if errors.Is(readErr, ErrProjectNotFound) {
			continue
		}

		if readErr != nil {
			return nil, readErr
		}

		projects = append(projects, project)
	}

	return projects, nil
}
