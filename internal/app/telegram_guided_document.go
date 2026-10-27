package app

import (
	"context"
	"encoding/json"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func (r telegramReceiver) guidedDocument(ctx context.Context, request telegramcore.OperationRequest, observed GuidedUseCaseInspection, kind string) (*tgmessage.DocumentResponse, error) {
	if kind == "results" {
		return telegramGuidedResultDocument(request.Target, observed)
	}

	if kind == "" {
		if !core.GuidedDispatchAllowed(*observed.Session) || len([]rune(observed.Question)) <= 600 {
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

	return &tgmessage.DocumentResponse{Message: message, Document: telegramDocumentForTarget(request.Target, "committed/"+original.Commit+"/"+original.Reference, recordDigest(payload), telegramDocumentName(kind, observed.Session.Topic), "text/markdown", payload)}, nil
}

func telegramGuidedQuestionDocument(target telegramcore.OperationTarget, observed GuidedUseCaseInspection) *tgmessage.DocumentResponse {
	payload := []byte(observed.Question)
	message := tgmessage.Blocks{Family: "43", InitialText: "Clarification for " + observed.Session.Topic, Content: tgmessage.DocumentExcerpt(payload), SecondaryText: "Complete current question attached. Reply here with its answer."}

	return &tgmessage.DocumentResponse{Message: message, Document: telegramDocumentForTarget(target, "exports/guided-questions/"+observed.Session.ID+".md", recordDigest(payload), telegramDocumentName("question", observed.Session.Topic), "text/markdown", payload)}
}

func telegramGuidedResultDocument(target telegramcore.OperationTarget, observed GuidedUseCaseInspection) (*tgmessage.DocumentResponse, error) {
	snapshot := guidedResultSnapshot(observed)

	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, err
	}

	message := tgmessage.Blocks{Family: "43", InitialText: "Result evidence for " + observed.Session.Topic, Content: telegramGuidedResultSummary(observed), SecondaryText: "Complete verified result export and pending later ideas attached."}
	filename := strings.TrimSuffix(telegramDocumentName("results", observed.Session.Topic), ".md") + ".json"

	return &tgmessage.DocumentResponse{Message: message, Document: telegramDocumentForTarget(target, "exports/guided-results/"+observed.Session.ID+".json", recordDigest(payload), filename, "application/json", payload)}, nil
}
