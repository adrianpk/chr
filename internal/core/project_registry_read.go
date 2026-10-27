package core

import (
	"context"
	"errors"

	"forge.adrianpk.com/chr/internal/core/database"
)

// ReadRegisteredProjects reads explicit identities without migrating or creating a registry.
func ReadRegisteredProjects(ctx context.Context, env database.DataDirectory, ids []string) (projects []Project, err error) {
	if len(ids) == 0 || len(ids) > 50 {
		return nil, errors.New("one to 50 explicit project identities are required")
	}

	db, err := database.OpenReadOnly(ctx, env)
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
