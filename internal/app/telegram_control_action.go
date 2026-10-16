package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) applyControlAction(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig) error {
	preview, err := telegramRetainedControl(claim.Request)
	if err != nil {
		return err
	}

	operations, ok := r.operations.(telegramControlService)
	if !ok || claim.Choice != preview.Request.Action {
		return r.finishControlAction(ctx, claim, WorkflowWorkStatus{}, ErrRecordUnsupported, config)
	}

	if err := r.validateRecordAuthority(ctx, claim.Request); err != nil {
		return r.finishControlAction(ctx, claim, WorkflowWorkStatus{}, err, config)
	}

	check := func(ctx context.Context) error {
		if !claim.Action.ExpiresAt.After(time.Now()) {
			return core.ErrWorkflowConflict
		}

		return r.validateRecordAuthority(ctx, claim.Request)
	}
	binding := core.WorkflowMutationBinding{RequestID: claim.Request.ID, Operator: claim.Request.Operator}

	status, err := operations.ApplyTelegramControl(ctx, claim.Request.Target, binding, preview.Request, check)
	if err != nil && status.ControlReceipt != nil {
		return err
	}

	if err != nil && !telegramControlExpectedFailure(err) {
		return err
	}

	return r.finishControlAction(ctx, claim, status, err, config)
}

func (r telegramReceiver) finishControlAction(ctx context.Context, claim core.TelegramActionClaim, status WorkflowWorkStatus, failure error, config core.TelegramReceiverConfig) error {
	state, receipt := "blocked", ""
	text := "Native control is unavailable or stale. Read current state before requesting a new control."

	if failure == nil {
		state = "confirmed"
		receipt = fmt.Sprintf("workflow:%s:revision:%d", status.ID, status.Revision)
		text = "Native control confirmed."

		if status.ControlReceipt != nil {
			receipt = status.ControlReceipt.ReceiptRef
			text += "\nNative request receipt: " + receipt
		}
	}

	if status.ID != "" {
		text += "\n" + telegramWorkProgressText(status)
	}

	facts, err := json.Marshal(status)
	if err != nil {
		return err
	}

	result := core.TelegramActionResult{State: state, ReceiptRef: receipt, Facts: facts, Text: text, ConfigDigest: config.Digest()}
	if state == "confirmed" {
		result.Invocation = &core.TelegramInvocationBinding{ID: status.ID, Revision: status.Revision}
		result.InvocationPhase = telegramWorkObservation(status).Phase
	}

	return r.store.FinishTelegramAction(ctx, claim, result, time.Now())
}

func telegramControlExpectedFailure(err error) bool {
	for _, expected := range []error{core.ErrWorkflowConflict, core.ErrWorkflowOwned, core.ErrWorkflowCapacity, core.ErrWorkflowStop, core.ErrWorkflowIntegrationBudget} {
		if errors.Is(err, expected) {
			return true
		}
	}

	return false
}
