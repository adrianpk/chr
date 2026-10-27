package message

import (
	"encoding/json"
	"errors"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// RetainedInput binds a serialized view to its stored text and document identity.
type RetainedInput struct{ Payload, Hash, Text, DocumentID string }

// DecodeRetained validates the original hash, version, rendering and attachment binding.
// A nil presentation denotes a legacy delivery without a stored presentation.
func DecodeRetained(input RetainedInput) (*Presentation, error) {
	if input.Hash != "" && contenthash.Sum([]byte(input.Payload)) != input.Hash {
		return nil, errors.New("retained telegram presentation hash mismatch")
	}

	presentation, err := decodeRetainedPayload(input)
	if err != nil {
		return nil, err
	}

	if input.DocumentID != "" && (presentation == nil || presentation.DocumentID != input.DocumentID || input.Hash == "") {
		return nil, ErrDocumentUnavailable
	}

	return presentation, nil
}

func decodeRetainedPayload(input RetainedInput) (*Presentation, error) {
	if input.Payload == "" {
		return nil, nil
	}

	var value Presentation
	if err := json.Unmarshal([]byte(input.Payload), &value); err != nil {
		return nil, err
	}

	if value.Version != 1 || (value.TemplateRevision != "text-v1" && value.TemplateRevision != "document-v1") {
		return nil, errors.New("unsupported telegram presentation version")
	}

	rendered := Render(value.Blocks)
	if value.TemplateRevision == "document-v1" {
		rendered = RenderDocument(value.Blocks)
	}

	if rendered.PlainFallback != value.PlainFallback || input.Text != value.PlainFallback || !slices.Equal(rendered.Entities, value.Entities) {
		return nil, errors.New("invalid retained telegram presentation")
	}

	return &value, nil
}
