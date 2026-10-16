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

	instruction := []rune(scope.Instruction)
	if len(instruction) > 1024 {
		instruction = append(instruction[:1024], []rune("\n[Instruction excerpt]")...)
	}

	body := ""
	if prepared.Mutation.Native != nil {
		body = string(prepared.Mutation.Native.Record.Body)
	}

	excerpt, truncated := authoringSourceExcerpt(body)
	if truncated {
		excerpt += "\n[Description excerpt; read the canonical source below.]"
	}

	return fmt.Sprintf("Native backlog ticket preview\nIntent: %s\nProject: %s\nNamespace: %s\nScope hash: %s\nCanonical source: %s\nDescription hash: %s\nContract hash: %s\nOriginal instruction:\n%s\n\nDescription:\n%s\nNo work invocation has started.", scope.Intent, scope.ProjectID, scope.Namespace, scope.Hash, scope.DescriptionRef, scope.DescriptionHash, scope.ContractHash, string(instruction), excerpt)
}

func telegramPreparedWorkFacts(result AuthoringWorkResult) (json.RawMessage, error) {
	result.Result.Record.Body = ""

	return json.Marshal(result)
}

func telegramPreparedWorkResultText(result AuthoringWorkResult) string {
	return fmt.Sprintf("Native backlog ticket prepared: %s\nCanonical record: %s\nScope hash: %s\nIntent: %s\nOwner receipt: %s\nNo work invocation has started.", result.Result.Record.Identity, result.Result.Record.Reference, result.Scope.Hash, result.Scope.Intent, result.Result.ReceiptRef)
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
