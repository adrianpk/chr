package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramHelpHTTPBody struct {
	Entities  []core.TelegramTextEntity `json:"entities"`
	Text      string                    `json:"text"`
	ParseMode string                    `json:"parse_mode"`
}

// Real HTTP and SQLite distinguish definite format rejection from uncertain delivery.
func TestTelegramHelpHTTP(t *testing.T) {
	tests := []struct {
		name, text, failure, state string
		status, attempts           int
		modes                      []string
	}{
		{"formatted", "/help", "", "", 0, 0, []string{"Markdown"}},
		{"unknown command", "/invent", "", "", 0, 0, []string{""}},
		{"plain reply", "/status", "", "", 0, 0, []string{""}},
		{"format fallback", "/help", "Bad Request: can't parse entities: unsupported entity", "", 400, 0, []string{"Markdown", ""}},
		{"ordinary rejection", "/help", "Bad Request: chat not found", "blocked", 400, 1, []string{"Markdown"}},
		{"rate limit", "/help", "Too Many Requests", "pending", 429, 1, []string{"Markdown"}},
		{"uncertain server", "/help", "Internal Server Error", "uncertain", 503, 1, []string{"Markdown"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payloads := make(chan telegramHelpHTTPBody, 4)
			server := telegramHelpHTTPServer(t, payloads, test.status, test.failure)
			receiver, _ := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})

			update := retainTelegramHelpInput(t, receiver, 1, test.text)
			if err := receiver.respond(t.Context(), update); err != nil {
				t.Fatal(err)
			}

			if err := receiver.drainDeliveries(t.Context()); err != nil {
				t.Fatal(err)
			}

			assertTelegramHelpHTTP(t, payloads, test.modes, test.text)
			assertTelegramHelpDelivery(t, receiver, test.state, test.attempts)

			if err := receiver.respond(t.Context(), update); err != nil {
				t.Fatal(err)
			}

			if err := receiver.drainDeliveries(t.Context()); err != nil || len(payloads) != 0 {
				t.Fatal("replay or deferred failure caused an immediate extra send", err)
			}
		})
	}
}

func telegramHelpHTTPServer(t *testing.T, payloads chan<- telegramHelpHTTPBody, status int, failure string) *httptest.Server {
	t.Helper()

	var rejected atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload telegramHelpHTTPBody
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}

		payloads <- payload

		if status != 0 && rejected.CompareAndSwap(false, true) {
			w.WriteHeader(status)

			if err := json.NewEncoder(w).Encode(map[string]any{"ok": false, "error_code": status, "description": failure, "parameters": map[string]int{"retry_after": 60}}); err != nil {
				t.Error(err)
			}

			return
		}

		if _, err := fmt.Fprint(w, `{"ok":true,"result":{"message_id":77}}`); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

// A pending confirmation cannot interpret /help as its answer or consume its binding.
func TestTelegramHelpConfirmation(t *testing.T) {
	f := telegramOperationAppFixture(t)
	action, prompt := telegramOperationOfferRecord(t, f)
	before := f.sessions(t)

	help := f.send(t, "/help", prompt.ID)
	if !strings.Contains(help.Text, "*Execution*") || !reflect.DeepEqual(before, f.sessions(t)) {
		t.Fatal("help was intercepted by a pending confirmation")
	}

	if text := f.send(t, "/answer "+action.Token+" confirm", 0).Text; !strings.Contains(text, "Confirmed record mutation") {
		t.Fatal("help changed the pending confirmation", text)
	}
}

// A real pending authoring question and provider usage remain unchanged by help.
func TestTelegramHelpQuestion(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
		return telegramAuthoringBridgeGeneration(t, input)
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare retry ceiling", 0)

	c := f.sessions(t)[0]
	f.send(t, "Preserve five attempts.", 0)
	f.send(t, "/done "+c.ID, 0)

	question := f.send(t, "/quiz "+c.ID, 0)
	before := telegramAuthoringBridgeSnapshot(t, f, c.ID)
	used := calls.Load()

	if text := f.send(t, "/help", question.ID).Text; !strings.Contains(text, "*Authoring*") {
		t.Fatal("help unavailable during pending question", text)
	}

	if !reflect.DeepEqual(before, telegramAuthoringBridgeSnapshot(t, f, c.ID)) || calls.Load() != used {
		t.Fatal("help answered the question or invoked the model")
	}
}

// Current authorization is checked again before a formatted response reaches HTTP.
func TestTelegramHelpRevocation(t *testing.T) {
	client := &telegramHelpTransportFailure{}
	receiver, config := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: client})

	update := retainTelegramHelpInput(t, receiver, 1, "/help")
	if err := receiver.respond(t.Context(), update); err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].ProjectIDs = []string{"other-project"}
	if err := core.SaveTelegramReceiverConfig(receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	if err := receiver.drainDeliveries(t.Context()); err != nil || client.calls != 0 {
		t.Fatal("revoked help reached the sender", client.calls, err)
	}

	assertTelegramHelpDelivery(t, receiver, "blocked", 0)
}

