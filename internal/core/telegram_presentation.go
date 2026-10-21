package core

import (
	"strings"
	"unicode/utf16"
)

// TelegramMessageBlocks contains optional visible blocks, in reading order. ContentKind
// selects literal Markdown, block code, inline code or an ordinary text/list body.
type TelegramMessageBlocks struct {
	Family        string `json:"family,omitempty"`
	InitialText   string `json:"initial_text,omitempty"`
	ContentKind   string `json:"content_kind,omitempty"`
	Content       string `json:"content,omitempty"`
	SecondaryText string `json:"secondary_text,omitempty"`
}

// TelegramTextEntity addresses the final visible text in UTF-16 code units.
// Explicit entities keep arbitrary dynamic delimiters out of a markup parser.
type TelegramTextEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

// TelegramPresentation freezes the composed text and its formatting together.
// Version-zero deliveries continue to use their original outbox columns.
type TelegramPresentation struct {
	Version          int    `json:"version"`
	TemplateRevision string `json:"template_revision"`
	TelegramMessageBlocks
	PlainFallback string               `json:"plain_fallback"`
	Entities      []TelegramTextEntity `json:"entities,omitempty"`
}

const telegramPresentationLimit = 1800

// RenderTelegramMessage bounds visible blocks before computing entity offsets.
// A shortened body never consumes the space reserved for its next instruction.
func RenderTelegramMessage(message TelegramMessageBlocks) TelegramPresentation {
	message = boundTelegramMessage(normalizeTelegramMessage(message))

	result := TelegramPresentation{Version: 1, TemplateRevision: "text-v1", TelegramMessageBlocks: message}

	result.PlainFallback = joinTelegramMessageBlocks(message)
	if message.Content != "" {
		offset := 0
		if message.InitialText != "" {
			offset = telegramVisibleUnits(message.InitialText) + 2
		}

		result.Entities = telegramContentEntities(message, offset)
	}

	return result
}

func normalizeTelegramMessage(message TelegramMessageBlocks) TelegramMessageBlocks {
	message.InitialText = strings.TrimSpace(strings.ToValidUTF8(message.InitialText, "�"))
	message.Content = strings.ToValidUTF8(message.Content, "�")

	message.SecondaryText = strings.TrimSpace(strings.ToValidUTF8(message.SecondaryText, "�"))
	if strings.TrimSpace(message.Content) == "" {
		message.Content = ""
	}

	return message
}

func boundTelegramMessage(message TelegramMessageBlocks) TelegramMessageBlocks {
	secondaryLimit := telegramPresentationLimit - 64
	if message.InitialText != "" || message.Content != "" {
		secondaryLimit = 600
	}

	message.SecondaryText = boundTelegramPresentationBlock(message.SecondaryText, secondaryLimit)

	remaining := telegramPresentationLimit - 64 - telegramVisibleUnits(message.SecondaryText)
	if message.SecondaryText != "" && (message.InitialText != "" || message.Content != "") {
		remaining -= 2
	}

	initialLimit := remaining
	if message.Content != "" && message.InitialText != "" {
		initialLimit = min(600, remaining-2)
	}

	message.InitialText = boundTelegramPresentationBlock(message.InitialText, initialLimit)

	remaining -= telegramVisibleUnits(message.InitialText)
	if message.InitialText != "" && message.Content != "" {
		remaining -= 2
	}

	message.Content = boundTelegramPresentationBlock(message.Content, remaining)

	return message
}

func joinTelegramMessageBlocks(message TelegramMessageBlocks) string {
	var blocks []string

	for _, block := range []string{message.InitialText, message.Content, message.SecondaryText} {
		if block != "" {
			blocks = append(blocks, block)
		}
	}

	return strings.Join(blocks, "\n\n")
}

func telegramContentEntities(message TelegramMessageBlocks, offset int) []TelegramTextEntity {
	if message.Content == "" {
		return nil
	}

	kind := ""

	switch message.ContentKind {
	case "markdown", "code":
		kind = "pre"
	case "inline_code":
		kind = "code"
	}

	if kind == "" {
		return nil
	}

	return []TelegramTextEntity{{Type: kind, Offset: offset, Length: telegramVisibleUnits(message.Content)}}
}

func telegramVisibleUnits(text string) int {
	return len(utf16.Encode([]rune(text)))
}

func boundTelegramPresentationBlock(text string, limit int) string {
	if telegramVisibleUnits(text) <= limit {
		return text
	}

	const marker = "\n[shortened]"

	return telegramResponsePrefix(text, max(0, limit-len(marker))) + marker
}
