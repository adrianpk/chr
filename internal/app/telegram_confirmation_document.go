package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) offerConfirmationDocument(ctx context.Context, request core.TelegramOperationRequest, offer core.TelegramActionOffer, family, text, complete string) error {
	payload := []byte(fmt.Sprintf("# Confirmation scope export\n\n%s\n\nProject: %s\nNamespace: %s\n\n%s\n\n## Complete prepared operation\n\nThis identified export retains the exact prepared input used by the action owner.\n\n```json\n%s\n```\n", text, request.Target.ProjectID, request.Target.Namespace, complete, request.PreparedInput))
	digest := sha256.Sum256(request.PreparedInput)
	parts := strings.SplitN(text, "\n\n", 2)

	message := core.TelegramMessageBlocks{Family: family, InitialText: parts[0]}
	if len(parts) == 2 {
		message.Content = parts[1]
	}

	if family == "25" {
		message.ContentKind = "markdown"
		message.Content = core.TelegramDocumentExcerpt([]byte(strings.TrimPrefix(complete, "## Complete proposed ticket\n\n")))
		message.InitialText += "\nCreates a backlog ticket; execution does not start."
	}

	for _, project := range offer.Input.Projects {
		if project.ID == request.Target.ProjectID {
			message.InitialText += "\nProject: " + project.Name

			break
		}
	}

	offer.Delivery = &core.TelegramDocumentResponse{Message: message, Document: telegramDocumentForTarget(request.Target, "exports/confirmation/"+request.ID+".md", hex.EncodeToString(digest[:]), "confirmation-scope.md", "text/markdown", payload)}

	_, err := r.store.OfferTelegramAction(ctx, offer, time.Now())
	if !errors.Is(err, core.ErrTelegramDocumentUnavailable) {
		return err
	}

	return r.unavailableConfirmationDocument(ctx, request, offer.Input.Config)
}

func (r telegramReceiver) unavailableConfirmationDocument(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
	text := "The complete confirmation scope is unavailable under the document or storage limits. No action was offered. Inspect the scope locally or request a smaller selection."

	return r.store.FinishTelegramOperation(ctx, request, "blocked", "", json.RawMessage(`{"state":"blocked","reason":"confirmation_document_unavailable"}`), text, config.Digest(), time.Now())
}

type telegramWorkConfirmationService interface {
	ReadTelegramWorkConfirmation(context.Context, core.TelegramOperationTarget, WorkflowBoundPreview) ([]core.WorkflowRecord, error)
}

func (r telegramReceiver) workConfirmationContents(ctx context.Context, request core.TelegramOperationRequest, preview WorkflowBoundPreview) (string, error) {
	operations, ok := r.operations.(telegramWorkConfirmationService)
	if !ok {
		return "", core.ErrTelegramDocumentUnavailable
	}

	records, err := operations.ReadTelegramWorkConfirmation(ctx, request.Target, preview)
	if err != nil {
		return "", err
	}

	var content strings.Builder
	content.WriteString("## Complete selected tickets and validation contracts\n")

	for _, record := range records {
		fmt.Fprintf(&content, "\n### %s\n\n%s\n\nExecution contract:\n\n```json\n%s\n```\n", record.Task.Key, record.Body, record.Domain["execution_contract"])
	}

	return content.String(), nil
}

func telegramQuestionDocument(target core.TelegramOperationTarget, status WorkflowWorkStatus) (core.TelegramDocumentResponse, error) {
	payload, err := json.MarshalIndent(status.Integration.Decision, "", "  ")
	if err != nil {
		return core.TelegramDocumentResponse{}, err
	}

	digest := sha256.Sum256(payload)
	parts := strings.SplitN(telegramControlStatusText(status), "\n\n", 2)

	message := core.TelegramMessageBlocks{Family: "37", InitialText: parts[0], SecondaryText: "Complete retained questions, options and bindings are attached as an identified JSON export."}
	if len(parts) == 2 {
		message.Content = parts[1]
	}

	return core.TelegramDocumentResponse{Message: message, Document: telegramDocumentForTarget(target, "exports/work-questions/"+status.ID+".json", hex.EncodeToString(digest[:]), "work-questions.json", "application/json", payload)}, nil
}
