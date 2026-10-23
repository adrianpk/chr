package app

import (
	"context"
	"encoding/json"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) guidedDocument(ctx context.Context, request core.TelegramOperationRequest, observed GuidedUseCaseInspection, kind string) (*core.TelegramDocumentResponse, error) {
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
	message := core.TelegramMessageBlocks{Family: "43", InitialText: "Original " + kind + " for " + observed.Session.Topic, ContentKind: "markdown", Content: core.TelegramDocumentExcerpt(payload), SecondaryText: "Complete committed original attached. The current guided step is retained."}

	return &core.TelegramDocumentResponse{Message: message, Document: telegramDocumentForTarget(request.Target, "committed/"+original.Commit+"/"+original.Reference, recordDigest(payload), telegramDocumentName(kind, observed.Session.Topic), "text/markdown", payload)}, nil
}

func telegramGuidedQuestionDocument(target core.TelegramOperationTarget, observed GuidedUseCaseInspection) *core.TelegramDocumentResponse {
	payload := []byte(observed.Question)
	message := core.TelegramMessageBlocks{Family: "43", InitialText: "Clarification for " + observed.Session.Topic, Content: core.TelegramDocumentExcerpt(payload), SecondaryText: "Complete current question attached. Reply here with its answer."}

	return &core.TelegramDocumentResponse{Message: message, Document: telegramDocumentForTarget(target, "exports/guided-questions/"+observed.Session.ID+".md", recordDigest(payload), telegramDocumentName("question", observed.Session.Topic), "text/markdown", payload)}
}

func telegramGuidedResultDocument(target core.TelegramOperationTarget, observed GuidedUseCaseInspection) (*core.TelegramDocumentResponse, error) {
	snapshot := telegramGuidedResultExport{Name: observed.Session.Topic, Stage: observed.Session.Guided.Stage, Status: observed.Session.Guided.Status, Problem: observed.ResultProblem, Ideas: observed.Ideas}
	if observed.Result != nil {
		snapshot.Complete = observed.Result.Complete
		snapshot.Deliveries = observed.Result.Receipts
		snapshot.Validations = observed.Result.Validations
		snapshot.Closures = observed.Result.Closures
	}

	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, err
	}

	message := core.TelegramMessageBlocks{Family: "43", InitialText: "Result evidence for " + observed.Session.Topic, Content: telegramGuidedResultSummary(observed), SecondaryText: "Complete verified result export and pending later ideas attached."}
	filename := strings.TrimSuffix(telegramDocumentName("results", observed.Session.Topic), ".md") + ".json"

	return &core.TelegramDocumentResponse{Message: message, Document: telegramDocumentForTarget(target, "exports/guided-results/"+observed.Session.ID+".json", recordDigest(payload), filename, "application/json", payload)}, nil
}

type telegramGuidedResultExport struct {
	Name, Stage, Status, Problem string
	Complete                     bool
	Deliveries                   []core.WorkflowResultReceipt
	Validations                  []core.WorkflowResultValidation
	Closures                     []core.WorkflowResultClosure
	Ideas                        []core.GuidedLaterIdea
}
