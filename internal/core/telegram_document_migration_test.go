package core

import (
	"encoding/json"
	"testing"
	"time"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// Upgrade retains both legacy Markdown and version-one text presentations with
// their original retry/format meaning; neither acquires an unrelated document.
func TestTelegramPresentationMigrationDocuments(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	assets := telegramRecoveryMigrationAssets(t)
	delete(assets, "assets/migration/sqlite/20261021000028-add_telegram_notification_presentation.sql")
	delete(assets, "assets/migration/sqlite/20261021000027-add_telegram_action_presentation.sql")
	delete(assets, "assets/migration/sqlite/20261021000026-add_telegram_document_snapshots.sql")

	store, err := NewTelegramReceiverStore(t.Context(), env, assets)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	config := telegramReceiverTestConfig()
	if err := store.BindReceiver(t.Context(), 12345, config, now); err != nil {
		t.Fatal(err)
	}

	updates := []TelegramReceivedUpdate{telegramRetainedTestUpdate(t, 1, "/help"), telegramRetainedTestUpdate(t, 2, "/status")}
	if err := store.RetainBatch(t.Context(), 12345, updates, now); err != nil {
		t.Fatal(err)
	}

	if err := store.QueueFormattedReadResponse(t.Context(), updates[0], "*Commands*", "Commands", config.Digest(), now); err != nil {
		t.Fatal(err)
	}

	if err := store.QueueReadResponse(t.Context(), updates[1], "Old status", config.Digest(), now); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(tgmessage.Render(tgmessage.Blocks{InitialText: "Old status"}))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.db.ExecContext(t.Context(), "UPDATE telegram_outbox SET presentation=? WHERE event_key='update:2'", string(payload)); err != nil {
		t.Fatal(err)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	checkTelegramDocumentMigration(t, openTelegramReceiverTestStore(t, env), now)
}

func checkTelegramDocumentMigration(t *testing.T, store *TelegramReceiverStore, now time.Time) {
	t.Helper()

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 2 {
		t.Fatal("upgrade lost old presentations", due, err)
	}

	for _, delivery := range due {
		if delivery.DocumentID != "" || delivery.Attempts != 0 {
			t.Fatal("upgrade invented a document or reset attempts", delivery)
		}

		checkTelegramDocumentMigrationRow(t, delivery)
	}
}

func checkTelegramDocumentMigrationRow(t *testing.T, delivery TelegramReceiverDelivery) {
	t.Helper()

	if delivery.Text == "*Commands*" && delivery.ParseMode != "Markdown" {
		t.Fatal("upgrade lost legacy formatting", delivery)
	}

	if delivery.Text == "Old status" && (delivery.Presentation == nil || delivery.Presentation.TemplateRevision != "text-v1") {
		t.Fatal("upgrade lost the original text manifest", delivery)
	}
}