func retainTelegramHelpInput(t *testing.T, receiver telegramReceiver, id int64, text string) core.TelegramReceivedUpdate {
	t.Helper()

	update := core.TelegramReceivedUpdate{UpdateID: id, Digest: fmt.Sprintf("%064x", id), Kind: "message", Peer: core.TelegramPeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001}, MessageID: id, ChatType: "private", Text: text}
	if err := receiver.store.RetainBatch(t.Context(), 12345, []core.TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	return update
}

func assertTelegramHelpHTTP(t *testing.T, payloads <-chan telegramHelpHTTPBody, modes []string, command string) {
	t.Helper()

	if len(payloads) != len(modes) {
		t.Fatalf("send attempts = %d, want %d", len(payloads), len(modes))
	}

	for _, mode := range modes {
		payload := <-payloads
		if payload.ParseMode != mode {
			t.Fatalf("parse mode = %q, want %q", payload.ParseMode, mode)
		}

		assertTelegramHelpBody(t, payload, command)
	}
}

func assertTelegramHelpBody(t *testing.T, payload telegramHelpHTTPBody, command string) {
	t.Helper()

	if command != "/help" && command != "/invent" {
		return
	}

	if command == "/invent" {
		if payload.Text != "I could not use that command. Send /help to see available commands." {
			t.Fatal("unknown command lost its focused help route")
		}

		return
	}

	want := telegramBasicHelp().Text
	if payload.ParseMode == "" {
		want = telegramBasicHelp().PlainText
	}

	if payload.Text != want {
		t.Fatal("outgoing help lost its complete retained body")
	}
}

// Revoked intake authority cannot enqueue help or invoke any application service.
func TestTelegramHelpUnauthorized(t *testing.T) {
	client := &telegramHelpTransportFailure{}
	receiver, config := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: client})
	update := retainTelegramHelpInput(t, receiver, 1, "/help")
	receiver.operations = telegramConversationNoReads{}

	config.AllowedUsers[0].TelegramUserID = 200002
	if err := core.SaveTelegramReceiverConfig(receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	if err := receiver.respond(t.Context(), update); err != nil {
		t.Fatal(err)
	}

	if err := receiver.drainDeliveries(t.Context()); err != nil || client.calls != 0 {
		t.Fatal("unauthorized help reached the sender", client.calls, err)
	}

	assertTelegramHelpDelivery(t, receiver, "", 0)
}

func assertTelegramHelpDelivery(t *testing.T, receiver telegramReceiver, state string, attempts int) {
	t.Helper()

	diagnostics, err := core.InspectTelegramReceiverDiagnostics(t.Context(), receiver.env, 12345)
	if err != nil {
		t.Fatal(err)
	}

	if state == "" {
		if len(diagnostics.Deliveries) != 0 {
			t.Fatal("acknowledged send remains unresolved", diagnostics.Deliveries)
		}

		return
	}

	if len(diagnostics.Deliveries) != 1 || diagnostics.Deliveries[0].State != state || diagnostics.Deliveries[0].Attempts != attempts {
		t.Fatal("failed send lost its state or attempt budget", diagnostics.Deliveries)
	}
}

type telegramHelpTransportFailure struct{ calls int }

func (c *telegramHelpTransportFailure) Do(*http.Request) (*http.Response, error) {
	c.calls++

	return nil, errors.New("fixture transport failure")
}

// A lost HTTP response never authorizes an immediate second plain send.
func TestTelegramHelpTransportLoss(t *testing.T) {
	client := &telegramHelpTransportFailure{}
	receiver, _ := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: client, BaseURL: "http://example.invalid"})

	update := retainTelegramHelpInput(t, receiver, 1, "/help")
	if err := receiver.respond(t.Context(), update); err != nil {
		t.Fatal(err)
	}

	if err := receiver.drainDeliveries(t.Context()); err != nil || client.calls != 1 {
		t.Fatal("transport loss caused immediate fallback", client.calls, err)
	}

	assertTelegramHelpDelivery(t, receiver, "uncertain", 1)
}

// Explicit help and reply-to-help keep the focused session and capture bindings intact.
func TestTelegramHelpConversation(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	f.send(t, "/project example-project", 0)

	prompt := f.send(t, "/prepare topic", 0)
	f.send(t, "retained capture", prompt.ID)

	before := f.sessions(t)
	f.receiver.operations = telegramConversationNoReads{}

	help := f.send(t, "/help", prompt.ID)
	if !strings.Contains(help.Text, "*Conversations*") || !reflect.DeepEqual(before, f.sessions(t)) {
		t.Fatal("help became capture or changed the session")
	}

	if !strings.Contains(f.send(t, "/help", help.ID).Text, "*System controls*") || !reflect.DeepEqual(before, f.sessions(t)) {
		t.Fatal("help response gained a task binding")
	}

	checkTelegramNoWork(t, f.receiver.env)
}
