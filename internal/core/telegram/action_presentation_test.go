package telegram

import (
	"strings"
	"testing"
)

// TestTelegramPresentationActionLabels keeps human button labels separate from
// the opaque choice keys consumed by the original action owner.
func TestTelegramPresentationActionLabels(t *testing.T) {
	for _, test := range []struct{ choice, label string }{
		{"start", "Start work"},
		{"develop", "Prepare ticket"},
		{"confirm", "Confirm change"},
		{"pause", "Pause work"},
		{"continue", "Continue one step"},
		{"cancel", "Cancel"},
	} {
		t.Run(test.choice, func(t *testing.T) {
			token := strings.Repeat("a", 32)

			keyboard, err := actionKeyboard(token, []string{test.choice})
			if err != nil {
				t.Fatal(err)
			}

			button := keyboard.Rows[0][0]
			if button.Text != test.label || button.CallbackData != "tga:"+token+":"+test.choice {
				t.Fatal("visible label changed the retained callback binding", button)
			}
		})
	}
}
