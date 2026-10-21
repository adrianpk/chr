package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramAuthoringExplainRootBinding rejects a registry rebind between
// routing and source preparation, before reading context or invoking a provider.
func TestTelegramAuthoringExplainRootBinding(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	telegramAuthoringExplainOverview(t, f)

	input := telegramOperationRetainInput(t, f, "/explain Describe retries.")

	target, err := f.receiver.store.ResolveTelegramOperationTarget(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	telegramAuthoringChangeAuthority(t, f, "root")

	operations := telegramOperationConcreteService(t, f.receiver)

	_, err = operations.PrepareTelegramExplanation(t.Context(), target, AuthoringExplanationRequest{ID: "bound-root", Question: "Describe retries."})
	if !errors.Is(err, core.ErrWorkflowConflict) {
		t.Fatal("changed registry root was entered", err)
	}
}

func telegramAuthoringExplainOverview(t *testing.T, f *telegramConversationAppFixture) string {
	t.Helper()

	operations := telegramOperationConcreteService(t, f.receiver)

	projects, err := operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal(projects, err)
	}

	path := filepath.Join(projects[0].Path, "README.md")
	if err := os.WriteFile(path, []byte("# Example\n\nPreserve the configured retry ceiling.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

// TestTelegramAuthoringExplainRouting proves slash and ordinary-language
// explanations use real canonical context and the same configured HTTP provider.
func TestTelegramAuthoringExplainRouting(t *testing.T) {
	for _, mode := range []string{"slash", "natural", "session"} {
		t.Run(mode, func(t *testing.T) { telegramAuthoringExplainRoute(t, mode) })
	}
}

func telegramAuthoringExplainRoute(t *testing.T, mode string) {
	t.Helper()
	t.Setenv(core.OpsUserEnv, "original")

	f := newTelegramConversationAppFixture(t)
	telegramAuthoringExplainOverview(t, f)

	var interpretations atomic.Int64

	calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
		return telegramAuthoringExplainReply(t, input)
	}, func(input ConversationIntentContext) ConversationIntent {
		interpretations.Add(1)

		return ConversationIntent{OperationID: "explain", Arguments: ConversationIntentArguments{Text: input.Message}}
	})
	message, before := telegramAuthoringExplainMessage(t, f, mode)

	notice := f.send(t, message, 0)
	if !strings.Contains(notice.Text, "configured retry ceiling") || !strings.Contains(string(telegramLatestDocument(t, f, "explanation").Payload), "Canonical sources:") {
		t.Fatal("missing canonical explanation", notice)
	}

	telegramAuthoringExplainCheckCalls(t, mode, calls.Load(), interpretations.Load())

	if mode == "session" && f.session(t, before.ID) != before {
		t.Fatal("explanation changed its conversation")
	}

	if mode != "session" && len(f.sessions(t)) != 0 {
		t.Fatal("explanation created a conversation")
	}

	telegramAuthoringBridgeNoWork(t, f)
}

func telegramAuthoringExplainReply(t *testing.T, input authoringModelInput) string {
	t.Helper()

	if input.Topic != "" {
		return telegramAuthoringServiceOutput(t, "Preserve the configured retry ceiling.", "")
	}

	if len(input.Sources) == 0 || !strings.Contains(input.Sources[0].Content, "retry") {
		t.Error("explanation received no real canonical context")
	}

	return `{"explanation":"The configured retry ceiling is retained."}`
}

func telegramAuthoringExplainMessage(t *testing.T, f *telegramConversationAppFixture, mode string) (string, core.TelegramConversation) {
	t.Helper()

	if mode == "natural" {
		return "Explain how retries are bounded.", core.TelegramConversation{}
	}

	if mode != "session" {
		return "/explain How are retries bounded?", core.TelegramConversation{}
	}

	f.send(t, "/prepare retry ceiling", 0)

	before := f.sessions(t)[0]
	f.send(t, "/done "+before.ID, 0)

	before = f.session(t, before.ID)

	return "/explain --conversation " + before.ID + " How are retries bounded?", before
}

func telegramAuthoringExplainCheckCalls(t *testing.T, mode string, calls, interpretations int64) {
	t.Helper()

	expectedCalls, expectedIntent := int64(1), int64(0)
	if mode == "session" {
		expectedCalls = 2
	}

	if mode == "natural" {
		expectedIntent = 1
	}

	if calls != expectedCalls || interpretations != expectedIntent {
		t.Fatal("unexpected model routing", calls, interpretations)
	}
}

// TestTelegramAuthoringExplainRecovery reopens after result retention but before
// source acknowledgement, preserving an independent edit and the original answer.
func TestTelegramAuthoringExplainRecovery(t *testing.T) {
	t.Setenv(core.OpsUserEnv, "original")

	f := newTelegramConversationAppFixture(t)
	readme := telegramAuthoringExplainOverview(t, f)
	calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
		if !strings.Contains(input.Sources[0].Content, "configured retry ceiling") {
			t.Error("retained sources changed")
		}

		return `{"explanation":"The original ceiling is retained."}`
	}, func(ConversationIntentContext) ConversationIntent {
		t.Error("slash or retained explanation was interpreted")

		return ConversationIntent{}
	})

	request, prepared := telegramAuthoringExplainReserve(t, f)
	telegramAuthoringExplainCache(t, f, request, prepared)

	if err := os.WriteFile(readme, []byte("Independent later project context.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	telegramAuthoringReopen(t, f)

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 1 || len(f.notices) != 1 {
		t.Fatal("recovery regenerated or redelivered", calls.Load(), len(f.notices))
	}

	notice := <-f.notices
	if !strings.Contains(notice.Text, "original ceiling") {
		t.Fatal("original answer lost", notice)
	}

	content, err := os.ReadFile(readme)
	if err != nil || string(content) != "Independent later project context.\n" {
		t.Fatal("recovery changed project context", string(content), err)
	}

	telegramAuthoringBridgeNoWork(t, f)
}

func telegramAuthoringExplainReserve(t *testing.T, f *telegramConversationAppFixture) (core.TelegramOperationRequest, PreparedAuthoringExplanation) {
	t.Helper()

	input := telegramOperationRetainInput(t, f, "/explain How are retries bounded?")

	command, recognized, err := parseTelegramExplanation(input.Update)
	if err != nil || !recognized {
		t.Fatal(command, err)
	}

	command.Input = input
	user, _ := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	operations := telegramOperationConcreteService(t, f.receiver)

	request, err := f.receiver.reserveExplanation(t.Context(), operations, user, command)
	if err != nil {
		t.Fatal(err)
	}

	prepared, err := telegramRetainedExplanation(request)
	if err != nil {
		t.Fatal(err)
	}

	return request, prepared
}

// TestTelegramAuthoringExplainAuthority blocks a generated explanation if the
// operator or project root changes during HTTP generation, without any mutation.
func TestTelegramAuthoringExplainAuthority(t *testing.T) {
	for _, change := range []string{"root", "operator"} {
		t.Run(change, func(t *testing.T) {
			f := newTelegramConversationAppFixture(t)
			telegramAuthoringExplainOverview(t, f)

			calls := telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
				telegramAuthoringChangeAuthority(t, f, change)

				return `{"explanation":"This response must not be delivered."}`
			}, telegramAuthoringBridgeCapture)

			notice := f.send(t, "/explain Describe retry behavior.", 0)
			if calls.Load() != 1 || strings.Contains(notice.Text, "must not be delivered") || len(f.sessions(t)) != 0 {
				t.Fatal("changed authority was accepted", calls.Load(), notice)
			}

			telegramAuthoringBridgeNoWork(t, f)
		})
	}
}

func telegramAuthoringExplainCache(t *testing.T, f *telegramConversationAppFixture, request core.TelegramOperationRequest, prepared PreparedAuthoringExplanation) {
	t.Helper()

	if err := f.receiver.store.BeginTelegramOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	operations := telegramOperationConcreteService(t, f.receiver)

	model, err := f.receiver.store.ResolveTelegramOperationModel(t.Context(), f.receiver.env, request, time.Now(), func(ctx context.Context) (json.RawMessage, error) {
		result, err := operations.ExplainPreparedTelegram(ctx, request.Target, prepared)
		if err != nil {
			return nil, err
		}

		return json.Marshal(result)
	})
	if err != nil || len(model.Response) == 0 {
		t.Fatal(model, err)
	}
}
