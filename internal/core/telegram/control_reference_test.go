package telegram

import "testing"

// TestTelegramControlReference rejects each foreign peer dimension without
// changing conversation focus or revision during exact control resolution.
func TestTelegramControlReference(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(input *ConversationInput) { input.Project = "alpha" })

	c := f.sessions(t)[0]
	peer := PeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001}

	for _, test := range []struct {
		name string
		peer PeerIdentity
	}{
		{name: "bot", peer: PeerIdentity{BotID: 2, ChatID: peer.ChatID, UserID: peer.UserID}},
		{name: "chat", peer: PeerIdentity{BotID: peer.BotID, ChatID: 2, UserID: peer.UserID}},
		{name: "topic", peer: PeerIdentity{BotID: peer.BotID, ChatID: peer.ChatID, TopicID: 2, UserID: peer.UserID}},
		{name: "user", peer: PeerIdentity{BotID: peer.BotID, ChatID: peer.ChatID, UserID: 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if conversationLookup, err := f.store.ConversationReference(t.Context(), test.peer, c.ID); err != nil || conversationLookup.Found {
				t.Fatal("foreign peer resolved a control conversation", conversationLookup.Found, err)
			}
		})
	}

	if conversationLookup2, err := f.store.ConversationReference(t.Context(), peer, c.ID); err != nil || !conversationLookup2.Found || conversationLookup2.Conversation.Revision != c.Revision {
		t.Fatal("exact reference changed the conversation", conversationLookup2.Conversation, conversationLookup2.Found, err)
	}
}
