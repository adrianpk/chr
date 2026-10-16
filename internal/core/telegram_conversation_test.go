package core

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type telegramConversationFixture struct {
	store *TelegramReceiverStore
	input TelegramConversationInput
	next  int64
}

func newTelegramConversationFixture(t *testing.T) *telegramConversationFixture {
	t.Helper()

	config := telegramReceiverTestConfig()
	config.AllowedUsers[0].ProjectIDs = []string{"alpha", "beta"}
	config.DefaultProjectID = ""

	return &telegramConversationFixture{store: openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t)), input: TelegramConversationInput{
		Config: config, Projects: []Project{{ID: "alpha", Name: "Alpha Project", Path: t.TempDir(), Enabled: true}, {ID: "beta", Name: "Beta Project", Path: t.TempDir(), Enabled: true}}, Intent: "prepare", Text: "Topic",
	}, next: 1}
}

func (f *telegramConversationFixture) apply(t *testing.T, action TelegramConversationAction, change func(*TelegramConversationInput)) {
	t.Helper()

	input := f.input
	input.Action = action
	input.Update = telegramRetainedTestUpdate(t, f.next, fmt.Sprintf("source %d", f.next))
	f.next++

	if change != nil {
		change(&input)
	}

	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ApplyTelegramConversation(t.Context(), input, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func (f *telegramConversationFixture) sessions(t *testing.T) []TelegramConversation {
	t.Helper()

	sessions, err := f.store.Conversations(t.Context(), TelegramPeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001})
	if err != nil {
		t.Fatal(err)
	}

	return sessions
}

func (f *telegramConversationFixture) lastResponse(t *testing.T) string {
	t.Helper()

	var text string
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT text FROM telegram_outbox ORDER BY rowid DESC LIMIT 1`).Scan(&text); err != nil {
		t.Fatal(err)
	}

	return text
}

// TestTelegramConversationProjects verifies each exact project selection precedence.
func TestTelegramConversationProjects(t *testing.T) {
	tests := []struct {
		name, explicit, selected, defaultID string
		only                                bool
		want                                string
	}{
		{name: "explicit ID", explicit: "beta", selected: "alpha", defaultID: "alpha", want: "beta"},
		{name: "exact name", explicit: "Beta Project", want: "beta"},
		{name: "selected", selected: "beta", defaultID: "alpha", want: "beta"},
		{name: "default", defaultID: "beta", want: "beta"},
		{name: "sole", only: true, want: "alpha"},
		{name: "ambiguous"},
		{name: "fuzzy refused", explicit: "Beta"},
		{name: "unauthorized", explicit: "private"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newTelegramConversationFixture(t)
			if test.selected != "" {
				f.apply(t, TelegramConversationSelect, func(i *TelegramConversationInput) { i.Project = test.selected })
			}

			f.input.Config.DefaultProjectID = test.defaultID
			if test.only {
				f.input.Config.AllowedUsers[0].ProjectIDs = []string{"alpha"}
			}

			f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = test.explicit })

			checkTelegramConversationProject(t, f, test.want)
		})
	}
}

// TestTelegramConversationBinding keeps namespace, project and intent independent of peer selection.
func TestTelegramConversationBinding(t *testing.T) {
	t.Setenv(OpsUserEnv, "first")

	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha"; i.Intent = "develop" })

	original := f.sessions(t)[0]
	f.apply(t, TelegramConversationSelect, func(i *TelegramConversationInput) { i.Project = "beta" })
	t.Setenv(OpsUserEnv, "second")
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Conversation = original.ID })

	sessions := f.sessions(t)
	if len(sessions) != 2 {
		t.Fatal(sessions)
	}

	for _, session := range sessions {
		if session.ProjectID != "alpha" {
			t.Fatal("bound conversation lost precedence", sessions)
		}
	}

	if f.session(t, original.ID) != original {
		t.Fatal("selection mutated existing conversation", sessions)
	}

	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Conversation = original.ID; i.Project = "beta" })

	if len(f.sessions(t)) != 2 || !strings.Contains(f.lastResponse(t), "conflicts") {
		t.Fatal("project mismatch was retargeted")
	}

	f.input.Projects[0].Path = t.TempDir()
	f.apply(t, TelegramConversationFocus, func(i *TelegramConversationInput) { i.Conversation = original.ID })

	if !strings.Contains(f.lastResponse(t), "root changed") {
		t.Fatal("changed root accepted")
	}
}

// TestTelegramConversationIsolation verifies exact group users, topics and allocation limits.
func TestTelegramConversationIsolation(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.input.Config.AllowedChatIDs = []int64{-100}
	f.input.Config.AllowedTopicIDs = []int64{1, 2}
	other := f.input.Config.AllowedUsers[0]
	other.TelegramUserID = 100002

	f.input.Config.AllowedUsers = append(f.input.Config.AllowedUsers, other)
	for _, peer := range []TelegramPeerIdentity{{12345, -100, 1, 100001}, {12345, -100, 1, 100002}, {12345, -100, 2, 100001}} {
		f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) {
			i.Project = "alpha"
			i.Update.Peer = peer
			i.Update.ChatType = "supergroup"
		})

		sessions, err := f.store.Conversations(t.Context(), peer)
		if err != nil || len(sessions) != 1 {
			t.Fatal("peer isolation", sessions, err)
		}
	}

	for range 9 {
		f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })
	}

	if len(f.sessions(t)) != 8 || !strings.Contains(f.lastResponse(t), "capacity") {
		t.Fatal("per-peer quota was exceeded")
	}
}

func checkTelegramConversationProject(t *testing.T, f *telegramConversationFixture, want string) {
	t.Helper()

	sessions := f.sessions(t)
	if want == "" {
		if len(sessions) != 0 {
			t.Fatal("ambiguous target allocated", sessions)
		}

		return
	}

	if len(sessions) != 1 || sessions[0].ProjectID != want {
		t.Fatal("wrong project", sessions)
	}
}

func (f *telegramConversationFixture) session(t *testing.T, id string) TelegramConversation {
	t.Helper()

	for _, c := range f.sessions(t) {
		if c.ID == id {
			return c
		}
	}

	t.Fatal("session unavailable", id)

	return TelegramConversation{}
}
