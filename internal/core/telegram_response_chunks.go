package core

import (
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// TelegramResponseParts renders plain text within both Telegram Unicode limits.
// Oversized output becomes a bounded summary with its retained canonical reference.
func TelegramResponseParts(text, reference string) []string {
	text = strings.ToValidUTF8(text, "�")

	reference = BoundTelegramReceiverText(reference)
	if len(reference) > 512 {
		reference = "retained request; use local diagnostics"
	}

	if text == "" {
		text = "No records."
	}

	parts := splitTelegramResponse(text)
	if len(parts) > 8 {
		return []string{BoundTelegramReceiverText("Summary: " + telegramResponsePrefix(text, 2800) + "\nFull output exceeds eight parts. Reference: " + reference)}
	}

	if len(parts) == 1 {
		return parts
	}

	for index := range parts {
		parts[index] = fmt.Sprintf("[%d/%d | %s]\n%s", index+1, len(parts), reference, parts[index])
	}

	return parts
}

func splitTelegramResponse(text string) []string {
	var parts []string
	for text != "" && len(parts) <= 8 {
		part := telegramResponsePrefix(text, 3200)
		parts = append(parts, part)
		text = text[len(part):]
	}

	return parts
}

func telegramResponsePrefix(text string, limit int) string {
	units, end := 0, 0
	for index, char := range text {
		units += utf16.RuneLen(char)
		if units > limit {
			break
		}

		end = index + utf8.RuneLen(char)
	}

	return text[:end]
}
