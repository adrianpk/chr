package app

import openaiquiz "forge.adrianpk.com/chr/internal/core/quiz/openai"

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramAuthoringIntentHTTP(t *testing.T, f *telegramConversationAppFixture, reply func(ConversationIntentContext) string) *atomic.Int64 {
	t.Helper()

	var calls atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)

		var request struct {
			Input string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}

		var input ConversationIntentContext
		if err := json.Unmarshal([]byte(request.Input), &input); err != nil {
			t.Error(err)
		}

		output := ""
		if input.Message == "" {
			output = telegramAuthoringServiceOutput(t, "Preserve the existing retry policy.", "")
		} else {
			output = reply(input)
		}

		if err := json.NewEncoder(w).Encode(map[string]string{"output_text": output}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)

	provider := core.NewConfiguredStructuredModel(true)
	provider.API = openaiquiz.OpenAIQuizModelAdapter{Host: core.QuizGenerationHost{}, APIKey: "fixture", BaseURL: server.URL, Client: server.Client()}

	operations, ok := f.receiver.operations.(service)
	if !ok {
		t.Fatal("fixture service is unavailable")
	}

	operations.conversationModel = provider
	f.receiver.operations = operations

	return &calls
}

func telegramAuthoringIntentJSON(t *testing.T, intent ConversationIntent) string {
	t.Helper()

	if intent.Unresolved == nil {
		intent.Unresolved = []string{}
	}

	data, err := json.Marshal(intent)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

// TestTelegramAuthoringIntentParity compares the same typed start/capture/done
// transitions from explicit controls and real HTTP-backed natural language.
func TestTelegramAuthoringIntentParity(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringIntentHTTP(t, f, func(input ConversationIntentContext) string {
		switch input.Message {
		case "prepare the timeout change":
			return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "prepare", Arguments: ConversationIntentArguments{Text: "timeout change"}})
		case "finish my capture":
			return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "done"})
		default:
			return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "capture", Arguments: ConversationIntentArguments{Text: input.Message}})
		}
	})
	explicit := f.send(t, "/prepare timeout change", 0)
	first := f.sessions(t)[0]

	captured := f.send(t, "keep the existing retry policy", explicit.ID)
	if !strings.Contains(captured.Text, "Details saved") {
		t.Fatalf("capture=%q", captured.Text)
	}

	f.send(t, "/done "+first.ID, 0)
	f.send(t, "prepare the timeout change", 0)

	second := f.sessions(t)[0]
	f.send(t, "keep the existing retry policy", 0)
	f.send(t, "finish my capture", 0)

	left, right := f.session(t, first.ID), f.session(t, second.ID)
	if left.Phase != "prepared" || left.AuthoringSessionRef == "" || right.AuthoringSessionRef == "" {
		t.Fatalf("missing canonical preparation: %+v %+v", left, right)
	}

	left.ID, right.ID = "", ""

	left.AuthoringSessionRef, right.AuthoringSessionRef = "", ""
	if left != right {
		t.Fatalf("explicit=%+v natural=%+v", left, right)
	}

	if calls.Load() != 6 {
		t.Fatalf("model calls=%d", calls.Load())
	}

	checkTelegramNoWork(t, f.receiver.env)
}

