package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// TelegramActionResult contains confirmed owner facts or an explicit blocked outcome.
type TelegramActionResult struct {
	State             string
	ReceiptRef        string
	Facts             json.RawMessage
	Text              string
	ConfigDigest      string
	ConversationPhase string
}

// FinishTelegramAction commits both requests, action consumption, input acknowledgement
// and result delivery atomically. A blocked result supersedes rather than consumes.
func (s *TelegramReceiverStore) FinishTelegramAction(ctx context.Context, claim TelegramActionClaim, result TelegramActionResult, now time.Time) error {
	if result.State != "confirmed" && result.State != "blocked" {
		return errTelegramReceiverConflict
	}

	if !validTelegramOperationFinish(result.State, result.ReceiptRef, result.Facts) {
		return errTelegramReceiverConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		retained, err := retainedTelegramActionClaim(ctx, tx, claim)
		if err != nil {
			return err
		}

		pending, err := telegramReadResponsePending(ctx, tx, retained.Answer.Source)
		if err != nil || !pending {
			return err
		}

		if retained.Action.State != "applying" {
			return errTelegramReceiverConflict
		}

		return finishTelegramActionResult(ctx, tx, retained, result, now)
	})
}

func retainedTelegramActionClaim(ctx context.Context, tx *sql.Tx, claim TelegramActionClaim) (TelegramActionClaim, error) {
	action, err := readTelegramAction(ctx, tx, claim.Action.Token, claim.Action.Peer)
	if err != nil {
		return claim, err
	}

	request, err := readTelegramOperationID(ctx, tx, action.RequestID)
	if err != nil {
		return claim, err
	}

	if action.AnswerRequestID != claim.Answer.ID || request.ID != claim.Request.ID || request.ArgumentDigest != claim.Request.ArgumentDigest || action.GuardDigest != telegramActionGuard(request) {
		return claim, errTelegramReceiverConflict
	}

	retained := TelegramActionClaim{Action: action, Request: request, Choice: claim.Choice}
	err = readTelegramActionClaimAnswer(telegramConversationTransaction{ctx: ctx, tx: tx}, &retained)

	return retained, err
}

func finishTelegramActionRequests(ctx context.Context, tx *sql.Tx, claim TelegramActionClaim, result TelegramActionResult, now time.Time) error {
	for _, request := range []TelegramOperationRequest{claim.Request, claim.Answer} {
		changed, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET state=?,domain_receipt_ref=NULLIF(?,''),result=?,updated_at=? WHERE id=? AND argument_digest=? AND state='applying'`, result.State, result.ReceiptRef, string(result.Facts), formatProjectTime(now), request.ID, request.ArgumentDigest)
		if err != nil {
			return err
		}

		if err := telegramRequireChanged(changed); err != nil {
			return err
		}
	}

	return nil
}

func finishTelegramActionState(ctx context.Context, tx *sql.Tx, claim TelegramActionClaim, state string, now time.Time) error {
	actionState := "consumed"
	if state != "confirmed" {
		actionState = "superseded"
	}

	changed, err := tx.ExecContext(ctx, `UPDATE telegram_actions SET state=?,updated_at=? WHERE token=? AND state='applying' AND consumed_by_request=?`, actionState, formatProjectTime(now), claim.Action.Token, claim.Answer.ID)
	if err != nil {
		return err
	}

	if err := telegramRequireChanged(changed); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_replies SET invalidated_at=? WHERE action_token=? AND invalidated_at IS NULL`, formatProjectTime(now), claim.Action.Token)

	return err
}

func finishTelegramActionResult(ctx context.Context, tx *sql.Tx, retained TelegramActionClaim, result TelegramActionResult, now time.Time) error {
	if result.ConversationPhase != "" && result.State != "confirmed" {
		return ErrWorkflowConflict
	}

	if err := projectTelegramAuthoringWorkPhase(ctx, tx, retained.Request, result.ConversationPhase, formatProjectTime(now)); err != nil {
		return err
	}

	if err := finishTelegramActionRequests(ctx, tx, retained, result, now); err != nil {
		return err
	}

	if err := finishTelegramActionState(ctx, tx, retained, result.State, now); err != nil {
		return err
	}

	if err := queueTelegramReferencedResponse(ctx, tx, retained.Answer.Source, result.Text, retained.Request.ID, result.ConfigDigest, now); err != nil {
		return err
	}

	response := retained.Request
	response.Source = retained.Answer.Source

	return bindTelegramOperationResponse(ctx, tx, response)
}
