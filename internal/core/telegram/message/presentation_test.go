package message

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"
)

// Optional blocks and literal delimiters must preserve the same visible text in
// rich and plain delivery, with offsets measured after Unicode-aware shortening.
func TestTelegramPresentationText(t *testing.T) {
	tests := []struct {
		name       string
		message    Blocks
		text, kind string
	}{
		{"acknowledgement", Blocks{InitialText: "Details saved."}, "Details saved.", ""},
		{"question first", Blocks{Content: "Which platform?", SecondaryText: "Reply to this message."}, "Which platform?\n\nReply to this message.", ""},
		{"literal Markdown", Blocks{InitialText: "Draft 😀", ContentKind: "markdown", Content: "# Title\n```go\nx := `a_b`\n```", SecondaryText: "Saved."}, "Draft 😀\n\n# Title\n```go\nx := `a_b`\n```\n\nSaved.", "pre"},
		{"inline code", Blocks{ContentKind: "inline_code", Content: "a`b\\c[*]"}, "a`b\\c[*]", "code"},
		{"ordinary text", Blocks{Content: "*plain* [text](x) <tag> & \\"}, "*plain* [text](x) <tag> & \\", ""},
		{"list", Blocks{ContentKind: "list", Content: "- Alpha\n- Beta"}, "- Alpha\n- Beta", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Render(test.message)
			if result.PlainFallback != test.text {
				t.Fatal("visible blocks or entity offsets changed", result)
			}

			if test.kind != "" && (len(result.Entities) != 1 || result.Entities[0].Type != test.kind || result.Entities[0].Offset != telegramPresentationContentOffset(result) || result.Entities[0].Length != len(utf16.Encode([]rune(test.message.Content)))) {
				t.Fatal("literal content lost its formatting", result)
			}
		})
	}
}

func telegramPresentationContentOffset(p Presentation) int {
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
			message := Blocks{InitialText: "Saved 😀", ContentKind: "markdown", Content: test.text, SecondaryText: "Send /done to finish."}

			result := Render(message)
			if !utf8.ValidString(result.PlainFallback) || len(utf16.Encode([]rune(result.PlainFallback))) > 1800 || !strings.HasSuffix(result.PlainFallback, message.SecondaryText) {
				t.Fatal("visible text exceeds its bounds or lost the next action", result)
			}

			if again := Render(result.Blocks); again.PlainFallback != result.PlainFallback || !slices.Equal(again.Entities, result.Entities) {
				t.Fatal("retained blocks are not deterministic")
			}
		})
	}
}
