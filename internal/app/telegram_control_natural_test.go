package app

import (
	"testing"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramControlPayload rejects extra model authority fields and malformed
// explicit grants or answers before any native mutation can be prepared.
func TestTelegramControlPayload(t *testing.T) {
	for _, test := range []struct {
		name, command string
		valid         bool
	}{
		{"grant", "/grant run submission 1 More capacity", true},
		{"negative grant", "/grant run submission -1 More capacity", false},
		{"overflow grant", "/grant run tokens 999999999999999999999 More capacity", false},
		{"unknown grant", "/grant run anything 1 More capacity", false},
		{"missing reason", "/grant run submission 1", false},
		{"answer", "/answer-decision run question resolve 1 Preserve intent", true},
		{"missing option", "/answer-decision run question resolve 0 Preserve intent", false},
		{"unknown answer", "/answer-decision run question approve 1 Preserve intent", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, recognized, err := parseTelegramControl(telegramcore.TelegramReceivedUpdate{Kind: "message", Text: test.command})
			if !recognized || (err == nil) != test.valid {
				t.Fatal("unexpected bounded control parse", recognized, err)
			}
		})
	}

	_, _, err := telegramIntentControl(telegramcore.TelegramReceivedUpdate{}, ConversationIntent{OperationID: "control", Arguments: ConversationIntentArguments{Action: "grant", Text: "run", RecordJSON: `{"kind":"submission","amount":1,"reason":"More capacity","actor":"model-chosen"}`}})
	if err == nil {
		t.Fatal("model supplied mutation authority")
	}
}
