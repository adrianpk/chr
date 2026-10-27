package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramWorkFailure(failure error) string {
	if errors.Is(failure, workflowstate.ErrConflict) {
		return "Work conflict: the displayed selection, policy, target or authority changed. Read current native work before requesting a new preview."
	}

	if errors.Is(failure, core.ErrWorkflowOwned) {
		return "This selection already has an execution owner. Inspect the retained work before requesting another start."
	}

	if errors.Is(failure, core.ErrWorkflowCapacity) {
		return "Execution capacity is occupied. This request has no confirmed new admission. Inspect active work before retrying."
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

	return r.store.FinishTelegramAction(ctx, claim, core.TelegramActionResult{Message: telegramOutcomeMessage("32", text), State: "blocked", Facts: facts, Text: text, ConfigDigest: config.Digest()}, time.Now())
}
