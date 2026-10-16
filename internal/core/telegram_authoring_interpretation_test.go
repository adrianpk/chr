package core

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// TestTelegramAuthoringInterpretation proves retained decisions and retry budgets
// survive reopen without another model choice or an open database transaction.
func TestTelegramAuthoringInterpretation(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)

	update := telegramRetainedTestUpdate(t, 1, "prepare a change")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	calls := 0
	invoke := func(ctx context.Context) (json.RawMessage, error) {
		calls++
		// A second write can commit while the external provider is running.
		extra := telegramRetainedTestUpdate(t, 2, "/status")
		if err := store.RetainBatch(ctx, 12345, []TelegramReceivedUpdate{extra}, time.Now()); err != nil {
			return nil, err
		}

		return json.RawMessage(`{"operation":"prepare"}`), nil
	}

	first, err := store.ResolveTelegramInterpretation(t.Context(), env, update, time.Now(), invoke)
	if err != nil || first.Attempts != 1 || len(first.Decision) == 0 {
		t.Fatalf("first=%+v err=%v", first, err)
	}

	reopened := openTelegramReceiverTestStore(t, env)

	replay, err := reopened.ResolveTelegramInterpretation(t.Context(), env, update, time.Now(), invoke)
	if err != nil || string(replay.Decision) != string(first.Decision) || calls != 1 {
		t.Fatalf("replay=%+v calls=%d err=%v", replay, calls, err)
	}

	changed := update

	changed.Digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err := store.ResolveTelegramInterpretation(t.Context(), env, changed, time.Now(), invoke); err == nil {
		t.Fatal("changed source was accepted")
	}
}

// TestTelegramAuthoringInterpretationBudget preserves five failures and rejects
// another owner while an external call is active, even after a future deadline.
func TestTelegramAuthoringInterpretationBudget(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)

	update := telegramRetainedTestUpdate(t, 1, "a request")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	var calls atomic.Int64

	invoke := func(context.Context) (json.RawMessage, error) {
		calls.Add(1)

		return nil, errors.New("secret-bearing provider detail")
	}

	for attempt := 1; attempt <= 5; attempt++ {
		reopened := openTelegramReceiverTestStore(t, env)

		result, err := reopened.ResolveTelegramInterpretation(t.Context(), env, update, time.Now().Add(time.Hour), invoke)
		if err != nil || result.Attempts != attempt || result.ErrorCode != "provider_unavailable" {
			t.Fatalf("attempt=%d result=%+v err=%v", attempt, result, err)
		}

		if _, err := reopened.ResolveTelegramInterpretation(t.Context(), env, update, time.Now(), invoke); err != nil {
			t.Fatal(err)
		}
	}

	result, err := store.ResolveTelegramInterpretation(t.Context(), env, update, time.Now().Add(24*time.Hour), invoke)
	if err != nil || result.Attempts != 5 || calls.Load() != 5 {
		t.Fatalf("result=%+v calls=%d err=%v", result, calls.Load(), err)
	}
}

// TestTelegramAuthoringInterpretationExclusion proves a wall-clock jump cannot
// claim an active provider operation and that another database owner stays usable.
func TestTelegramAuthoringInterpretationExclusion(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)

	update := telegramRetainedTestUpdate(t, 1, "a request")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)

	go func() {
		_, err := store.ResolveTelegramInterpretation(t.Context(), env, update, time.Now(), func(context.Context) (json.RawMessage, error) {
			close(entered)
			<-release

			return json.RawMessage(`{}`), nil
		})
		done <- err
	}()

	<-entered

	other := openTelegramReceiverTestStore(t, env)
	_, err := other.ResolveTelegramInterpretation(t.Context(), env, update, time.Now().Add(24*time.Hour), func(context.Context) (json.RawMessage, error) { return nil, errors.New("must not run") })

	close(release)

	if err == nil {
		t.Error("active interpretation claim was stolen")
	}

	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// TestTelegramAuthoringInterpretationSource rejects forged peer/text bindings and
// deterministic controls before the external model callback can run.
func TestTelegramAuthoringInterpretationSource(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)

	source := telegramRetainedTestUpdate(t, 1, "ordinary request")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{source}, time.Now()); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		change func(*TelegramReceivedUpdate)
	}{
		{"peer", func(u *TelegramReceivedUpdate) { u.Peer.UserID++ }},
		{"topic", func(u *TelegramReceivedUpdate) { u.Peer.TopicID++ }},
		{"text", func(u *TelegramReceivedUpdate) { u.Text = "another request" }},
		{"slash", func(u *TelegramReceivedUpdate) { u.Text = "/status" }},
		{"callback", func(u *TelegramReceivedUpdate) { u.Kind = "callback_query" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := source
			tt.change(&changed)

			_, err := store.ResolveTelegramInterpretation(t.Context(), env, changed, time.Now(), func(context.Context) (json.RawMessage, error) {
				t.Error("forged source invoked model")

				return json.RawMessage(`{}`), nil
			})
			if err == nil {
				t.Fatal("forged or deterministic source accepted")
			}
		})
	}
}
