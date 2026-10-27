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
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func (r telegramReceiver) offerConfirmationDocument(ctx context.Context, request telegramcore.OperationRequest, offer telegramcore.ActionOffer, image telegramConfirmationDocument) error {
	payload := []byte(fmt.Sprintf("# Confirmation scope export\n\n%s\n\nProject: %s\nNamespace: %s\n\n%s\n\n## Complete prepared operation\n\nThis identified export retains the exact prepared input used by the action owner.\n\n```json\n%s\n```\n", image.Text, request.Target.ProjectID, request.Target.Namespace, image.Complete, request.PreparedInput))
	digest := sha256.Sum256(request.PreparedInput)
	parts := strings.SplitN(image.Text, "\n\n", 2)

	message := tgmessage.Blocks{Family: image.Family, InitialText: parts[0]}
	if len(parts) == 2 {
		message.Content = parts[1]
	}

	if image.Family == "25" {
		message.ContentKind = "markdown"
		message.Content = tgmessage.DocumentExcerpt([]byte(message.Content))
		message.InitialText += "\nCreates a backlog ticket; execution does not start."
	}

	for _, project := range offer.Input.Projects {
		if project.ID == request.Target.ProjectID {
			message.InitialText += "\nProject: " + project.Name

			break
		}
	}

	offer.Delivery = &tgmessage.DocumentResponse{Message: message, Document: telegramDocumentForTarget(request.Target, telegramDocumentSource{Reference: "exports/confirmation/" + request.ID + ".md", Hash: hex.EncodeToString(digest[:]), Filename: "confirmation-scope.md", MediaType: "text/markdown", Payload: payload})}

	_, err := r.store.OfferAction(ctx, offer, time.Now())
	if !errors.Is(err, tgmessage.ErrDocumentUnavailable) {
		return err
	}

	return r.unavailableConfirmationDocument(ctx, request, offer.Input.Config)
}

func (r telegramReceiver) unavailableConfirmationDocument(ctx context.Context, request telegramcore.OperationRequest, config telegramcore.ReceiverConfig) error {
	text := "The complete confirmation scope is unavailable under the document or storage limits. No action was offered. Inspect the scope locally or request a smaller selection."

	return r.store.FinishOperation(ctx, request, telegramcore.OperationCompletion{State: "blocked", ReceiptRef: "", Facts: json.RawMessage(`{"state":"blocked","reason":"confirmation_document_unavailable"}`), Text: text, ConfigDigest: config.Digest(), At: time.Now()})
}

type telegramWorkConfirmationService interface {
	ReadTelegramWorkConfirmation(context.Context, telegramcore.OperationTarget, WorkflowBoundPreview) ([]core.WorkflowRecord, error)
}

type telegramWorkConfirmation struct {
	Complete string
	Titles   []string
}

func (r telegramReceiver) workConfirmationContents(ctx context.Context, request telegramcore.OperationRequest, preview WorkflowBoundPreview) (telegramWorkConfirmation, error) {
	operations, ok := r.operations.(telegramWorkConfirmationService)
	if !ok {
		return telegramWorkConfirmation{}, tgmessage.ErrDocumentUnavailable
	}

	records, err := operations.ReadTelegramWorkConfirmation(ctx, request.Target, preview)
	if err != nil {
		return telegramWorkConfirmation{}, err
	}

	var (
		content strings.Builder
		titles  []string
	)

	content.WriteString("## Complete selected tickets and validation contracts\n")

	for _, record := range records {
		title := telegramWorkflowRecordTitle(record)
		titles = append(titles, title+" ("+record.Task.Key+")")
		fmt.Fprintf(&content, "\n### %s — %s\n\n%s\n\nExecution contract:\n\n```json\n%s\n```\n", record.Task.Key, title, record.Body, record.Domain["execution_contract"])
	}

	return telegramWorkConfirmation{Complete: content.String(), Titles: titles}, nil
}

func telegramQuestionDocument(target telegramcore.OperationTarget, status WorkflowWorkStatus) (tgmessage.DocumentResponse, error) {
	payload, err := json.MarshalIndent(status.Integration, "", "  ")
	if err != nil {
		return tgmessage.DocumentResponse{}, err
	}

	digest := sha256.Sum256(payload)
	parts := strings.SplitN(telegramControlStatusText(status), "\n\n", 2)

	message := tgmessage.Blocks{Family: "37", InitialText: parts[0], SecondaryText: "Complete retained questions, options and bindings are attached as an identified JSON export."}
	if len(parts) == 2 {
		message.Content = parts[1]
	}

	return tgmessage.DocumentResponse{Message: message, Document: telegramDocumentForTarget(target, telegramDocumentSource{Reference: "exports/work-questions/" + status.ID + ".json", Hash: hex.EncodeToString(digest[:]), Filename: "work-questions.json", MediaType: "application/json", Payload: payload})}, nil
}

type telegramConfirmationDocument struct{ Family, Text, Complete string }
