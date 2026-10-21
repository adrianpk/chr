package app

import (
	"context"
	"encoding/json"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) finishOriginalDocument(ctx context.Context, operations telegramExplanationService, request core.TelegramOperationRequest, prepared PreparedAuthoringExplanation, config core.TelegramReceiverConfig) error {
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
	response := core.TelegramDocumentResponse{
		Message:  core.TelegramMessageBlocks{Family: "24", InitialText: "Original " + string(artifact.Ref.Kind) + " document.", ContentKind: "markdown", Content: core.TelegramDocumentExcerpt(payload), SecondaryText: "Complete original document attached."},
		Document: telegramDocumentForTarget(request.Target, artifact.Ref.Path, artifact.SourceHash, string(artifact.Ref.Kind)+".md", "text/markdown", payload),
	}
	request.Delivery = &response

	return r.store.FinishTelegramOperation(ctx, request, "confirmed", "authoring-document:"+request.ID, result, response.Message.InitialText, current.Digest(), time.Now())
}
