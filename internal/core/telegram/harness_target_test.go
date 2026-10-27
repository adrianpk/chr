package telegram

import (
	"testing"
	"time"
)

// Harness reads require an explicit selected project, including when a default
// exists. Repeated inspection must not consume conversation allocation capacity.
func TestTelegramHarnessSelectionTarget(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.input.Update = telegramRetainedTestUpdate(t, 99, "/harness")

	f.input.Config.DefaultProjectID = "alpha"
	if _, err := f.store.ResolveTelegramHarnessTarget(t.Context(), f.input, time.Now()); err == nil {
		t.Fatal("default project replaced explicit selection")
	}

	f.apply(t, TelegramConversationSelect, func(input *TelegramConversationInput) { input.Project = "alpha" })

	for range 12 {
		target, err := f.store.ResolveTelegramHarnessTarget(t.Context(), f.input, time.Now())
		if err != nil || target.ProjectID != "alpha" || len(target.ProjectBindingDigest) != 64 || target.ConversationID != "" {
			t.Fatal("selected project was not bound", target, err)
		}
	}

	if len(f.sessions(t)) != 0 {
		t.Fatal("inspection allocated a confirmation conversation")
	}
}

// Explicit overrides, stale replies and foreign peers cannot replace the
// current selected project's authority at the configuration boundary.
func TestTelegramHarnessSelectionScope(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.input.Update = telegramRetainedTestUpdate(t, 99, "/harness")
	f.apply(t, TelegramConversationSelect, func(input *TelegramConversationInput) { input.Project = "alpha" })

	for _, test := range []struct {
		name   string
		change func(*TelegramConversationInput)
	}{
		{"project override", func(input *TelegramConversationInput) { input.Project = "beta" }},
		{"conversation override", func(input *TelegramConversationInput) { input.Conversation = "foreign-session" }},
		{"namespace override", func(input *TelegramConversationInput) { input.Namespace = "foreign" }},
		{"reply target", func(input *TelegramConversationInput) { input.Update.ReplyToMessageID = 10 }},
		{"foreign peer", func(input *TelegramConversationInput) { input.Update.Peer.UserID++ }},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := f.input
			test.change(&input)

			if _, err := f.store.ResolveTelegramHarnessTarget(t.Context(), input, time.Now()); err == nil {
				t.Fatal("foreign input resolved the selected project")
			}
		})
	}
}
