package message

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// ResponseParts bounds newly composed responses to one short view.
// Existing multipart rows retain their stored text, order and retry guards.
func ResponseParts(text, reference string) []string {
	text = strings.ToValidUTF8(text, "�")
	if strings.TrimSpace(text) == "" {
		text = "No records."
	}

	message := Blocks{Family: "44", Content: text}
	if VisibleUnits(text) > Limit-64 {
		message.SecondaryText = "Output shortened. Narrow your request or inspect the original operation locally.\nReference: " + BoundBlock(strings.ToValidUTF8(reference, "�"), 256)
	}

	return []string{Render(message).PlainFallback}
}

// SplitResponse retains at most nine compatibility chunks with the existing unit limit.
func SplitResponse(text string) []string {
	var parts []string
	for text != "" && len(parts) <= 8 {
		part := Prefix(text, 3200)
		parts = append(parts, part)
		text = text[len(part):]
	}

	return parts
}

// Prefix retains complete runes within the UTF-16 unit limit.
func Prefix(text string, limit int) string {
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
