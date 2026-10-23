package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramConversationNotice struct {
	ID       int64
	Text     string
	Keyboard json.RawMessage
}
type telegramConversationAppFixture struct {
	receiver  telegramReceiver
	updates   chan json.RawMessage
	notices   chan telegramConversationNotice
	documents chan telegramDocumentUpload
	next      int64
	last      json.RawMessage
}

func newTelegramConversationAppFixture(t *testing.T) *telegramConversationAppFixture {
	t.Helper()

	f := &telegramConversationAppFixture{updates: make(chan json.RawMessage, 1), notices: make(chan telegramConversationNotice, 100), documents: make(chan telegramDocumentUpload, 100), next: 1}

	var messageID atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/getUpdates") {
			payload := json.RawMessage(`[]`)
			select {
			case payload = <-f.updates:
			default:
			}

			if err := json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": payload}); err != nil {
				t.Error(err)
			}

			return
		}

		var body struct {
			Text     string          `json:"text"`
			Keyboard json.RawMessage `json:"reply_markup"`
		}

		document := strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data")
		if document {
			upload := readTelegramDocumentUpload(t, r)
			f.documents <- upload

			body.Text = upload.Fields["caption"]
		} else if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}

		id := messageID.Add(1)
		f.notices <- telegramConversationNotice{ID: id, Text: body.Text, Keyboard: body.Keyboard}

		result := map[string]any{"message_id": id}
		if document {
			result["document"] = map[string]string{"file_id": "original-document"}
		}

		if err := json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": result}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)

	f.receiver, _ = telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})
	addTelegramAppTestProjects(t, f.receiver.env)
	seedTelegramSavedHandoff(t, f.receiver.env)

	return f
}

func (f *telegramConversationAppFixture) send(t *testing.T, text string, reply int64) telegramConversationNotice {
	t.Helper()

	message := map[string]any{"message_id": f.next, "from": map[string]any{"id": 100001}, "chat": map[string]any{"id": 100001, "type": "private"}, "text": text}
	if reply > 0 {
		message["reply_to_message"] = map[string]any{"message_id": reply, "text": "quoted text has no authority"}
	}

	data, err := json.Marshal([]any{map[string]any{"update_id": f.next, "message": message}})
	if err != nil {
		t.Fatal(err)
	}

	f.next++

	f.last = data
	f.updates <- data

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	select {
	case notice := <-f.notices:
		return notice
	default:
		return telegramConversationNotice{}
	}
}

func (f *telegramConversationAppFixture) sessions(t *testing.T) []core.TelegramConversation {
	t.Helper()

	sessions, err := f.receiver.store.Conversations(t.Context(), core.TelegramPeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001})
	if err != nil {
		t.Fatal(err)
	}

	return sessions
}

func (f *telegramConversationAppFixture) session(t *testing.T, id string) core.TelegramConversation {
	t.Helper()

	for _, c := range f.sessions(t) {
		if c.ID == id {
			return c
		}
	}

	t.Fatal("session unavailable", id)

	return core.TelegramConversation{}
}

// TestTelegramConversationJourney exercises real HTTP decoding, command dispatch, SQLite and reply delivery.
func TestTelegramConversationJourney(t *testing.T) {
	f := newTelegramConversationAppFixture(t)

	help := f.send(t, "/help", 0).Text
	for _, command := range telegramConversationRegistry() {
		if !strings.Contains(help, command.Name) {
			t.Fatal("help missing registered command", command.Name)
		}
	}

	f.send(t, "/project example-project", 0)

	firstPrompt := f.send(t, "/prepare first topic", 0)
	first := f.sessions(t)[0]
	f.send(t, "first part", firstPrompt.ID)
	f.send(t, "/develop second topic", 0)

	var second core.TelegramConversation

	for _, c := range f.sessions(t) {
		if c.ID != first.ID {
			second = c
		}
	}

	checkTelegramConversationRouting(t, f, first.ID, second.ID)
	checkTelegramConversationArtifacts(t, f)
	f.send(t, "/done "+second.ID, 0)
	f.send(t, "/reply "+second.ID+" late", 0)
	f.send(t, "/cancel "+first.ID, 0)

	if f.session(t, second.ID).Phase != "prepared" || f.session(t, second.ID).AuthoringSessionRef == "" || f.session(t, first.ID).Capture != "" {
		t.Fatal("completion/cancel crossed sessions")
	}

	if !strings.Contains(f.send(t, "/sessions", 0).Text, second.ID) {
		t.Fatal("session list missing target")
	}

	checkTelegramNoWork(t, f.receiver.env)
}

func checkTelegramConversationRouting(t *testing.T, f *telegramConversationAppFixture, first, second string) {
	t.Helper()

	if !strings.Contains(f.send(t, "ambiguous", 0).Text, "Select a session") {
		t.Fatal("ambiguous input chose a session")
	}

	f.send(t, "/focus "+first, 0)
	f.send(t, "second part", 0)

	notice := f.send(t, "/reply "+second+" explicit part", 0)
	f.send(t, "reply part", notice.ID)

	if f.session(t, first).Capture != "first part\nsecond part" || f.session(t, second).Capture != "explicit part\nreply part" {
		t.Fatal("targeting precedence failed", f.sessions(t))
	}

	before := f.session(t, second)
	f.updates <- f.last

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	if len(f.notices) != 0 || f.session(t, second) != before {
		t.Fatal("source replay duplicated output/capture")
	}

	if !strings.Contains(f.send(t, "/dump "+second, 0).Text, "explicit part\nreply part") {
		t.Fatal("capture view missing retained parts")
	}
}

func checkTelegramConversationArtifacts(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	projects, err := f.receiver.operations.ListRegisteredProjects([]string{"example-project", "disabled-project", "private-project"})
	if err != nil {
		t.Fatal(err)
	}

	for _, project := range projects {
		files, err := os.ReadDir(project.Path)
		if err != nil || len(files) != 0 {
			t.Fatal("conversation changed project artifacts", project.ID, files, err)
		}
	}
}

// TestTelegramConversationWireLimit checks the live decoder's inclusive 16-KiB boundary.
func TestTelegramConversationWireLimit(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	f.send(t, "/prepare --project example-project bounded topic", 0)

	c := f.sessions(t)[0]
	f.send(t, strings.Repeat("x", 16*1024), 0)

	if len(f.session(t, c.ID).Capture) != 16*1024 {
		t.Fatal("maximum message rejected")
	}

	if f.send(t, strings.Repeat("x", 16*1024+1), 0).ID != 0 {
		t.Fatal("oversized message dispatched")
	}

	if len(f.session(t, c.ID).Capture) != 16*1024 {
		t.Fatal("oversized message appended")
	}
}

// TestTelegramConversationRootChange rechecks the registry before a queued append.
func TestTelegramConversationRootChange(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	f.send(t, "/prepare --project example-project root topic", 0)

	c := f.sessions(t)[0]

	registry, err := core.NewProjectRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := registry.UpdateProjectPath(t.Context(), "example-project", t.TempDir()); err != nil {
		t.Fatal(err)
	}

	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(f.send(t, "must not append", 0).Text, "root changed") {
		t.Fatal("stale root not diagnosed")
	}

	if f.session(t, c.ID).Capture != "" {
		t.Fatal("stale root changed capture")
	}
}
