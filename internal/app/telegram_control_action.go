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

	if err := r.validateRecordTargetAuthority(ctx, claim.Request, true); err != nil {
		return r.controlRecoveryFailure(ctx, claim, err, config)
	}

	binding := core.WorkflowMutationBinding{RequestID: claim.Request.ID, Operator: claim.Request.Operator}

	status, found, err := operations.ReadTelegramControlReceipt(ctx, claim.Request.Target, binding, preview.Request)
	if found {
		return r.completeControlAction(ctx, claim, status, err, config)
	}

	if err != nil {
		return r.controlRecoveryFailure(ctx, claim, err, config)
	}

	if claim.ReceiptOnly {
		return r.staleActionDiagnostic(ctx, claim.ResponseSource, core.TelegramActionAnswer{Token: claim.Action.Token, Choice: claim.Choice}, config)
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
	status, err = operations.ApplyTelegramControl(ctx, claim.Request.Target, binding, preview.Request, check)

	return r.completeControlAction(ctx, claim, status, err, config)
}

func (r telegramReceiver) completeControlAction(ctx context.Context, claim core.TelegramActionClaim, status WorkflowWorkStatus, err error, config core.TelegramReceiverConfig) error {
	if err != nil && status.ControlReceipt != nil {
		return &telegramControlReceiptPending{failure: err}
	}

	if err != nil && !telegramControlExpectedFailure(err) {
		return err
	}

	return r.finishControlAction(ctx, claim, status, err, config)
}

func (r telegramReceiver) controlRecoveryFailure(ctx context.Context, claim core.TelegramActionClaim, err error, config core.TelegramReceiverConfig) error {
	if !telegramControlExpectedFailure(err) {
		return err
	}

	if claim.ReceiptOnly {
		return r.actionDiagnostic(ctx, claim.ResponseSource, config)
	}

	return r.finishControlAction(ctx, claim, WorkflowWorkStatus{}, err, config)
}

// A native receipt proves acceptance while the inbox still needs its outcome.
// The receiver may retry expected domain barriers without stopping intake.
type telegramControlReceiptPending struct{ failure error }

func (e *telegramControlReceiptPending) Error() string { return e.failure.Error() }

func (e *telegramControlReceiptPending) Unwrap() error { return e.failure }

func (r telegramReceiver) finishControlAction(ctx context.Context, claim core.TelegramActionClaim, status WorkflowWorkStatus, failure error, config core.TelegramReceiverConfig) error {
	state, receipt := "blocked", ""
	text := "Native control is unavailable or stale. Read current state before requesting a new control."

	if failure == nil {
		state = "confirmed"
		receipt = fmt.Sprintf("workflow:%s:revision:%d", status.ID, status.Revision)
		text = "Control confirmed: " + telegramControlTitle(claim.Choice) + "."

		if status.ControlReceipt != nil {
			receipt = status.ControlReceipt.ReceiptRef
			text += "\nThe original control receipt is retained."
		}
	}

	if status.ID != "" {
		text += "\n" + telegramWorkProgressText(status)
	}

	facts, err := json.Marshal(status)
	if err != nil {
		return err
	}

	result := core.TelegramActionResult{Message: telegramOutcomeMessage("36", text), State: state, ReceiptRef: receipt, Facts: facts, Text: text, ConfigDigest: config.Digest()}
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
