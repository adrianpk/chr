package app

import (
	"context"
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramCaptureIntentTestProvider struct{}

func (telegramCaptureIntentTestProvider) InvokeStructuredModel(_ context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
	var input ConversationIntentContext
	if err := json.Unmarshal(request.Input, &input); err != nil {
		return nil, err
	}

	return json.Marshal(ConversationIntent{OperationID: "capture", Unresolved: []string{}, Arguments: ConversationIntentArguments{Text: input.Message}})
}
