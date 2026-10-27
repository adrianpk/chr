package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TelegramActionResult contains confirmed owner facts or an explicit blocked outcome.
type TelegramActionResult struct {
	Delivery          *TelegramDocumentResponse
	Guided            *TelegramGuidedProgress
	Message           *TelegramMessageBlocks
	Invocation        *TelegramInvocationBinding
	InvocationPhase   string
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

		source, pending, err := telegramActionResponseSource(ctx, tx, retained, result)
		if err != nil || !pending {
			return err
		}

		if retained.Action.State != "applying" && !confirmedTelegramControlReceipt(retained, result) {
			return errTelegramReceiverConflict
		}

		retained.ResponseSource = source

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

	if action.AnswerRequestID != claim.Answer.ID || request.ID != claim.Request.ID || request.ArgumentDigest != claim.Request.ArgumentDigest || !telegramActionDigestMatches(action, request) {
		return claim, errTelegramReceiverConflict
	}

	retained := TelegramActionClaim{Action: action, Request: request, Choice: claim.Choice, ResponseSource: claim.ResponseSource}
	err = readTelegramActionClaimAnswer(telegramConversationTransaction{ctx: ctx, tx: tx}, &retained)

	return retained, err
}

func finishTelegramActionRequests(ctx context.Context, tx *sql.Tx, claim TelegramActionClaim, result TelegramActionResult, now time.Time) error {
	recovery := confirmedTelegramControlReceipt(claim, result)
	for _, request := range []TelegramOperationRequest{claim.Request, claim.Answer} {
		changed, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET state=?,domain_receipt_ref=NULLIF(?,''),result=?,updated_at=? WHERE id=? AND argument_digest=? AND (state='applying' OR (? AND state='blocked'))`, result.State, result.ReceiptRef, string(result.Facts), database.FormatTime(now), request.ID, request.ArgumentDigest, recovery)
		if err != nil {
			return err
		}

		if err := telegramRequireChanged(changed); err != nil {
			return err
		}
	}

	return nil
}

func finishTelegramActionState(ctx context.Context, tx *sql.Tx, claim TelegramActionClaim, result TelegramActionResult, now time.Time) error {
	actionState := "consumed"
	if result.State != "confirmed" {
		actionState = "superseded"
	}

	changed, err := tx.ExecContext(ctx, `UPDATE telegram_actions SET state=?,updated_at=? WHERE token=? AND (state='applying' OR (? AND state IN ('expired','superseded'))) AND consumed_by_request=?`, actionState, database.FormatTime(now), claim.Action.Token, confirmedTelegramControlReceipt(claim, result), claim.Answer.ID)
	if err != nil {
		return err
	}

	if err := telegramRequireChanged(changed); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_replies SET invalidated_at=? WHERE action_token=? AND invalidated_at IS NULL`, database.FormatTime(now), claim.Action.Token)

	return err
}

func finishTelegramActionResult(ctx context.Context, tx *sql.Tx, retained TelegramActionClaim, result TelegramActionResult, now time.Time) error {
	if err := finishTelegramHarnessConversation(ctx, tx, retained.Request, result.State, now); err != nil {
		return err
	}

	if err := projectTelegramWorkBinding(ctx, tx, retained.Request, result, database.FormatTime(now)); err != nil {
		return err
	}

	if result.ConversationPhase != "" && result.State != "confirmed" {
		return ErrWorkflowConflict
	}

	if err := projectTelegramAuthoringWorkPhase(ctx, tx, retained.Request, result.ConversationPhase, database.FormatTime(now)); err != nil {
		return err
	}

	if err := finishTelegramActionRequests(ctx, tx, retained, result, now); err != nil {
		return err
	}

	if err := finishTelegramActionState(ctx, tx, retained, result, now); err != nil {
		return err
	}

	return finishTelegramActionResponse(ctx, tx, retained, result, now)
}

func finishTelegramActionResponse(ctx context.Context, tx *sql.Tx, retained TelegramActionClaim, result TelegramActionResult, now time.Time) error {
	var err error

	retained, err = projectTelegramGuidedAction(ctx, tx, retained, result, now)
	if err != nil {
		return err
	}

	if err := queueTelegramActionResultResponse(ctx, tx, retained, result, now); err != nil {
		return err
	}

	response := retained.Request

	response.Source = retained.ResponseSource
	if err := bindTelegramOperationResponse(ctx, tx, response); err != nil {
		return err
	}

	if result.Guided != nil {
		_, err := tx.ExecContext(ctx, `UPDATE telegram_outbox SET question_binding=? WHERE request_id=? AND action_token IS NULL`, result.Guided.Binding, retained.Request.ID)

		return err
	}

	return nil
}

func telegramActionResponseSource(ctx context.Context, tx *sql.Tx, claim TelegramActionClaim, result TelegramActionResult) (TelegramReceivedUpdate, bool, error) {
	original := claim.Answer.Source

	pending, err := telegramReadResponsePending(ctx, tx, original)
	if err != nil || pending {
		return original, pending, err
	}

	if !confirmedTelegramControlReceipt(claim, result) || claim.ResponseSource.Peer != original.Peer || claim.ResponseSource.UpdateID == 0 || claim.Action.State == "consumed" {
		return original, false, nil
	}

	pending, err = telegramReadResponsePending(ctx, tx, claim.ResponseSource)

	return claim.ResponseSource, pending, err
}

func confirmedTelegramControlReceipt(claim TelegramActionClaim, result TelegramActionResult) bool {
	return claim.Request.OperationID == "work.control" && claim.Answer.ID != "" && result.State == "confirmed" && result.ReceiptRef == workflowMutationReceiptPath(claim.Request.ID)
}

func queueTelegramActionResultResponse(ctx context.Context, tx *sql.Tx, retained TelegramActionClaim, result TelegramActionResult, now time.Time) error {
	if result.Delivery != nil {
		response := retained.Request
		response.Source = retained.ResponseSource
		response.Delivery = result.Delivery

		return queueTelegramOperationResult(ctx, tx, response, result.Text, result.ReceiptRef, result.ConfigDigest, now)
	}

	if result.Message != nil {
		return queueTelegramPresentation(ctx, tx, retained.ResponseSource, *result.Message, result.ConfigDigest, now)
	}

	return queueTelegramReferencedResponse(ctx, tx, retained.ResponseSource, result.Text, retained.Request.ID, result.ConfigDigest, now)
}
