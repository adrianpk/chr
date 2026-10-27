package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramRecoveryRemote detects a competing remote poller at the real
// getUpdates boundary and blocks restart without probing or taking ownership.
func TestTelegramRecoveryRemote(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)

		if r.URL.Path != "/bot12345:example_test_token_123456789/getUpdates" {
			t.Error("unexpected takeover operation")
		}

		w.WriteHeader(http.StatusConflict)

		if _, err := w.Write([]byte(`{"ok":false,"error_code":409,"description":"private remote receiver details"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	receiver, _ := telegramAppTestReceiver(t, telegramcore.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})

	_, err := receiver.step(t.Context())
	if err != nil && strings.Contains(err.Error(), "private") {
		t.Fatal("remote details leaked")
	}

	if err == nil || ExitCode(telegramCommandFailure(err)) != 78 {
		t.Fatal("remote conflict did not require correction", err)
	}

	if _, err := receiver.step(t.Context()); err == nil {
		t.Fatal("blocked receiver polled again")
	}

	checkTelegramRecoveryRemoteState(t, receiver, calls.Load())
}

func checkTelegramRecoveryRemoteState(t *testing.T, receiver telegramReceiver, calls int32) {
	t.Helper()

	state, err := receiver.store.Receiver(t.Context(), receiver.botID)
	if err != nil || state.Health != "blocked" || state.LastErrorCode != "receiver_conflict" || state.RetryAttempt != 1 || calls != 1 {
		t.Fatal("remote receiver conflict not retained", state, calls, err)
	}
}
