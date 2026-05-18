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

const userDatabaseName = "chc.db"

type ProjectRegistry interface {
	AddProject(ctx context.Context, project Project) error
	ListProjects(ctx context.Context) ([]Project, error)
	GetProjectByName(ctx context.Context, name string) (Project, error)
	Close() error
}

type ProjectRegistryOption func(*sqliteProjectRegistry)

type sqliteProjectRegistry struct {
	db *sql.DB
}

func NewProjectRegistry(ctx context.Context, env UserEnvironment, assetsFS fs.FS) (ProjectRegistry, error) {
	if err := os.MkdirAll(env.DataDir(), 0o755); err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open project registry database: %w", err)
	}

	db.SetMaxOpenConns(1)

	migrator := NewMigrator(assetsFS)
	if err := migrator.Migrate(ctx, db); err != nil {
		return nil, closeDBWithError(db, fmt.Errorf("migrate project registry database: %w", err))
	}

	return &sqliteProjectRegistry{db: db}, nil
}

func NewProjectRegistryFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (ProjectRegistry, error) {
	migrator := NewMigrator(assetsFS)
	if err := migrator.Migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("migrate project registry database: %w", err)
	}

	return &sqliteProjectRegistry{db: db}, nil
}

func (r *sqliteProjectRegistry) AddProject(ctx context.Context, project Project) error {
	if err := validateProject(project); err != nil {
		return err
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO projects (
    id,
    name,
    path,
    enabled,
    registered_at,
    last_indexed_at
) VALUES (?, ?, ?, ?, ?, ?)`,
		project.ID,
		project.Name,
		project.Path,
		boolToInt(project.Enabled),
		formatProjectTime(project.RegisteredAt),
		formatOptionalProjectTime(project.LastIndexedAt),
	)
	if err != nil {
		return fmt.Errorf("add project: %w", err)
	}

	return nil
}

func (r *sqliteProjectRegistry) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, path, enabled, registered_at, last_indexed_at
FROM projects
ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked through rows.Err below.
		_ = rows.Close()
	}()

	projects := []Project{}

	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list projects rows: %w", err)
	}

	return projects, nil
}

func (r *sqliteProjectRegistry) GetProjectByName(ctx context.Context, name string) (Project, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, path, enabled, registered_at, last_indexed_at
FROM projects
WHERE name = ?`,
		name,
	)

	project, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, fmt.Errorf("project not found: %s", name)
	}

	if err != nil {
		return Project{}, err
	}

	return project, nil
}

func (r *sqliteProjectRegistry) Close() error {
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("close project registry database: %w", err)
	}

	return nil
}

type projectScanner interface {
	Scan(dest ...any) error
}

func scanProject(scanner projectScanner) (Project, error) {
	var project Project

	var enabled int

	var registeredAt string

	var lastIndexedAt sql.NullString

	err := scanner.Scan(
		&project.ID,
		&project.Name,
		&project.Path,
		&enabled,
		&registeredAt,
		&lastIndexedAt,
	)
	if err != nil {
		return Project{}, fmt.Errorf("scan project: %w", err)
	}

	parsedRegisteredAt, err := parseProjectTime(registeredAt)
	if err != nil {
		return Project{}, err
	}

	project.Enabled = enabled != 0
	project.RegisteredAt = parsedRegisteredAt

	if lastIndexedAt.Valid {
		parsedLastIndexedAt, err := parseProjectTime(lastIndexedAt.String)
		if err != nil {
			return Project{}, err
		}

		project.LastIndexedAt = &parsedLastIndexedAt
	}

	return project, nil
}

func validateProject(project Project) error {
	if project.ID == "" {
		return fmt.Errorf("project id is required")
	}

	if project.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if project.Path == "" {
		return fmt.Errorf("project path is required")
	}

	if project.RegisteredAt.IsZero() {
		return fmt.Errorf("project registered timestamp is required")
	}

	return nil
}

func formatProjectTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

func formatOptionalProjectTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return formatProjectTime(*value)
}

func parseProjectTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse project timestamp: %w", err)
	}

	return parsed, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}

func closeDBWithError(db *sql.DB, err error) error {
	if closeErr := db.Close(); closeErr != nil {
		return fmt.Errorf("%w; close database failed: %w", err, closeErr)
	}

	return err
}
