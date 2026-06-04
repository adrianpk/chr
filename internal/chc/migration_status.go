package chc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite"
)

// MigrationStatus describes a package boundary value.
type MigrationStatus struct {
	DatabasePath      string
	DatabaseExists    bool
	SchemaTableExists bool
	Available         []Migration
	Applied           []AppliedMigration
	Pending           []Migration
	Unknown           []AppliedMigration
}

// AppliedMigration describes a package boundary value.
type AppliedMigration struct {
	Version   string
	Name      string
	AppliedAt string
}

// MigrationStatusChecker describes a package boundary value.
type MigrationStatusChecker struct {
	assetsFS fs.FS
}

// NewMigrationStatusChecker constructs the corresponding package service or value.
func NewMigrationStatusChecker(assetsFS fs.FS) MigrationStatusChecker {
	return MigrationStatusChecker{assetsFS: assetsFS}
}

// UserDatabasePath executes the corresponding package operation.
func UserDatabasePath(env UserEnvironment) string {
	return filepath.Join(env.DataDir(), userDatabaseName)
}

// Check executes the corresponding package operation.
func (c MigrationStatusChecker) Check(
	ctx context.Context,
	env UserEnvironment,
) (MigrationStatus, error) {
	err := ctx.Err()
	if err != nil {
		return MigrationStatus{}, err
	}

	available, err := ReadSQLiteMigrations(c.assetsFS)
	if err != nil {
		return MigrationStatus{}, err
	}

	status := MigrationStatus{
		DatabasePath: UserDatabasePath(env),
		Available:    available,
	}

	_, err = os.Stat(status.DatabasePath)
	if err != nil {
		if os.IsNotExist(err) {
			status.Pending = append([]Migration(nil), available...)

			return status, nil
		}

		return MigrationStatus{}, fmt.Errorf("inspect user database: %w", err)
	}

	status.DatabaseExists = true

	db, err := sql.Open("sqlite", status.DatabasePath)
	if err != nil {
		return MigrationStatus{}, fmt.Errorf("open user database: %w", err)
	}

	defer func() {
		//nolint:errcheck // Read-only status result is returned above.
		_ = db.Close()
	}()

	status.SchemaTableExists, err = schemaMigrationsTableExists(ctx, db)
	if err != nil {
		return MigrationStatus{}, err
	}

	if !status.SchemaTableExists {
		status.Pending = append([]Migration(nil), available...)

		return status, nil
	}

	status.Applied, err = listAppliedMigrations(ctx, db)
	if err != nil {
		return MigrationStatus{}, err
	}

	status.Pending = pendingMigrations(available, status.Applied)
	status.Unknown = unknownAppliedMigrations(available, status.Applied)

	return status, nil
}

func schemaMigrationsTableExists(ctx context.Context, db *sql.DB) (bool, error) {
	var name string

	err := db.QueryRowContext(
		ctx,
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'",
	).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("inspect schema migrations table: %w", err)
	}

	return true, nil
}

func listAppliedMigrations(ctx context.Context, db *sql.DB) ([]AppliedMigration, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT version, name, applied_at
FROM schema_migrations
ORDER BY version`,
	)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked below.
		_ = rows.Close()
	}()

	applied := []AppliedMigration{}

	for rows.Next() {
		var migration AppliedMigration

		err = rows.Scan(&migration.Version, &migration.Name, &migration.AppliedAt)
		if err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}

		applied = append(applied, migration)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("list applied migration rows: %w", err)
	}

	return applied, nil
}

func pendingMigrations(available []Migration, applied []AppliedMigration) []Migration {
	appliedVersions := appliedMigrationVersions(applied)
	pending := []Migration{}

	for _, migration := range available {
		if _, ok := appliedVersions[migration.Version]; !ok {
			pending = append(pending, migration)
		}
	}

	return pending
}

func unknownAppliedMigrations(available []Migration, applied []AppliedMigration) []AppliedMigration {
	availableVersions := availableMigrationVersions(available)
	unknown := []AppliedMigration{}

	for _, migration := range applied {
		if _, ok := availableVersions[migration.Version]; !ok {
			unknown = append(unknown, migration)
		}
	}

	sort.SliceStable(unknown, func(i, j int) bool {
		return unknown[i].Version < unknown[j].Version
	})

	return unknown
}

func availableMigrationVersions(migrations []Migration) map[string]struct{} {
	versions := make(map[string]struct{}, len(migrations))

	for _, migration := range migrations {
		versions[migration.Version] = struct{}{}
	}

	return versions
}

func appliedMigrationVersions(migrations []AppliedMigration) map[string]struct{} {
	versions := make(map[string]struct{}, len(migrations))

	for _, migration := range migrations {
		versions[migration.Version] = struct{}{}
	}

	return versions
}
