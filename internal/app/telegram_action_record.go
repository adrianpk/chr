package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) applyRecordAction(ctx context.Context, claim telegramcore.ActionClaim, config telegramcore.ReceiverConfig) error {
	operations, ok := r.operations.(telegramRecordService)
	if !ok || !strings.HasPrefix(claim.Request.OperationID, "records.") || claim.Choice != "confirm" {
		return r.finishRecordActionFailure(ctx, claim, ErrRecordUnsupported, config)
	}

	prepared, err := r.prepareRecordAction(ctx, claim)
	if errors.Is(err, workflowstate.ErrConflict) {
		return r.finishRecordActionFailure(ctx, claim, err, config)
	}

	if err != nil {
		return err
	}

	applyContext, cancel := context.WithDeadline(ctx, claim.Action.ExpiresAt)
	defer cancel()

	result, err := operations.ApplyTelegramRecord(applyContext, claim.Request.Target, prepared)
	if errors.Is(err, workflowstate.ErrConflict) || errors.Is(err, ErrRecordUnsupported) {
		return r.finishRecordActionFailure(ctx, claim, err, config)
	}

	if err != nil {
		return err
	}

	result.Record.Body = ""

	facts, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return r.store.FinishAction(ctx, claim, telegramcore.ActionResult{Message: telegramOutcomeMessage("29", telegramRecordResultText(result)), State: telegramcore.OperationConfirmed, ReceiptRef: result.ReceiptRef, Facts: facts, Text: telegramRecordResultText(result), ConfigDigest: config.Digest()}, time.Now())
}

func (r telegramReceiver) finishRecordActionFailure(ctx context.Context, claim telegramcore.ActionClaim, failure error, config telegramcore.ReceiverConfig) error {
	text := telegramRecordFailure(failure)

	facts, err := json.Marshal(struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}{"blocked", text})
	if err != nil {
		return err
	}

	return r.store.FinishAction(ctx, claim, telegramcore.ActionResult{Message: telegramOutcomeMessage("30", text), State: telegramcore.OperationBlocked, Facts: facts, Text: text, ConfigDigest: config.Digest()}, time.Now())
}

func (r telegramReceiver) prepareRecordAction(ctx context.Context, claim telegramcore.ActionClaim) (PreparedRecordMutation, error) {
	if !claim.Action.ExpiresAt.After(time.Now()) {
		return PreparedRecordMutation{}, workflowstate.ErrConflict
	}

	if err := r.validateRecordAuthority(ctx, claim.Request); err != nil {
		return PreparedRecordMutation{}, err
	}

	var prepared PreparedRecordMutation
	if err := json.Unmarshal(claim.Request.PreparedInput, &prepared); err != nil {
		return PreparedRecordMutation{}, err
	}

	if !telegramPreparedRecordBinding(prepared, claim.Request) {
		return PreparedRecordMutation{}, workflowstate.ErrConflict
	}

	if err := r.store.BeginOperation(ctx, claim.Request, time.Now()); err != nil {
		return PreparedRecordMutation{}, err
	}

	return prepared, nil
}
