package app

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramActionHelp() string {
	return "/answer <action-token> <choice> — Answer the exact displayed action; replying or using its button has the same binding"
}

func (r telegramReceiver) respondAction(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	answer, recognized, err := r.store.ResolveTelegramActionAnswer(ctx, update)
	if !recognized {
		return false, nil
	}

	if errors.Is(err, core.ErrTelegramActionUnavailable) {
		return true, r.actionDiagnostic(ctx, update, config)
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
		return true, r.actionDiagnostic(ctx, update, config)
	}

	if err != nil {
		return true, err
	}

	switch claim.Action.State {
	case "expired", "superseded":
		return true, r.actionDiagnostic(ctx, update, config)
	case "consumed":
		return true, r.store.QueueReadResponse(ctx, update, "Action already confirmed. Request: "+claim.Request.ID, config.Digest(), time.Now())
	default:
		return true, r.dispatchTelegramAction(ctx, claim, config)
	}
}

func (r telegramReceiver) dispatchTelegramAction(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig) error {
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
	return r.store.QueueReadResponse(ctx, update, "This action is unavailable, expired or stale for this exact peer and prompt. Read the current state before requesting a new action.", config.Digest(), time.Now())
}
