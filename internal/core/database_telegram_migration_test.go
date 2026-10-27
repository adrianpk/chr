package core

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramReceiverMigration preserves the registry and rolls back an interrupted additive migration.
func TestTelegramReceiverMigration(t *testing.T) {
	env := telegramReceiverTestEnv(t)

	registry, err := projectcore.NewRegistry(t.Context(), env, testMigrationFS())
	if err != nil {
		t.Fatal(err)
	}

	project := projectcore.Project{ID: "example-project", Name: "Example", Path: t.TempDir(), Enabled: true, RegisteredAt: time.Now()}
	if err := registry.AddProject(t.Context(), project); err != nil {
		t.Fatal(err)
	}

	assets := testMigrationFS()

	assets["assets/migration/sqlite/20261016000016-add_telegram_receiver.sql"] = &fstest.MapFile{Data: []byte("-- +migrate Up\nCREATE TABLE telegram_receivers(bot_id INTEGER); invalid SQL;\n-- +migrate Down\nDROP TABLE telegram_receivers;")}
	if store, err := telegramcore.NewTelegramReceiverStore(t.Context(), env, assets); err == nil {
		t.Fatal("failed migration accepted", store.Close())
	}

	inspection, err := database.OpenReadOnly(t.Context(), env)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := inspection.Close(); err != nil {
			t.Error(err)
		}
	}()

	var tableCount int
	if err := inspection.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM sqlite_master WHERE name='telegram_receivers'").Scan(&tableCount); err != nil || tableCount != 0 {
		t.Fatal("partial migration remained", tableCount, err)
	}

	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}

	store := openTelegramReceiverTestStore(t, env)

	checkTelegramRegistryMigration(t, store, env, project)
}

func checkTelegramRegistryMigration(t *testing.T, store *telegramFixtureStore, env UserEnvironment, project projectcore.Project) {
	t.Helper()

	var name string
	if err := store.db.QueryRowContext(t.Context(), "SELECT name FROM projects WHERE id=?", project.ID).Scan(&name); err != nil || name != project.Name {
		t.Fatal("migration lost registry", err)
	}

	if _, err := store.db.ExecContext(t.Context(), "INSERT INTO schema_migrations VALUES('99999999999999','future','2026-01-01T00:00:00Z')"); err != nil {
		t.Fatal(err)
	}

	if newer, err := telegramcore.NewTelegramReceiverStore(t.Context(), env, os.DirFS("../../cmd/chr")); err == nil || !strings.Contains(err.Error(), "unknown") {
		if newer != nil {
			t.Error(newer.Close())
		}

		t.Fatal("future schema accepted", err)
	}
}
