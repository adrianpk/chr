package core

import (
	"strings"
	"testing"
	"time"
)

func acknowledgeTelegramConversation(t *testing.T, f *telegramConversationFixture, messageID int64) {
	t.Helper()

	var id string
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT id FROM telegram_outbox ORDER BY rowid DESC LIMIT 1`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ReserveDelivery(t.Context(), id, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.DeliverySent(t.Context(), id, messageID, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func bindTelegramTestQuestion(t *testing.T, f *telegramConversationFixture, c TelegramConversation, questionID string) {
	t.Helper()

	input := f.input
	input.Update = telegramRetainedTestUpdate(t, 999, "binding")
	input.Conversation = c.ID

	question := TelegramConversationQuestion{SessionRef: "canonical-session-" + c.ID, QuestionID: questionID, Revision: 1, ContentDigest: strings.Repeat("a", 64)}
	if err := f.store.BindTelegramConversationQuestion(t.Context(), input, c.Revision, question, time.Now()); err != nil {
		t.Fatal(err)
	}
}

// TestTelegramConversationReplies binds definite acknowledgements to exact revisions and sessions.
func TestTelegramConversationReplies(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

	first := f.sessions(t)[0]
	bindTelegramTestQuestion(t, f, first, "question-a")
	f.apply(t, TelegramConversationShow, func(i *TelegramConversationInput) { i.Conversation = first.ID })
	acknowledgeTelegramConversation(t, f, 100)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "beta" })

	var second TelegramConversation

	for _, c := range f.sessions(t) {
		if c.ID != first.ID {
			second = c
		}
	}

	bindTelegramTestQuestion(t, f, second, "question-b")
	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Update.ReplyToMessageID = 100; i.Text = "answer a" })

	a, b := f.session(t, first.ID), f.session(t, second.ID)
	if a.Capture != "answer a" || b.QuestionBinding == "" || b.Capture != "" {
		t.Fatal("reply crossed sessions", a, b)
	}

	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) {
		i.Update.ReplyToMessageID = 100
		i.Text = "stale"
		i.Conversation = second.ID
	})

	if !strings.Contains(f.lastResponse(t), "unavailable") || f.session(t, first.ID).Capture != "answer a" {
		t.Fatal("stale reply fell back to another session")
	}

	f.apply(t, TelegramConversationShow, func(i *TelegramConversationInput) { i.Conversation = second.ID })
	acknowledgeTelegramConversation(t, f, 101)

	second = f.session(t, second.ID)
	bindTelegramTestQuestion(t, f, second, "replacement")
	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Update.ReplyToMessageID = 101; i.Text = "old question" })

	if !strings.Contains(f.lastResponse(t), "unavailable") {
		t.Fatal("replaced question answered")
	}
}

// TestTelegramConversationReplyGuards refuses cross-user/topic, mismatched and terminal prompts.
func TestTelegramConversationReplyGuards(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

	c := f.sessions(t)[0]
	acknowledgeTelegramConversation(t, f, 200)
	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) {
		i.Update.ReplyToMessageID = 200
		i.Project = "beta"
		i.Text = "mismatch"
	})

	if !strings.Contains(f.lastResponse(t), "conflicts") {
		t.Fatal("reply project mismatch accepted")
	}

	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) {
		i.Update.ReplyToMessageID = 200
		i.Conversation = "different"
		i.Text = "mismatch"
	})

	if !strings.Contains(f.lastResponse(t), "conflicts") {
		t.Fatal("reply conversation mismatch accepted")
	}

	f.apply(t, TelegramConversationDone, func(i *TelegramConversationInput) { i.Conversation = c.ID })
	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Update.ReplyToMessageID = 200; i.Text = "late" })

	if f.sessions(t)[0].Capture != "" {
		t.Fatal("terminal reply accepted")
	}
}

// TestTelegramConversationFocus preserves ambiguity until a peer explicitly chooses a session.
func TestTelegramConversationFocus(t *testing.T) {
	f := newTelegramConversationFixture(t)
	for range 2 {
		f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })
	}

	sessions := f.sessions(t)
	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Text = "ambiguous" })

	if !strings.Contains(f.lastResponse(t), "Select a session") {
		t.Fatal("ambiguous text guessed a session")
	}

	f.apply(t, TelegramConversationFocus, func(i *TelegramConversationInput) { i.Conversation = sessions[1].ID })
	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Text = "focused" })

	if f.session(t, sessions[1].ID).Capture != "focused" {
		t.Fatal("focus not used")
	}

	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Conversation = sessions[0].ID; i.Text = "explicit" })

	if f.session(t, sessions[0].ID).Capture != "explicit" {
		t.Fatal("explicit target lost precedence")
	}
}
