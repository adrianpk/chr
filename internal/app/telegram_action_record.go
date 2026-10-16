package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) applyRecordAction(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig) error {
	operations, ok := r.operations.(telegramRecordService)
	if !ok || !strings.HasPrefix(claim.Request.OperationID, "records.") || claim.Choice != "confirm" {
		return r.finishRecordActionFailure(ctx, claim, ErrRecordUnsupported, config)
	}

	prepared, err := r.prepareRecordAction(ctx, claim)
	if errors.Is(err, core.ErrWorkflowConflict) {
		return r.finishRecordActionFailure(ctx, claim, err, config)
	}

	if err != nil {
		return err
	}

	applyContext, cancel := context.WithDeadline(ctx, claim.Action.ExpiresAt)
	defer cancel()

	result, err := operations.ApplyTelegramRecord(applyContext, claim.Request.Target, prepared)
	if errors.Is(err, core.ErrWorkflowConflict) || errors.Is(err, ErrRecordUnsupported) {
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

	return r.store.FinishTelegramAction(ctx, claim, core.TelegramActionResult{State: "confirmed", ReceiptRef: result.ReceiptRef, Facts: facts, Text: "Confirmed record mutation.\nRequest: " + claim.Request.ID + "\n" + string(facts), ConfigDigest: config.Digest()}, time.Now())
}

func (r telegramReceiver) finishRecordActionFailure(ctx context.Context, claim core.TelegramActionClaim, failure error, config core.TelegramReceiverConfig) error {
	text := telegramRecordFailure(failure)

	facts, err := json.Marshal(struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}{"blocked", text})
	if err != nil {
		return err
	}

	return r.store.FinishTelegramAction(ctx, claim, core.TelegramActionResult{State: "blocked", Facts: facts, Text: text, ConfigDigest: config.Digest()}, time.Now())
}

func (r telegramReceiver) prepareRecordAction(ctx context.Context, claim core.TelegramActionClaim) (PreparedRecordMutation, error) {
	if !claim.Action.ExpiresAt.After(time.Now()) {
		return PreparedRecordMutation{}, core.ErrWorkflowConflict
	}

	if err := r.validateRecordAuthority(ctx, claim.Request); err != nil {
		return PreparedRecordMutation{}, err
	}

	var prepared PreparedRecordMutation
	if err := json.Unmarshal(claim.Request.PreparedInput, &prepared); err != nil {
		return PreparedRecordMutation{}, err
	}

	if !telegramPreparedRecordBinding(prepared, claim.Request) {
		return PreparedRecordMutation{}, core.ErrWorkflowConflict
	}

	if err := r.store.BeginTelegramOperation(ctx, claim.Request, time.Now()); err != nil {
		return PreparedRecordMutation{}, err
	}

	return prepared, nil
}
