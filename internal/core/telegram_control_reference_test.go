package core

import "testing"

// TestTelegramControlReference rejects each foreign peer dimension without
// changing conversation focus or revision during exact control resolution.
func TestTelegramControlReference(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(input *TelegramConversationInput) { input.Project = "alpha" })

	c := f.sessions(t)[0]
	peer := TelegramPeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001}

	for _, test := range []struct {
		name string
		peer TelegramPeerIdentity
	}{
		{name: "bot", peer: TelegramPeerIdentity{BotID: 2, ChatID: peer.ChatID, UserID: peer.UserID}},
		{name: "chat", peer: TelegramPeerIdentity{BotID: peer.BotID, ChatID: 2, UserID: peer.UserID}},
		{name: "topic", peer: TelegramPeerIdentity{BotID: peer.BotID, ChatID: peer.ChatID, TopicID: 2, UserID: peer.UserID}},
		{name: "user", peer: TelegramPeerIdentity{BotID: peer.BotID, ChatID: peer.ChatID, UserID: 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, found, err := f.store.TelegramConversationReference(t.Context(), test.peer, c.ID); err != nil || found {
				t.Fatal("foreign peer resolved a control conversation", found, err)
			}
		})
	}

	if current, found, err := f.store.TelegramConversationReference(t.Context(), peer, c.ID); err != nil || !found || current.Revision != c.Revision {
		t.Fatal("exact reference changed the conversation", current, found, err)
	}
}
