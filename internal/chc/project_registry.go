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

// ErrProjectNotFound identifies a package boundary error.
var ErrProjectNotFound = errors.New("project not found")

// ProjectRegistry defines a package boundary interface.
type ProjectRegistry interface {
	AddProject(ctx context.Context, project Project) error
	ListProjects(ctx context.Context) ([]Project, error)
	GetProjectByID(ctx context.Context, id string) (Project, error)
	GetProjectByName(ctx context.Context, name string) (Project, error)
	GetProjectByPath(ctx context.Context, path string) (Project, error)
	DisableProject(ctx context.Context, name string) (Project, error)
	UpdateProjectPath(ctx context.Context, name, path string) (Project, error)
	Close() error
}

// ProjectRegistryOption defines a package boundary type.
type ProjectRegistryOption func(*sqliteProjectRegistry)

type sqliteProjectRegistry struct {
	db *sql.DB
}

// NewProjectRegistry constructs the corresponding package service or value.
func NewProjectRegistry(ctx context.Context, env UserEnvironment, assetsFS fs.FS) (ProjectRegistry, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open project registry database: %w", err)
	}

	err = configureSQLiteDatabase(ctx, db)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	migrator := NewMigrator(assetsFS)

	err = migrator.Migrate(ctx, db)
	if err != nil {
		return nil, closeDBWithError(db, fmt.Errorf("migrate project registry database: %w", err))
	}

	return &sqliteProjectRegistry{db: db}, nil
}

// NewProjectRegistryFromDB constructs the corresponding package service or value.
func NewProjectRegistryFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (ProjectRegistry, error) {
	migrator := NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate project registry database: %w", err)
	}

	return &sqliteProjectRegistry{db: db}, nil
}

// AddProject executes the corresponding package operation.
func (r *sqliteProjectRegistry) AddProject(ctx context.Context, project Project) error {
	err := validateProject(project)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
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

// ListProjects lists the requested data.
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

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("list projects rows: %w", err)
	}

	return projects, nil
}

// GetProjectByName executes the corresponding package operation.
func (r *sqliteProjectRegistry) GetProjectByName(ctx context.Context, name string) (Project, error) {
	return r.getProject(ctx, "name", name)
}

// GetProjectByID executes the corresponding package operation.
func (r *sqliteProjectRegistry) GetProjectByID(ctx context.Context, id string) (Project, error) {
	return r.getProject(ctx, "id", id)
}

// GetProjectByPath executes the corresponding package operation.
func (r *sqliteProjectRegistry) GetProjectByPath(ctx context.Context, path string) (Project, error) {
	return r.getProject(ctx, "path", path)
}

// DisableProject executes the corresponding package operation.
func (r *sqliteProjectRegistry) DisableProject(ctx context.Context, name string) (Project, error) {
	project, err := r.GetProjectByName(ctx, name)
	if err != nil {
		return Project{}, err
	}

	_, err = r.db.ExecContext(ctx, "UPDATE projects SET enabled = 0 WHERE id = ?", project.ID)
	if err != nil {
		return Project{}, fmt.Errorf("disable project: %w", err)
	}

	project.Enabled = false

	return project, nil
}

// UpdateProjectPath executes the corresponding package operation.
func (r *sqliteProjectRegistry) UpdateProjectPath(ctx context.Context, name, path string) (Project, error) {
	if path == "" {
		return Project{}, fmt.Errorf("project path is required")
	}

	project, err := r.GetProjectByName(ctx, name)
	if err != nil {
		return Project{}, err
	}

	_, err = r.db.ExecContext(ctx, "UPDATE projects SET path = ? WHERE id = ?", path, project.ID)
	if err != nil {
		return Project{}, fmt.Errorf("update project path: %w", err)
	}

	project.Path = path

	return project, nil
}

func (r *sqliteProjectRegistry) getProject(ctx context.Context, field, value string) (Project, error) {
	column, err := projectLookupColumn(field)
	if err != nil {
		return Project{}, err
	}

	row := r.db.QueryRowContext(
		ctx,
		fmt.Sprintf(`SELECT id, name, path, enabled, registered_at, last_indexed_at
FROM projects
WHERE %s = ?`, column),
		value,
	)

	project, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, fmt.Errorf("%w: %s=%s", ErrProjectNotFound, field, value)
	}

	if err != nil {
		return Project{}, err
	}

	return project, nil
}

func projectLookupColumn(field string) (string, error) {
	switch field {
	case "id", "name", "path":
		return field, nil
	default:
		return "", fmt.Errorf("unsupported project lookup field: %s", field)
	}
}

// Close executes the corresponding package operation.
func (r *sqliteProjectRegistry) Close() error {
	err := r.db.Close()
	if err != nil {
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
	closeErr := db.Close()
	if closeErr != nil {
		return fmt.Errorf("%w; close database failed: %w", err, closeErr)
	}

	return err
}
