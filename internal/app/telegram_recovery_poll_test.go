package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramRecoveryPollEpoch checks the actual HTTP poll after a long
// outage. It requests no offset, keeps local pending input and retains a sparse ID.
func TestTelegramRecoveryPollEpoch(t *testing.T) {
	polled := make(chan map[string]json.RawMessage, 1)

	server := telegramRecoveryEpochServer(t, polled)
	defer server.Close()

	receiver, _ := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})
	receiver.slots = newTelegramActivationSlots()

	expireTelegramRecoveryCursor(t, receiver)

	if _, err := receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	request := <-polled
	if _, ok := request["offset"]; ok {
		t.Fatal("long outage retained a stale offset", request)
	}

	if _, ok := request["drop_pending_updates"]; ok {
		t.Fatal("retention recovery dropped pending updates")
	}

	checkTelegramRecoveryPollPending(t, receiver)

	state, err := receiver.store.Receiver(t.Context(), 12345)
	if err != nil || state.Epoch != 2 || state.RetentionGapAt.IsZero() || state.NextUpdateID == nil || *state.NextUpdateID != 6 {
		t.Fatal("new epoch did not retain the earliest available source", state, err)
	}
}

func telegramRecoveryEpochServer(t *testing.T, polled chan<- map[string]json.RawMessage) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}

		polled <- request

		if _, err := w.Write([]byte(`{"ok":true,"result":[{"update_id":5,"message":{"message_id":5,"from":{"id":100001},"chat":{"id":100001,"type":"private"},"text":"/help"}}]}`)); err != nil {
			t.Error(err)
		}
	}))
}

func expireTelegramRecoveryCursor(t *testing.T, receiver telegramReceiver) {
	t.Helper()

	db, err := sql.Open("sqlite", core.UserDatabasePath(receiver.env))
	if err != nil {
		t.Fatal(err)
	}

	_, writeErr := db.ExecContext(t.Context(), `UPDATE telegram_receivers SET next_update_id=9001,last_poll_at=? WHERE bot_id=12345`, time.Now().Add(-25*time.Hour).UTC().Format(time.RFC3339Nano))

	_, seedErr := db.ExecContext(t.Context(), `INSERT INTO telegram_updates(bot_id,update_id,cursor_epoch,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at) VALUES(12345,499,1,?,'message',100001,0,100001,499,'private',?,'/status','pending','',?)`, strings.Repeat("a", 64), time.Now().Add(-25*time.Hour).UTC().Format(time.RFC3339Nano), time.Now().Add(-25*time.Hour).UTC().Format(time.RFC3339Nano))
	if err := errors.Join(writeErr, seedErr, db.Close()); err != nil {
		t.Fatal(err)
	}
}

func checkTelegramRecoveryPollPending(t *testing.T, receiver telegramReceiver) {
	t.Helper()

	pending, err := receiver.store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 2 || pending[0].UpdateID != 5 || pending[1].UpdateID != 499 {
		t.Fatal("new epoch lost locally pending input", pending, err)
	}
}
