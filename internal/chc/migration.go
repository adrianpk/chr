package chc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const sqliteMigrationDir = "assets/migration/sqlite"

const (
	migrateUpMarker   = "-- +migrate Up"
	migrateDownMarker = "-- +migrate Down"
)

type Migration struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

type Migrator interface {
	Migrate(ctx context.Context, db *sql.DB) error
}

type migrator struct {
	assetsFS fs.FS
	now      func() time.Time
}

type MigratorOption func(*migrator)

func NewMigrator(assetsFS fs.FS, opts ...MigratorOption) Migrator {
	migrator := &migrator{
		assetsFS: assetsFS,
		now:      time.Now,
	}

	for _, opt := range opts {
		opt(migrator)
	}

	return migrator
}

func WithMigratorClock(now func() time.Time) MigratorOption {
	return func(migrator *migrator) {
		if now != nil {
			migrator.now = now
		}
	}
}

func (m *migrator) Migrate(ctx context.Context, db *sql.DB) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := ensureSchemaMigrations(ctx, db); err != nil {
		return err
	}

	migrations, err := ReadSQLiteMigrations(m.assetsFS)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if err := ctx.Err(); err != nil {
			return err
		}

		applied, err := migrationApplied(ctx, db, migration.Version)
		if err != nil {
			return err
		}

		if applied {
			continue
		}

		if err := m.applyMigration(ctx, db, migration); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchemaMigrations(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
    version text PRIMARY KEY,
    name text NOT NULL,
    applied_at text NOT NULL
)`)
	if err != nil {
		return fmt.Errorf("ensure schema migrations table: %w", err)
	}

	return nil
}

func (m *migrator) applyMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
		return rollbackWithError(tx, fmt.Errorf("apply migration %s: %w", migration.Version, err))
	}

	appliedAt := m.now().UTC().Format(time.RFC3339)
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
		migration.Version,
		migration.Name,
		appliedAt,
	)
	if err != nil {
		return rollbackWithError(tx, fmt.Errorf("record migration %s: %w", migration.Version, err))
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.Version, err)
	}

	return nil
}

func ReadSQLiteMigrations(assetsFS fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(assetsFS, sqliteMigrationDir)
	if err != nil {
		return nil, fmt.Errorf("read sqlite migrations: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		migration, err := readMigration(assetsFS, entry.Name())
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func readMigration(assetsFS fs.FS, name string) (Migration, error) {
	version, migrationName, err := parseMigrationName(name)
	if err != nil {
		return Migration{}, err
	}

	path := filepath.ToSlash(filepath.Join(sqliteMigrationDir, name))
	content, err := fs.ReadFile(assetsFS, path)
	if err != nil {
		return Migration{}, fmt.Errorf("read migration %s: %w", name, err)
	}

	upSQL, downSQL, err := splitMigrationSQL(string(content))
	if err != nil {
		return Migration{}, fmt.Errorf("parse migration %s: %w", name, err)
	}

	return Migration{
		Version: version,
		Name:    migrationName,
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}, nil
}

func parseMigrationName(filename string) (string, string, error) {
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	version, name, ok := strings.Cut(filename, "-")
	if !ok || version == "" || name == "" {
		return "", "", fmt.Errorf("invalid migration filename: %s", filename)
	}

	return version, name, nil
}

func splitMigrationSQL(content string) (string, string, error) {
	upStart := strings.Index(content, migrateUpMarker)
	downStart := strings.Index(content, migrateDownMarker)

	if upStart < 0 {
		return "", "", errors.New("missing up section")
	}

	if downStart < 0 {
		return "", "", errors.New("missing down section")
	}

	if downStart < upStart {
		return "", "", errors.New("down section appears before up section")
	}

	upSQL := strings.TrimSpace(content[upStart+len(migrateUpMarker) : downStart])
	downSQL := strings.TrimSpace(content[downStart+len(migrateDownMarker):])

	if upSQL == "" {
		return "", "", errors.New("empty up section")
	}

	if downSQL == "" {
		return "", "", errors.New("empty down section")
	}

	return upSQL, downSQL, nil
}

func migrationApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var found string
	err := db.QueryRowContext(
		ctx,
		"SELECT version FROM schema_migrations WHERE version = ?",
		version,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("inspect migration %s: %w", version, err)
	}

	return true, nil
}

func rollbackWithError(tx *sql.Tx, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return fmt.Errorf("%w; rollback failed: %w", err, rollbackErr)
	}

	return err
}
