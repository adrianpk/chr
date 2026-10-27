package telegram

import (
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// Existing plain and legacy-Markdown deliveries keep their exact text, fallback
// choice, send uncertainty and attempt counters after the additive migration.
func TestTelegramPresentationFormatMigration(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	assets := telegramRecoveryMigrationAssets(t)
	delete(assets, "assets/migration/sqlite/20261021000028-add_telegram_notification_presentation.sql")
	delete(assets, "assets/migration/sqlite/20261021000027-add_telegram_action_presentation.sql")
	delete(assets, "assets/migration/sqlite/20261021000025-add_telegram_text_presentation.sql")

	store, err := NewTelegramReceiverStore(t.Context(), env, assets)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	config := telegramReceiverTestConfig()
	if err := store.BindReceiver(t.Context(), 12345, config, now); err != nil {
		t.Fatal(err)
	}

	for _, item := range []struct {
		id, text, mode, plain string
		fallback              int
	}{
		{"old-plain", "Original plain", "", "", 0},
		{"old-help", "*Original help*", "Markdown", "Original help", 1},
	} {
		_, err := store.db.ExecContext(t.Context(), `INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,state,attempts,total_attempts,parse_mode,plain_text,plain_fallback,created_at,updated_at)
 VALUES(?,12345,100001,0,100001,?,?,'result',?,'uncertain',2,2,?,?,?,?,?)`, item.id, item.id, config.Digest(), item.text, item.mode, item.plain, item.fallback, database.FormatTime(now), database.FormatTime(now))
		if err != nil {
			t.Fatal(err)
		}
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	checkTelegramPresentationLegacy(t, openTelegramReceiverTestStore(t, env), now)
}

func checkTelegramPresentationLegacy(t *testing.T, store *TelegramReceiverStore, now time.Time) {
	t.Helper()

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 2 {
		t.Fatal("migration lost existing deliveries", due, err)
	}

	type legacyBody struct {
		text, mode, plain string
		fallback          bool
	}

	expected := map[string]legacyBody{
		"old-help":  {"*Original help*", "Markdown", "Original help", true},
		"old-plain": {"Original plain", "", "", false},
	}

	for _, delivery := range due {
		if delivery.Presentation != nil || delivery.State != "uncertain" || delivery.Attempts != 2 {
			t.Fatal("migration changed original delivery identity", delivery)
		}

		actual := legacyBody{delivery.Text, delivery.ParseMode, delivery.PlainText, delivery.PlainFallback}
		if actual != expected[delivery.ID] {
			t.Fatal("migration changed legacy text, format or fallback", delivery)
		}
	}
}
