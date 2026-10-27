package app

import (
	"context"
	"encoding/json"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func (r telegramReceiver) finishOriginalDocument(ctx context.Context, operations telegramExplanationService, request telegramcore.OperationRequest, prepared PreparedAuthoringExplanation, config telegramcore.ReceiverConfig) error {
	artifact, err := operations.ReadPreparedTelegramDocument(ctx, request.Target, prepared)
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	current, err := r.validateAuthoringAuthority(ctx, request, false)
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	result, err := json.Marshal(struct {
		SourceRef  string `json:"source_ref"`
		SourceHash string `json:"source_hash"`
	}{artifact.Ref.Path, artifact.SourceHash})
	if err != nil {
		return err
	}

	payload := []byte(artifact.Body)
	response := tgmessage.DocumentResponse{
		Message:  tgmessage.Blocks{Family: "24", InitialText: "Original " + string(artifact.Ref.Kind) + " document.", ContentKind: "markdown", Content: tgmessage.DocumentExcerpt(payload), SecondaryText: "Complete original document attached."},
		Document: telegramDocumentForTarget(request.Target, telegramDocumentSource{Reference: artifact.Ref.Path, Hash: artifact.SourceHash, Filename: string(artifact.Ref.Kind) + ".md", MediaType: "text/markdown", Payload: payload}),
	}
	request.Delivery = &response

	return r.store.FinishOperation(ctx, request, telegramcore.OperationCompletion{State: telegramcore.OperationConfirmed, ReceiptRef: "authoring-document:" + request.ID, Facts: result, Text: response.Message.InitialText, ConfigDigest: current.Digest(), At: time.Now()})
}
