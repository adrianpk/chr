package telegram

import (
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"
)

// Capture diagnostics must preserve prior text at real persistence boundaries;
// accepted dynamic text retains its bytes while the visible response stays short.
func TestTelegramPresentationCoverageCapture(t *testing.T) {
	for _, item := range []telegramCapturePresentationCase{
		{"ordinary", "literal `*_` 😀\n\n", "Details saved", 1, false, true},
		{"near part limit", "last detail", "Near the capture limit: 28 of 32 parts", 27, false, true},
		{"invalid encoding", string([]byte{0xff}), "valid UTF-8 text of at most 16 KiB", 1, false, false},
		{"oversized detail", strings.Repeat("x", 16385), "Saved details are unchanged", 1, false, false},
		{"finished capture", "late text", "This capture is finished", 1, true, false},
	} {
		t.Run(item.name, func(t *testing.T) { checkTelegramCapturePresentation(t, item) })
	}
}

type telegramCapturePresentationCase struct {
	name, text, want string
	parts            int
	finish, accepted bool
}

func checkTelegramCapturePresentation(t *testing.T, item telegramCapturePresentationCase) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

	for range item.parts {
		f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Text = "retained detail" })
	}

	if item.finish {
		f.apply(t, TelegramConversationDone, nil)
	}

	before := f.sessions(t)[0]
	f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Text = item.text; i.Conversation = before.ID })

	after := f.sessions(t)[0]
	if item.accepted {
		if after.Capture != before.Capture+"\n"+item.text {
			t.Fatal("accepted input lost original bytes", after.Capture)
		}
	} else if after != before {
		t.Fatal("rejected detail changed saved state", before, after)
	}

	text := f.lastResponse(t)
	if !strings.Contains(text, item.want) || !utf8.ValidString(text) || len(utf16.Encode([]rune(text))) > 1800 {
		t.Fatal("capture outcome is misleading or unbounded", text)
	}
}
