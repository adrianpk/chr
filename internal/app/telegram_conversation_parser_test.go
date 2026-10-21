package app

import (
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramConversationCommands checks deterministic grammar without consulting semantic resolution.
func TestTelegramConversationCommands(t *testing.T) {
	tests := []struct {
		name, text        string
		recognized, valid bool
	}{
		{"plain", "raw\ntext", true, true},
		{"project name", "/project Exact Name", true, true},
		{"prepare project", "/prepare --project example-project topic", true, true},
		{"develop", "/develop topic", true, true},
		{"missing topic", "/prepare", true, false},
		{"unknown option", "/prepare --shell run", true, false},
		{"focus", "/focus opaque-id", true, true},
		{"missing focus", "/focus", true, false},
		{"reply", "/reply opaque-id message", true, true},
		{"missing reply", "/reply opaque-id", true, false},
		{"list arguments", "/sessions extra", true, false},
		{"dump", "/dump", true, true},
		{"done", "/done opaque-id", true, true},
		{"cancel", "/cancel opaque-id", true, true},
		{"unknown slash", "/invent do everything", false, true},
		{"unavailable authoring", "/quiz", false, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, recognized, diagnostic := telegramConversationInput(core.TelegramReceivedUpdate{Kind: "message", Text: test.text})
			if recognized != test.recognized || (diagnostic == "") != test.valid {
				t.Fatal("incorrect grammar", recognized, diagnostic)
			}
		})
	}
}

type telegramConversationNoReads struct{}

func (telegramConversationNoReads) ListRegisteredProjects([]string) ([]core.Project, error) {
	panic("unexpected registry read")
}

func (telegramConversationNoReads) InspectTelegramReceiver(int64) (core.TelegramReceiverState, error) {
	panic("unexpected receiver read")
}

// TestTelegramConversationUnknown returns registry help without invoking any service or model operation.
func TestTelegramConversationUnknown(t *testing.T) {
	receiver, config := telegramAppTestReceiver(t, core.TelegramReceiverAPI{})
	receiver.operations = telegramConversationNoReads{}
	update := core.TelegramReceivedUpdate{UpdateID: 1, Digest: strings.Repeat("a", 64), Kind: "message", Peer: core.TelegramPeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001}, MessageID: 1, ChatType: "private", Text: "/invent"}

	text, err := telegramReadResponse(receiver.operations, config.AllowedUsers[0], update)
	if err != nil || !strings.Contains(text.Text, "could not use that command") || !strings.Contains(text.Text, "/help") {
		t.Fatal("unknown command did not return help", text, err)
	}
}
