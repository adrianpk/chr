package telegram

import (
	"strings"
	"testing"
	"time"
)

// Both renderings are frozen atomically, survive reopen and ignore source replay.
func TestTelegramHelpOutbox(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	now := time.Now().UTC()

	update := telegramRetainedTestUpdate(t, 1, "/help")
	if err := store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	for _, bodies := range [][2]string{{"*Commands*\n`/original`", "Commands\n/original"}, {"`/changed`", "/changed"}} {
		if err := store.QueueFormattedReadResponse(t.Context(), update, bodies[0], bodies[1], telegramReceiverTestConfig().Digest(), now); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)
	delivery := checkTelegramHelpSnapshot(t, store, now)
	selectTelegramHelpFallback(t, store, delivery.ID, now)

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)
	checkTelegramHelpFallback(t, store, delivery.ID, now)
}

func checkTelegramHelpSnapshot(t *testing.T, store *ReceiverStore, now time.Time) ReceiverDelivery {
	t.Helper()

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 {
		t.Fatal("lost or duplicated retained help", due, err)
	}

	delivery := due[0]
	if delivery.Text != "*Commands*\n`/original`" || delivery.PlainText != "Commands\n/original" || delivery.ParseMode != "Markdown" || delivery.PlainFallback {
		t.Fatal("help snapshot changed across reopen/replay", delivery)
	}

	return delivery
}

func selectTelegramHelpFallback(t *testing.T, store *ReceiverStore, id string, now time.Time) {
	t.Helper()

	if err := store.ReserveDelivery(t.Context(), id, now); err != nil {
		t.Fatal(err)
	}

	if selected, err := store.SelectPlainFallback(t.Context(), id, now); err != nil || !selected {
		t.Fatal("definite rejection did not select fallback", selected, err)
	}

	if selected, err := store.SelectPlainFallback(t.Context(), id, now); err != nil || selected {
		t.Fatal("fallback was selected twice", selected, err)
	}
}

func checkTelegramHelpFallback(t *testing.T, store *ReceiverStore, id string, now time.Time) {
	t.Helper()

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 || !due[0].PlainFallback || due[0].PlainText != "Commands\n/original" || due[0].Attempts != 1 {
		t.Fatal("fallback lost snapshot or attempt budget", due, err)
	}

	if err := store.ReserveDelivery(t.Context(), id, now); err != nil {
		t.Fatal(err)
	}

	if err := store.DeliverySent(t.Context(), id, 42, now); err != nil {
		t.Fatal(err)
	}

	checkTelegramHelpAttempts(t, store, id)
}

func checkTelegramHelpAttempts(t *testing.T, store *ReceiverStore, id string) {
	t.Helper()

	var attempts, lifetime, bindings int
	if err := store.db.QueryRowContext(t.Context(), "SELECT attempts,total_attempts FROM telegram_outbox WHERE id=?", id).Scan(&attempts, &lifetime); err != nil || attempts != 2 || lifetime != 2 {
		t.Fatal("actual send attempts were not retained", attempts, lifetime, err)
	}

	if err := store.db.QueryRowContext(t.Context(), "SELECT count(*) FROM telegram_replies WHERE message_id=42").Scan(&bindings); err != nil || bindings != 0 {
		t.Fatal("help created a task reply binding", bindings, err)
	}
}

// Invalid formatted replies cannot acknowledge input or allocate a partial outbox.
func TestTelegramHelpOutboxBounds(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	update := telegramRetainedTestUpdate(t, 1, "/help")

	now := time.Now()
	if err := store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	for _, bodies := range [][2]string{{strings.Repeat("a", 3201), "plain"}, {"formatted", strings.Repeat("😀", 1601)}, {"", "plain"}} {
		if err := store.QueueFormattedReadResponse(t.Context(), update, bodies[0], bodies[1], telegramReceiverTestConfig().Digest(), now); err == nil {
			t.Fatal("invalid help snapshot accepted")
		}
	}

	pending, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 1 {
		t.Fatal("invalid response acknowledged its source", pending, err)
	}
}
