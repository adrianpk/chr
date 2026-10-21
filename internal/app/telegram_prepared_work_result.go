package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramPreparedWorkPreview(prepared PreparedAuthoringWork) string {
	scope := prepared.Scope
	title, body := "the retained request", ""
	if prepared.Mutation.Native != nil {
		body = string(prepared.Mutation.Native.Record.Body)
		if value := prepared.Mutation.Native.Record.Metadata["title"]; len(value) > 0 {
			if err := json.Unmarshal(value, &title); err != nil {
				title = "the retained request"
			}
		}
	}
	excerpt := core.TelegramDocumentExcerpt([]byte(body))
	return fmt.Sprintf("Prepare ticket %s in %s?\n\n%s\n\nOriginal request: %s\nConfirming authorizes this exact scope and creates the ticket in the backlog. Execution does not start.", title, scope.ProjectID, excerpt, core.TelegramDocumentExcerpt([]byte(scope.Instruction)))
}

func telegramPreparedWorkFacts(result AuthoringWorkResult) (json.RawMessage, error) {
	result.Result.Record.Body = ""

	return json.Marshal(result)
}

func telegramPreparedWorkResultText(result AuthoringWorkResult) string {
	return fmt.Sprintf("Ticket prepared in backlog: %s — %s.\n\nExecution has not started. Use /work ticket %s to review its exact start confirmation.\nCanonical record: %s", result.Result.Record.Identity, result.Result.Record.Title, result.Result.Record.Identity, result.Result.Record.Reference)
}

func (r telegramReceiver) finishPreparedWork(ctx context.Context, request core.TelegramOperationRequest, result AuthoringWorkResult, config core.TelegramReceiverConfig) error {
	if result.Scope.Binding != (core.WorkflowMutationBinding{RequestID: request.ID, Operator: request.Operator}) || result.Result.ReceiptRef == "" {
		return core.ErrWorkflowConflict
	}

	facts, err := telegramPreparedWorkFacts(result)
	if err != nil {
		return err
	}

	text := telegramPreparedWorkResultText(result)

	err = r.store.FinishTelegramAuthoring(ctx, request, core.TelegramAuthoringProgress{SessionRef: result.Scope.SessionRef}, result.Result.ReceiptRef, facts, text, config.Digest(), time.Now())
	if errors.Is(err, core.ErrWorkflowConflict) {
		return r.store.FinishTelegramOperation(ctx, request, "confirmed", result.Result.ReceiptRef, facts, "Canonical work preparation is confirmed; the advanced conversation is unchanged.\n"+text, config.Digest(), time.Now())
	}

	return err
}
