package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramAppTestReceiver(t *testing.T, api core.TelegramReceiverAPI) (telegramReceiver, core.TelegramReceiverConfig) {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	config := core.TelegramReceiverConfig{
		Enabled: true, BotTokenEnv: fmt.Sprintf("CHR_TEST_TELEGRAM_%x", sha256.Sum256([]byte(env.ConfigDir()))), PollTimeoutSeconds: 1,
		AllowedUsers: []core.TelegramAllowedUser{{TelegramUserID: 100001, OperatorID: "maintainer", ProjectIDs: []string{"example-project", "disabled-project"}}},
	}

	token := "12345:example_test_token_123456789"
	if err := core.SaveTelegramReceiverToken(env, config, token, false); err != nil {
		t.Fatal(err)
	}

	if err := core.SaveTelegramReceiverConfig(env, config, false); err != nil {
		t.Fatal(err)
	}

	seedTelegramAppSchema(t, env)

	store, err := core.NewTelegramReceiverStore(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})

	if err := store.BindReceiver(t.Context(), 12345, config, time.Now()); err != nil {
		t.Fatal(err)
	}

	operations := newService(t.Context(), env, commandProjectContext{}, testAssetsFS())
	operations.conversationModel = telegramCaptureIntentTestProvider{}

	return telegramReceiver{env: env, api: api, store: store, operations: operations, botID: 12345, token: token}, config
}

func addTelegramAppTestProjects(t *testing.T, env core.UserEnvironment) {
	t.Helper()

	registry, err := core.NewProjectRegistry(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"example-project", "disabled-project", "private-project"} {
		project := core.Project{ID: id, Name: id, Path: t.TempDir(), Enabled: id != "disabled-project", RegisteredAt: time.Now()}
		if err := registry.AddProject(t.Context(), project); err != nil {
			t.Fatal(err)
		}
	}

	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestTelegramReceiverReadJourney uses real HTTP and SQLite to prove authorized reads, durable replay and no work admission.
func TestTelegramReceiverReadJourney(t *testing.T) {
	sent := make(chan string, 8)

	server := telegramReadTestServer(t, sent)
	defer server.Close()

	receiver, _ := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})
	addTelegramAppTestProjects(t, receiver.env)
	seedTelegramSavedHandoff(t, receiver.env)

	for range 2 {
		if _, err := receiver.step(t.Context()); err != nil {
			t.Fatal(err)
		}
	}

	if len(sent) != 3 {
		t.Fatal("read replay duplicated or lost responses", len(sent))
	}

	help, projects, work := <-sent, <-sent, <-sent
	if !strings.Contains(help, "/status") || projects != "example-project — example-project" || !strings.Contains(work, "unavailable command") {
		t.Fatal("incorrect bounded read responses", help, projects, work)
	}

	state, err := receiver.store.Receiver(t.Context(), 12345)
	if err != nil || state.NextUpdateID == nil || *state.NextUpdateID != 311 {
		t.Fatal("complete batch was not retained", state, err)
	}

	checkTelegramNoWork(t, receiver.env)
}

func checkTelegramNoWork(t *testing.T, env core.UserEnvironment) {
	t.Helper()

	records, err := core.NewRemoteCommandRecordRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	commands, readErr := records.ListRemoteCommandRecords(t.Context())
	if err := records.Close(); err != nil {
		t.Fatal(err)
	}

	if readErr != nil || len(commands) != 1 || commands[0].Status != core.RemoteCommandRecordAccepted {
		t.Fatal("receiver changed historical command handoffs", readErr)
	}

	executions, err := core.NewExecutionRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	runs, readErr := executions.ListProjectExecutions(t.Context(), "example-project")
	if err := executions.Close(); err != nil {
		t.Fatal(err)
	}

	if readErr != nil || len(runs) != 0 {
		t.Fatal("receiver admitted work from a read or historical handoff", readErr)
	}
}

