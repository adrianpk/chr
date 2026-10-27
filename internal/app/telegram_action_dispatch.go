package app

import (
	"context"
	"errors"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func (r telegramReceiver) respondAction(ctx context.Context, user telegramcore.TelegramAllowedUser, config telegramcore.TelegramReceiverConfig, update telegramcore.TelegramReceivedUpdate) (bool, error) {
	answer, recognized, err := r.store.ResolveTelegramActionAnswer(ctx, update)
	if !recognized {
		return false, r.checkCaptureReplyLane(update)
	}

	if errors.Is(err, telegramcore.ErrTelegramActionUnavailable) {
		return true, r.staleActionDiagnostic(ctx, update, answer, config)
	}

	if err != nil {
		return true, err
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	claim, err := r.store.ClaimTelegramAction(ctx, telegramcore.TelegramConversationInput{Update: update, Config: config, Projects: projects}, answer, time.Now())
	if errors.Is(err, telegramcore.ErrTelegramActionUnavailable) {
		return true, r.staleActionDiagnostic(ctx, update, answer, config)
	}

	if err != nil {
		return true, err
	}

	return true, r.respondActionClaim(ctx, update, claim, config)
}

func (r telegramReceiver) respondActionClaim(ctx context.Context, update telegramcore.TelegramReceivedUpdate, claim telegramcore.TelegramActionClaim, config telegramcore.TelegramReceiverConfig) error {
	if err := r.checkActionLane(claim); err != nil {
		return err
	}

	if claim.ReceiptOnly {
		return r.applyControlAction(ctx, claim, config)
	}

	switch claim.Action.State {
	case "expired", "superseded":
		return r.staleActionDiagnostic(ctx, update, telegramcore.TelegramActionAnswer{Token: claim.Action.Token, Choice: claim.Choice}, config)
	case "consumed":
		return r.store.QueuePresentationResponse(ctx, update, *telegramOutcomeMessage("40", telegramActionReplayText(claim.Request)), config.Digest(), time.Now())
	default:
		return r.dispatchTelegramAction(ctx, claim, config)
	}
}

func (r telegramReceiver) dispatchTelegramAction(ctx context.Context, claim telegramcore.TelegramActionClaim, config telegramcore.TelegramReceiverConfig) error {
	if claim.Request.OperationID == "guided.step" {
		return r.applyGuidedAction(ctx, claim, config)
	}

	if claim.Request.OperationID == telegramcore.TelegramHarnessSelectionOperation {
		return r.applyHarnessAction(ctx, claim, config)
	}

	if claim.Request.OperationID == "work.control" {
		return r.applyControlAction(ctx, claim, config)
	}

	if claim.Request.OperationID == "work.admit" {
		return r.applyWorkAdmission(ctx, claim, config)
	}

	if claim.Request.OperationID == "authoring.work" {
		return r.applyPreparedWorkAction(ctx, claim, config)
	}

	return r.applyRecordAction(ctx, claim, config)
}

func (r telegramReceiver) actionDiagnostic(ctx context.Context, update telegramcore.TelegramReceivedUpdate, config telegramcore.TelegramReceiverConfig) error {
	return r.store.QueuePresentationResponse(ctx, update, *telegramOutcomeMessage("40", telegramActionUnavailableText), config.Digest(), time.Now())
}
