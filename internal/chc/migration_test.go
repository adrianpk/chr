package chc

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"
	"time"

	_ "modernc.org/sqlite"
)

func TestReadSQLiteMigrationsParsesAndSortsAssets(t *testing.T) {
	migrations, err := ReadSQLiteMigrations(fstest.MapFS{
		"assets/migration/sqlite/20260518000002-second.sql": {
			Data: []byte("-- +migrate Up\nCREATE TABLE second (id text);\n-- +migrate Down\nDROP TABLE second;"),
		},
		"assets/migration/sqlite/20260518000001-first.sql": {
			Data: []byte("-- +migrate Up\nCREATE TABLE first (id text);\n-- +migrate Down\nDROP TABLE first;"),
		},
	})
	if err != nil {
		t.Fatalf("ReadSQLiteMigrations returned error: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("len(migrations) = %d, want 2", len(migrations))
	}

	if migrations[0].Version != "20260518000001" {
		t.Fatalf("first version = %q, want 20260518000001", migrations[0].Version)
	}

	if migrations[1].Version != "20260518000002" {
		t.Fatalf("second version = %q, want 20260518000002", migrations[1].Version)
	}
}

func TestMigratorAppliesPendingMigrations(t *testing.T) {
	db := openTestSQLite(t)
	migrator := NewMigrator(testMigrationFS(), WithMigratorClock(func() time.Time {
		return time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	}))

	if err := migrator.Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	assertTableExists(t, db, "projects")
	assertMigrationApplied(t, db, "20260518000001", "create_projects", "2026-05-18T12:30:00Z")
}

func TestMigratorSkipsAppliedMigrations(t *testing.T) {
	db := openTestSQLite(t)
	migrator := NewMigrator(testMigrationFS())

	if err := migrator.Migrate(context.Background(), db); err != nil {
		t.Fatalf("first Migrate returned error: %v", err)
	}

	if err := migrator.Migrate(context.Background(), db); err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}

	var count int

	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", "20260518000001").Scan(&count)
	if err != nil {
		t.Fatalf("count migrations: %v", err)
	}

	if count != 1 {
		t.Fatalf("migration count = %d, want 1", count)
	}
}

func openTestSQLite(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close sqlite: %v", err)
		}
	})

	return db
}

func testMigrationFS() fstest.MapFS {
	return fstest.MapFS{
		"assets/migration/sqlite/20260518000001-create_projects.sql": {
			Data: []byte(`-- +migrate Up
CREATE TABLE projects (
    id text PRIMARY KEY,
    name text NOT NULL UNIQUE,
    path text NOT NULL UNIQUE,
    enabled integer NOT NULL,
    registered_at text NOT NULL,
    last_indexed_at text
);

-- +migrate Down
DROP TABLE IF EXISTS projects;
`),
		},
	}
}

func assertTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()

	var found string

	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?",
		name,
	).Scan(&found)
	if err != nil {
		t.Fatalf("table %s does not exist: %v", name, err)
	}
}

func assertMigrationApplied(t *testing.T, db *sql.DB, version, name, appliedAt string) {
	t.Helper()

	var gotName string

	var gotAppliedAt string

	err := db.QueryRow(
		"SELECT name, applied_at FROM schema_migrations WHERE version = ?",
		version,
	).Scan(&gotName, &gotAppliedAt)
	if err != nil {
		t.Fatalf("query migration: %v", err)
	}

	if gotName != name {
		t.Fatalf("migration name = %q, want %q", gotName, name)
	}

	if gotAppliedAt != appliedAt {
		t.Fatalf("applied_at = %q, want %q", gotAppliedAt, appliedAt)
	}
}