func seedTelegramSavedHandoff(t *testing.T, env core.UserEnvironment) {
	t.Helper()

	update := core.TelegramUpdate{Message: core.TelegramUpdateMessage{
		From: core.TelegramSenderUser{ID: 100001}, Chat: core.TelegramChat{ID: 100001},
		Text: "@example-project loop start --status ready --max-iterations 1",
	}}

	record, err := core.RemoteCommandRecordFromTelegramUpdate(update, time.Now())
	if err != nil || record.Status != core.RemoteCommandRecordAccepted {
		t.Fatal("historical handoff fixture is invalid", err)
	}

	repository, err := core.NewRemoteCommandRecordRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	if err := repository.CreateRemoteCommandRecord(t.Context(), record); err != nil {
		t.Fatal(err)
	}

	if err := repository.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestTelegramReceiverRevocation blocks a retained read response after project authority changes.
func TestTelegramReceiverRevocation(t *testing.T) {
	receiver, config := telegramAppTestReceiver(t, core.TelegramReceiverAPI{})

	update := core.TelegramReceivedUpdate{UpdateID: 1, Digest: strings.Repeat("a", 64), Kind: "message", Peer: core.TelegramPeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001}, MessageID: 1, ChatType: "private", Text: "/projects"}
	if err := receiver.store.RetainBatch(t.Context(), 12345, []core.TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := receiver.store.QueueReadResponse(t.Context(), update, "previous project list", config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].ProjectIDs = []string{"other-project"}
	if err := core.SaveTelegramReceiverConfig(receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	if err := receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	due, err := receiver.store.DueDeliveries(t.Context(), 12345, time.Now().Add(time.Hour))
	if err != nil || len(due) != 0 {
		t.Fatal("revoked configuration retained a deliverable response", err)
	}
}

// TestTelegramReceiverShutdown cancels an owned long poll and releases both locks without touching another service.
func TestTelegramReceiverShutdown(t *testing.T) {
	polling := make(chan struct{})

	release := make(chan struct{})

	server := telegramShutdownTestServer(t, polling, release)
	defer server.Close()
	defer close(release)

	receiver, _ := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	finished := make(chan error, 1)
	go func() { finished <- serveTelegramReceiver(ctx, receiver.env, testAssetsFS(), receiver.api) }()

	select {
	case <-polling:
	case err := <-finished:
		t.Fatal("receiver failed before polling", err)
	case <-time.After(3 * time.Second):
		t.Fatal("receiver did not poll")
	}

	cancel()

	select {
	case err := <-finished:
		if err != nil {
			t.Fatal("shutdown failed", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("owned poll did not shut down")
	}

	checkTelegramShutdownReleased(t, receiver.env)
}

func checkTelegramShutdownReleased(t *testing.T, env core.UserEnvironment) {
	t.Helper()

	lock, err := core.AcquireTelegramReceiverLock(env, 12345)
	if err != nil {
		t.Fatal("shutdown retained bot ownership", err)
	}

	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(core.UserDatabasePath(env)); err != nil {
		t.Fatal("shutdown lost durable state", err)
	}
}

func telegramShutdownTestServer(t *testing.T, polling chan struct{}, release <-chan struct{}) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := `{"id":12345,"is_bot":true}`

		switch {
		case strings.HasSuffix(r.URL.Path, "/getWebhookInfo"):
			result = `{"url":""}`
		case strings.HasSuffix(r.URL.Path, "/getUpdates"):
			close(polling)

			select {
			case <-r.Context().Done():
			case <-release:
			}

			return
		}

		if _, err := fmt.Fprintf(w, `{"ok":true,"result":%s}`, result); err != nil {
			t.Error(err)
		}
	}))
}

func telegramReadTestServer(t *testing.T, sent chan<- string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sendMessage") {
			receiveTelegramTestMessage(t, w, r, sent)

			return
		}

		if err := json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": telegramReadTestUpdates()}); err != nil {
			t.Error(err)
		}
	}))
}

func receiveTelegramTestMessage(t *testing.T, w http.ResponseWriter, r *http.Request, sent chan<- string) {
	t.Helper()

	var message struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		t.Error(err)
	}

	if message.ChatID != 100001 {
		t.Error("denied peer received a response")
	}

	sent <- message.Text

	if _, err := fmt.Fprint(w, `{"ok":true,"result":{"message_id":55}}`); err != nil {
		t.Error(err)
	}
}

func telegramReadTestUpdates() []any {
	var updates []any

	for index, text := range []string{"/help", "/projects", "/unavailable-work ticket EX-1", "private denied body"} {
		user := 100001
		if index == 3 {
			user++
		}

		updates = append(updates, map[string]any{"update_id": 10 + index*100, "message": map[string]any{"message_id": index + 1, "from": map[string]any{"id": user}, "chat": map[string]any{"id": user, "type": "private"}, "text": text}})
	}

	return updates
}
