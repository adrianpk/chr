package app

import (
	"context"
	"encoding/json"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramCaptureIntentTestProvider struct{}

func (telegramCaptureIntentTestProvider) InvokeStructuredModel(_ context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
	if strings.HasPrefix(request.Purpose, "authoring_") {
		return json.Marshal(authoringModelOutput{Body: "Preserve the captured requirements."})
	}

	var input ConversationIntentContext
	if err := json.Unmarshal(request.Input, &input); err != nil {
		return nil, err
	}

	return json.Marshal(ConversationIntent{OperationID: "capture", Unresolved: []string{}, Arguments: ConversationIntentArguments{Text: input.Message}})
}
