package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	openaiquiz "forge.adrianpk.com/chr/internal/core/quiz/openai"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramAuthoringBridgeHTTP(t *testing.T, f *telegramConversationAppFixture, generate func(authoringModelInput) string, intent func(ConversationIntentContext) ConversationIntent) *atomic.Int64 {
	t.Helper()

	var generations atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var envelope struct {
			Input string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
			t.Error(err)
		}

		output := telegramAuthoringBridgeReply(t, envelope.Input, generate, intent, &generations)

		if err := json.NewEncoder(w).Encode(map[string]string{"output_text": output}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)

	provider := core.NewConfiguredStructuredModel(true)
	provider.API = openaiquiz.Adapter{Host: core.QuizGenerationHost{}, APIKey: "fixture", BaseURL: server.URL, Client: server.Client()}
	operations := telegramOperationConcreteService(t, f.receiver)
	operations.conversationModel = provider
	f.receiver.operations = operations

	return &generations
}

func telegramAuthoringBridgeReply(t *testing.T, payload string, generate func(authoringModelInput) string, intent func(ConversationIntentContext) ConversationIntent, generations *atomic.Int64) string {
	t.Helper()

	var input ConversationIntentContext
	if err := json.Unmarshal([]byte(payload), &input); err != nil {
		t.Error(err)
	}

	if input.Message != "" {
		return telegramAuthoringIntentJSON(t, intent(input))
	}

	var modelInput authoringModelInput
	if err := json.Unmarshal([]byte(payload), &modelInput); err != nil {
		t.Error(err)
	}

	generations.Add(1)

	return generate(modelInput)
}

func telegramAuthoringBridgeCapture(input ConversationIntentContext) ConversationIntent {
	return ConversationIntent{OperationID: "capture", Arguments: ConversationIntentArguments{Text: input.Message}}
}

func telegramAuthoringBridgeSnapshot(t *testing.T, f *telegramConversationAppFixture, conversation string) authoringcore.Snapshot {
	t.Helper()

	c := f.session(t, conversation)

	id, err := telegramAuthoringSessionID(c.AuthoringSessionRef)
	if err != nil {
		t.Fatal(err)
	}

	operations := telegramOperationConcreteService(t, f.receiver)

	snapshot, err := operations.ObserveTelegramAuthoring(t.Context(), telegramcore.OperationTarget{ProjectID: c.ProjectID, Namespace: c.Namespace}, id)
	if err != nil {
		t.Fatal(err)
	}

	return snapshot
}

func telegramAuthoringBridgeNoWork(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	checkTelegramNoWork(t, f.receiver.env)

	operations := telegramOperationConcreteService(t, f.receiver)

	projects, err := operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal("project missing", projects, err)
	}

	operations.projectContext = commandProjectContext{ProjectID: projects[0].ID, ProjectDir: projects[0].Path}
	telegramAuthoringServiceNoWork(t, operations)
}

// TestTelegramAuthoringBridgeJourney traverses real Telegram/model HTTP and
// SQLite boundaries to generate, answer and refine canonical artifacts once.
func TestTelegramAuthoringBridgeJourney(t *testing.T) {
	t.Setenv(opsnamespace.UserEnv, "original")

	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
		return telegramAuthoringBridgeGeneration(t, input)
	}, telegramAuthoringBridgeCapture)

	f.send(t, "/prepare retry ceiling", 0)

	c := f.sessions(t)[0]
	f.send(t, "  Preserve existing defaults.\n\n", 0)

	if text := f.send(t, "/done "+c.ID, 0).Text; !strings.Contains(text, "saved as a dump") {
		t.Fatal("capture did not generate a dump", text)
	}

	t.Setenv(opsnamespace.UserEnv, "changed-after-capture")

	question := f.send(t, "/quiz "+c.ID, 0)
	if !strings.Contains(question.Text, "Which retry ceiling") {
		t.Fatal("real question was not delivered", question)
	}

	f.send(t, "  Five attempts.\n\n", question.ID)
	f.send(t, "/spec "+c.ID, 0)
	f.send(t, "/refine spec "+c.ID+" retain the same scope and clarify validation", 0)
	f.send(t, "/plan "+c.ID, 0)

	telegramAuthoringBridgeCheckJourney(t, f, c.ID, calls.Load())
	telegramAuthoringBridgeCheckReplay(t, f, c.ID, calls)
	telegramAuthoringBridgeNoWork(t, f)
}

