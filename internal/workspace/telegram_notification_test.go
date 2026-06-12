package workspace

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRenderTelegramNotificationMessageUsesSummaryAndDetail(t *testing.T) {
	event := validNotificationEvent()

	message := RenderTelegramNotificationMessage(event)

	if !strings.Contains(message.Text, event.Payload.Summary) {
		t.Fatalf("message text = %q, want summary", message.Text)
	}

	if !strings.Contains(message.Text, event.Payload.Detail) {
		t.Fatalf("message text = %q, want detail", message.Text)
	}
}

func TestTelegramSenderSendsMessage(t *testing.T) {
	var gotPath string

	var gotPayload struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		gotPath = request.URL.Path

		err := json.NewDecoder(request.Body).Decode(&gotPayload)
		if err != nil {
			t.Fatalf("decode request payload: %v", err)
		}

		_, err = writer.Write([]byte(`{"ok":true}`))
		if err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	sender := TelegramSender{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	result, err := sender.SendMessage(context.Background(), "token-1", TelegramMessage{
		ChatID: "123456",
		Text:   "hello",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", result.StatusCode)
	}

	if gotPath != "/bottoken-1/sendMessage" {
		t.Fatalf("path = %q, want Telegram sendMessage endpoint", gotPath)
	}

	if gotPayload.ChatID != "123456" || gotPayload.Text != "hello" {
		t.Fatalf("payload = %#v, want chat id and text", gotPayload)
	}
}

func TestTelegramSenderReturnsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)

		_, err := writer.Write([]byte(`{"ok":false}`))
		if err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	sender := TelegramSender{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	result, err := sender.SendMessage(context.Background(), "token-1", TelegramMessage{
		ChatID: "123456",
		Text:   "hello",
	})
	if err == nil {
		t.Fatal("SendMessage returned nil error")
	}

	if result.StatusCode != http.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want 400", result.StatusCode)
	}
}

func TestTelegramSenderValidatesInputs(t *testing.T) {
	sender := NewTelegramSender(nil)

	for _, test := range []struct {
		name    string
		token   string
		message TelegramMessage
		wantErr string
	}{
		{name: "token", token: "", message: TelegramMessage{ChatID: "1", Text: "hello"}, wantErr: "telegram bot token is required"},
		{name: "chat id", token: "token", message: TelegramMessage{Text: "hello"}, wantErr: "telegram chat id is required"},
		{name: "text", token: "token", message: TelegramMessage{ChatID: "1"}, wantErr: "telegram message text is required"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := sender.SendMessage(context.Background(), test.token, test.message)
			if err == nil {
				t.Fatal("SendMessage returned nil error")
			}

			if err.Error() != test.wantErr {
				t.Fatalf("SendMessage error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestRenderTelegramNotificationMessageUsesSummaryOnly(t *testing.T) {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	event := NotificationEvent{
		ID:        "notification-1",
		ProjectID: "project-1",
		Type:      NotificationEventManualTest,
		Severity:  NotificationSeverityInfo,
		Status:    NotificationStatusPending,
		Source:    NotificationSourceManual,
		Payload: NotificationPayload{
			Summary: "hello",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	message := RenderTelegramNotificationMessage(event)

	if message.Text != "hello" {
		t.Fatalf("Text = %q, want summary only", message.Text)
	}
}
