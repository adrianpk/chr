package app

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestTelegramAuthoringExplainFailure retains the original intent and canonical
// context across malformed-output retries, exhausts five calls, and keeps status usable.
func TestTelegramAuthoringExplainFailure(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	readme := telegramAuthoringExplainOverview(t, f)

	var interpretations atomic.Int64

	calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
		if len(input.Sources) != 1 || !strings.Contains(input.Sources[0].Content, "configured retry ceiling") {
			t.Error("retry replaced the retained source image")
		}

		return `{"explanation":null}`
	}, func(input ConversationIntentContext) ConversationIntent {
		interpretations.Add(1)

		return ConversationIntent{OperationID: "explain", Arguments: ConversationIntentArguments{Text: input.Message}}
	})

	notice := f.send(t, "Explain retry behavior.", 0)
	if notice.Text != "" || calls.Load() != 1 {
		t.Fatal("malformed result was accepted", notice, calls.Load())
	}

	if err := os.WriteFile(readme, []byte("Independent replacement context.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	telegramAuthoringReopen(t, f)
	telegramAuthoringExplainExhaust(t, f)

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	telegramAuthoringExplainCheckExhaustion(t, f, calls.Load(), interpretations.Load())

	telegramAuthoringExplainCheckStatus(t, f)
}

func telegramAuthoringExplainExhaust(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	updates, err := f.receiver.store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(updates) != 1 {
		t.Fatal(updates, err)
	}

	operationLookup, err := f.receiver.store.Operation(t.Context(), updates[0], 0)

	request, found := operationLookup.Request, operationLookup.Found
	if err != nil || !found {
		t.Fatal(request, err)
	}

	prepared, err := telegramRetainedExplanation(request)
	if err != nil {
		t.Fatal(err)
	}

	operations := telegramOperationConcreteService(t, f.receiver)
	for attempt := 2; attempt <= 5; attempt++ {
		model, err := f.receiver.store.ResolveOperationModel(t.Context(), f.receiver.env, request, time.Now().Add(time.Hour), func(ctx context.Context) (json.RawMessage, error) {
			result, err := operations.ExplainPreparedTelegram(ctx, request.Target, prepared)
			if err != nil {
				return nil, err
			}

			return json.Marshal(result)
		})
		if err != nil || model.Attempts != attempt {
			t.Fatal(model, err)
		}
	}
}

func telegramAuthoringExplainCheckStatus(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	notice := f.send(t, "/status", 0)
	if notice.Text == "" {
		t.Fatal("provider failure blocked deterministic status")
	}

	if len(f.sessions(t)) != 0 {
		t.Fatal("explanation failure created a conversation")
	}

	telegramAuthoringBridgeNoWork(t, f)
}

func telegramAuthoringExplainCheckExhaustion(t *testing.T, f *telegramConversationAppFixture, calls, interpretations int64) {
	t.Helper()

	if len(f.notices) != 1 {
		t.Fatal("missing exhaustion diagnostic", len(f.notices))
	}

	notice := <-f.notices
	if !strings.Contains(notice.Text, "exhausted five") || calls != 5 || interpretations != 1 {
		t.Fatal("budget or intent restarted", notice, calls, interpretations)
	}
}