func telegramAuthoringBridgeGeneration(t *testing.T, input authoringModelInput) string {
	t.Helper()

	if input.Target == authoringcore.ArtifactKindQuiz {
		return telegramAuthoringServiceOutput(t, "", "Which retry ceiling applies?")
	}

	if input.Target == authoringcore.ArtifactKindSpec && len(input.Answers) != 1 {
		t.Error("canonical answer missing from generation")
	}

	return telegramAuthoringServiceOutput(t, "## Requirements\n\nRetain five attempts.\n\n## Validation\n\nCheck exhaustion after reopening.", "")
}

func telegramAuthoringBridgeCheckJourney(t *testing.T, f *telegramConversationAppFixture, conversation string, calls int64) {
	t.Helper()

	snapshot := telegramAuthoringBridgeSnapshot(t, f, conversation)
	if len(snapshot.Artifacts) != 5 || len(snapshot.Events) != 2 || calls != 5 || snapshot.Question != nil || snapshot.Session.Namespace != "original" {
		t.Fatalf("journey snapshot=%+v calls=%d", snapshot, calls)
	}

	if !strings.Contains(snapshot.Artifacts[0].Body, "/prepare retry ceiling") {
		t.Fatal("original instruction was replaced by model text")
	}

	if !strings.Contains(snapshot.Artifacts[0].Body, "  Preserve existing defaults.\n\n") {
		t.Fatal("canonical raw capture lost user formatting")
	}

	if snapshot.Events[1].Answer.Text != "  Five attempts.\n\n" {
		t.Fatal("canonical answer lost user formatting")
	}
}

func telegramAuthoringBridgeCheckReplay(t *testing.T, f *telegramConversationAppFixture, conversation string, calls *atomic.Int64) {
	t.Helper()

	before := telegramAuthoringBridgeSnapshot(t, f, conversation).SourceHash

	f.updates <- f.last

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	if len(f.notices) != 0 || calls.Load() != 5 || telegramAuthoringBridgeSnapshot(t, f, conversation).SourceHash != before {
		t.Fatal("source replay repeated generation or delivery")
	}
}

// TestTelegramAuthoringBridgeClarification answers a real missing-information
// prompt and completes its requested specification without a second spec command.
func TestTelegramAuthoringBridgeClarification(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
		if len(input.Answers) == 0 {
			return telegramAuthoringServiceOutput(t, "", "Which timeout is required?")
		}

		if input.Answers[0].Answer != "Keep sixty seconds." {
			t.Error("actual bound answer was replaced")
		}

		return telegramAuthoringServiceOutput(t, "## Requirements\n\nKeep sixty seconds.", "")
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare timeouts", 0)

	c := f.sessions(t)[0]

	question := f.send(t, "/spec "+c.ID, 0)
	if !strings.Contains(question.Text, "Which timeout") {
		t.Fatal("missing real clarification", question)
	}

	f.send(t, "/reply "+c.ID+" Keep sixty seconds.", 0)

	snapshot := telegramAuthoringBridgeSnapshot(t, f, c.ID)
	if snapshot.Question != nil || len(snapshot.Session.Outputs) != 1 || snapshot.Session.Outputs[0].Kind != authoringcore.ArtifactKindSpec || calls.Load() != 2 {
		t.Fatalf("answer did not complete the requested artifact: %+v calls=%d", snapshot, calls.Load())
	}

	before := snapshot.SourceHash

	f.send(t, "stale answer", question.ID)

	if telegramAuthoringBridgeSnapshot(t, f, c.ID).SourceHash != before || calls.Load() != 2 {
		t.Fatal("stale prompt changed its artifact")
	}

	telegramAuthoringBridgeNoWork(t, f)
}
