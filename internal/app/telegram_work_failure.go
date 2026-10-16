package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramWorkFailure(failure error) string {
	if errors.Is(failure, core.ErrWorkflowConflict) {
		return "Work conflict: the displayed selection, policy, target or authority changed. Read current native work before requesting a new preview."
	}

	if errors.Is(failure, core.ErrWorkflowOwned) || errors.Is(failure, core.ErrWorkflowCapacity) {
		return "Native work is already owned or capacity is occupied. Inspect the existing invocation before requesting another admission."
	}

	return "Native work is unavailable. Inspect the project's workflow readiness; this request has no confirmed admission."
}

func (r telegramReceiver) finishWorkFailure(ctx context.Context, request core.TelegramOperationRequest, state string, failure error, config core.TelegramReceiverConfig) error {
	text := telegramWorkFailure(failure)

	facts, err := json.Marshal(struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}{state, text})
	if err != nil {
		return err
	}

	return r.store.FinishTelegramOperation(ctx, request, state, "", facts, text, config.Digest(), time.Now())
}

func (r telegramReceiver) finishWorkActionFailure(ctx context.Context, claim core.TelegramActionClaim, failure error, config core.TelegramReceiverConfig) error {
	text := telegramWorkFailure(failure)

	facts, err := json.Marshal(struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}{"blocked", text})
	if err != nil {
		return err
	}

	return r.store.FinishTelegramAction(ctx, claim, core.TelegramActionResult{State: "blocked", Facts: facts, Text: text, ConfigDigest: config.Digest()}, time.Now())
}
