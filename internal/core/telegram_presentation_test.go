package core

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

// Optional blocks and literal delimiters must preserve the same visible text in
// rich and plain delivery, with offsets measured after Unicode-aware shortening.
func TestTelegramPresentationText(t *testing.T) {
	tests := []struct {
		name       string
		message    TelegramMessageBlocks
		text, kind string
	}{
		{"acknowledgement", TelegramMessageBlocks{InitialText: "Details saved."}, "Details saved.", ""},
		{"question first", TelegramMessageBlocks{Content: "Which platform?", SecondaryText: "Reply to this message."}, "Which platform?\n\nReply to this message.", ""},
		{"literal Markdown", TelegramMessageBlocks{InitialText: "Draft 😀", ContentKind: "markdown", Content: "# Title\n```go\nx := `a_b`\n```", SecondaryText: "Saved."}, "Draft 😀\n\n# Title\n```go\nx := `a_b`\n```\n\nSaved.", "pre"},
		{"inline code", TelegramMessageBlocks{ContentKind: "inline_code", Content: "a`b\\c[*]"}, "a`b\\c[*]", "code"},
		{"ordinary text", TelegramMessageBlocks{Content: "*plain* [text](x) <tag> & \\"}, "*plain* [text](x) <tag> & \\", ""},
		{"list", TelegramMessageBlocks{ContentKind: "list", Content: "- Alpha\n- Beta"}, "- Alpha\n- Beta", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := RenderTelegramMessage(test.message)
			if result.PlainFallback != test.text {
				t.Fatal("visible blocks or entity offsets changed", result)
			}

			if test.kind != "" && (len(result.Entities) != 1 || result.Entities[0].Type != test.kind || result.Entities[0].Offset != telegramPresentationContentOffset(result) || result.Entities[0].Length != len(utf16.Encode([]rune(test.message.Content)))) {
				t.Fatal("literal content lost its formatting", result)
			}
		})
	}
}

func telegramPresentationContentOffset(p TelegramPresentation) int {
	if p.InitialText == "" {
		return 0
	}

	return len(utf16.Encode([]rune(p.InitialText))) + 2
}

// Truncation preserves surrogate pairs and the next instruction. Re-rendering
// retained blocks cannot append another marker or shift the literal body.
func TestTelegramPresentationTextBounds(t *testing.T) {
	for _, test := range []struct{ name, text string }{
		{"astral", strings.Repeat("😀", 1801)},
		{"markup", strings.Repeat("`*_\\<&", 500)},
		{"invalid UTF8", strings.Repeat("\xff", 2000)},
	} {
		t.Run(test.name, func(t *testing.T) {
			message := TelegramMessageBlocks{InitialText: "Saved 😀", ContentKind: "markdown", Content: test.text, SecondaryText: "Send /done to finish."}

			result := RenderTelegramMessage(message)
			if !utf8.ValidString(result.PlainFallback) || len(utf16.Encode([]rune(result.PlainFallback))) > 1800 || !strings.HasSuffix(result.PlainFallback, message.SecondaryText) {
				t.Fatal("visible text exceeds its bounds or lost the next action", result)
			}

			if again := RenderTelegramMessage(result.TelegramMessageBlocks); again.PlainFallback != result.PlainFallback || !slices.Equal(again.Entities, result.Entities) {
				t.Fatal("retained blocks are not deterministic")
			}
		})
	}
}

// Real SQLite proves atomic acknowledgement and immutable source replay/reopen.
func TestTelegramPresentationFormatSnapshot(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	now := time.Now()

	update := telegramRetainedTestUpdate(t, 1, "show capture")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	original := TelegramMessageBlocks{ContentKind: "markdown", Content: "# Original\n```literal```"}
	for _, message := range []TelegramMessageBlocks{original, {Content: "replacement"}} {
		if err := store.QueuePresentationResponse(t.Context(), update, message, telegramReceiverTestConfig().Digest(), now); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 || due[0].Presentation == nil || due[0].Text != original.Content {
		t.Fatal("snapshot changed or duplicated", due, err)
	}

	checkTelegramPresentationFallback(t, store, due[0].ID, original.Content, now)
}

func checkTelegramPresentationFallback(t *testing.T, store *TelegramReceiverStore, id, original string, now time.Time) {
	t.Helper()

	selectTelegramHelpFallback(t, store, id, now)

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 || !due[0].PlainFallback || due[0].PlainText != original || due[0].Attempts != 1 {
		t.Fatal("fallback changed content or attempt allowance", due, err)
	}

	payload, err := json.Marshal(TelegramPresentation{Version: 2})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.db.ExecContext(t.Context(), "UPDATE telegram_outbox SET presentation=? WHERE id=?", string(payload), id); err != nil {
		t.Fatal(err)
	}

	if _, err := store.DueDeliveries(t.Context(), 12345, now); err == nil {
		t.Fatal("future presentation accepted")
	}
}
