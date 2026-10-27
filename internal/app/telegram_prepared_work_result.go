package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramPreparedWorkPreview(prepared PreparedAuthoringWork) string {
	scope := prepared.Scope

	title, body := "the retained request", ""

	if prepared.Mutation.Native != nil {
		record := prepared.Mutation.Native.Record
		body = string(record.Body)
		title = telegramWorkflowRecordTitle(record)

		var contract core.GeneratedWorkflowTask
		if err := json.Unmarshal(record.Domain["execution_contract"], &contract); err == nil && contract.Body != "" {
			body = contract.Body
		}
	}

	excerpt := tgmessage.DocumentExcerpt([]byte(body))

	return fmt.Sprintf("Prepare ticket: %s\n\n%s\n\nOriginal request: %s\nConfirming authorizes this exact scope and creates the ticket in the backlog. Execution does not start.", title, excerpt, tgmessage.DocumentExcerpt([]byte(scope.Instruction)))
}

func telegramPreparedWorkFacts(result AuthoringWorkResult) (json.RawMessage, error) {
	result.Result.Record.Body = ""

	return json.Marshal(result)
}

func telegramPreparedWorkResultText(result AuthoringWorkResult) string {
	return fmt.Sprintf("Ticket prepared in backlog: %s — %s.\n\nExecution has not started. Use /work ticket %s to review its exact start confirmation.", result.Result.Record.Identity, result.Result.Record.Title, result.Result.Record.Identity)
}

func (r telegramReceiver) finishPreparedWork(ctx context.Context, request telegramcore.OperationRequest, result AuthoringWorkResult, config telegramcore.ReceiverConfig) error {
	if result.Scope.Binding != (workflowstate.MutationBinding{RequestID: request.ID, Operator: request.Operator}) || result.Result.ReceiptRef == "" {
		return workflowstate.ErrConflict
	}

	facts, err := telegramPreparedWorkFacts(result)
	if err != nil {
		return err
	}

	text := telegramPreparedWorkResultText(result)
	request.Message = telegramOutcomeMessage("26", text)

	err = r.store.FinishAuthoring(ctx, request, telegramcore.AuthoringProgress{SessionRef: result.Scope.SessionRef}, result.Result.ReceiptRef, facts, text, config.Digest(), time.Now())
	if errors.Is(err, workflowstate.ErrConflict) {
		request.Message = telegramOutcomeMessage("26", "Ticket preparation is confirmed; the advanced conversation is unchanged.\n"+text)

		return r.store.FinishOperation(ctx, request, "confirmed", result.Result.ReceiptRef, facts, "Canonical work preparation is confirmed; the advanced conversation is unchanged.\n"+text, config.Digest(), time.Now())
	}

	return err
}

func telegramWorkflowRecordTitle(record core.WorkflowRecord) string {
	var title string
	if err := json.Unmarshal(record.Metadata["title"], &title); err == nil && title != "" {
		return title
	}

	var contract core.GeneratedWorkflowTask
	if err := json.Unmarshal(record.Domain["execution_contract"], &contract); err == nil && contract.Title != "" {
		return contract.Title
	}

	return "the retained request"
}
