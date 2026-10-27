package app

import (
	"reflect"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// Help must be readable by category rather than one unstructured command list.
func TestTelegramHelpPresentation(t *testing.T) {
	text := renderTelegramHelp(telegramHelpMetadata(), "").Text
	for _, heading := range []string{"Execution", "Tickets", "Projects", "Authoring", "Conversations", "System controls"} {
		if !strings.Contains(text, "*"+heading+"*\n") {
			t.Fatalf("missing category %q in help:\n%s", heading, text)
		}
	}

	if strings.Contains(text, " — ") || strings.Contains(text, "```") {
		t.Fatal("help contains descriptions or a large code block")
	}
}

// The live projection covers every command owner, including aliases and overloads.
func TestTelegramHelpMetadata(t *testing.T) {
	entries := telegramHelpMetadata()

	want := len(telegramReadRegistry()) + len(telegramConversationRegistry()) + len(telegramAuthoringRegistry()) + len(telegramControlCommands()) + len(telegramRecordRegistry()) + 5
	if len(entries) != want {
		t.Fatalf("metadata entries = %d, want %d", len(entries), want)
	}

	reply := renderTelegramHelp(telegramHelpMetadata(), "")
	for _, entry := range entries {
		if !slices.Contains(strings.Split(reply.PlainText, "\n"), entry.Syntax) || !strings.Contains(reply.Text, "`"+entry.Syntax+"`") {
			t.Fatalf("canonical syntax missing: %q", entry.Syntax)
		}
	}

	if len(tgmessage.ResponseParts(reply.Text, "help")) != 1 || len(tgmessage.ResponseParts(reply.PlainText, "help")) != 1 {
		t.Fatal("current catalogue splits during actual response retention")
	}

	t.Logf("%d entries; formatted %d UTF-16 units; plain %d UTF-16 units", len(entries), len(utf16.Encode([]rune(reply.Text))), len(utf16.Encode([]rune(reply.PlainText))))
}

// Injected definitions prove ordering, multiplicity and exact syntax without a shadow catalogue.
func TestTelegramHelpRendering(t *testing.T) {
	entries := []telegramHelpEntry{
		{"two", "Projects", "/z [value=*_]"},
		{"one", "Execution", "  /same\t<value>  "},
		{"duplicate", "Execution", "  /same\t<value>  "},
		{"new", "", "/unknown"},
	}
	original := slices.Clone(entries)
	reply := renderTelegramHelp(entries, "Unknown command.\n")

	want := "Unknown command.\n*Execution*\n`  /same\t<value>  `\n`  /same\t<value>  `\n\n*Projects*\n`/z [value=*_]`\n\n*Other*\n`/unknown`"
	if reply.Text != want || !reflect.DeepEqual(entries, original) {
		t.Fatalf("incorrect rendering or mutated input: %#v", reply)
	}

	slices.Reverse(entries)

	if renderTelegramHelp(entries, "Unknown command.\n") != reply {
		t.Fatal("input ordering changed the output")
	}

	entries[0].Syntax = "/replacement"
	if !strings.Contains(renderTelegramHelp(entries, "").PlainText, "/replacement") {
		t.Fatal("changed definition did not change help")
	}

	if renderTelegramHelp(nil, "") != (telegramReadReply{}) {
		t.Fatal("empty metadata produced a formatted reply")
	}
}

// Plain fallback preserves source punctuation rather than reverse-engineering markup.
func TestTelegramHelpEscaping(t *testing.T) {
	tests := []struct{ name, syntax, formatted string }{
		{"inline", "/command [some_value=*]", "`/command [some_value=*]`"},
		{"backticks", "/command `value`", "/command \\`value\\`"},
		{"backslash", `/command c:\value`, `/command c:\\value`},
		{"unicode", "/command 😀", "`/command 😀`"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reply := renderTelegramHelp([]telegramHelpEntry{{"injected", "Projects", test.syntax}}, "")
			if reply.Text != "*Projects*\n"+test.formatted || reply.PlainText != "Projects\n"+test.syntax {
				t.Fatalf("incorrect formatted/plain bodies: %#v", reply)
			}
		})
	}
}

// Both Unicode limits and markup expansion must fail as a complete bounded diagnostic.
func TestTelegramHelpOverflow(t *testing.T) {
	for _, test := range []struct{ name, syntax string }{
		{"pipeline limit", strings.Repeat("a", 3200)},
		{"UTF16 limit", strings.Repeat("😀", 1600)},
		{"escaping expansion", strings.Repeat("\\", 1700)},
	} {
		t.Run(test.name, func(t *testing.T) {
			reply := renderTelegramHelp([]telegramHelpEntry{{"large", "Other", test.syntax}}, "Unknown command.\n")
			if reply.Text != "Unknown command.\n"+telegramHelpOverflow || reply.PlainText != "" || !telegramHelpFits(reply.Text) {
				t.Fatalf("oversize help became a partial catalogue: %#v", reply)
			}
		})
	}
}
