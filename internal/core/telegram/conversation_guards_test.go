package telegram

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestTelegramConversationReplyPeers refuses another user or topic even with the same bot message ID.
func TestTelegramConversationReplyPeers(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.input.Config.AllowedChatIDs = []int64{-100}
	f.input.Config.AllowedTopicIDs = []int64{1, 2}
	other := f.input.Config.AllowedUsers[0]
	other.TelegramUserID = 100002
	f.input.Config.AllowedUsers = append(f.input.Config.AllowedUsers, other)
	owner := PeerIdentity{12345, -100, 1, 100001}

	f.apply(t, ConversationStart, func(i *ConversationInput) {
		i.Project = "alpha"
		i.Update.Peer = owner
		i.Update.ChatType = "supergroup"
	})
	acknowledgeTelegramConversation(t, f, 300)

	for _, peer := range []PeerIdentity{{12345, -100, 1, 100002}, {12345, -100, 2, 100001}} {
		f.apply(t, ConversationCapture, func(i *ConversationInput) {
			i.Text = "crossed"
			i.Update.Peer = peer
			i.Update.ChatType = "supergroup"
			i.Update.ReplyToMessageID = 300
		})

		if !strings.Contains(f.lastResponse(t), "unavailable") {
			t.Fatal("cross-peer reply accepted", peer)
		}
	}

	sessions, err := f.store.Conversations(t.Context(), owner)
	if err != nil || len(sessions) != 1 || sessions[0].Capture != "" {
		t.Fatal("owner session changed", sessions, err)
	}
}

// TestTelegramConversationReplyRetention limits usable associations and rejects an unacknowledged prompt.
func TestTelegramConversationReplyRetention(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(i *ConversationInput) { i.Project = "alpha" })

	c := f.sessions(t)[0]
	f.apply(t, ConversationCapture, func(i *ConversationInput) { i.Update.ReplyToMessageID = 777; i.Text = "before acknowledgement" })

	if !strings.Contains(f.lastResponse(t), "unavailable") {
		t.Fatal("unacknowledged reply routed")
	}

	for n := range 65 {
		f.apply(t, ConversationShow, func(i *ConversationInput) { i.Conversation = c.ID })
		acknowledgeTelegramConversation(t, f, int64(1000+n))
	}

	var count int
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_replies WHERE conversation_id=? AND invalidated_at IS NULL`, c.ID).Scan(&count); err != nil || count != 64 {
		t.Fatal("reply retention bound", count, err)
	}

	f.apply(t, ConversationCapture, func(i *ConversationInput) { i.Update.ReplyToMessageID = 1000; i.Text = "pruned" })

	if f.session(t, c.ID).Capture != "" {
		t.Fatal("pruned association usable")
	}

	f.apply(t, ConversationCapture, func(i *ConversationInput) { i.Update.ReplyToMessageID = 1064; i.Text = "current" })

	if f.session(t, c.ID).Capture != "current" {
		t.Fatal("current association unavailable")
	}
}

// TestTelegramConversationBotQuota enforces one bot-wide allocation bound across distinct peers.
func TestTelegramConversationBotQuota(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.input.Config.AllowedChatIDs = []int64{-100}

	f.input.Config.AllowedUsers = nil
	for n := range 17 {
		f.input.Config.AllowedUsers = append(f.input.Config.AllowedUsers, AllowedUser{TelegramUserID: int64(100001 + n), OperatorID: fmt.Sprintf("operator-%d", n), ProjectIDs: []string{"alpha"}})
	}

	for n := range 129 {
		f.apply(t, ConversationStart, func(i *ConversationInput) {
			i.Project = "alpha"
			i.Update.Peer = PeerIdentity{12345, -100, 0, int64(100001 + n/8)}
			i.Update.ChatType = "supergroup"
		})
	}

	var count int
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_conversations`).Scan(&count); err != nil || count != 128 {
		t.Fatal("bot allocation bound", count, err)
	}

	if !strings.Contains(f.lastResponse(t), "capacity") {
		t.Fatal("no allocation diagnostic")
	}
}

// TestTelegramConversationQuestionCAS refuses a replaced observed question revision.
func TestTelegramConversationQuestionCAS(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(i *ConversationInput) { i.Project = "alpha" })

	c := f.sessions(t)[0]
	bindTelegramTestQuestion(t, f, c, "first")

	input := f.input
	input.Conversation = c.ID
	input.Update = telegramRetainedTestUpdate(t, 999, "question")

	question := ConversationQuestion{SessionRef: "session", QuestionID: "replacement", Revision: 2, ContentDigest: strings.Repeat("b", 64)}
	if err := f.store.BindConversationQuestion(t.Context(), input, c.Revision, question, time.Now()); err == nil {
		t.Fatal("stale revision accepted")
	}

	if !strings.Contains(f.session(t, c.ID).QuestionBinding, "first") {
		t.Fatal("question replaced without guard")
	}

	f.apply(t, ConversationDone, nil)
	f.apply(t, ConversationCancel, nil)

	if f.session(t, c.ID).Phase != "canceled" {
		t.Fatal("unstarted finished capture cannot be canceled")
	}
}

// TestTelegramConversationConcurrent replays one source through independent database handles without duplicate capture.
func TestTelegramConversationConcurrent(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(i *ConversationInput) { i.Project = "alpha" })

	second := openTelegramReceiverTestStore(t, f.env)
	input := f.input
	input.Update = telegramRetainedTestUpdate(t, 900, "concurrent part")

	input.Action, input.Text = ConversationCapture, "concurrent part"
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)

	for _, store := range []*ReceiverStore{f.store, second} {
		go func() { results <- store.ApplyConversation(t.Context(), input, time.Now()) }()
	}

	firstError, secondError := <-results, <-results
	if firstError != nil || secondError != nil {
		t.Fatal("concurrent replay failed", firstError, secondError)
	}

	if f.sessions(t)[0].Capture != "concurrent part" {
		t.Fatal("duplicate concurrent append")
	}

	var count int
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_outbox`).Scan(&count); err != nil || count != 2 {
		t.Fatal("duplicate concurrent response", count, err)
	}
}

// TestTelegramConversationProjectNames refuses ambiguous names and unauthorized defaults.
func TestTelegramConversationProjectNames(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*ConversationInput)
	}{
		{"ambiguous name", func(i *ConversationInput) {
			i.Projects[0].Name = "Same"
			i.Projects[1].Name = "Same"
			i.Project = "Same"
		}},
		{"unauthorized default", func(i *ConversationInput) { i.Config.DefaultProjectID = "private" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newTelegramConversationFixture(t)
			f.apply(t, ConversationStart, test.configure)

			if len(f.sessions(t)) != 0 {
				t.Fatal("ambiguous project guessed")
			}
		})
	}
}
