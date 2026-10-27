package telegram

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func telegramOperationTestRequest(t *testing.T, store *TelegramReceiverStore, id int64) TelegramOperationRequest {
	t.Helper()

	update := telegramRetainedTestUpdate(t, id, "/tickets create")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	request, err := store.ReserveTelegramOperation(t.Context(), TelegramOperationRequest{Source: update, Operator: "maintainer", Target: TelegramOperationTarget{ProjectID: "example-project", Namespace: "default", ProjectBindingDigest: strings.Repeat("a", 64)}, OperationID: "records.ticket.create", Arguments: json.RawMessage(`{"title":"A task","synopsis":"An outcome"}`)}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	return request
}

// TestTelegramOperationRequestRecovery keeps one immutable request across preparation, acknowledgement and send crashes.
func TestTelegramOperationRequestRecovery(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	request := telegramOperationTestRequest(t, store, 5)

	input := json.RawMessage(`{"reserved_id":"PRJ-1"}`)
	if err := store.PrepareTelegramOperation(t.Context(), request, input, time.Now()); err != nil {
		t.Fatal(err)
	}

	request.PreparedInput = input
	if err := store.BeginTelegramOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)

	retained, found, err := store.TelegramOperation(t.Context(), request.Source, 0)
	if err != nil || !found || retained.ID != request.ID || retained.State != "applying" || string(retained.PreparedInput) != string(input) {
		t.Fatal("request lost frozen owner input", retained, err)
	}

	checkTelegramOperationFinish(t, store, retained)
}

func checkTelegramOperationFinish(t *testing.T, store *TelegramReceiverStore, request TelegramOperationRequest) {
	t.Helper()

	now := time.Now()
	for range 2 {
		if err := store.FinishTelegramOperation(t.Context(), request, "confirmed", "run/receipts/original.json", json.RawMessage(`{"identity":"PRJ-1"}`), "Confirmed PRJ-1", telegramReceiverTestConfig().Digest(), now); err != nil {
			t.Fatal(err)
		}
	}

	retained, _, err := store.TelegramOperation(t.Context(), request.Source, 0)
	if err != nil || retained.State != "confirmed" || retained.ReceiptRef != "run/receipts/original.json" {
		t.Fatal("confirmed request changed", retained, err)
	}

	deliveries, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(deliveries) != 1 {
		t.Fatal("result delivery duplicated", deliveries, err)
	}
}

// TestTelegramOperationRequestConflict retains failed request arguments and refuses prepared-input corruption.
func TestTelegramOperationRequestConflict(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	request := telegramOperationTestRequest(t, store, 1)
	changed := request

	changed.Arguments = json.RawMessage(`{"title":"Replacement"}`)
	if _, err := store.ReserveTelegramOperation(t.Context(), changed, time.Now()); err == nil {
		t.Fatal("changed request arguments accepted")
	}

	if err := store.PrepareTelegramOperation(t.Context(), request, json.RawMessage(`{"identity":"PRJ-1"}`), time.Now()); err != nil {
		t.Fatal(err)
	}

	if _, err := store.db.ExecContext(t.Context(), `UPDATE telegram_requests SET prepared_input='{"identity":"PRJ-2"}' WHERE id=?`, request.ID); err != nil {
		t.Fatal(err)
	}

	if _, _, err := store.TelegramOperation(t.Context(), request.Source, 0); err == nil {
		t.Fatal("corrupt prepared input accepted")
	}
}

// TestTelegramOperationRequestRollback preserves application intent when the outbox cannot retain its result.
func TestTelegramOperationRequestRollback(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	request := telegramOperationTestRequest(t, store, 1)

	request.PreparedInput = json.RawMessage(`{"identity":"PRJ-1"}`)
	if err := store.PrepareTelegramOperation(t.Context(), request, request.PreparedInput, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := store.BeginTelegramOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	_, err := store.db.ExecContext(t.Context(), `WITH RECURSIVE n(i) AS (VALUES(1) UNION ALL SELECT i+1 FROM n WHERE i<1024) INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,state,created_at,updated_at) SELECT 'capacity-'||i,12345,100001,0,100001,'capacity-'||i,?,'result','retained','pending','now','now' FROM n`, telegramReceiverTestConfig().Digest())
	if err != nil {
		t.Fatal(err)
	}

	if err := store.FinishTelegramOperation(t.Context(), request, "confirmed", "receipt", json.RawMessage(`{}`), "confirmed", telegramReceiverTestConfig().Digest(), time.Now()); err == nil {
		t.Fatal("outbox capacity did not abort acknowledgement")
	}

	retained, _, err := store.TelegramOperation(t.Context(), request.Source, 0)
	if err != nil || retained.State != "applying" {
		t.Fatal("failed acknowledgement changed request", retained, err)
	}

	updates, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(updates) != 1 {
		t.Fatal("failed acknowledgement lost source", updates, err)
	}
}

// TestTelegramOperationRejectedBinding keeps terminal failures immutable and rejects new requests from acknowledged input.
func TestTelegramOperationRejectedBinding(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))

	request := telegramOperationTestRequest(t, store, 1)
	if err := store.FinishTelegramOperation(t.Context(), request, "rejected", "", json.RawMessage(`{"reason":"unsupported"}`), "Unsupported", telegramReceiverTestConfig().Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	request.Arguments = json.RawMessage(`{"title":"Changed after rejection"}`)
	if _, err := store.ReserveTelegramOperation(t.Context(), request, time.Now()); err == nil {
		t.Fatal("rejected identity rebound")
	}

	request.Ordinal = 1
	if _, err := store.ReserveTelegramOperation(t.Context(), request, time.Now()); err == nil {
		t.Fatal("acknowledged input emitted a new operation")
	}
}
