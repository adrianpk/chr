package telegram

import (
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// A pending plain reply created before format support remains unchanged after migration.
func TestTelegramHelpLegacyMigration(t *testing.T) {
	env := telegramReceiverTestEnv(t)

	store, err := NewReceiverStore(t.Context(), env, telegramPreHelpAssets(t))
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	config := telegramReceiverTestConfig()
	if err := store.BindReceiver(t.Context(), 12345, config, now); err != nil {
		t.Fatal(err)
	}
	// Seed the old representation directly: new queue code requires the new columns.
	_, err = store.db.ExecContext(t.Context(), `INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,state,created_at,updated_at)
 VALUES('legacy-help',12345,100001,0,100001,'update:1',?,'result','legacy plain help','pending',?,?)`, config.Digest(), database.FormatTime(now), database.FormatTime(now))
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 {
		t.Fatal("format migration changed an old pending plain reply", due, err)
	}

	assertTelegramLegacyReply(t, due[0])
}

func assertTelegramLegacyReply(t *testing.T, delivery ReceiverDelivery) {
	t.Helper()

	if delivery.Text != "legacy plain help" || delivery.ParseMode != "" || delivery.PlainText != "" || delivery.PlainFallback {
		t.Fatal("format migration changed an old pending plain reply", delivery)
	}
}

func telegramPreHelpAssets(t *testing.T) fstest.MapFS {
	t.Helper()

	assets := os.DirFS("../../../cmd/chr")

	paths, err := fs.Glob(assets, "assets/migration/sqlite/*.sql")
	if err != nil {
		t.Fatal(err)
	}

	previous := fstest.MapFS{}

	for _, path := range paths {
		if strings.Contains(path, "20261017000024-") {
			continue
		}

		data, err := fs.ReadFile(assets, path)
		if err != nil {
			t.Fatal(err)
		}

		previous[path] = &fstest.MapFile{Data: data}
	}

	return previous
}
