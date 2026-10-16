package app

import (
	"context"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

const telegramActionUnavailableText = "This action is unavailable, expired or stale for this exact peer and prompt. Read the current state before requesting a new action."

func (r telegramReceiver) staleActionDiagnostic(ctx context.Context, update core.TelegramReceivedUpdate, answer core.TelegramActionAnswer, config core.TelegramReceiverConfig) error {
	text := telegramActionUnavailableText
	if status, ok := r.staleControlStatus(ctx, update, answer); ok {
		text += "\nCurrent state:\n" + telegramWorkProgressText(status)
	}

	return r.store.QueueReadResponse(ctx, update, text, config.Digest(), time.Now())
}

func (r telegramReceiver) staleControlStatus(ctx context.Context, update core.TelegramReceivedUpdate, answer core.TelegramActionAnswer) (WorkflowWorkStatus, bool) {
	request, err := r.store.TelegramActionRequest(ctx, update.Peer, answer)
	if err != nil || request.OperationID != "work.control" {
		return WorkflowWorkStatus{}, false
	}

	if err := r.validateRecordTargetAuthority(ctx, request, true); err != nil {
		return WorkflowWorkStatus{}, false
	}

	preview, err := telegramRetainedControl(request)
	if err != nil {
		return WorkflowWorkStatus{}, false
	}

	operations, ok := r.operations.(telegramControlService)
	if !ok {
		return WorkflowWorkStatus{}, false
	}

	status, err := operations.ReadTelegramControl(ctx, request.Target, preview.Request.ID)

	return status, err == nil
}
