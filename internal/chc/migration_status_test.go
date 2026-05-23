package chc

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrationStatusReportsPendingWhenDatabaseIsMissing(t *testing.T) {
	env := testMigrationStatusEnvironment(t)

	status, err := NewMigrationStatusChecker(testMigrationFS()).Check(context.Background(), env)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if status.DatabaseExists {
		t.Fatal("DatabaseExists = true, want false")
	}

	if len(status.Pending) != 7 {
		t.Fatalf("Pending len = %d, want 7", len(status.Pending))
	}
}

func TestMigrationStatusReportsCleanDatabase(t *testing.T) {
	env := testMigrationStatusEnvironment(t)
	db := openMigrationStatusDB(t, env)

	err := NewMigrator(testMigrationFS()).Migrate(context.Background(), db)
	if err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	status, err := NewMigrationStatusChecker(testMigrationFS()).Check(context.Background(), env)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if !status.DatabaseExists {
		t.Fatal("DatabaseExists = false, want true")
	}

	if len(status.Applied) != 7 {
		t.Fatalf("Applied len = %d, want 7", len(status.Applied))
	}

	if len(status.Pending) != 0 {
		t.Fatalf("Pending len = %d, want 0", len(status.Pending))
	}

	if len(status.Unknown) != 0 {
		t.Fatalf("Unknown len = %d, want 0", len(status.Unknown))
	}
}

func TestMigrationStatusReportsUnknownAppliedMigration(t *testing.T) {
	env := testMigrationStatusEnvironment(t)
	db := openMigrationStatusDB(t, env)

	err := ensureSchemaMigrations(context.Background(), db)
	if err != nil {
		t.Fatalf("ensureSchemaMigrations returned error: %v", err)
	}

	_, err = db.Exec(`INSERT INTO schema_migrations (version, name, applied_at)
VALUES ('99999999999999', 'future', '2026-05-21T12:00:00Z')`)
	if err != nil {
		t.Fatalf("insert unknown migration: %v", err)
	}

	status, err := NewMigrationStatusChecker(testMigrationFS()).Check(context.Background(), env)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if len(status.Unknown) != 1 {
		t.Fatalf("Unknown len = %d, want 1", len(status.Unknown))
	}

	if status.Unknown[0].Version != "99999999999999" {
		t.Fatalf("Unknown version = %q, want future version", status.Unknown[0].Version)
	}
}

func testMigrationStatusEnvironment(t *testing.T) UserEnvironment {
	t.Helper()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	return env
}

func openMigrationStatusDB(t *testing.T, env UserEnvironment) *sql.DB {
	t.Helper()

	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}

	t.Cleanup(func() {
		err = db.Close()
		if err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})

	return db
}
