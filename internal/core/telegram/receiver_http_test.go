package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestTelegramReceiverHTTP exercises real HTTP readiness, sparse intake and topic delivery.
func TestTelegramReceiverHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { telegramReceiverHTTPHandler(t, w, r) }))
	defer server.Close()

	api := ReceiverAPI{Client: server.Client(), BaseURL: server.URL}

	token := "12345:example_test_token_123456789"
	if bot, err := api.CheckConnection(t.Context(), token); err != nil || bot != 12345 {
		t.Fatal("readiness failed", bot, err)
	}

	offset := int64(9001)

	updates, err := api.Poll(t.Context(), token, &offset, 1)
	if err != nil || len(updates) != 1 || updates[0].UpdateID != 9005 || updates[0].Text != "/help" {
		t.Fatal("intake failed", updates, err)
	}

	if id, err := api.Send(t.Context(), token, PeerIdentity{ChatID: -100, TopicID: 7}, "hello"); err != nil || id != 88 {
		t.Fatal("delivery failed", id, err)
	}
}

// TestTelegramReceiverHTTPFailures checks fail-closed API semantics and sanitized errors.
func TestTelegramReceiverHTTPFailures(t *testing.T) {
	cases := []struct {
		name, body, code string
		status           int
		permanent        bool
	}{
		{"credentials", `{"ok":false,"error_code":401,"description":"private-token"}`, "credentials_or_access", 401, true},
		{"conflict", `{"ok":false,"error_code":409}`, "receiver_conflict", 409, true},
		{"rate limit", `{"ok":false,"error_code":429,"parameters":{"retry_after":20}}`, "rate_limited", 429, false},
		{"server", `{"ok":false,"error_code":503}`, "request_failed", 503, false},
		{"api failure", `{"ok":false,"error_code":400}`, "request_failed", 200, true},
		{"oversized", strings.Repeat("x", receiverBodyLimit+1), "response_limit_or_read_failure", 200, false},
		{"trailing JSON", `{"ok":true,"result":[]} {}`, "invalid_response", 200, true},
		{"missing ID", `{"ok":true,"result":[{"message":{}}]}`, "invalid_response", 200, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) { checkTelegramHTTPFailure(t, test.body, test.status, test.code, test.permanent) })
	}
}

func checkTelegramHTTPFailure(t *testing.T, body string, status int, code string, permanent bool) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)

		if _, err := fmt.Fprint(w, body); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	api := ReceiverAPI{Client: server.Client(), BaseURL: server.URL}
	_, err := api.Poll(t.Context(), "12345:example_test_token_123456789", nil, 1)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != code || apiErr.Permanent != permanent || strings.Contains(err.Error(), "private-token") {
		t.Fatalf("unexpected safe API result: %v", err)
	}

	if code == "rate_limited" && apiErr.RetryAfter != 20*time.Second {
		t.Fatal("retry_after discarded")
	}
}

// TestTelegramReceiverWebhook refuses startup without removing or changing the webhook.
func TestTelegramReceiverWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := `{"id":12345,"is_bot":true}`
		if strings.HasSuffix(r.URL.Path, "/getWebhookInfo") {
			result = `{"url":"https://example.com/webhook"}`
		} else if !strings.HasSuffix(r.URL.Path, "/getMe") {
			t.Error("unexpected readiness mutation", r.URL.Path)
		}

		if _, err := fmt.Fprintf(w, `{"ok":true,"result":%s}`, result); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	_, err := (ReceiverAPI{Client: server.Client(), BaseURL: server.URL}).CheckConnection(t.Context(), "12345:example_test_token_123456789")

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "webhook_conflict" {
		t.Fatal("webhook was not refused", err)
	}
}

// TestTelegramReceiverLockProcess proves cross-process exclusion and release without deleting lock files.
func TestTelegramReceiverLockProcess(t *testing.T) {
	if root := os.Getenv("CHR_TEST_TELEGRAM_LOCK_ROOT"); root != "" {
		env := telegramTestEnvironment{root: root}

		if lock, err := AcquireReceiverLock(env, 12345); err == nil {
			t.Fatal("second process acquired live bot lock", lock.Close())
		}

		return
	}

	root := t.TempDir()

	env := telegramTestEnvironment{root: root}

	lock, err := AcquireReceiverLock(env, 12345)
	if err != nil {
		t.Fatal(err)
	}

	child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestTelegramReceiverLockProcess$")

	child.Env = append(os.Environ(), "CHR_TEST_TELEGRAM_LOCK_ROOT="+root)
	output, childErr := child.CombinedOutput()

	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}

	if childErr != nil {
		t.Fatal("cross-process lock failed", string(output), childErr)
	}

	reopened, err := AcquireReceiverLock(env, 12345)
	if err != nil {
		t.Fatal("released lock stayed owned", err)
	}

	if err := reopened.Close(); err != nil {
		t.Fatal(err)
	}
}

func telegramReceiverHTTPHandler(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	var payload map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Error(err)
	}

	result := `{"ok":true,"result":{"id":12345,"is_bot":true,"username":"example_bot"}}`

	switch {
	case strings.HasSuffix(r.URL.Path, "/getWebhookInfo"):
		result = `{"ok":true,"result":{"url":""}}`
	case strings.HasSuffix(r.URL.Path, "/getUpdates"):
		if !telegramHTTPTestPollBounds(payload) {
			t.Error("incorrect bounded polling request")
		}

		result = `{"ok":true,"result":[{"update_id":9005,"message":{"message_id":4,"from":{"id":100001,"is_bot":false},"chat":{"id":100001,"type":"private"},"text":"/help","entities":[]}}]}`
	case strings.HasSuffix(r.URL.Path, "/sendMessage"):
		if string(payload["message_thread_id"]) != "7" || payload["parse_mode"] != nil {
			t.Error("topic routing or plain text contract violated")
		}

		result = `{"ok":true,"result":{"message_id":88}}`
	}

	if _, err := fmt.Fprint(w, result); err != nil {
		t.Error(err)
	}
}

func telegramHTTPTestPollBounds(payload map[string]json.RawMessage) bool {
	return string(payload["offset"]) == "9001" && string(payload["limit"]) == "50" && string(payload["timeout"]) == "1"
}
