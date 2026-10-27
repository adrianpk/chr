package telegram

import (
	"encoding/json"
	"testing"
	"time"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// Real SQLite proves atomic acknowledgement and immutable source replay/reopen.
func TestTelegramPresentationFormatSnapshot(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	now := time.Now()

	update := telegramRetainedTestUpdate(t, 1, "show capture")
	if err := store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	original := tgmessage.Blocks{ContentKind: "markdown", Content: "# Original\n```literal```"}
	for _, message := range []tgmessage.Blocks{original, {Content: "replacement"}} {
		if err := store.QueuePresentationResponse(t.Context(), update, message, telegramReceiverTestConfig().Digest(), now); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 || due[0].Presentation == nil || due[0].Text != original.Content {
		t.Fatal("snapshot changed or duplicated", due, err)
	}

	checkTelegramPresentationFallback(t, store, due[0].ID, original.Content, now)
}

func checkTelegramPresentationFallback(t *testing.T, store *ReceiverStore, id, original string, now time.Time) {
	t.Helper()

	selectTelegramHelpFallback(t, store, id, now)

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 || !due[0].PlainFallback || due[0].PlainText != original || due[0].Attempts != 1 {
		t.Fatal("fallback changed content or attempt allowance", due, err)
	}

	payload, err := json.Marshal(tgmessage.Presentation{Version: 2})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.db.ExecContext(t.Context(), "UPDATE telegram_outbox SET presentation=? WHERE id=?", string(payload), id); err != nil {
		t.Fatal(err)
	}

	if _, err := store.DueDeliveries(t.Context(), 12345, now); err == nil {
		t.Fatal("future presentation accepted")
	}
}
