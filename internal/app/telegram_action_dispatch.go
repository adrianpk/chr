package app

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) respondAction(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	answer, recognized, err := r.store.ResolveTelegramActionAnswer(ctx, update)
	if !recognized {
		return false, r.checkCaptureReplyLane(update)
	}

	if errors.Is(err, core.ErrTelegramActionUnavailable) {
		return true, r.staleActionDiagnostic(ctx, update, answer, config)
	}

	if err != nil {
		return true, err
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	claim, err := r.store.ClaimTelegramAction(ctx, core.TelegramConversationInput{Update: update, Config: config, Projects: projects}, answer, time.Now())
	if errors.Is(err, core.ErrTelegramActionUnavailable) {
		return true, r.staleActionDiagnostic(ctx, update, answer, config)
	}

	if err != nil {
		return true, err
	}

	return true, r.respondActionClaim(ctx, update, claim, config)
}

func (r telegramReceiver) respondActionClaim(ctx context.Context, update core.TelegramReceivedUpdate, claim core.TelegramActionClaim, config core.TelegramReceiverConfig) error {
	if err := r.checkActionLane(claim); err != nil {
		return err
	}

	if claim.ReceiptOnly {
		return r.applyControlAction(ctx, claim, config)
	}

	switch claim.Action.State {
	case "expired", "superseded":
		return r.staleActionDiagnostic(ctx, update, core.TelegramActionAnswer{Token: claim.Action.Token, Choice: claim.Choice}, config)
	case "consumed":
		return r.store.QueueReadResponse(ctx, update, "Action already confirmed. Request: "+claim.Request.ID, config.Digest(), time.Now())
	default:
		return r.dispatchTelegramAction(ctx, claim, config)
	}
}

func (r telegramReceiver) dispatchTelegramAction(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig) error {
	if claim.Request.OperationID == core.TelegramHarnessSelectionOperation {
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

func (r telegramReceiver) actionDiagnostic(ctx context.Context, update core.TelegramReceivedUpdate, config core.TelegramReceiverConfig) error {
	return r.store.QueueReadResponse(ctx, update, telegramActionUnavailableText, config.Digest(), time.Now())
}
