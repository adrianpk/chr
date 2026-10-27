package app

import (
	"context"
	"encoding/json"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const telegramGuidedOvertakenAck = "Your original request is confirmed. The conversation has advanced; its current step is retained. Send Show current step to read the latest prompt."

// A committed effect can outlive its original display fence. Confirm only its
// retained request; never refresh the command, overwrite a newer prompt or grant
// reply authority to this historical acknowledgement.
func (r telegramReceiver) finishOvertakenGuided(ctx context.Context, request telegramcore.TelegramOperationRequest, progress telegramcore.TelegramGuidedProgress, facts json.RawMessage, config telegramcore.TelegramReceiverConfig) error {
	if err := r.validateOvertakenGuided(ctx, request, progress.SessionRef); err != nil {
		return err
	}

	request.Message = telegramOutcomeMessage("43", telegramGuidedOvertakenAck)
	if request.Delivery != nil {
		request.Delivery.Message.SecondaryText = telegramGuidedOvertakenAck
	}

	return r.store.FinishTelegramOperation(ctx, request, "confirmed", progress.SessionRef, facts, telegramGuidedOvertakenAck, config.Digest(), time.Now())
}

func (r telegramReceiver) finishOvertakenGuidedAction(ctx context.Context, claim telegramcore.TelegramActionClaim, result telegramcore.TelegramActionResult) error {
	if err := r.validateOvertakenGuided(ctx, claim.Request, result.ReceiptRef); err != nil {
		return err
	}

	result.Guided = nil

	result.Text, result.Message = telegramGuidedOvertakenAck, telegramOutcomeMessage("43", telegramGuidedOvertakenAck)
	if result.Delivery != nil {
		result.Delivery.Message.SecondaryText = telegramGuidedOvertakenAck
	}

	return r.store.FinishTelegramAction(ctx, claim, result, time.Now())
}

func (r telegramReceiver) validateOvertakenGuided(ctx context.Context, request telegramcore.TelegramOperationRequest, sessionRef string) error {
	config, err := r.validateAuthoringAuthority(ctx, request, true)
	if err != nil {
		return err
	}

	id, err := telegramAuthoringSessionID(sessionRef)
	if err != nil {
		return err
	}

	if err := r.validateGuidedRunBinding(ctx, request, config, id); err != nil {
		return err
	}

	current, found, err := r.store.TelegramConversationReference(ctx, request.Source.Peer, request.Target.ConversationID)
	if err != nil {
		return err
	}

	if !found || current.Revision <= request.Target.ConversationRevision {
		return workflowstate.ErrConflict
	}

	return nil
}
