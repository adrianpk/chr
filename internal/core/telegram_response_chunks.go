package core

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// TelegramResponseParts bounds newly composed responses to one short view.
// Existing multipart rows retain their stored text, order and retry guards.
func TelegramResponseParts(text, reference string) []string {
	text = strings.ToValidUTF8(text, "�")
	if strings.TrimSpace(text) == "" {
		text = "No records."
	}

	message := TelegramMessageBlocks{Family: "44", Content: text}
	if telegramVisibleUnits(text) > telegramPresentationLimit-64 {
		message.SecondaryText = "Output shortened. Narrow your request or inspect the original operation locally.\nReference: " + boundTelegramPresentationBlock(strings.ToValidUTF8(reference, "�"), 256)
	}

	return []string{RenderTelegramMessage(message).PlainFallback}
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
