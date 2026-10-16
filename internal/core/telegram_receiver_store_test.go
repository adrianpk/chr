package core

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func openTelegramReceiverTestStore(t *testing.T, env UserEnvironment) *TelegramReceiverStore {
	t.Helper()
	store, err := NewTelegramReceiverStore(t.Context(), env, os.DirFS("../../cmd/chr"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := store.BindReceiver(t.Context(), 12345, telegramReceiverTestConfig(), time.Now()); err != nil {
		t.Fatal(err)
	}

	return store
}

func telegramRetainedTestUpdate(t *testing.T, id int64, text string) TelegramReceivedUpdate {
	t.Helper()
	data := fmt.Sprintf(`{"update_id":%d,"message":{"message_id":1,"from":{"id":100001},"chat":{"id":100001,"type":"private"},"text":%q}}`, id, text)
	update, err := decodeTelegramReceiverUpdate([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	return AuthorizeTelegramReceiverUpdate(telegramReceiverTestConfig(), 12345, update)
}

// TestTelegramReceiverBatch proves sparse deduplication and all-or-nothing cursor changes after conflict.
func TestTelegramReceiverBatch(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	first := telegramRetainedTestUpdate(t, 9, "/help")
	second := telegramRetainedTestUpdate(t, 9000, "/status")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{first, second, first}, time.Now()); err != nil {
		t.Fatal(err)
	}

	conflict := telegramRetainedTestUpdate(t, 9, "/projects")
	newer := telegramRetainedTestUpdate(t, 10000, "/help")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{newer, conflict}, time.Now()); err == nil {
		t.Fatal("changed source digest accepted")
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)
	state, err := store.Receiver(t.Context(), 12345)
	if err != nil || state.NextUpdateID == nil || *state.NextUpdateID != 9001 {
		t.Fatal("unretained input acknowledged", state, err)
	}

	pending, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 2 || pending[0].UpdateID != 9 || pending[1].UpdateID != 9000 {
		t.Fatal("duplicate or partially retained batch", pending, err)
	}
}

// TestTelegramReceiverIntakeCapacity leaves the entire over-capacity batch and cursor unacknowledged.
func TestTelegramReceiverIntakeCapacity(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	for batch := range 11 {
		updates := make([]TelegramReceivedUpdate, 0, 100)
		for i := range 100 {
			updates = append(updates, telegramRetainedTestUpdate(t, int64(batch*100+i), "/help"))
		}

		err := store.RetainBatch(t.Context(), 12345, updates, time.Now())
		if (batch == 10) != (err != nil) {
			t.Fatal("incorrect capacity outcome", batch, err)
		}
	}

	state, err := store.Receiver(t.Context(), 12345)
	if err != nil || state.NextUpdateID == nil || *state.NextUpdateID != 1000 {
		t.Fatal("capacity failure advanced cursor", state, err)
	}
}

// TestTelegramReceiverOutbox survives reopen, deduplicates read responses and retains uncertain send attempts.
func TestTelegramReceiverOutbox(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	now := time.Now().UTC().Truncate(time.Second)
	update := telegramRetainedTestUpdate(t, 1, "/help")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if err := store.QueueReadResponse(t.Context(), update, "help", now); err != nil {
			t.Fatal(err)
		}
	}

	deliveries, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(deliveries) != 1 {
		t.Fatal("read response was not unique", deliveries, err)
	}

	delivery := deliveries[0]
	if err := store.ReserveDelivery(t.Context(), delivery.ID, now); err != nil {
		t.Fatal(err)
	}

	if err := store.DeliveryFailed(t.Context(), delivery, &TelegramAPIError{Code: "transport_failure", Uncertain: true}, now); err != nil {
		t.Fatal(err)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)
	deliveries, err = store.DueDeliveries(t.Context(), 12345, now.Add(time.Minute))
	if err != nil || len(deliveries) != 1 || deliveries[0].ID != delivery.ID || deliveries[0].State != "uncertain" || deliveries[0].Attempts != 1 {
		t.Fatal("send history was reset", deliveries, err)
	}

	checkTelegramDeliveryCompletion(t, store, deliveries[0], now.Add(time.Minute))
}

func checkTelegramDeliveryCompletion(t *testing.T, store *TelegramReceiverStore, delivery TelegramReceiverDelivery, now time.Time) {
	t.Helper()
	if err := store.ReserveDelivery(t.Context(), delivery.ID, now); err != nil {
		t.Fatal(err)
	}

	if err := store.DeliverySent(t.Context(), delivery.ID, 55, now); err != nil {
		t.Fatal(err)
	}

	pending, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 0 {
		t.Fatal("processed input became pending", err)
	}

	due, err := store.DueDeliveries(t.Context(), 12345, now.Add(time.Hour))
	if err != nil || len(due) != 0 {
		t.Fatal("confirmed delivery repeated", err)
	}
}

// TestTelegramReceiverPollBudget preserves exhaustion and retry_after across connection reopen.
func TestTelegramReceiverPollBudget(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	now := time.Now().UTC().Truncate(time.Second)
	for attempt := range 5 {
		when := now.Add(time.Duration(attempt) * time.Minute)
		if err := store.ReservePoll(t.Context(), 12345, when); err != nil {
			t.Fatal(err)
		}

		if err := store.PollFailed(t.Context(), 12345, &TelegramAPIError{Code: "rate_limited", RetryAfter: 30 * time.Second}, when); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)
	if err := store.ReservePoll(t.Context(), 12345, now.Add(time.Hour)); err == nil {
		t.Fatal("restart cleared poll exhaustion")
	}

	state, err := store.Receiver(t.Context(), 12345)
	if err != nil || state.RetryAttempt != 5 || !state.RetryAt.Equal(now.Add(4*time.Minute+30*time.Second)) {
		t.Fatal("retained retry state changed", state, err)
	}
}

// TestTelegramReceiverMigration preserves the registry and rolls back an interrupted additive migration.
func TestTelegramReceiverMigration(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	registry, err := NewProjectRegistry(t.Context(), env, testMigrationFS())
	if err != nil {
		t.Fatal(err)
	}

	project := Project{ID: "example-project", Name: "Example", Path: t.TempDir(), Enabled: true, RegisteredAt: time.Now()}
	if err := registry.AddProject(t.Context(), project); err != nil {
		t.Fatal(err)
	}

	assets := testMigrationFS()
	assets["assets/migration/sqlite/20261016000016-add_telegram_receiver.sql"] = &fstest.MapFile{Data: []byte("-- +migrate Up\nCREATE TABLE telegram_receivers(bot_id INTEGER); invalid SQL;\n-- +migrate Down\nDROP TABLE telegram_receivers;")}
	if store, err := NewTelegramReceiverStore(t.Context(), env, assets); err == nil {
		t.Fatal("failed migration accepted", store.Close())
	}

	var tableCount int
	if err := registry.(*sqliteProjectRegistry).db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM sqlite_master WHERE name='telegram_receivers'").Scan(&tableCount); err != nil || tableCount != 0 {
		t.Fatal("partial migration remained", tableCount, err)
	}

	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}

	store := openTelegramReceiverTestStore(t, env)
	var name string
	if err := store.db.QueryRowContext(t.Context(), "SELECT name FROM projects WHERE id=?", project.ID).Scan(&name); err != nil || name != project.Name {
		t.Fatal("migration lost registry", err)
	}

	if _, err := store.db.ExecContext(t.Context(), "INSERT INTO schema_migrations VALUES('99999999999999','future','2026-01-01T00:00:00Z')"); err != nil {
		t.Fatal(err)
	}

	if newer, err := NewTelegramReceiverStore(t.Context(), env, os.DirFS("../../cmd/chr")); err == nil || !strings.Contains(err.Error(), "unknown") {
		if newer != nil {
			t.Error(newer.Close())
		}

		t.Fatal("future schema accepted", err)
	}
}
