// Package message owns bounded Telegram presentation and complete document values.
package message

import (
	"strings"
	"unicode/utf16"
)

// Blocks contains optional visible blocks, in reading order. ContentKind
// selects literal Markdown, block code, inline code or an ordinary text/list body.
type Blocks struct {
	Family        string `json:"family,omitempty"`
	InitialText   string `json:"initial_text,omitempty"`
	ContentKind   string `json:"content_kind,omitempty"`
	Content       string `json:"content,omitempty"`
	SecondaryText string `json:"secondary_text,omitempty"`
}

// Entity addresses the final visible text in UTF-16 code units.
// Explicit entities keep arbitrary dynamic delimiters out of a markup parser.
type Entity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

// Presentation freezes the composed text and its formatting together.
// Version-zero deliveries continue to use their original outbox columns.
type Presentation struct {
	Version          int    `json:"version"`
	TemplateRevision string `json:"template_revision"`
	Blocks
	ActionRef     string   `json:"action_ref,omitempty"`
	ScopeDigest   string   `json:"scope_digest,omitempty"`
	ActionChoices []string `json:"action_choices,omitempty"`
	ActionLabels  []string `json:"action_labels,omitempty"`
	DocumentID    string   `json:"document_id,omitempty"`
	DocumentHash  string   `json:"document_hash,omitempty"`
	PlainFallback string   `json:"plain_fallback"`
	Entities      []Entity `json:"entities,omitempty"`
}

// Limit bounds the existing text presentation in UTF-16 units.
const Limit = 1800

// Render bounds visible blocks before computing entity offsets.
// A shortened body never consumes the space reserved for its next instruction.
func Render(message Blocks) Presentation {
	return render(message, Limit, "text-v1")
}

func render(message Blocks, limit int, revision string) Presentation {
	message = bound(normalize(message), limit)

	result := Presentation{Version: 1, TemplateRevision: revision, Blocks: message}

	result.PlainFallback = joinBlocks(message)
	if message.Content != "" {
		offset := 0
		if message.InitialText != "" {
			offset = VisibleUnits(message.InitialText) + 2
		}

		result.Entities = contentEntities(message, offset)
	}

	return result
}

func normalize(message Blocks) Blocks {
	message.InitialText = strings.TrimSpace(strings.ToValidUTF8(message.InitialText, "�"))
	message.Content = strings.ToValidUTF8(message.Content, "�")

	message.SecondaryText = strings.TrimSpace(strings.ToValidUTF8(message.SecondaryText, "�"))
	if strings.TrimSpace(message.Content) == "" {
		message.Content = ""
	}

	return message
}

func bound(message Blocks, limit int) Blocks {
	blockLimit := blockLimit(limit)

	secondaryLimit := limit - 64
	if message.InitialText != "" || message.Content != "" {
		secondaryLimit = blockLimit
	}

	message.SecondaryText = BoundBlock(message.SecondaryText, secondaryLimit)

	remaining := limit - 64 - VisibleUnits(message.SecondaryText)
	if message.SecondaryText != "" && (message.InitialText != "" || message.Content != "") {
		remaining -= 2
	}

	initialLimit := remaining
	if message.Content != "" && message.InitialText != "" {
		initialLimit = min(blockLimit, remaining-2)
	}

	message.InitialText = BoundBlock(message.InitialText, initialLimit)

	remaining -= VisibleUnits(message.InitialText)
	if message.InitialText != "" && message.Content != "" {
		remaining -= 2
	}

	message.Content = BoundBlock(message.Content, remaining)

	return message
}

func joinBlocks(message Blocks) string {
	var blocks []string

	for _, block := range []string{message.InitialText, message.Content, message.SecondaryText} {
		if block != "" {
			blocks = append(blocks, block)
		}
	}

	return strings.Join(blocks, "\n\n")
}

func contentEntities(message Blocks, offset int) []Entity {
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

	return []Entity{{Type: kind, Offset: offset, Length: VisibleUnits(message.Content)}}
}

// VisibleUnits counts text in the UTF-16 units used by Telegram.
func VisibleUnits(text string) int {
	return len(utf16.Encode([]rune(text)))
}

// BoundBlock shortens a visible block within the existing unit budget.
func BoundBlock(text string, limit int) string {
	if VisibleUnits(text) <= limit {
		return text
	}

	const marker = "\n[shortened]"

	return Prefix(text, max(0, limit-len(marker))) + marker
}

func blockLimit(limit int) int {
	if limit < Limit {
		return 150
	}

	return 600
}