// TestTelegramAuthoringIntentInvalid rejects malformed/unknown proposals and
// invented authority without effects; unknown slash commands never invoke a model.
func TestTelegramAuthoringIntentInvalid(t *testing.T) {
	tests := []struct{ name, payload string }{
		{"unknown_operation", `{"operation_id":"shell","arguments":{},"unresolved":[],"explanation":""}`},
		{"unknown_field", `{"operation_id":"prepare","arguments":{"text":"change","operator":"admin"},"unresolved":[],"explanation":""}`},
		{"duplicate_field", `{"operation_id":"prepare","operation_id":"develop","arguments":{},"unresolved":[],"explanation":""}`},
		{"trailing_payload", `{"operation_id":"help","arguments":{},"unresolved":[],"explanation":""} {}`},
		{"null", `null`},
		{"null_arguments", `{"operation_id":"help","arguments":null,"unresolved":[],"explanation":""}`},
		{"missing_arguments", `{"operation_id":"help","unresolved":[],"explanation":""}`},
		{"malformed", `oops`},
		{"unknown_project", `{"operation_id":"prepare","arguments":{"text":"change","project":"private-project"},"unresolved":[],"explanation":""}`},
		{"invented_scope", `{"operation_id":"projects","arguments":{"project":"private-project"},"unresolved":[],"explanation":""}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTelegramConversationAppFixture(t)
			calls := telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string { return tt.payload })
			notice := f.send(t, "prepare a change", 0)
			// Every rejected proposal produces a correction without an effect.
			if notice.Text == "" {
				t.Fatal("missing correction")
			}

			if len(f.sessions(t)) != 0 {
				t.Fatal("invalid proposal created a conversation")
			}

			f.send(t, "/unknown danger", 0)
			f.send(t, "/help", 0)

			if calls.Load() != 1 {
				t.Fatalf("deterministic controls called model: %d", calls.Load())
			}

			checkTelegramNoWork(t, f.receiver.env)
		})
	}
}

// TestTelegramAuthoringIntentRecordParity sends equivalent record creation through
// the real HTTP resolver and verifies one shared-owner ticket and retained target.
func TestTelegramAuthoringIntentRecordParity(t *testing.T) {
	f := telegramOperationAppFixture(t)
	calls := telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string {
		return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "ticket", Arguments: ConversationIntentArguments{Action: "create", RecordJSON: `{"title":"Natural ticket","synopsis":"Prepared by request"}`}})
	})

	notice := f.send(t, "create a ticket named Natural ticket with synopsis Prepared by request", 0)
	if !strings.Contains(notice.Text, "PRJ-1") {
		t.Fatalf("notice=%q", notice.Text)
	}

	f.send(t, `/ticket create {"title":"Explicit ticket","synopsis":"Prepared by request"}`, 0)

	repeated := f.last
	f.updates <- repeated

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	service, ok := f.receiver.operations.(service)
	if !ok {
		t.Fatal("missing fixture service")
	}

	projects, err := service.ListRegisteredProjects([]string{"example-project"})
	if err != nil {
		t.Fatal(err)
	}

	service.projectContext = commandProjectContext{ProjectDir: projects[0].Path, ProjectID: "example-project"}

	page, err := service.ReadRecords(t.Context(), "default", RecordQuery{Kind: RecordTicket})
	if err != nil || len(page.Records) != 2 || calls.Load() != 1 {
		t.Fatalf("page=%+v calls=%d err=%v", page, calls.Load(), err)
	}

	checkTelegramNoWork(t, f.receiver.env)
}

// TestTelegramAuthoringIntentRegistry keeps resolver IDs identical to available
// deterministic controls, with capture as the ordinary-message transition.
func TestTelegramAuthoringIntentRegistry(t *testing.T) {
	ids := map[string]bool{}
	for _, operation := range telegramIntentRegistry() {
		if operation.ID == "" || ids[operation.ID] {
			t.Fatalf("duplicate/empty operation: %+v", operation)
		}

		ids[operation.ID] = true
	}

	for _, command := range telegramConversationRegistry() {
		if !ids[strings.TrimPrefix(command.Name, "/")] {
			t.Fatalf("missing %s", command.Name)
		}
	}

	for _, command := range telegramRecordRegistry() {
		if !ids[strings.TrimPrefix(command.Name, "/")] {
			t.Fatalf("missing %s", command.Name)
		}
	}

	for _, command := range telegramReadRegistry() {
		if !ids[strings.TrimPrefix(command.Command, "/")] {
			t.Fatalf("missing %s", command.Command)
		}
	}
}

// TestTelegramAuthoringIntentRootChange rechecks the frozen registry binding after
// an in-flight model call; a replacement root cannot receive the proposed write.
func TestTelegramAuthoringIntentRootChange(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	replacement := t.TempDir()
	telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string {
		registry, err := projectcore.NewRegistry(t.Context(), f.receiver.env, testAssetsFS())
		if err != nil {
			t.Error(err)

			return "{}"
		}

		if _, err := registry.UpdateProjectPath(t.Context(), "example-project", replacement); err != nil {
			t.Error(err)
		}

		if err := registry.Close(); err != nil {
			t.Error(err)
		}

		return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "prepare", Arguments: ConversationIntentArguments{Text: "bounded change"}})
	})

	notice := f.send(t, "prepare a bounded change", 0)
	if !strings.Contains(notice.Text, "prior target or arguments are no longer valid") || len(f.sessions(t)) != 0 {
		t.Fatalf("notice=%q sessions=%+v", notice.Text, f.sessions(t))
	}

	checkTelegramNoWork(t, f.receiver.env)
}

// TestTelegramAuthoringIntentRevocation prevents a provider result from changing
// conversation state after authorization is revoked during its call.
func TestTelegramAuthoringIntentRevocation(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string {
		config, err := telegramcore.LoadReceiverConfig(f.receiver.env)
		if err != nil {
			t.Error(err)

			return "{}"
		}

		config.AllowedUsers[0].TelegramUserID = 100002
		if err := telegramcore.SaveReceiverConfig(f.receiver.env, config, true); err != nil {
			t.Error(err)
		}

		return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "prepare", Arguments: ConversationIntentArguments{Text: "bounded change"}})
	})

	notice := f.send(t, "prepare a bounded change", 0)
	if notice.Text != "" || len(f.sessions(t)) != 0 {
		t.Fatalf("revoked input changed state: %q", notice.Text)
	}

	checkTelegramNoWork(t, f.receiver.env)
}

// TestTelegramAuthoringIntentProviderFailure preserves the pending instruction and
// prior capture while deterministic help/status controls remain usable.
func TestTelegramAuthoringIntentProviderFailure(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	f.send(t, "/prepare retained context", 0)

	prior := f.sessions(t)[0]

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer server.Close()

	provider := core.NewConfiguredStructuredModel(true)
	provider.API = openaiquiz.OpenAIQuizModelAdapter{Host: core.QuizGenerationHost{}, APIKey: "fixture", BaseURL: server.URL, Client: server.Client()}

	operations, ok := f.receiver.operations.(service)
	if !ok {
		t.Fatal("missing fixture service")
	}

	operations.conversationModel = provider
	f.receiver.operations = operations
	f.send(t, "keep this new context", 0)

	if notice := f.send(t, "/status", 0); !strings.Contains(notice.Text, "Receiver:") {
		t.Fatalf("status=%q", notice.Text)
	}

	if current := f.session(t, prior.ID); current != prior {
		t.Fatal("provider failure changed prior conversation")
	}

	pending, err := f.receiver.store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 1 || pending[0].Text != "keep this new context" {
		t.Fatalf("pending=%+v err=%v", pending, err)
	}

	checkTelegramNoWork(t, f.receiver.env)
}

// TestTelegramAuthoringIntentReplayScope freezes model choice and namespace before
// dispatch; reopening and changing OPS_USER cannot regenerate or redirect it.
func TestTelegramAuthoringIntentReplayScope(t *testing.T) {
	t.Setenv("OPS_USER", "original")

	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string {
		return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "prepare", Arguments: ConversationIntentArguments{Text: "retained scope"}})
	})
	input := telegramOperationRetainInput(t, f, "prepare retained scope")
	operations := telegramOperationConcreteService(t, f.receiver)

	first, err := f.receiver.resolveIntent(t.Context(), operations, input)
	if err != nil {
		t.Fatal(err)
	}

	reopened, err := telegramcore.NewReceiverStore(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reopened.Close(); err != nil {
			t.Error(err)
		}
	}()

	f.receiver.store = reopened

	t.Setenv("OPS_USER", "changed")

	replay, err := f.receiver.resolveIntent(t.Context(), operations, input)
	if err != nil || string(replay.Decision) != string(first.Decision) {
		t.Fatalf("replay=%+v calls=%d err=%v", replay, calls.Load(), err)
	}

	if err := f.receiver.applyIntentDecision(t.Context(), input.Update, replay.Decision); err != nil {
		t.Fatal(err)
	}

	sessions := f.sessions(t)
	if len(sessions) != 1 || sessions[0].Namespace != "original" {
		t.Fatalf("sessions=%+v", sessions)
	}

	if calls.Load() != 1 {
		t.Fatalf("provider repeated: %d", calls.Load())
	}

	checkTelegramNoWork(t, f.receiver.env)
}
