package app

import (
	"context"
	"encoding/json"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func (r telegramReceiver) guidedDocument(ctx context.Context, request telegramcore.OperationRequest, observed GuidedUseCaseInspection, kind string) (*tgmessage.DocumentResponse, error) {
	if kind == "results" {
		return telegramGuidedResultDocument(request.Target, observed)
	}

	if kind == "" {
		if !authoringcore.GuidedDispatchAllowed(*observed.Session) || len([]rune(observed.Question)) <= 600 {
			return nil, nil
		}

		return telegramGuidedQuestionDocument(request.Target, observed), nil
	}

	operations, ok := r.operations.(telegramGuidedService)
	if !ok {
		return nil, ErrRecordUnsupported
	}

	original, err := operations.ReadTelegramGuidedOriginal(ctx, request.Target, observed.Session.ID, kind)
	if err != nil {
		return nil, err
	}

	payload := []byte(original.Content)
	message := tgmessage.Blocks{Family: "43", InitialText: "Original " + kind + " for " + observed.Session.Topic, ContentKind: "markdown", Content: tgmessage.DocumentExcerpt(payload), SecondaryText: "Complete committed original attached. The current guided step is retained."}

	return &tgmessage.DocumentResponse{Message: message, Document: telegramDocumentForTarget(request.Target, telegramDocumentSource{Reference: "committed/" + original.Commit + "/" + original.Reference, Hash: recordDigest(payload), Filename: telegramDocumentName(kind, observed.Session.Topic), MediaType: "text/markdown", Payload: payload})}, nil
}

func telegramGuidedQuestionDocument(target telegramcore.OperationTarget, observed GuidedUseCaseInspection) *tgmessage.DocumentResponse {
	payload := []byte(observed.Question)
	message := tgmessage.Blocks{Family: "43", InitialText: "Clarification for " + observed.Session.Topic, Content: tgmessage.DocumentExcerpt(payload), SecondaryText: "Complete current question attached. Reply here with its answer."}

	return &tgmessage.DocumentResponse{Message: message, Document: telegramDocumentForTarget(target, telegramDocumentSource{Reference: "exports/guided-questions/" + observed.Session.ID + ".md", Hash: recordDigest(payload), Filename: telegramDocumentName("question", observed.Session.Topic), MediaType: "text/markdown", Payload: payload})}
}

func telegramGuidedResultDocument(target telegramcore.OperationTarget, observed GuidedUseCaseInspection) (*tgmessage.DocumentResponse, error) {
	snapshot := guidedResultSnapshot(observed)

	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, err
	}

	message := tgmessage.Blocks{Family: "43", InitialText: "Result evidence for " + observed.Session.Topic, Content: telegramGuidedResultSummary(observed), SecondaryText: "Complete verified result export and pending later ideas attached."}
	filename := strings.TrimSuffix(telegramDocumentName("results", observed.Session.Topic), ".md") + ".json"

	return &tgmessage.DocumentResponse{Message: message, Document: telegramDocumentForTarget(target, telegramDocumentSource{Reference: "exports/guided-results/" + observed.Session.ID + ".json", Hash: recordDigest(payload), Filename: filename, MediaType: "application/json", Payload: payload})}, nil
}
