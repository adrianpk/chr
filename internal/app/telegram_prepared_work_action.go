package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) applyPreparedWorkAction(ctx context.Context, claim telegramcore.ActionClaim, config telegramcore.ReceiverConfig) error {
	operations, ok := r.operations.(telegramPreparedWorkService)
	if !ok || claim.Choice != "develop" {
		return r.finishPreparedWorkActionFailure(ctx, claim, config, ErrRecordUnsupported)
	}

	current, err := r.validateAuthoringAuthority(ctx, claim.Request, false)
	if err != nil {
		return r.finishPreparedWorkActionFailure(ctx, claim, config, err)
	}

	prepared, err := telegramRetainedPreparedWork(claim.Request)
	if err != nil {
		return r.finishPreparedWorkActionFailure(ctx, claim, current, err)
	}

	if prepared.Authorization != nil {
		return r.finishPreparedWorkActionFailure(ctx, claim, current, workflowstate.ErrConflict)
	}

	grant, err := core.AuthorizeAuthoringWork(prepared.Scope, workflowstate.MutationBinding{RequestID: claim.Answer.ID, Operator: claim.Answer.Operator}, prepared.Scope.Hash)
	if err != nil {
		return r.finishPreparedWorkActionFailure(ctx, claim, current, err)
	}

	prepared.Authorization = &grant

	return r.dispatchPreparedWorkAction(ctx, operations, claim, prepared, current)
}

func (r telegramReceiver) dispatchPreparedWorkAction(ctx context.Context, operations telegramPreparedWorkService, claim telegramcore.ActionClaim, prepared PreparedAuthoringWork, config telegramcore.ReceiverConfig) error {
	applyContext, cancel := context.WithDeadline(ctx, claim.Action.ExpiresAt)
	defer cancel()

	result, err := operations.ApplyTelegramAuthoringWork(applyContext, claim.Request.Target, prepared, func(ctx context.Context) error {
		_, err := r.validateAuthoringAuthority(ctx, claim.Request, false)

		return err
	})
	if errors.Is(err, workflowstate.ErrConflict) || errors.Is(err, ErrRecordUnsupported) || errors.Is(err, core.ErrAuthoringAuthorityRequired) {
		return r.finishPreparedWorkActionFailure(ctx, claim, config, err)
	}

	if err != nil {
		return err
	}

	facts, err := telegramPreparedWorkFacts(result)
	if err != nil {
		return err
	}

	phase := ""
	if _, err := r.validateAuthoringAuthority(ctx, claim.Request, false); err == nil {
		phase = "prepared"
	}

	return r.store.FinishAction(ctx, claim, telegramcore.ActionResult{Message: telegramOutcomeMessage("26", telegramPreparedWorkResultText(result)), State: "confirmed", ReceiptRef: result.Result.ReceiptRef, Facts: facts, Text: telegramPreparedWorkResultText(result), ConfigDigest: config.Digest(), ConversationPhase: phase}, time.Now())
}

func (r telegramReceiver) finishPreparedWorkActionFailure(ctx context.Context, claim telegramcore.ActionClaim, config telegramcore.ReceiverConfig, failure error) error {
	text := telegramAuthoringFailure(failure)

	facts, err := json.Marshal(map[string]string{"state": "blocked", "message": text})
	if err != nil {
		return err
	}

	return r.store.FinishAction(ctx, claim, telegramcore.ActionResult{Message: telegramOutcomeMessage("27", text), State: "blocked", Facts: facts, Text: text, ConfigDigest: config.Digest()}, time.Now())
}
