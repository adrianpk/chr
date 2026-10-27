package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	modelcore "forge.adrianpk.com/chr/internal/core/model"
	openaiquiz "forge.adrianpk.com/chr/internal/core/quiz/openai"
)

func telegramAuthoringModelRequest() StructuredModelRequest {
	return StructuredModelRequest{ID: "intent-test", Purpose: "conversation_intent", Input: json.RawMessage(`{"message":"explain"}`), Schema: map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{}, "required": []string{}}, Instructions: "Return JSON only."}
}

// TestTelegramAuthoringModelBoundary uses the configured HTTP provider and rejects
// malformed/oversized output, transport failures and oversized context.
func TestTelegramAuthoringModelBoundary(t *testing.T) {
	tests := []struct {
		name, output string
		status       int
		wantErr      bool
	}{
		{"valid", `{"operation_id":"help"}`, 200, false},
		{"malformed", `{oops`, 200, true},
		{"oversized", `{"text":"` + strings.Repeat("a", 65536) + `"}`, 200, true},
		{"unavailable", "secret error", 503, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := telegramAuthoringModelServer(t, tt.status, tt.output)
			defer server.Close()

			provider := NewConfiguredStructuredModel(true)
			provider.API = openaiquiz.Adapter{Host: harnesshost.QuizGeneration{}, APIKey: "fixture", BaseURL: server.URL, Client: server.Client()}

			_, err := provider.InvokeStructuredModel(t.Context(), telegramAuthoringModelRequest())
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v", err)
			}

			if err != nil && strings.Contains(err.Error(), "secret error") {
				t.Fatal("provider detail leaked")
			}
		})
	}

	provider := NewConfiguredStructuredModel(true)
	request := telegramAuthoringModelRequest()

	request.Input = json.RawMessage(`{"text":"` + strings.Repeat("a", 128*1024) + `"}`)
	if _, err := provider.InvokeStructuredModel(t.Context(), request); err == nil {
		t.Fatal("oversized context accepted")
	}
}

// TestTelegramAuthoringModelCapacity proves two shared calls, bounded cancellation
// and slot reuse without permitting a third in-flight request.
func TestTelegramAuthoringModelCapacity(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})

	server := telegramAuthoringWaitingServer(t, entered, release)
	defer server.Close()

	provider := NewConfiguredStructuredModel(true)
	provider.API = openaiquiz.Adapter{Host: harnesshost.QuizGeneration{}, APIKey: "fixture", BaseURL: server.URL, Client: server.Client()}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan error, 2)

	for range 2 {
		go func() { _, err := provider.InvokeStructuredModel(ctx, telegramAuthoringModelRequest()); done <- err }()
	}

	for range 2 {
		<-entered
	}

	if _, err := provider.InvokeStructuredModel(t.Context(), telegramAuthoringModelRequest()); !errors.Is(err, modelcore.ErrCapacity) {
		t.Errorf("third request err=%v", err)
	}

	cancel()
	close(release)

	for range 2 {
		if err := <-done; err == nil {
			t.Error("canceled model call succeeded")
		}
	}

	shortCtx, shortCancel := context.WithTimeout(t.Context(), time.Nanosecond)
	defer shortCancel()

	if _, err := provider.InvokeStructuredModel(shortCtx, telegramAuthoringModelRequest()); err == nil || errors.Is(err, modelcore.ErrCapacity) {
		t.Fatalf("released slot err=%v", err)
	}
}

func telegramAuthoringModelServer(t *testing.T, status int, output string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" || r.Header.Get("Authorization") != "Bearer fixture" {
			t.Error("wrong provider request")
		}

		w.WriteHeader(status)

		if err := json.NewEncoder(w).Encode(map[string]any{"output_text": output}); err != nil {
			t.Error(err)
		}
	}))
}

func telegramAuthoringWaitingServer(t *testing.T, entered chan struct{}, release <-chan struct{}) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}

		select {
		case <-release:
		case <-r.Context().Done():
		}

		if err := json.NewEncoder(w).Encode(map[string]string{"output_text": "{}"}); err != nil && r.Context().Err() == nil {
			t.Error(err)
		}
	}))
}
