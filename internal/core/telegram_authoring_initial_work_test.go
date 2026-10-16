package core

import (
	"encoding/json"
	"testing"
	"time"
)

func telegramAuthoringInitialWork(t *testing.T, f *telegramConversationFixture, id int64, stale bool) (TelegramOperationRequest, error) {
	t.Helper()

	input := f.input
	input.Conversation = f.sessions(t)[0].ID

	input.Update = telegramRetainedTestUpdate(t, id, "/prepare-work "+input.Conversation)
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	target, err := f.store.ResolveTelegramOperationTarget(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if stale {
		target.ConversationRevision++
	}

	return f.store.ReserveTelegramOperation(t.Context(), TelegramOperationRequest{Source: input.Update, Operator: "maintainer", Target: target, OperationID: "authoring.work", Arguments: json.RawMessage(`{"original":"instruction"}`)}, time.Now())
}

// TestTelegramAuthoringInitialWork pins one first scope even after failure and
// reopening, so a later request cannot reuse an original instruction for expansion.
func TestTelegramAuthoringInitialWork(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

	first, err := telegramAuthoringInitialWork(t, f, 90, false)
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.FinishTelegramOperation(t.Context(), first, "blocked", "", json.RawMessage(`{}`), "Correct the scope.", f.input.Config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	second, err := telegramAuthoringInitialWork(t, f, 91, false)
	if err != nil {
		t.Fatal(err)
	}

	store := openTelegramReceiverTestStore(t, f.env)
	for _, test := range []struct {
		name    string
		request TelegramOperationRequest
		initial bool
	}{{"first", first, true}, {"later", second, false}} {
		t.Run(test.name, func(t *testing.T) {
			initial, err := store.IsInitialTelegramAuthoringWork(t.Context(), test.request)
			if err != nil || initial != test.initial {
				t.Fatal("initial instruction was rebound", initial, err)
			}
		})
	}
}

// TestTelegramAuthoringInitialRollback rolls back both a stale reservation and
// its first-scope marker; the next valid request still owns the first binding.
func TestTelegramAuthoringInitialRollback(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

	if _, err := telegramAuthoringInitialWork(t, f, 90, true); err == nil {
		t.Fatal("stale scope reservation succeeded")
	}

	request, err := telegramAuthoringInitialWork(t, f, 91, false)
	if err != nil {
		t.Fatal(err)
	}

	initial, err := f.store.IsInitialTelegramAuthoringWork(t.Context(), request)
	if err != nil || !initial {
		t.Fatal("failed reservation consumed first scope", initial, err)
	}
}
