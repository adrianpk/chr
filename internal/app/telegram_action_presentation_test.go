package app

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf16"
)

// TestTelegramPresentationActionWork checks the actual HTTP preview against the
// native owner: human scope and consequences are visible before any admission.
func TestTelegramPresentationActionWork(t *testing.T) {
	f, operations, record := telegramWorkFixture(t)

	action, prompt := telegramWorkOffer(t, f, "/work ticket "+record.Task.Key)
	for _, part := range []string{record.Task.Key, "Selected tickets:", "Publication:", "Limits:", "validation", "Confirming admits"} {
		if !strings.Contains(prompt.Text, part) {
			t.Fatal("start preview omitted scope or consequence", part, prompt)
		}
	}

	for _, audit := range []string{"Snapshot:", "Policy revision:", "Target commit:"} {
		if strings.Contains(prompt.Text, audit) {
			t.Fatal("routine caption exposed an audit field", audit)
		}
	}

	checkTelegramConfirmationUpload(t, f, record.Domain["execution_contract"], action.Token)
	checkTelegramWorkRunCount(t, operations, 0)
}

// TestTelegramPresentationControlPreview distinguishes accepting a pause request
// from quiescence and verifies that presenting the control does not change work.
func TestTelegramPresentationControlPreview(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)
	before := checkTelegramWorkRunCount(t, operations, 1)

	prompt := f.send(t, "/pause "+ref.Invocation.ID, 0)
	if !strings.Contains(prompt.Text, "Request pause") || !strings.Contains(prompt.Text, "worker may still be stopping") {
		t.Fatal("pause preview claimed a completed stop", prompt)
	}

	after := checkTelegramWorkRunCount(t, operations, 1)
	if after.Revision != before.Revision || after.StopIntent != before.StopIntent {
		t.Fatal("preview changed the native invocation")
	}
}

func checkTelegramConfirmationUpload(t *testing.T, f *telegramConversationAppFixture, contract []byte, token string) {
	t.Helper()

	select {
	case upload := <-f.documents:
		if upload.Filename != "confirmation-scope.md" || !bytes.Contains(upload.Payload, contract) || len(utf16.Encode([]rune(upload.Fields["caption"]))) > 900 {
			t.Fatal("confirmation omitted the complete contract or exceeded its caption bound", upload.Filename)
		}

		checkTelegramConfirmationKeyboard(t, upload.Fields["reply_markup"], token)
	default:
		t.Fatal("start confirmation did not include its complete scope document")
	}
}

func checkTelegramConfirmationKeyboard(t *testing.T, raw, token string) {
	t.Helper()

	var keyboard struct {
		Rows [][]struct {
			Text string `json:"text"`
			Data string `json:"callback_data"`
		} `json:"inline_keyboard"`
	}
	if err := json.Unmarshal([]byte(raw), &keyboard); err != nil {
		t.Fatal(err)
	}

	if len(keyboard.Rows) != 1 || len(keyboard.Rows[0]) != 1 || keyboard.Rows[0][0].Text != "Start work" || keyboard.Rows[0][0].Data != "tga:"+token+":start" {
		t.Fatal("upload changed the original action binding", keyboard)
	}
}
