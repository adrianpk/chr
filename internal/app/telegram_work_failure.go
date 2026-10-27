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

func (r telegramReceiver) finishWorkFailure(ctx context.Context, request telegramcore.OperationRequest, state string, failure error, config telegramcore.ReceiverConfig) error {
	text := telegramWorkFailure(failure)

	facts, err := json.Marshal(struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}{state, text})
	if err != nil {
		return err
	}

	return r.store.FinishOperation(ctx, request, telegramcore.OperationCompletion{State: state, ReceiptRef: "", Facts: facts, Text: text, ConfigDigest: config.Digest(), At: time.Now()})
}

func (r telegramReceiver) finishWorkActionFailure(ctx context.Context, claim telegramcore.ActionClaim, failure error, config telegramcore.ReceiverConfig) error {
	text := telegramWorkFailure(failure)

	facts, err := json.Marshal(struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}{"blocked", text})
	if err != nil {
		return err
	}

	return r.store.FinishAction(ctx, claim, telegramcore.ActionResult{Message: telegramOutcomeMessage("32", text), State: "blocked", Facts: facts, Text: text, ConfigDigest: config.Digest()}, time.Now())
}
