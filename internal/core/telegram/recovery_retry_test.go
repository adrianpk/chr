package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// TestTelegramRecoveryRetry preserves transport exhaustion through reopen and
// starts exactly one explicit, attributable cycle without changing its source.
func TestTelegramRecoveryRetry(t *testing.T) {
	for _, kind := range []string{"poll", "delivery"} {
		t.Run(kind, func(t *testing.T) { checkTelegramRecoveryRetry(t, kind) })
	}
}

func checkTelegramRecoveryRetry(t *testing.T, kind string) {
	t.Helper()

	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	now := time.Now().UTC().Truncate(time.Second)

	request := seedTelegramRecoveryRetry(t, store, kind, now)

	for attempt := range 5 {
		when := now.Add(time.Duration(attempt) * 4 * time.Minute)
		failTelegramRecoveryTransport(t, store, request, when)

		if err := store.Close(); err != nil {
			t.Fatal(err)
		}

		store = openTelegramReceiverTestStore(t, env)
	}

	before, err := store.Receiver(t.Context(), 12345)
	if err != nil {
		t.Fatal(err)
	}

	checkTelegramRecoveryExhausted(t, store, request, now.Add(time.Hour))
	checkTelegramRecoveryRetryLock(t, store, request)

	if err := store.RetryTransport(t.Context(), request, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	if err := store.RetryTransport(t.Context(), request, now.Add(time.Hour)); err == nil {
		t.Fatal("duplicate local retry created another cycle")
	}

	checkTelegramRecoveryRetryResult(t, store, request, before)
}

func checkTelegramRecoveryRetryResult(t *testing.T, store *ReceiverStore, request TransportRetry, before ReceiverState) {
	t.Helper()

	after, err := store.Receiver(t.Context(), 12345)
	if err != nil || after.Epoch != before.Epoch || *after.NextUpdateID != *before.NextUpdateID {
		t.Fatal("retry moved source identity", after, err)
	}

	checkTelegramRecoveryCycle(t, store, request)

	var attempts int
	if err := store.db.QueryRowContext(t.Context(), `SELECT interpretation_attempt FROM telegram_updates WHERE bot_id=12345 AND update_id=1`).Scan(&attempts); err != nil || attempts != 4 {
		t.Fatal("transport retry reset an application model budget", attempts, err)
	}
}

func seedTelegramRecoveryRetry(t *testing.T, store *ReceiverStore, kind string, now time.Time) TransportRetry {
	t.Helper()

	update := telegramRetainedTestUpdate(t, 1, "/status")
	if err := store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	execTelegramRecoverySQL(t, store, `UPDATE telegram_updates SET interpretation_attempt=4,interpretation_error_code='provider_unavailable' WHERE bot_id=12345 AND update_id=1`)

	request := TransportRetry{BotID: 12345, Poll: kind == "poll", Actor: "maintainer"}
	if kind == "delivery" {
		if err := store.QueueReadResponse(t.Context(), update, "original result", telegramReceiverTestConfig().Digest(), now); err != nil {
			t.Fatal(err)
		}

		deliveries, err := store.DueDeliveries(t.Context(), 12345, now)
		if err != nil || len(deliveries) != 1 {
			t.Fatal(deliveries, err)
		}

		request.DeliveryID = deliveries[0].ID
	}

	return request
}

func failTelegramRecoveryTransport(t *testing.T, store *ReceiverStore, request TransportRetry, now time.Time) {
	t.Helper()

	failure := &APIError{Code: "rate_limited", RetryAfter: 3 * time.Minute}

	if request.Poll {
		if err := store.ReservePoll(t.Context(), 12345, now); err != nil {
			t.Fatal(err)
		}

		if err := store.PollFailed(t.Context(), 12345, failure, now); err != nil {
			t.Fatal(err)
		}

		return
	}

	deliveries, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(deliveries) != 1 {
		t.Fatal(deliveries, err)
	}

	if err := store.ReserveDelivery(t.Context(), request.DeliveryID, now); err != nil {
		t.Fatal(err)
	}

	if err := store.DeliveryFailed(t.Context(), deliveries[0], failure, now); err != nil {
		t.Fatal(err)
	}
}

func checkTelegramRecoveryExhausted(t *testing.T, store *ReceiverStore, request TransportRetry, now time.Time) {
	t.Helper()

	if request.Poll {
		if err := store.ReservePoll(t.Context(), 12345, now); err == nil {
			t.Fatal("restart reset poll exhaustion")
		}

		return
	}

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 0 {
		t.Fatal("restart reset delivery exhaustion", due, err)
	}
}

func checkTelegramRecoveryRetryLock(t *testing.T, store *ReceiverStore, request TransportRetry) {
	t.Helper()

	lock, err := AcquireConfigurationLock(store.env)
	if err != nil {
		t.Fatal(err)
	}

	retryErr := store.RetryTransport(t.Context(), request, time.Now())
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}

	if retryErr == nil {
		t.Fatal("local retry raced a live receiver")
	}
}

func checkTelegramRecoveryCycle(t *testing.T, store *ReceiverStore, request TransportRetry) {
	t.Helper()

	var (
		attempt, cycle, total int
		actor, when           string
	)

	query := `SELECT retry_attempt,retry_cycle,total_retry_attempts,retry_actor,COALESCE(retry_started_at,'') FROM telegram_receivers WHERE bot_id=?`

	var id any = request.BotID
	if !request.Poll {
		query = `SELECT attempts,retry_cycle,total_attempts,retry_actor,COALESCE(retry_started_at,'') FROM telegram_outbox WHERE id=?`
		id = request.DeliveryID
	}

	if err := store.db.QueryRowContext(t.Context(), query, id).Scan(&attempt, &cycle, &total, &actor, &when); err != nil {
		t.Fatal(err)
	}

	if attempt != 0 || cycle != 2 || total != 5 || actor != "maintainer" || when == "" {
		t.Fatal("retry lost lifetime or attribution", attempt, cycle, total, actor, when)
	}
}

// TestTelegramRecoveryModels reopens SQLite after every failed model call.
// Exhausted interpretation and operation generation never call the model again.
func TestTelegramRecoveryModels(t *testing.T) {
	for _, kind := range []string{"interpretation", "operation"} {
		t.Run(kind, func(t *testing.T) { checkTelegramRecoveryModels(t, kind) })
	}
}

func checkTelegramRecoveryModels(t *testing.T, kind string) {
	t.Helper()

	f, request := telegramAuthoringModelFixture(t)
	calls := 0
	invoke := func(context.Context) (json.RawMessage, error) {
		calls++

		return nil, errors.New("private provider detail")
	}

	for attempt := 1; attempt <= 6; attempt++ {
		if err := f.store.Close(); err != nil {
			t.Fatal(err)
		}

		f.store = openTelegramReceiverTestStore(t, f.env)
		now := time.Now().Add(time.Duration(attempt) * time.Hour)
		count, code := 0, ""

		switch kind {
		case "interpretation":
			result, err := f.store.ResolveInterpretation(t.Context(), f.env, request.Source, now, invoke)
			if err != nil {
				t.Fatal(err)
			}

			count, code = result.Attempts, result.ErrorCode
		case "operation":
			result, err := f.store.ResolveOperationModel(t.Context(), f.env, request, now, invoke)
			if err != nil {
				t.Fatal(err)
			}

			count, code = result.Attempts, result.ErrorCode
		}

		if count != min(attempt, 5) || calls != min(attempt, 5) || code != "provider_unavailable" {
			t.Fatal("model restart changed bounded retries", count, calls, code)
		}
	}
}
